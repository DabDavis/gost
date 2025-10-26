package parser

import (
	"strconv"
	"strings"
)

// feed processes PTY bytes and routes between plain text and ANSI states.
func (s *System) feed(data string) {
	for _, r := range data {
		switch s.state {
		case stateText:
			if r == 0x1b { // ESC
				s.state = stateEsc
				s.escBuf.Reset()
			} else {
				s.putChar(r)
			}

		case stateEsc:
			if r == '[' {
				s.state = stateCSI
				s.escBuf.Reset()
			} else {
				s.state = stateText
			}

		case stateCSI:
			s.escBuf.WriteRune(r)
			if r >= '@' && r <= '~' {
				s.execCSI(s.escBuf.String())
				s.state = stateText
			}
		}
	}
}

// execCSI executes ANSI control sequences (cursor moves, erases, colors).
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
		}

	case 'K': // erase line variants
		switch nums[0] {
		case 0: // ESC[K
			s.eraseToLineEnd()
		case 1: // ESC[1K
			s.eraseFromLineStart()
		case 2: // ESC[2K
			s.eraseFullLine()
		}

	case 'm': // colors
		s.execSGR(nums)
	}

	s.clipCursor()
	s.syncCursor()
}

// execSGR applies simple color attributes (SGR codes 0–47).
func (s *System) execSGR(nums []int) {
	if len(nums) == 0 {
		nums = []int{0}
	}
	for _, n := range nums {
		switch {
		case n == 0:
			s.fg, s.bg = 7, 0
		case n >= 30 && n <= 37:
			s.fg = n - 30
		case n >= 40 && n <= 47:
			s.bg = n - 40
		}
	}
}

