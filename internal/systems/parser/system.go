package parser

import (
	"gost/internal/components"
	"gost/internal/events"
)

// System interprets PTY output and updates the TermBuffer
// by parsing printable characters and ANSI escape sequences.
type System struct {
	bus    *events.Bus
	sub    <-chan events.Event
	buffer *components.TermBuffer

	cx, cy int // cursor position
	fg, bg int // colors
	state  int // parser state
	escBuf stringBuilder
}

// parser states
const (
	stateText = iota
	stateEsc
	stateCSI
)

// NewSystem creates a new ANSI-aware parser for terminal output.
func NewSystem(bus *events.Bus) *System {
	s := &System{
		bus:    bus,
		sub:    bus.Subscribe("pty_output"),
		buffer: components.NewTermBuffer(80, 24),
		fg:     7,
		bg:     0,
	}
	return s
}

// UpdateECS processes PTY output events and updates the terminal buffer.
func (s *System) UpdateECS() {
	select {
	case evt := <-s.sub:
		if text, ok := evt.(string); ok {
			s.feed(text)
			s.syncCursor()
			s.bus.Publish("term_updated", s.buffer)
		}
	default:
	}
}

// Buffer exposes the active terminal buffer.
func (s *System) Buffer() *components.TermBuffer { return s.buffer }

// Reset clears the parser state and buffer.
func (s *System) Reset() {
	s.buffer.Clear()
	s.cx, s.cy = 0, 0
	s.fg, s.bg = 7, 0
	s.state = stateText
}

