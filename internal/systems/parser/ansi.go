package parser

import (
	"strconv"
	"strings"
)

// feed consumes PTY text and interprets ANSI / VT escape sequences.
func (s *System) feed(data string) {
	for _, r := range data {
		switch s.state {

		// --- Normal text ---
		case stateText:
			if r == 0x1b { // ESC
				s.state = stateEsc
				s.escBuf.Reset()
			} else {
				s.putChar(r)
			}

		// --- ESC prefix ---
		case stateEsc:
			switch r {
			case '[':
				s.state = stateCSI
				s.escBuf.Reset()
			case ']':
				s.state = stateOsc
				s.escBuf.Reset()
			case '7': // DECSC (Save cursor)
				s.savedX, s.savedY = s.cx, s.cy
				s.state = stateText
			case '8': // DECRC (Restore cursor)
				s.cx, s.cy = s.savedX, s.savedY
				s.clipCursor()
				s.syncCursor()
				s.state = stateText
			default:
				// Unrecognized single-char ESC — ignore
				s.state = stateText
			}

		// --- CSI: ESC [ ... command ---
		case stateCSI:
			s.escBuf.WriteRune(r)
			if r >= '@' && r <= '~' {
				s.execCSI(s.escBuf.String())
				s.state = stateText
			}

		// --- OSC: ESC ] ... BEL or ESC \ terminates ---
		case stateOsc:
			if r == '\a' {
				s.execOSC(s.escBuf.String())
				s.state = stateText
			} else if r == '\\' && len(s.escBuf.buf) > 0 && s.escBuf.buf[len(s.escBuf.buf)-1] == 0x1b {
				s.execOSC(s.escBuf.String())
				s.state = stateText
			} else {
				s.escBuf.WriteRune(r)
			}
		}
	}
}

// execCSI interprets CSI (ESC [) control sequences.
func (s *System) execCSI(seq string) {
	if len(seq) == 0 {
		return
	}
	cmd := seq[len(seq)-1]
	params := strings.Split(seq[:len(seq)-1], ";")

	var nums []int
	for _, p := range params {
		if p == "" {
			nums = append(nums, 0)
		} else if n, err := strconv.Atoi(p); err == nil {
			nums = append(nums, n)
		}
	}
	if len(nums) == 0 {
		nums = []int{0}
	}

	switch cmd {

	case 'A': // cursor up
		s.cy -= nums[0]
	case 'B': // cursor down
		s.cy += nums[0]
	case 'C': // cursor right
		s.cx += nums[0]
	case 'D': // cursor left
		s.cx -= nums[0]
	case 'H', 'f': // move to row;col
		y, x := 1, 1
		if len(nums) >= 2 {
			y, x = nums[0], nums[1]
		}
		s.cx, s.cy = x-1, y-1

	case 'J': // clear screen
		if nums[0] == 2 {
			s.buffer.Clear()
			s.cx, s.cy = 0, 0
		}
	case 'K': // erase line variants
		switch nums[0] {
		case 0:
			s.eraseToLineEnd()
		case 1:
			s.eraseFromLineStart()
		case 2:
			s.eraseFullLine()
		}
	case 'm': // colors
		s.execSGR(nums)
	default:
		// unsupported CSI, ignore gracefully
	}

	s.clipCursor()
	s.syncCursor()
}

// execOSC handles OSC (Operating System Command) sequences.
func (s *System) execOSC(seq string) {
	// OSC examples:
	//   ESC ] 0;title BEL   → set title
	//   ESC ] 2;title BEL   → set icon/title
	if strings.HasPrefix(seq, "0;") || strings.HasPrefix(seq, "2;") {
		// ignore safely (title updates)
		return
	}
	// unsupported OSC codes ignored safely
}

// execSGR applies color attributes and resets.
func (s *System) execSGR(nums []int) {
	if len(nums) == 0 {
		nums = []int{0}
	}
	for _, n := range nums {
		switch {
		case n == 0: // reset
			s.fg, s.bg = 7, 0
		case n >= 30 && n <= 37: // standard fg
			s.fg = n - 30
		case n >= 40 && n <= 47: // standard bg
			s.bg = n - 40
		case n >= 90 && n <= 97: // bright fg
			s.fg = (n - 90) + 8
		case n >= 100 && n <= 107: // bright bg
			s.bg = (n - 100) + 8
		// 256/24-bit color placeholders (for future)
		case n == 38 || n == 48:
			// skip next params (advanced colors)
		default:
			// ignore unsupported attributes (bold, underline, etc.)
		}
	}
}

