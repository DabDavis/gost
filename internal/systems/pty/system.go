package pty

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/creack/pty"
	"gost/internal/events"
	"gost/internal/systems/input"
)

// -----------------------------------------------------------------------------
// PTY Global Context
// -----------------------------------------------------------------------------

var globalPTY struct {
	mu sync.Mutex
	f  *os.File
}

// -----------------------------------------------------------------------------
// PTY System
// -----------------------------------------------------------------------------

type System struct {
	bus     *events.Bus
	shell   string
	started bool
	mu      sync.Mutex
	cmdKill func()
}

// NewSystem wires PTY with input + bus.
func NewSystem(bus *events.Bus) *System {
	ps := &System{
		bus:   bus,
		shell: defaultShell(),
	}

	// keyboard → PTY
	input.WriteToPTY = func(b []byte) {
		globalPTY.mu.Lock()
		defer globalPTY.mu.Unlock()
		if globalPTY.f != nil {
			if _, err := globalPTY.f.Write(b); err != nil {
				log.Println("[PTY] write error:", err)
			}
		}
	}

	ps.subscribeConfigChanges()
	return ps
}

func (s *System) UpdateECS() {
	if s.started {
		return
	}
	s.started = true
	s.launchShell()
}

// -----------------------------------------------------------------------------
// Shell Launch
// -----------------------------------------------------------------------------

func (s *System) launchShell() {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, cmd, err := startShell(s.shell)
	if err != nil {
		log.Println("[PTY] start failed:", err)
		s.bus.Publish("pty_restart_failed", err.Error())
		return
	}

	globalPTY.mu.Lock()
	globalPTY.f = f
	globalPTY.mu.Unlock()

	s.cmdKill = func() {
		_ = cmd.Process.Kill()
		_ = f.Close()
	}

	log.Println("[PTY] started shell:", s.shell)
	s.bus.Publish("pty_restarted", s.shell)

	startResizeWatcher(f)
	go s.readLoop(f, cmd)
}

// -----------------------------------------------------------------------------
// Shell + IO
// -----------------------------------------------------------------------------

func startShell(shell string) (*os.File, *exec.Cmd, error) {
	cmd := exec.Command(shell, "-i")
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	f, err := pty.Start(cmd)
	if err != nil {
		return nil, nil, err
	}
	return f, cmd, nil
}

func (s *System) readLoop(f *os.File, cmd *exec.Cmd) {
	defer func() {
		log.Println("[PTY] session ended.")
		s.bus.Publish("pty_ended", nil)
	}()

	reader := bufio.NewReader(f)
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			out := make([]byte, n)
			copy(out, buf[:n])
			s.bus.Publish("pty_output", out)
		}
		if err != nil {
			if err != io.EOF {
				log.Println("[PTY] read error:", err)
			}
			return
		}
	}
}

// -----------------------------------------------------------------------------
// Resize watcher
// -----------------------------------------------------------------------------

// startResizeWatcher adjusts PTY size on SIGWINCH.
func startResizeWatcher(f *os.File) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGWINCH)
	go func() {
		for range sigs {
			if err := pty.InheritSize(os.Stdin, f); err != nil {
				log.Println("[PTY] resize error:", err)
			}
		}
	}()
	_ = pty.InheritSize(os.Stdin, f)
}

// -----------------------------------------------------------------------------
// Config + Helpers
// -----------------------------------------------------------------------------

func (s *System) subscribeConfigChanges() {
	if s.bus == nil {
		return
	}
	sub := s.bus.Subscribe("config_shell_updated")
	go func() {
		for evt := range sub {
			if path, ok := evt.(string); ok && path != "" {
				s.restartShell(path)
			}
		}
	}()
}

func (s *System) restartShell(newShell string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.shell == newShell {
		return
	}
	log.Printf("[PTY] restarting shell: %s → %s", s.shell, newShell)
	if s.cmdKill != nil {
		s.cmdKill()
		s.cmdKill = nil
	}
	s.shell = newShell
	go s.launchShell()
}

func defaultShell() string {
	if sh := os.Getenv("SHELL"); sh != "" {
		return sh
	}
	return "/bin/bash"
}

