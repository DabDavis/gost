package render

import "gost/internal/components"

// Viewport combines scrollback and live terminal buffer
// into a unified slice of glyph lines ready for rendering.
type Viewport struct {
    scrollback *components.Scrollback
    term       *components.TermBuffer
    offset     int // scrollback offset
}

// NewViewport initializes a viewport with both live and history sources.
func NewViewport(sb *components.Scrollback, term *components.TermBuffer) *Viewport {
    return &Viewport{
        scrollback: sb,
        term:       term,
        offset:     0,
    }
}

// SetOffset updates the current scrollback offset (0 = live view).
func (v *Viewport) SetOffset(offset int) {
    v.offset = offset
}

// Compose builds the visible glyph matrix given the terminal height.
// It merges scrollback + term buffer into a continuous visual frame.
func (v *Viewport) Compose(height int) [][]components.Glyph {
    if v.term == nil {
        return nil
    }

    lines := [][]components.Glyph{}

    // If scrollback is visible, prepend history lines
    if v.scrollback != nil && v.offset > 0 {
        sbLines := v.scrollback.GetVisibleLines(v.offset, v.term.Height)
        lines = append(lines, sbLines...)
    }

    // Always append live terminal content safely
    v.term.RLock()
    lines = append(lines, v.term.Cells...)
    v.term.RUnlock()

    // Clamp to visible height
    if len(lines) > height {
        lines = lines[len(lines)-height:]
    }
    return lines
}

