package scrollback

import "log"

// subscribeConfigChanges listens for hot-reload updates to scroll speed.
func (s *System) subscribeConfigChanges() {
	sub := s.bus.Subscribe("config_changed")
	go func() {
		for evt := range sub {
			if cfg, ok := evt.(interface {
				GetScrollStep() int
			}); ok {
				step := cfg.GetScrollStep()
				if step <= 0 {
					step = 5
				}
				s.scrollStep = step
				log.Printf("[Scrollback] Scroll step updated to %d lines/event", step)
			}
		}
	}()
}

