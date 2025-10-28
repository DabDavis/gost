package render

import (
    "image/color"
    "sync"

    "github.com/hajimehoshi/ebiten/v2"
    "golang.org/x/image/font"
    "golang.org/x/image/font/basicfont"

    "gost/internal/components"
    "gost/internal/ecs"
    "gost/internal/events"
)

// System draws the composed terminal view using cached backgrounds and font.
// It listens to configuration and scrollback events dynamically.
type System struct {
    scrollDirty bool
    mu          sync.RWMutex

    bus        *events.Bus
    term       *components.TermBuffer
    scrollback *components.Scrollback
    viewport   *Viewport

    fontFace font.Face
    cellW, cellH int

    scrollOffset int // scrollback offset (updated by event listener)
    fgPalette    [8]colorRGBA
    bgPalette    [8]colorRGBA
    bgTiles      [8]*ebiten.Image
    bgColor      color.Color

    configSub <-chan events.Event // config change listener
    offsetSub <-chan events.Event // scroll offset listener
}

// NewSystem initializes the renderer and prepares background tiles.
func NewSystem(bus *events.Bus) *System {
    r := &System{
        bus:       bus,
        fontFace:  basicfont.Face7x13,
        cellW:     7,
        cellH:     14,
        fgPalette: defaultFgPalette(),
        bgPalette: defaultBgPalette(),
        bgColor:   color.Black,
    }

    r.precacheTiles()
    return r
}

// Layout defines logical render size based on terminal cell grid.
func (r *System) Layout(outW, outH int) (int, int) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    if r.term != nil {
        return r.term.Width * r.cellW, r.term.Height * r.cellH
    }
    return 640, 384
}

// UpdateECS is called each tick by the ECS world.
// Rendering happens separately in Draw(), so this is a no-op.
func (r *System) UpdateECS() {
    // no update logic required for rendering
}

// AttachTerm links the render system to the live terminal buffer.
func (r *System) AttachTerm(term *components.TermBuffer) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.term = term
}

// AttachScrollback links the render system to the scrollback buffer.
func (r *System) AttachScrollback(sb *components.Scrollback) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.scrollback = sb
}

// Buffer safely returns the terminal buffer (used by Selection).
func (r *System) Buffer() *components.TermBuffer {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.term
}

// Interface check for ECS compliance.
var _ ecs.System = (*System)(nil)

