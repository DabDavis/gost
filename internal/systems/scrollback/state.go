package scrollback

import "sync"

// The state file isolates offset logic and synchronization.

// mutex guards concurrent access to offset and scrollback state.
var mu sync.RWMutex

// adjustOffset modifies the current scroll offset and notifies the renderer.
func (s *System) adjustOffset(delta int) {
	mu.Lock()
	defer mu.Unlock()

	newOffset := s.offset + delta
	if newOffset < 0 {
		newOffset = 0
	}

	if newOffset != s.offset {
		s.offset = newOffset
		s.bus.Publish("scroll_offset_changed", newOffset)
		s.bus.Publish("render_redraw", nil)
	}
}

// clampOffset ensures the offset stays within valid bounds.
func (s *System) clampOffset() {
	mu.Lock()
	defer mu.Unlock()

	count := s.scrollback.Count()
	if count <= 0 {
		if s.offset != 0 {
			s.offset = 0
			s.bus.Publish("scroll_offset_changed", 0)
			s.bus.Publish("render_redraw", nil)
		}
		return
	}

	if s.offset > count {
		s.offset = count
		s.bus.Publish("scroll_offset_changed", s.offset)
		s.bus.Publish("render_redraw", nil)
	}
}

// Offset returns the current scrollback offset safely.
func (s *System) Offset() int {
	mu.RLock()
	defer mu.RUnlock()
	return s.offset
}

// Reset clears any manual scrollback offset and restores live view.
func (s *System) Reset() {
	mu.Lock()
	defer mu.Unlock()

	if s.offset != 0 {
		s.offset = 0
		s.bus.Publish("scroll_offset_changed", 0)
		s.bus.Publish("render_redraw", nil)
	}
}

