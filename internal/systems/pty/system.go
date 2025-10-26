package pty

import (
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

    // Start background read loop
    go func() {
        buf := make([]byte, 1024)
        for {
            n, err := f.Read(buf)
            if n > 0 {
                s.bus.Publish("pty_output", string(buf[:n]))
            }
            if err != nil {
                if err == io.EOF {
                    break
                }
                log.Println("[PTY] read error:", err)
                break
            }
        }
    }()
}

