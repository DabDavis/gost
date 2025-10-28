package scrollback

import (
    "gost/internal/components"
    "gost/internal/events"
)

// System manages scrollback offset and input events.
type System struct {
    bus        *events.Bus
    term       *components.TermBuffer
    scrollback *components.Scrollback

    offset     int // current scroll offset (0 = live)
    scrollStep int // lines per scroll tick
}

// NewSystem creates a new scrollback manager.
func NewSystem(bus *events.Bus, term *components.TermBuffer, sb *components.Scrollback) *System {
    s := &System{
        bus:        bus,
        term:       term,
        scrollback: sb,
        offset:     0,
        scrollStep: 3,
    }
    s.subscribeScrollEvents()
    return s
}

// UpdateECS is called every frame (no-op for now).
func (s *System) UpdateECS() {}

// scrollUp moves the viewport up through scrollback.
func (s *System) scrollUp(lines int) {
    if s.scrollback == nil || s.scrollback.Count() == 0 {
        return
    }
    s.offset += lines
    if s.offset > s.scrollback.Count()-1 {
        s.offset = s.scrollback.Count() - 1
    }
    s.bus.Publish("scroll_offset_changed", s.offset)
}

// scrollDown moves the viewport down toward live output.
func (s *System) scrollDown(lines int) {
    s.offset -= lines
    if s.offset <= 0 {
        s.offset = 0
        s.bus.Publish("scroll_reset", nil)
    } else {
        s.bus.Publish("scroll_offset_changed", s.offset)
    }
}

// Reset clears scrollback view.
func (s *System) Reset() {
    s.offset = 0
    s.bus.Publish("scroll_reset", nil)
}

// subscribeScrollEvents hooks mouse + keyboard events.
func (s *System) subscribeScrollEvents() {
    if s.bus == nil {
        return
    }

    subScrollUp := s.bus.Subscribe("scroll_up")
    subScrollDown := s.bus.Subscribe("scroll_down")
    subScrollReset := s.bus.Subscribe("scroll_reset_request")
    subScrollPageUp := s.bus.Subscribe("scroll_page_up")
    subScrollPageDown := s.bus.Subscribe("scroll_page_down")

    go func() {
        for range subScrollUp {
            s.scrollUp(s.scrollStep)
        }
    }()
    go func() {
        for range subScrollDown {
            s.scrollDown(s.scrollStep)
        }
    }()
    go func() {
        for range subScrollPageUp {
            s.scrollUp(s.term.Height) // page scroll up
        }
    }()
    go func() {
        for range subScrollPageDown {
            s.scrollDown(s.term.Height) // page scroll down
        }
    }()
    go func() {
        for range subScrollReset {
            s.Reset()
        }
    }()
}

