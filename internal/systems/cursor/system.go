package cursor

import (
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"gost/internal/components"
	"gost/internal/events"
)

// System handles cursor rendering & blink logic.
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
		bus:   bus,
		cellW: cellW,
		cellH: cellH,
		style: defaultCursorStyle(),
		blinkVisible: true,
		lastBlink:    time.Now(),
	}

	cs.subscribeTermUpdates()
	cs.subscribeConfigChanges()
	return cs
}

// UpdateECS handles blinking state toggling.
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

// DrawCursor renders the cursor shape if visible.
func (c *System) DrawCursor(screen *ebiten.Image, cx, cy int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.blinkVisible || c.term == nil {
		return
	}
	tb := c.term
	tbX, tbY := tb.GetCursor()

	if cx < 0 || cy < 0 || cy >= tb.Height || cx >= tb.Width {
		cx, cy = tbX, tbY
	}

	switch c.style.Shape {
	case "underline":
		c.drawUnderline(screen, cx, cy)
	default:
		c.drawBlock(screen, cx, cy)
	}
}

// Thread-safe getter for term buffer
func (c *System) getTerm() *components.TermBuffer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.term
}
// Draw is the public entrypoint for ECS rendering.
// It wraps DrawCursor with thread-safe visibility handling.
func (c *System) Draw(screen *ebiten.Image) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    if !c.blinkVisible || c.term == nil {
        return
    }

    cx, cy := c.term.GetCursor()
    switch c.style.Shape {
    case "underline":
        c.drawUnderline(screen, cx, cy)
    default:
        c.drawBlock(screen, cx, cy)
    }
}
