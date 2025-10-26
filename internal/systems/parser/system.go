package parser

import (
	"log"

	"gost/internal/components"
	"gost/internal/events"
)

// --- Parser states ---
const (
	stateText = iota
	stateEsc
	stateCSI
	stateOsc // Operating System Command (ESC ])
)

// System consumes PTY output and updates the terminal buffer.
type System struct {
	bus    *events.Bus
	buffer *components.TermBuffer
	sub    <-chan events.Event

	state  int
	escBuf stringBuilder

	cx, cy       int // cursor position
	fg, bg       int // current color attributes
	savedX, savedY int // saved cursor (for ESC7/ESC8)
}

// NewSystem subscribes to PTY output and initializes parser state.
func NewSystem(bus *events.Bus, tb *components.TermBuffer) *System {
	ps := &System{
		bus:    bus,
		buffer: tb,
		sub:    bus.Subscribe("pty_output"),
		fg:     7,
		bg:     0,
	}
	return ps
}

// UpdateECS processes any queued PTY output.
func (s *System) UpdateECS() {
	select {
	case evt := <-s.sub:
		data, ok := evt.([]byte)
		if !ok {
			return
		}
		s.feed(string(data))
	default:
		// nothing queued
	}
}

// Reset clears parser state.
func (s *System) Reset() {
	s.state = stateText
	s.cx, s.cy = 0, 0
	s.fg, s.bg = 7, 0
	s.savedX, s.savedY = 0, 0
	s.escBuf.Reset()
	log.Println("[Parser] reset state")
}

