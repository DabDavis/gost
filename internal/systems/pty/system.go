package pty

import (
	"log"
	"os"
	"sync"

	"gost/internal/events"
	"gost/internal/systems/input"
)

// -----------------------------------------------------------------------------
// PTY System — manages the interactive shell session.
// -----------------------------------------------------------------------------

type System struct {
	bus     *events.Bus
	shell   string
	started bool
	mu      sync.Mutex
	cmdKill func() // deferred cleanup
}

// NewSystem links PTY with input and initializes write hook.
func NewSystem(bus *events.Bus) *System {
	ps := &System{
		bus:   bus,
		shell: defaultShell(),
	}

	// Link keyboard → PTY write handler
	input.WriteToPTY = func(b []byte) {
		globalPTY.mu.Lock()
		defer globalPTY.mu.Unlock()
		if globalPTY.f != nil {
			if _, err := globalPTY.f.Write(b); err != nil {
				log.Println("[PTY] write error:", err)
			}
		}
	}

	// Watch config updates
	ps.subscribeConfigChanges()

	return ps
}

// UpdateECS ensures the shell starts only once.
func (s *System) UpdateECS() {
	if s.started {
		return
	}
	s.started = true
	s.launchShell()
}

// launchShell starts the PTY shell and its read loop.
func (s *System) launchShell() {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, cmd, err := startShell(s.shell)
	if err != nil {
		log.Println("[PTY] start failed:", err)
		s.bus.Publish("pty_restart_failed", err.Error())
		return
	}

	// Keep references
	globalPTY.f = f
	s.cmdKill = func() {
		_ = cmd.Process.Kill()
		_ = f.Close()
	}
	log.Println("[PTY] started shell:", s.shell)
	s.bus.Publish("pty_restarted", s.shell)

	startResizeWatcher(f, 7, 14)
	go s.readLoop(f, cmd)
}

// restartShell cleanly restarts when shell path changes.
func (s *System) restartShell(newShell string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.shell == newShell {
		return
	}

	log.Printf("[PTY] restarting shell: %s → %s", s.shell, newShell)

	// Terminate existing PTY
	if s.cmdKill != nil {
		s.cmdKill()
		s.cmdKill = nil
	}

	s.shell = newShell
	go s.launchShell()
}

// defaultShell returns a safe fallback.
func defaultShell() string {
	if sh := os.Getenv("SHELL"); sh != "" {
		return sh
	}
	return "/bin/bash"
}

