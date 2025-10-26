package cursor

import (
    "image/color"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "gost/internal/events"
    "gost/internal/components"
)

// System draws a static cursor at the parser’s current position.
type System struct {
    bus    *events.Bus
    term   *components.TermBuffer
    cellW  int
    cellH  int
}

// NewSystem subscribes to "term_updated" events to track the buffer.
func NewSystem(bus *events.Bus, cellW, cellH int) *System {
    cs := &System{
        bus:   bus,
        cellW: cellW,
        cellH: cellH,
    }

    sub := bus.Subscribe("term_updated")
    go func() {
        for evt := range sub {
            if tb, ok := evt.(*components.TermBuffer); ok {
                cs.term = tb
            }
        }
    }()
    return cs
}

// UpdateECS does nothing — static cursor, no blinking.
func (c *System) UpdateECS() {}

// DrawCursor renders the cursor as a solid rectangle or underline.
func (c *System) DrawCursor(screen *ebiten.Image, cx, cy int) {
    if c.term == nil {
        return
    }

    if cx < 0 || cy < 0 || cy >= c.term.Height || cx >= c.term.Width {
        return
    }

    // --- Choose cursor style ---
    // solid block:
    cursorColor := color.RGBA{200, 200, 200, 180}
    ebitenutil.DrawRect(
        screen,
        float64(cx*c.cellW),
        float64(cy*c.cellH),
        float64(c.cellW),
        float64(c.cellH),
        cursorColor,
    )

    // --- OR underline style ---
    // cursorColor := color.RGBA{255, 255, 255, 200}
    // ebitenutil.DrawRect(
    //     screen,
    //     float64(cx*c.cellW),
    //     float64((cy+1)*c.cellH-2),
    //     float64(c.cellW),
    //     2,
    //     cursorColor,
    // )
}

