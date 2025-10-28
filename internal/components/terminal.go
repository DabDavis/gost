package components

import (
	"sync"

	"gost/internal/events"
)

// -----------------------------------------------------------------------------
// Glyph
// -----------------------------------------------------------------------------

// Glyph represents a single cell in the terminal grid.
type Glyph struct {
	Rune rune // Unicode codepoint
	Fg   int  // Foreground color index (0–7)
	Bg   int  // Background color index (0–7)
}

// -----------------------------------------------------------------------------
// TermBuffer
// -----------------------------------------------------------------------------

// TermBuffer stores the live visible screen contents of the terminal.
// It supports concurrent access and event-based scrollback integration.
type TermBuffer struct {
	mu               sync.RWMutex
	Width, Height    int
	Cells            [][]Glyph
	CursorX, CursorY int

	bus *events.Bus
}

// NewTermBuffer allocates a clean terminal grid.
func NewTermBuffer(width, height int) *TermBuffer {
	tb := &TermBuffer{
		Width:  width,
		Height: height,
		Cells:  make([][]Glyph, height),
	}
	for y := range tb.Cells {
		tb.Cells[y] = make([]Glyph, width)
	}
	tb.Clear()
	return tb
}

// AttachBus links the terminal to an event bus for async scrollback notifications.
func (tb *TermBuffer) AttachBus(bus *events.Bus) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.bus = bus
}

// Clear resets all cells to blank space.
func (tb *TermBuffer) Clear() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	for y := 0; y < tb.Height; y++ {
		for x := 0; x < tb.Width; x++ {
			tb.Cells[y][x] = Glyph{Rune: ' ', Fg: 7, Bg: 0}
		}
	}
	tb.CursorX, tb.CursorY = 0, 0
}

// SetRune writes a rune at (x, y).
func (tb *TermBuffer) SetRune(x, y int, r rune, fg, bg int) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if x < 0 || y < 0 || y >= tb.Height || x >= tb.Width {
		return
	}
	tb.Cells[y][x] = Glyph{Rune: r, Fg: fg, Bg: bg}
}

// GetRune returns a glyph at (x, y), or blank if out of bounds.
func (tb *TermBuffer) GetRune(x, y int) Glyph {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	if x < 0 || y < 0 || y >= tb.Height || x >= tb.Width {
		return Glyph{Rune: ' ', Fg: 7, Bg: 0}
	}
	return tb.Cells[y][x]
}

// ScrollUp shifts all lines up, clears the bottom row,
// and publishes the top line to scrollback if attached.
func (tb *TermBuffer) ScrollUp() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.Height == 0 || tb.Width == 0 {
		return
	}

	top := make([]Glyph, tb.Width)
	copy(top, tb.Cells[0])

	// Scroll up
	copy(tb.Cells, tb.Cells[1:])
	tb.Cells[tb.Height-1] = make([]Glyph, tb.Width)
	for i := range tb.Cells[tb.Height-1] {
		tb.Cells[tb.Height-1][i] = Glyph{Rune: ' ', Fg: 7, Bg: 0}
	}

	// Cursor safety
	if tb.CursorY > 0 {
		tb.CursorY--
	}

	// Fire event for scrollback capture
	if tb.bus != nil {
		line := make([]Glyph, len(top))
		copy(line, top)
		go tb.bus.Publish("term_scrolled", line)
	}
}

// Resize adjusts the terminal grid, preserving contents.
func (tb *TermBuffer) Resize(newW, newH int) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	newCells := make([][]Glyph, newH)
	for y := 0; y < newH; y++ {
		newCells[y] = make([]Glyph, newW)
		if y < tb.Height {
			copy(newCells[y], tb.Cells[y][:min(newW, tb.Width)])
		}
	}

	tb.Width, tb.Height = newW, newH
	tb.Cells = newCells

	if tb.CursorY >= newH {
		tb.CursorY = newH - 1
	}
	if tb.CursorX >= newW {
		tb.CursorX = newW - 1
	}
}

// Safe cursor helpers
func (tb *TermBuffer) SetCursor(x, y int) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.CursorX, tb.CursorY = x, y
}

func (tb *TermBuffer) GetCursor() (int, int) {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.CursorX, tb.CursorY
}

// Locking exposure (for advanced systems only)
func (tb *TermBuffer) Lock()   { tb.mu.Lock() }
func (tb *TermBuffer) Unlock() { tb.mu.Unlock() }
func (tb *TermBuffer) RLock()  { tb.mu.RLock() }
func (tb *TermBuffer) RUnlock() { tb.mu.RUnlock() }

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// -----------------------------------------------------------------------------
// Scrollback
// -----------------------------------------------------------------------------

// Scrollback preserves lines that scrolled off the terminal display.
// It supports concurrent reads by render + selection systems.
type Scrollback struct {
	mu    sync.RWMutex
	Lines [][]Glyph
	Max   int
}

// NewScrollback allocates a history buffer with the given maximum lines.
func NewScrollback(max int) *Scrollback {
	return &Scrollback{
		Lines: make([][]Glyph, 0, max),
		Max:   max,
	}
}

// PushLine appends a line, dropping the oldest if full.
func (sb *Scrollback) PushLine(line []Glyph) {
	if sb == nil || len(line) == 0 {
		return
	}

	sb.mu.Lock()
	defer sb.mu.Unlock()

	copyLine := make([]Glyph, len(line))
	copy(copyLine, line)

	if len(sb.Lines) >= sb.Max {
		copy(sb.Lines, sb.Lines[1:])
		sb.Lines[len(sb.Lines)-1] = copyLine
	} else {
		sb.Lines = append(sb.Lines, copyLine)
	}
}

// Count returns total lines in scrollback.
func (sb *Scrollback) Count() int {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return len(sb.Lines)
}

// GetLine fetches a historical line (0 = oldest).
func (sb *Scrollback) GetLine(index int) []Glyph {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	if index < 0 || index >= len(sb.Lines) {
		return nil
	}
	line := make([]Glyph, len(sb.Lines[index]))
	copy(line, sb.Lines[index])
	return line
}

// GetVisibleLines returns a deep copy of visible lines given an offset + height.
func (sb *Scrollback) GetVisibleLines(offset, height int) [][]Glyph {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	n := len(sb.Lines)
	if n == 0 {
		return nil
	}

	// Clamp
	if offset > n {
		offset = n
	} else if offset < 0 {
		offset = 0
	}

	start := n - offset - height
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > n {
		end = n
	}

	lines := make([][]Glyph, 0, end-start)
	for _, src := range sb.Lines[start:end] {
		row := make([]Glyph, len(src))
		copy(row, src)
		lines = append(lines, row)
	}
	return lines
}

// Clear erases all history.
func (sb *Scrollback) Clear() {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.Lines = sb.Lines[:0]
}

