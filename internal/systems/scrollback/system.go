package scrollback

import (
	"gost/internal/components"
	"gost/internal/ecs"
	"gost/internal/events"
)

// -----------------------------------------------------------------------------
// Scrollback System — manages off-screen history and navigation
// -----------------------------------------------------------------------------

type System struct {
	bus        *events.Bus
	buffer     *components.TermBuffer
	scrollback *components.Scrollback

	offset     int
	scrollStep int

	termScrollSub   <-chan events.Event
	scrollUpSub     <-chan events.Event
	scrollDownSub   <-chan events.Event
	scrollPageUpSub <-chan events.Event
	scrollPageDnSub <-chan events.Event
}

// NewSystem connects the TermBuffer and Scrollback to the ECS bus.
func NewSystem(bus *events.Bus, buffer *components.TermBuffer, sb *components.Scrollback) *System {
	s := &System{
		bus:             bus,
		buffer:          buffer,
		scrollback:      sb,
		scrollStep:      5,
		termScrollSub:   bus.Subscribe("term_scrolled"),
		scrollUpSub:     bus.Subscribe("scroll_up"),
		scrollDownSub:   bus.Subscribe("scroll_down"),
		scrollPageUpSub: bus.Subscribe("scroll_page_up"),
		scrollPageDnSub: bus.Subscribe("scroll_page_down"),
	}

	go s.listenEvents()
	s.subscribeConfigChanges() // live hotreload integration

	return s
}

func (s *System) UpdateECS() {
	s.clampOffset()
}

var _ ecs.System = (*System)(nil)

