package parser

import (
	"log"
	"strconv"
	"strings"
	"unicode"

	"gost/internal/components"
	"gost/internal/events"
)

// -----------------------------------------------------------------------------
// ANSI Parser State Machine
// -----------------------------------------------------------------------------

const (
	stateText = iota
	stateEsc
	stateCSI
	stateOsc
)

// System consumes PTY output and updates the terminal buffer.
type System struct {
	bus    *events.Bus
	buffer *components.TermBuffer
	sub    <-chan events.Event

	state  int
	escBuf stringBuilder

	cx, cy         int // cursor position
	fg, bg         int // current color attributes
	savedX, savedY int // saved cursor for ESC7/ESC8
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

// -----------------------------------------------------------------------------
// ECS Integration
// -----------------------------------------------------------------------------

func (s *System) UpdateECS() {
	select {
	case evt := <-s.sub:
		if data, ok := evt.([]byte); ok {
			s.feed(string(data))
		}
	default:
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

// -----------------------------------------------------------------------------
// Input Feed
// -----------------------------------------------------------------------------

func (s *System) feed(input string) {
	for _, r := range input {
		switch s.state {
		case stateText:
			switch r {
			case '\x1b':
				s.state = stateEsc
			default:
				s.putChar(r)
			}
		case stateEsc:
			switch r {
			case '[':
				s.state = stateCSI
				s.escBuf.Reset()
			case ']':
				s.state = stateOsc
				s.escBuf.Reset()
			case '7': // Save cursor
				s.savedX, s.savedY = s.cx, s.cy
				s.state = stateText
			case '8': // Restore cursor
				s.cx, s.cy = s.savedX, s.savedY
				s.clipCursor()
				s.syncCursor()
				s.state = stateText
			default:
				s.state = stateText
			}
		case stateCSI:
			if (r >= '0' && r <= '9') || r == ';' {
				s.escBuf.WriteRune(r)
				continue
			}
			s.executeCSI(r)
			s.state = stateText
		case stateOsc:
			if r == '\x07' { // BEL terminates OSC
				s.state = stateText
			} else {
				s.escBuf.WriteRune(r)
			}
		}
	}
}

// -----------------------------------------------------------------------------
// Character Output
// -----------------------------------------------------------------------------

func (s *System) putChar(r rune) {
	switch r {
	case '\r': // carriage return
		s.cx = 0
	case '\n': // newline
		s.cy++
		if s.cy >= s.buffer.Height {
			s.buffer.ScrollUp()
			s.cy = s.buffer.Height - 1
		}
	case '\b': // backspace
		if s.cx > 0 {
			s.cx--
			s.buffer.SetRune(s.cx, s.cy, ' ', s.fg, s.bg)
		}
	default:
		if unicode.IsPrint(r) {
			if s.cx >= s.buffer.Width {
				s.cx = 0
				s.cy++
				if s.cy >= s.buffer.Height {
					s.buffer.ScrollUp()
					s.cy = s.buffer.Height - 1
				}
			}
			s.buffer.SetRune(s.cx, s.cy, r, s.fg, s.bg)
			s.cx++
		}
	}
	s.clipCursor()
	s.syncCursor()
}

// -----------------------------------------------------------------------------
// CSI (Control Sequence Introducer) Commands
// -----------------------------------------------------------------------------

func (s *System) executeCSI(final rune) {
	args := s.parseArgs(s.escBuf.String())
	switch final {
	case 'A': // Cursor Up
		n := s.argOr(args, 0, 1)
		s.cy -= n
	case 'B': // Cursor Down
		n := s.argOr(args, 0, 1)
		s.cy += n
	case 'C': // Cursor Right
		n := s.argOr(args, 0, 1)
		s.cx += n
	case 'D': // Cursor Left
		n := s.argOr(args, 0, 1)
		s.cx -= n
	case 'H', 'f': // Cursor Position
		y := s.argOr(args, 0, 1) - 1
		x := s.argOr(args, 1, 1) - 1
		s.cx, s.cy = x, y
	case 'J': // Erase in Display
		s.eraseDisplay(s.argOr(args, 0, 0))
	case 'K': // Erase in Line
		mode := s.argOr(args, 0, 0)
		switch mode {
		case 0:
			s.eraseToLineEnd()
		case 1:
			s.eraseFromLineStart()
		case 2:
			s.eraseFullLine()
		}
	case 'm': // SGR (Select Graphic Rendition)
		s.applySGR(args)
	default:
		// unrecognized sequence
	}
	s.clipCursor()
	s.syncCursor()
}

// -----------------------------------------------------------------------------
// SGR (Select Graphic Rendition)
// -----------------------------------------------------------------------------

func (s *System) applySGR(args []int) {
	if len(args) == 0 {
		s.fg, s.bg = 7, 0
		return
	}
	for i := 0; i < len(args); i++ {
		code := args[i]
		switch {
		case code == 0:
			s.fg, s.bg = 7, 0
		case code >= 30 && code <= 37:
			s.fg = code - 30
		case code >= 40 && code <= 47:
			s.bg = code - 40
		case code == 39:
			s.fg = 7
		case code == 49:
			s.bg = 0
		case code == 38 || code == 48:
			if i+2 < len(args) && args[i+1] == 5 {
				if code == 38 {
					s.fg = args[i+2]
				} else {
					s.bg = args[i+2]
				}
				i += 2
			}
		}
	}
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func (s *System) parseArgs(seq string) []int {
	if seq == "" {
		return nil
	}
	parts := strings.Split(seq, ";")
	args := make([]int, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			args = append(args, 0)
			continue
		}
		if n, err := strconv.Atoi(p); err == nil {
			args = append(args, n)
		}
	}
	return args
}

func (s *System) argOr(args []int, idx, def int) int {
	if idx < len(args) {
		return args[idx]
	}
	return def
}

// -----------------------------------------------------------------------------
// Erase & Cursor Management
// -----------------------------------------------------------------------------

func (s *System) eraseDisplay(mode int) {
	switch mode {
	case 0: // from cursor to end
		for y := s.cy; y < s.buffer.Height; y++ {
			startX := 0
			if y == s.cy {
				startX = s.cx
			}
			for x := startX; x < s.buffer.Width; x++ {
				s.buffer.SetRune(x, y, ' ', s.fg, s.bg)
			}
		}
	case 1: // from start to cursor
		for y := 0; y <= s.cy; y++ {
			endX := s.buffer.Width - 1
			if y == s.cy {
				endX = s.cx
			}
			for x := 0; x <= endX; x++ {
				s.buffer.SetRune(x, y, ' ', s.fg, s.bg)
			}
		}
	case 2: // entire screen
		s.buffer.Clear()
		s.cx, s.cy = 0, 0
	}
}

func (s *System) eraseToLineEnd() {
	for x := s.cx; x < s.buffer.Width; x++ {
		s.buffer.SetRune(x, s.cy, ' ', s.fg, s.bg)
	}
}

func (s *System) eraseFromLineStart() {
	for x := 0; x <= s.cx; x++ {
		s.buffer.SetRune(x, s.cy, ' ', s.fg, s.bg)
	}
}

func (s *System) eraseFullLine() {
	for x := 0; x < s.buffer.Width; x++ {
		s.buffer.SetRune(x, s.cy, ' ', s.fg, s.bg)
	}
	s.cx = 0
}

func (s *System) clipCursor() {
	if s.cx < 0 {
		s.cx = 0
	}
	if s.cy < 0 {
		s.cy = 0
	}
	if s.cx >= s.buffer.Width {
		s.cx = s.buffer.Width - 1
	}
	if s.cy >= s.buffer.Height {
		s.cy = s.buffer.Height - 1
	}
}

func (s *System) syncCursor() {
	s.buffer.SetCursor(s.cx, s.cy)
}

// -----------------------------------------------------------------------------
// Lightweight String Builder
// -----------------------------------------------------------------------------

type stringBuilder struct{ buf []rune }

func (b *stringBuilder) Reset()           { b.buf = b.buf[:0] }
func (b *stringBuilder) WriteRune(r rune) { b.buf = append(b.buf, r) }
func (b *stringBuilder) String() string   { return string(b.buf) }

