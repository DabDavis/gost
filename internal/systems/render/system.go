package render

import (
    "image/color"
    "sync"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/text"
    "golang.org/x/image/font"
    "golang.org/x/image/font/basicfont"

    "gost/internal/events"
    "gost/internal/components"
)

// System draws the terminal buffer using a cached monospace bitmap font.
type System struct {
    bus      *events.Bus
    term     *components.TermBuffer
    mu       sync.RWMutex

    fontFace font.Face
    cellW, cellH int

    fgPalette, bgPalette [8]color.Color
    bgTiles              [8]*ebiten.Image // cached 8 background color tiles
}

// NewSystem initializes the renderer and subscribes to terminal updates.
func NewSystem(bus *events.Bus) *System {
    rs := &System{
        bus:      bus,
        fontFace: basicfont.Face7x13,
        cellW:    7,
        cellH:    14,
        fgPalette: [8]color.Color{
            color.RGBA{0, 0, 0, 255},       // black
            color.RGBA{205, 49, 49, 255},   // red
            color.RGBA{13, 188, 121, 255},  // green
            color.RGBA{229, 229, 16, 255},  // yellow
            color.RGBA{36, 114, 200, 255},  // blue
            color.RGBA{188, 63, 188, 255},  // magenta
            color.RGBA{17, 168, 205, 255},  // cyan
            color.RGBA{229, 229, 229, 255}, // white
        },
        bgPalette: [8]color.Color{
            color.RGBA{0, 0, 0, 255},
            color.RGBA{60, 0, 0, 255},
            color.RGBA{0, 60, 0, 255},
            color.RGBA{60, 60, 0, 255},
            color.RGBA{0, 0, 60, 255},
            color.RGBA{60, 0, 60, 255},
            color.RGBA{0, 60, 60, 255},
            color.RGBA{80, 80, 80, 255},
        },
    }

    // Precache reusable background tiles
    for i := range rs.bgPalette {
        tile := ebiten.NewImage(rs.cellW, rs.cellH)
        tile.Fill(rs.bgPalette[i])
        rs.bgTiles[i] = tile
    }

    // Subscribe to term updates
    sub := bus.Subscribe("term_updated")
    go func() {
        for evt := range sub {
            if tb, ok := evt.(*components.TermBuffer); ok {
                rs.mu.Lock()
                rs.term = tb
                rs.mu.Unlock()
            }
        }
    }()

    return rs
}

// Buffer exposes the current terminal buffer (for selection/cursor systems).
func (r *System) Buffer() *components.TermBuffer {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.term
}

// UpdateECS (no-op for render system).
func (r *System) UpdateECS() {}

// Draw renders the terminal buffer efficiently using cached backgrounds.
func (r *System) Draw(screen *ebiten.Image) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    screen.Fill(color.Black)
    if r.term == nil {
        return
    }

    for y := 0; y < r.term.Height; y++ {
        for x := 0; x < r.term.Width; x++ {
            g := r.term.Cells[y][x]
            bgIndex := g.Bg % 8
            fgIndex := g.Fg % 8

            // --- Draw cached background tile ---
            op := &ebiten.DrawImageOptions{}
            op.GeoM.Translate(float64(x*r.cellW), float64(y*r.cellH))
            screen.DrawImage(r.bgTiles[bgIndex], op)

            // --- Draw character ---
            if g.Rune == 0 || g.Rune == ' ' {
                continue
            }
            fg := r.fgPalette[fgIndex]
            text.Draw(screen, string(g.Rune), r.fontFace,
                x*r.cellW, y*r.cellH+r.cellH-2, fg)
        }
    }
}

// Layout defines logical screen size (in pixels) based on terminal buffer.
func (r *System) Layout(outW, outH int) (int, int) {
    if r.term != nil {
        return r.term.Width * r.cellW, r.term.Height * r.cellH
    }
    return 640, 384
}

