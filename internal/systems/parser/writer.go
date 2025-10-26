package parser

import "unicode"

// putChar handles printable output and control characters.
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

// eraseToLineEnd implements ESC[K.
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

// lightweight rune builder (avoids allocations)
type stringBuilder struct{ buf []rune }

func (b *stringBuilder) Reset()           { b.buf = b.buf[:0] }
func (b *stringBuilder) WriteRune(r rune) { b.buf = append(b.buf, r) }
func (b *stringBuilder) String() string   { return string(b.buf) }

