package input

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// handleMouseScroll checks the mouse wheel delta and publishes scroll events.
func (s *System) handleMouseScroll(now time.Time) {
	_, dy := ebiten.Wheel()

	if dy == 0 {
		return
	}

	// Simple debounce: only one event per frame
	if dy > 0 {
		s.bus.Publish("scroll_up", nil)
	} else if dy < 0 {
		s.bus.Publish("scroll_down", nil)
	}
}

