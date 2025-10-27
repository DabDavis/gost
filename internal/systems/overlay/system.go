package overlay

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"gost/internal/ecs"
)

// -----------------------------------------------------------------------------
// Overlay System — compositing manager for multiple visual layers
// -----------------------------------------------------------------------------

type System struct {
	mu     sync.RWMutex
	layers []Drawable
}

// NewSystem creates an empty overlay compositor.
func NewSystem() *System {
	return &System{
		layers: make([]Drawable, 0, 8), // small prealloc
	}
}

// AddLayer registers a Drawable in z-order sequence.
// Layers are drawn in the same order they’re added.
func (o *System) AddLayer(d Drawable) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.layers = append(o.layers, d)
}

// RemoveLayer removes a layer reference (safe, non-panic).
func (o *System) RemoveLayer(target Drawable) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for i, l := range o.layers {
		if l == target {
			o.layers = append(o.layers[:i], o.layers[i+1:]...)
			break
		}
	}
}

// UpdateECS runs fade/animation updates for overlays that implement ECSSystem.
func (o *System) UpdateECS() {
	o.mu.RLock()
	defer o.mu.RUnlock()
	for _, l := range o.layers {
		if sys, ok := l.(ecs.System); ok {
			sys.UpdateECS()
		}
	}
}

// Draw iterates through all layers and renders them in sequence.
func (o *System) Draw(screen *ebiten.Image) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	for _, layer := range o.layers {
		layer.Draw(screen)
	}
}

// Compile-time interface check
var _ ecs.System = (*System)(nil)

