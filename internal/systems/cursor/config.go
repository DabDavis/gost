package cursor

import (
	"time"

	"gost/internal/components"
)

// subscribeTermUpdates watches for "term_updated" events to attach the buffer.
func (c *System) subscribeTermUpdates() {
	sub := c.bus.Subscribe("term_updated")
	go func() {
		for evt := range sub {
			if tb, ok := evt.(*components.TermBuffer); ok {
				c.mu.Lock()
				c.term = tb
				c.mu.Unlock()
			}
		}
	}()
}

// subscribeConfigChanges watches for "config_changed" and applies style dynamically.
func (c *System) subscribeConfigChanges() {
	sub := c.bus.Subscribe("config_changed")
	go func() {
		for evt := range sub {
			// Expected interface provided by config.RootConfig
			if cfg, ok := evt.(interface {
				GetCursorColor() interface{ RGBA() (r, g, b, a uint32) }
				GetCursorShape() string
				GetCursorBlink() bool
				GetCursorBlinkRate() int
			}); ok {
				c.mu.Lock()
				c.style.Color = cfg.GetCursorColor()
				c.style.Shape = cfg.GetCursorShape()
				c.style.Blink = cfg.GetCursorBlink()
				c.style.BlinkRate = time.Duration(cfg.GetCursorBlinkRate()) * time.Millisecond
				c.mu.Unlock()
			}
		}
	}()
}

