package components

import "sync"

// Glyph represents a single character cell on the terminal screen.
type Glyph struct {
	Rune rune // Unicode codepoint to render
	Fg   int  // Foreground color index (0–7)
	Bg   int  // Background color index (0–7)
}

// TermBuffer represents the full 2D terminal screen —
// storing all visible characters and their attributes.
type TermBuffer struct {
	mu                sync.RWMutex
	Width, Height     int
	Cells             [][]Glyph
	CursorX, CursorY  int
}

// NewTermBuffer allocates a new blank terminal buffer of given size.
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

// Clear wipes the screen, resetting all cells to a blank space.
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

// SetRune writes a rune at (x, y) with the given colors.
func (tb *TermBuffer) SetRune(x, y int, r rune, fg, bg int) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if x < 0 || y < 0 || y >= tb.Height || x >= tb.Width {
		return
	}
	tb.Cells[y][x] = Glyph{Rune: r, Fg: fg, Bg: bg}
}

// GetRune retrieves a glyph from (x, y), returning a blank cell if out of range.
func (tb *TermBuffer) GetRune(x, y int) Glyph {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	if x < 0 || y < 0 || y >= tb.Height || x >= tb.Width {
		return Glyph{Rune: ' ', Fg: 7, Bg: 0}
	}
	return tb.Cells[y][x]
}

// ScrollUp shifts all lines up by one and clears the last row.
func (tb *TermBuffer) ScrollUp() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	copy(tb.Cells, tb.Cells[1:])
	tb.Cells[tb.Height-1] = make([]Glyph, tb.Width)
	for x := range tb.Cells[tb.Height-1] {
		tb.Cells[tb.Height-1][x] = Glyph{Rune: ' ', Fg: 7, Bg: 0}
	}
	if tb.CursorY > 0 {
		tb.CursorY--
	}
}

// Resize adjusts the terminal dimensions, preserving existing content.
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

// min utility
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

