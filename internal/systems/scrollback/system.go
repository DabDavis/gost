package scrollback

import (
	"sync"

	"gost/internal/components"
	"gost/internal/events"
)

// System connects the TermBuffer to Scrollback via the event bus,
// capturing lines that scroll off the visible area and managing user scroll state.
type System struct {
	bus        *events.Bus
	buffer     *components.TermBuffer
	scrollback *components.Scrollback

	mu     sync.RWMutex // protects offset
	offset int          // number of lines scrolled up in history

	scrollUpSub   <-chan events.Event
	scrollDownSub <-chan events.Event
	termScrollSub <-chan events.Event
}

// NewSystem creates and registers the scrollback system.
// It listens for "term_scrolled" (buffer overflow) and
// "scroll_up"/"scroll_down" (user scroll commands) events.
func NewSystem(bus *events.Bus, buffer *components.TermBuffer, sb *components.Scrollback) *System {
	s := &System{
		bus:        bus,
		buffer:     buffer,
		scrollback: sb,
	}

	// Subscribe to terminal scrolls
	s.termScrollSub = bus.Subscribe("term_scrolled")

	// Subscribe to user input scroll commands
	s.scrollUpSub = bus.Subscribe("scroll_up")
	s.scrollDownSub = bus.Subscribe("scroll_down")

	// --- Capture scrolled lines into history ---
	go func() {
		for evt := range s.termScrollSub {
			if line, ok := evt.([]components.Glyph); ok {
				sb.PushLine(line)
			}
		}
	}()

	// --- Handle user scroll commands ---
	go func() {
		for {
			select {
			case <-s.scrollUpSub:
				s.adjustOffset(+5)
			case <-s.scrollDownSub:
				s.adjustOffset(-5)
			}
		}
	}()

	return s
}

// UpdateECS runs every tick and clamps scroll offset within history bounds.
// Only publishes "scroll_offset_changed" when the offset actually changes.
func (s *System) UpdateECS() {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := s.scrollback.Count()
	if count <= 0 {
		if s.offset != 0 {
			s.offset = 0
			s.bus.Publish("scroll_offset_changed", 0)
		}
		return
	}

	clamped := s.offset
	if clamped > count {
		clamped = count
	}
	if clamped < 0 {
		clamped = 0
	}

	if clamped != s.offset {
		s.offset = clamped
		s.bus.Publish("scroll_offset_changed", clamped)
	}
}

// Offset returns the current scrollback offset (thread-safe).
func (s *System) Offset() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.offset
}

// Reset clears any manual scrollback offset (returns to live view).
func (s *System) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.offset != 0 {
		s.offset = 0
		s.bus.Publish("scroll_offset_changed", 0)
	}
}

// adjustOffset increments offset safely and notifies renderer lazily.
func (s *System) adjustOffset(delta int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	newOffset := s.offset + delta
	if newOffset < 0 {
		newOffset = 0
	}
	if newOffset != s.offset {
		s.offset = newOffset
		s.bus.Publish("scroll_offset_changed", newOffset)
	}
}

