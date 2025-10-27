package pty

import "log"

// subscribeConfigChanges watches for config reloads and restarts PTY if needed.
func (s *System) subscribeConfigChanges() {
	sub := s.bus.Subscribe("config_changed")
	go func() {
		for evt := range sub {
			if cfg, ok := evt.(interface {
				GetDefaultShell() string
			}); ok {
				newShell := cfg.GetDefaultShell()
				if newShell != "" && newShell != s.shell {
					log.Printf("[PTY] Detected shell change: %s → %s", s.shell, newShell)
					s.restartShell(newShell)
				}
			}
		}
	}()
}

