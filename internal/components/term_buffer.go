package components

import (
	"sync"

	"gost/internal/events"
)

// Glyph represents a single character cell on the terminal screen.
type Glyph struct {
	Rune rune // Unicode codepoint to render
	Fg   int  // Foreground color index (0–7)
	Bg   int  // Background color index (0–7)
}

// TermBuffer represents the full 2D terminal screen —
// storing all visible characters and their attributes.
type TermBuffer struct {
	mu               sync.RWMutex
	Width, Height    int
	Cells            [][]Glyph
	CursorX, CursorY int

	bus *events.Bus // optional event bus for scrollback integration
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

// AttachBus links an event bus to the TermBuffer, enabling it to emit events.
func (tb *TermBuffer) AttachBus(bus *events.Bus) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.bus = bus
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

// ScrollUp shifts all lines up by one, clears the last row,
// and emits the top line as a "term_scrolled" event if a bus is attached.
func (tb *TermBuffer) ScrollUp() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.Height == 0 || tb.Width == 0 {
		return
	}

	// Copy the top line before shifting
	topLine := make([]Glyph, tb.Width)
	copy(topLine, tb.Cells[0])

	// Scroll contents up
	copy(tb.Cells, tb.Cells[1:])
	tb.Cells[tb.Height-1] = make([]Glyph, tb.Width)
	for x := range tb.Cells[tb.Height-1] {
		tb.Cells[tb.Height-1][x] = Glyph{Rune: ' ', Fg: 7, Bg: 0}
	}

	// Move cursor up one if needed
	if tb.CursorY > 0 {
		tb.CursorY--
	}

	// Emit scroll event if bus is attached — async to avoid blocking
	if tb.bus != nil {
		lineCopy := make([]Glyph, len(topLine))
		copy(lineCopy, topLine)
		go tb.bus.Publish("term_scrolled", lineCopy)
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

// --- Safe cursor access ---

// SetCursor safely updates cursor position.
func (tb *TermBuffer) SetCursor(x, y int) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.CursorX = x
	tb.CursorY = y
}

// GetCursor safely retrieves cursor position.
func (tb *TermBuffer) GetCursor() (int, int) {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.CursorX, tb.CursorY
}

// --- Locking wrappers ---
//
// These methods expose safe locking access for external systems (e.g., render viewport)
// without exposing the internal mutex directly. Use them sparingly — most callers should
// prefer the higher-level SetRune/GetRune API.

func (tb *TermBuffer) Lock()   { tb.mu.Lock() }
func (tb *TermBuffer) Unlock() { tb.mu.Unlock() }
func (tb *TermBuffer) RLock()  { tb.mu.RLock() }
func (tb *TermBuffer) RUnlock() { tb.mu.RUnlock() }

// --- Helpers ---

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

