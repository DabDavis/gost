package pty

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
	"gost/internal/events"
	"gost/internal/systems/input"
)

// globalPTY holds the live PTY file descriptor and lock.
var globalPTY struct {
	f  *os.File
	mu sync.Mutex
}

// System runs a shell inside a pseudoterminal and bridges it to the event bus.
type System struct {
	bus     *events.Bus
	started bool
}

// NewSystem starts a new PTY manager and links it to the input system.
func NewSystem(bus *events.Bus) *System {
	ps := &System{bus: bus}

	// Link the modular input system’s writer
	input.WriteToPTY = func(b []byte) {
		globalPTY.mu.Lock()
		defer globalPTY.mu.Unlock()
		if globalPTY.f != nil {
			_, err := globalPTY.f.Write(b)
			if err != nil {
				log.Println("[PTY] write error:", err)
			}
		}
	}

	return ps
}

// UpdateECS ensures the PTY is started once per session.
func (s *System) UpdateECS() {
	if s.started {
		return
	}
	s.started = true

	// Choose user shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	cmd := exec.Command(shell)
	f, err := pty.Start(cmd)
	if err != nil {
		log.Println("pty start failed:", err)
		return
	}

	globalPTY.f = f
	log.Println("[PTY] started shell:", shell)

	// Background read loop
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := f.Read(buf)
			if n > 0 {
				out := string(buf[:n])
				s.bus.Publish("pty_output", out)

				// Detect manual "exit" input to trigger shutdown
				if bytes.Contains(bytes.ToLower(buf[:n]), []byte("exit")) {
					log.Println("[PTY] 'exit' detected, shutting down GoST...")
					s.bus.Publish("system_exit", nil)
					return
				}
			}
			if err != nil {
				if err == io.EOF {
					log.Println("[PTY] shell exited (EOF), shutting down GoST...")
					s.bus.Publish("system_exit", nil)
					return
				}
				log.Println("[PTY] read error:", err)
				s.bus.Publish("system_exit", nil)
				return
			}
		}
	}()
}

