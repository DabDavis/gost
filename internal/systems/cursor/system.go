package cursor

import (
	"fmt"
	"image/color"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"gost/internal/components"
	"gost/internal/events"
	"gost/internal/ecs"
)

// -----------------------------------------------------------------------------
// Cursor Style Definition
// -----------------------------------------------------------------------------

type cursorStyle struct {
	Shape     string        // "block" or "underline"
	Color     color.Color   // cursor color
	Blink     bool          // blinking enabled
	BlinkRate time.Duration // blink interval
}

func defaultCursorStyle() cursorStyle {
	return cursorStyle{
		Shape:     "block",
		Color:     color.RGBA{255, 255, 255, 255},
		Blink:     true,
		BlinkRate: 600 * time.Millisecond,
	}
}

// -----------------------------------------------------------------------------
// Cursor System
// -----------------------------------------------------------------------------

type System struct {
	bus *events.Bus
	mu  sync.RWMutex

	term *components.TermBuffer
	cellW, cellH int

	style        cursorStyle
	blinkVisible bool
	lastBlink    time.Time
}

// NewSystem creates a new cursor system with defaults and subscribes to events.
func NewSystem(bus *events.Bus, cellW, cellH int) *System {
	cs := &System{
		bus:          bus,
		cellW:        cellW,
		cellH:        cellH,
		style:        defaultCursorStyle(),
		blinkVisible: true,
		lastBlink:    time.Now(),
	}
	cs.subscribeTermUpdates()
	cs.subscribeConfigChanges()
	return cs
}

// -----------------------------------------------------------------------------
// ECS Methods
// -----------------------------------------------------------------------------

func (c *System) UpdateECS() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.style.Blink {
		c.blinkVisible = true
		return
	}

	if time.Since(c.lastBlink) > c.style.BlinkRate {
		c.blinkVisible = !c.blinkVisible
		c.lastBlink = time.Now()
	}
}

func (c *System) Draw(screen *ebiten.Image) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.term == nil || !c.blinkVisible {
		return
	}

	cx, cy := c.term.GetCursor()
	if cx < 0 || cy < 0 || cy >= c.term.Height || cx >= c.term.Width {
		return
	}

	switch c.style.Shape {
	case "underline":
		c.drawUnderline(screen, cx, cy)
	default:
		c.drawBlock(screen, cx, cy)
	}
}

var _ ecs.System = (*System)(nil)

// -----------------------------------------------------------------------------
// Drawing Helpers
// -----------------------------------------------------------------------------

func (c *System) drawBlock(screen *ebiten.Image, cx, cy int) {
	ebitenutil.DrawRect(
		screen,
		float64(cx*c.cellW),
		float64(cy*c.cellH),
		float64(c.cellW),
		float64(c.cellH),
		c.style.Color,
	)
}

func (c *System) drawUnderline(screen *ebiten.Image, cx, cy int) {
	h := float64(c.cellH) / 6
	y := float64((cy+1)*c.cellH) - h
	ebitenutil.DrawRect(
		screen,
		float64(cx*c.cellW),
		y,
		float64(c.cellW),
		h,
		c.style.Color,
	)
}

// -----------------------------------------------------------------------------
// Config Integration
// -----------------------------------------------------------------------------

func (c *System) subscribeConfigChanges() {
	if c.bus == nil {
		return
	}
	sub := c.bus.Subscribe("cursor_config_changed")
	go func() {
		for evt := range sub {
			if cfg, ok := evt.(map[string]interface{}); ok {
				c.applyConfig(cfg)
			}
		}
	}()
}

func (c *System) applyConfig(cfg map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if shape, ok := cfg["shape"].(string); ok {
		c.style.Shape = shape
	}
	if blink, ok := cfg["blink"].(bool); ok {
		c.style.Blink = blink
	}
	if rate, ok := cfg["blink_rate"].(float64); ok {
		c.style.BlinkRate = time.Duration(rate) * time.Millisecond
	}
	if col, ok := cfg["color"].(color.Color); ok {
		c.style.Color = col
	}
}

// -----------------------------------------------------------------------------
// Term Buffer Integration
// -----------------------------------------------------------------------------

func (c *System) subscribeTermUpdates() {
	if c.bus == nil {
		return
	}
	sub := c.bus.Subscribe("term_updated")
	go func() {
		for evt := range sub {
			if tb, ok := evt.(*components.TermBuffer); ok {
				c.mu.Lock()
				c.term = tb
				c.mu.Unlock()
			}
		}
	}()
}

func (c *System) AttachTerm(term *components.TermBuffer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.term = term
}

// -----------------------------------------------------------------------------
// Utilities
// -----------------------------------------------------------------------------

func (c *System) getTerm() *components.TermBuffer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.term
}

func (c *System) Describe() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fmt.Sprintf("Cursor[%s, blink=%v, color=%v]",
		c.style.Shape, c.style.Blink, c.style.Color)
}

