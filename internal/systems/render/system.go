package render

import (
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"

	"github.com/hajimehoshi/ebiten/v2"

	"gost/internal/events"
	"gost/internal/components"
)

// System draws the terminal buffer (and scrollback) using cached backgrounds and font.
type System struct {
	bus        *events.Bus
	term       *components.TermBuffer
	scrollback *components.Scrollback
	mu         sync.RWMutex

	fontFace font.Face
	cellW, cellH int

	scrollOffset int // lines scrolled into history
	scrollDirty  bool // if true, return to live view on next keypress

	fgPalette, bgPalette [8]colorRGBA
	bgTiles              [8]*ebiten.Image
}

// NewSystem initializes renderer, subscribes to term updates & scroll offset changes.
func NewSystem(bus *events.Bus) *System {
	rs := &System{
		bus:      bus,
		fontFace: basicfont.Face7x13,
		cellW:    7,
		cellH:    14,
		fgPalette: defaultFgPalette(),
		bgPalette: defaultBgPalette(),
	}

	rs.precacheTiles()
	rs.subscribeEvents()

	return rs
}

// AttachScrollback links renderer to Scrollback buffer (optional).
func (r *System) AttachScrollback(sb *components.Scrollback) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scrollback = sb
}

// Buffer exposes the current terminal buffer.
func (r *System) Buffer() *components.TermBuffer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.term
}

// UpdateECS no-op (all drawing is in Draw()).
func (r *System) UpdateECS() {}

// Layout defines logical screen size in pixels.
func (r *System) Layout(outW, outH int) (int, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.term != nil {
		return r.term.Width * r.cellW, r.term.Height * r.cellH
	}
	return 640, 384
}

