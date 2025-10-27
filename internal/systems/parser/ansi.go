package parser

import (
	"strconv"
	"strings"
)

// feed consumes PTY text and interprets ANSI / VT escape sequences.
func (s *System) feed(data string) {
	for _, r := range data {
		switch s.state {

		// ------------------------------------------------------------
		// Normal text mode
		// ------------------------------------------------------------
		case stateText:
			if r == 0x1b { // ESC
				s.state = stateEsc
				s.escBuf.Reset()
			} else {
				s.putChar(r)
			}

		// ------------------------------------------------------------
		// ESC prefix — begin special sequence
		// ------------------------------------------------------------
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
				s.state = stateText // unknown ESC
			}

		// ------------------------------------------------------------
		// CSI: ESC [ ... command
		// ------------------------------------------------------------
		case stateCSI:
			s.escBuf.WriteRune(r)
			if r >= '@' && r <= '~' {
				s.execCSI(s.escBuf.String())
				s.state = stateText
			}

		// ------------------------------------------------------------
		// OSC: ESC ] ... terminated by BEL or ESC \
		// ------------------------------------------------------------
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

// -----------------------------------------------------------------------------
// CSI HANDLER
// -----------------------------------------------------------------------------

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

	case 'm': // SGR (Select Graphic Rendition)
		s.execSGR(nums)

	default:
		// unsupported CSI, ignore gracefully
	}

	s.clipCursor()
	s.syncCursor()
}

// -----------------------------------------------------------------------------
// OSC HANDLER
// -----------------------------------------------------------------------------

// execOSC handles OSC (Operating System Command) sequences like ESC ] 0;title BEL.
func (s *System) execOSC(seq string) {
	if strings.HasPrefix(seq, "0;") || strings.HasPrefix(seq, "2;") {
		// ignore window title updates
		return
	}
	// unsupported OSC codes are ignored safely
}

// -----------------------------------------------------------------------------
// SGR (Select Graphic Rendition) — Color + Attribute Handling
// -----------------------------------------------------------------------------

// execSGR applies color attributes (8,16,256,truecolor) and resets.
func (s *System) execSGR(nums []int) {
	if len(nums) == 0 {
		nums = []int{0}
	}

	i := 0
	for i < len(nums) {
		n := nums[i]

		switch {
		// --- Reset ---
		case n == 0:
			s.fg, s.bg = 7, 0

		// --- Standard 8 colors (30–37, 40–47) ---
		case n >= 30 && n <= 37:
			s.fg = n - 30
		case n >= 40 && n <= 47:
			s.bg = n - 40

		// --- Bright 16 colors (90–97, 100–107) ---
		case n >= 90 && n <= 97:
			s.fg = (n - 90) + 8
		case n >= 100 && n <= 107:
			s.bg = (n - 100) + 8

		// --- Extended color (256 or Truecolor) ---
		case n == 38 || n == 48:
			if i+1 < len(nums) {
				mode := nums[i+1]

				switch mode {
				case 5: // 256-color: ESC[38;5;n or ESC[48;5;n
					if i+2 < len(nums) {
						idx := clampColorIndex(nums[i+2])
						if n == 38 {
							s.fg = idx
						} else {
							s.bg = idx
						}
						i += 2
					}

				case 2: // Truecolor: ESC[38;2;r;g;b
					if i+4 < len(nums) {
						r, g, b := nums[i+2], nums[i+3], nums[i+4]
						idx := rgbTo256(r, g, b)
						if n == 38 {
							s.fg = idx
						} else {
							s.bg = idx
						}
						i += 4
					}
				}
			}

		// --- Ignore bold, underline, etc. ---
		default:
			// future styling: 1=bold,4=underline,etc.
		}

		i++
	}
}

// clampColorIndex bounds a color index to valid 0–255.
func clampColorIndex(n int) int {
	if n < 0 {
		return 0
	}
	if n > 255 {
		return 255
	}
	return n
}

// rgbTo256 approximates truecolor RGB to nearest 256-color index (6×6×6 cube).
func rgbTo256(r, g, b int) int {
	clamp := func(v int) int {
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return v
	}
	r, g, b = clamp(r), clamp(g), clamp(b)
	r6 := int(float64(r) / 255 * 5)
	g6 := int(float64(g) / 255 * 5)
	b6 := int(float64(b) / 255 * 5)
	return 16 + (36*r6) + (6*g6) + b6
}

