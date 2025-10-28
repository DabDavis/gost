package overlay

import (
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"gost/internal/events"
)

// -----------------------------------------------------------------------------
// SelectionLayer — persistent rectangular highlight renderer
// -----------------------------------------------------------------------------

type SelectionLayer struct {
	bus   *events.Bus
	mu    sync.RWMutex
	bounds map[string]int
	active bool
	cellW, cellH int
	color color.Color
}

// NewSelectionLayer creates a highlight overlay that listens for selection events.
func NewSelectionLayer(bus *events.Bus, cellW, cellH int) *SelectionLayer {
	sl := &SelectionLayer{
		bus:   bus,
		cellW: cellW,
		cellH: cellH,
		color: color.RGBA{80, 120, 255, 100}, // translucent blue highlight
		bounds: map[string]int{"x1": 0, "y1": 0, "x2": 0, "y2": 0},
	}

	sl.subscribeSelectionEvents()
	return sl
}

// -----------------------------------------------------------------------------
// Event Wiring
// -----------------------------------------------------------------------------

func (s *SelectionLayer) subscribeSelectionEvents() {
	if s.bus == nil {
		return
	}

	changedSub := s.bus.Subscribe("selection_changed")
	finishedSub := s.bus.Subscribe("selection_finished")
	clearedSub := s.bus.Subscribe("selection_cleared")

	go func() {
		for evt := range changedSub {
			if b, ok := evt.(map[string]int); ok {
				s.setBounds(b)
				s.setActive(true)
			}
		}
	}()
	go func() {
		for evt := range finishedSub {
			if b, ok := evt.(map[string]int); ok {
				s.setBounds(b)
				s.setActive(true)
			}
		}
	}()
	go func() {
		for range clearedSub {
			s.setActive(false)
		}
	}()
}

// -----------------------------------------------------------------------------
// Internal Helpers
// -----------------------------------------------------------------------------

func (s *SelectionLayer) setBounds(b map[string]int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bounds = b
}

func (s *SelectionLayer) setActive(active bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = active
}

// -----------------------------------------------------------------------------
// Draw
// -----------------------------------------------------------------------------

// Draw renders a translucent selection rectangle if active.
func (s *SelectionLayer) Draw(screen *ebiten.Image) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.active {
		return
	}

	b := s.bounds
	x1 := b["x1"]
	y1 := b["y1"]
	x2 := b["x2"]
	y2 := b["y2"]

	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			x0 := float64(x * s.cellW)
			y0 := float64(y * s.cellH)
			ebitenutil.DrawRect(screen, x0, y0, float64(s.cellW), float64(s.cellH), s.color)
		}
	}
}

