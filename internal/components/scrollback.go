package components

import "sync"

// Scrollback stores historical terminal lines that have scrolled off-screen.
// It is safe for concurrent access by render and scrollback systems.
type Scrollback struct {
	mu    sync.RWMutex
	Lines [][]Glyph
	Max   int // maximum number of stored lines
}

// NewScrollback creates a new scrollback buffer with the given capacity.
func NewScrollback(max int) *Scrollback {
	return &Scrollback{
		Lines: make([][]Glyph, 0, max),
		Max:   max,
	}
}

// PushLine appends a copy of a scrolled line to the history.
// Older lines are dropped when capacity is reached.
func (sb *Scrollback) PushLine(line []Glyph) {
	if sb == nil || len(line) == 0 {
		return
	}

	sb.mu.Lock()
	defer sb.mu.Unlock()

	// Copy to prevent shared slice mutation
	cpy := make([]Glyph, len(line))
	copy(cpy, line)

	// Append line
	if len(sb.Lines) >= sb.Max {
		// Drop oldest line (FIFO)
		copy(sb.Lines, sb.Lines[1:])
		sb.Lines[len(sb.Lines)-1] = cpy
	} else {
		sb.Lines = append(sb.Lines, cpy)
	}
}

// Count returns the number of lines currently stored.
func (sb *Scrollback) Count() int {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return len(sb.Lines)
}

// GetLine safely retrieves a specific historical line (0 = oldest).
// Returns nil if out of range.
func (sb *Scrollback) GetLine(index int) []Glyph {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	if index < 0 || index >= len(sb.Lines) {
		return nil
	}

	cpy := make([]Glyph, len(sb.Lines[index]))
	copy(cpy, sb.Lines[index])
	return cpy
}

// GetVisibleLines returns a slice of up to `height` lines starting
// `offset` lines from the newest entry (for renderer scrollback view).
func (sb *Scrollback) GetVisibleLines(offset, height int) [][]Glyph {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	n := len(sb.Lines)
	if n == 0 {
		return nil
	}

	// Clamp offset within history
	if offset > n {
		offset = n
	} else if offset < 0 {
		offset = 0
	}

	// Compute visible range
	start := n - offset - height
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > n {
		end = n
	}

	// Copy the requested window (deep copy of each line)
	lines := make([][]Glyph, 0, end-start)
	for _, src := range sb.Lines[start:end] {
		dst := make([]Glyph, len(src))
		copy(dst, src)
		lines = append(lines, dst)
	}
	return lines
}

// Clear removes all stored lines.
func (sb *Scrollback) Clear() {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.Lines = sb.Lines[:0]
}

