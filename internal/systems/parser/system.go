package parser

import (
    "strconv"
    "strings"
    "unicode"

    "gost/internal/events"
    "gost/internal/components"
)

// System interprets PTY output and updates the TermBuffer
// by parsing printable characters and ANSI escape sequences.
type System struct {
    bus    *events.Bus
    sub    <-chan events.Event
    buffer *components.TermBuffer
    cx, cy int
    fg, bg int
    state  int
    escBuf strings.Builder
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
        bus: bus,
        sub: bus.Subscribe("pty_output"),
        buffer: components.NewTermBuffer(80, 24),
        fg: 7,
        bg: 0,
    }
    return s
}

// UpdateECS processes any new PTY output events.
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

// feed processes raw PTY output byte-by-byte, handling ANSI state transitions.
func (s *System) feed(data string) {
    for _, r := range data {
        switch s.state {
        case stateText:
            if r == 0x1b {
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
            // command terminator (A–Z or @–~)
            if r >= '@' && r <= '~' {
                seq := s.escBuf.String()
                s.execCSI(seq)
                s.state = stateText
            }
        }
    }
    s.syncCursor()
}

// execCSI executes a parsed ANSI control sequence.
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
        if s.cy < 0 {
            s.cy = 0
        }

    case 'B': // cursor down
        s.cy += nums[0]
        if s.cy >= s.buffer.Height {
            s.cy = s.buffer.Height - 1
        }

    case 'C': // cursor right
        s.cx += nums[0]
        if s.cx >= s.buffer.Width {
            s.cx = s.buffer.Width - 1
        }

    case 'D': // cursor left
        s.cx -= nums[0]
        if s.cx < 0 {
            s.cx = 0
        }

    case 'H', 'f': // move cursor to row;col
        y, x := 1, 1
        if len(nums) >= 2 {
            y, x = nums[0], nums[1]
        }
        s.cx, s.cy = x-1, y-1
        s.clipCursor()

    case 'J': // erase screen
        if nums[0] == 2 {
            s.buffer.Clear()
        }

    case 'K': // erase to end of line (used by backspace redraw)
        if s.cx > 0 {
            s.cx-- // backspace-aware
        }
        for x := s.cx; x < s.buffer.Width; x++ {
            s.buffer.Cells[s.cy][x] = components.Glyph{
                Rune: ' ',
                Fg:   s.fg,
                Bg:   s.bg,
            }
        }

    case 'm': // SGR (color/style)
        s.applySGR(nums)
    }

    s.syncCursor()
}

// applySGR applies text attributes like colors.
func (s *System) applySGR(nums []int) {
    if len(nums) == 0 {
        nums = []int{0}
    }
    for _, n := range nums {
        switch {
        case n == 0:
            s.fg, s.bg = 7, 0 // reset
        case n >= 30 && n <= 37:
            s.fg = n - 30
        case n >= 40 && n <= 47:
            s.bg = n - 40
        }
    }
}

// putChar writes a printable character at the current cursor position.
func (s *System) putChar(r rune) {
    switch r {
    case '\r':
        s.cx = 0
    case '\n':
        s.cy++
        if s.cy >= s.buffer.Height {
            s.scrollUp()
            s.cy = s.buffer.Height - 1
        }
    default:
        if unicode.IsPrint(r) {
            if s.cx >= s.buffer.Width {
                s.cx = 0
                s.cy++
                if s.cy >= s.buffer.Height {
                    s.scrollUp()
                    s.cy = s.buffer.Height - 1
                }
            }
            s.buffer.Cells[s.cy][s.cx] = components.Glyph{
                Rune: r,
                Fg:   s.fg,
                Bg:   s.bg,
            }
            s.cx++
        }
    }
    s.syncCursor()
}

// scrollUp scrolls the terminal content by one line.
func (s *System) scrollUp() {
    copy(s.buffer.Cells, s.buffer.Cells[1:])
    s.buffer.Cells[s.buffer.Height-1] = make([]components.Glyph, s.buffer.Width)
}

// clipCursor clamps the cursor within valid screen bounds.
func (s *System) clipCursor() {
    if s.cx < 0 {
        s.cx = 0
    }
    if s.cx >= s.buffer.Width {
        s.cx = s.buffer.Width - 1
    }
    if s.cy < 0 {
        s.cy = 0
    }
    if s.cy >= s.buffer.Height {
        s.cy = s.buffer.Height - 1
    }
}

// syncCursor keeps the buffer’s cursor fields up to date.
func (s *System) syncCursor() {
    s.buffer.CursorX = s.cx
    s.buffer.CursorY = s.cy
}

