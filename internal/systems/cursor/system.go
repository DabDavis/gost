package cursor

import (
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"gost/internal/events"
	"gost/internal/components"
)

// System draws a static cursor at the parser’s current position.
type System struct {
	bus   *events.Bus
	mu    sync.RWMutex // protects term pointer
	term  *components.TermBuffer
	cellW int
	cellH int
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
				cs.mu.Lock()
				cs.term = tb
				cs.mu.Unlock()
			}
		}
	}()
	return cs
}

// UpdateECS does nothing — static cursor, no blinking.
func (c *System) UpdateECS() {}

// getTerm safely reads the current term pointer.
func (c *System) getTerm() *components.TermBuffer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.term
}

// DrawCursor renders the cursor as a solid rectangle or underline.
func (c *System) DrawCursor(screen *ebiten.Image, cx, cy int) {
	tb := c.getTerm()
	if tb == nil {
		return
	}

	tbX, tbY := tb.GetCursor() // now uses thread-safe getter
	if cx < 0 || cy < 0 || cy >= tb.Height || cx >= tb.Width {
		cx, cy = tbX, tbY
	}

	cursorColor := color.RGBA{200, 200, 200, 180}

	ebitenutil.DrawRect(
		screen,
		float64(cx*c.cellW),
		float64(cy*c.cellH),
		float64(c.cellW),
		float64(c.cellH),
		cursorColor,
	)
}

