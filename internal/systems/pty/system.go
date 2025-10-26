package pty

import (
	"log"
	"os"

	"gost/internal/events"
	"gost/internal/systems/input"
)

// System runs a shell inside a pseudoterminal and bridges it to the event bus.
type System struct {
	bus     *events.Bus
	started bool
}

// NewSystem links PTY with input and initializes write hook.
func NewSystem(bus *events.Bus) *System {
	ps := &System{bus: bus}

	input.WriteToPTY = func(b []byte) {
		globalPTY.mu.Lock()
		defer globalPTY.mu.Unlock()
		if globalPTY.f != nil {
			if _, err := globalPTY.f.Write(b); err != nil {
				log.Println("[PTY] write error:", err)
			}
		}
	}

	return ps
}

// UpdateECS starts the shell once per session.
func (s *System) UpdateECS() {
	if s.started {
		return
	}
	s.started = true

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	f, cmd, err := startShell(shell)
	if err != nil {
		log.Println("[PTY] start failed:", err)
		return
	}

	globalPTY.f = f
	log.Println("[PTY] started shell:", shell)

	startResizeWatcher(f, 7, 14)
	go s.readLoop(f, cmd)
}

