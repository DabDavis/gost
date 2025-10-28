package overlay

import (
    "github.com/hajimehoshi/ebiten/v2"
    "gost/internal/events"
)

// ConfigMessageOverlay is a simple layer that can later display config info or messages.
type ConfigMessageOverlay struct {
    bus *events.Bus
}

// NewConfigMessageOverlay creates a new overlay layer that listens for config messages.
func NewConfigMessageOverlay(bus *events.Bus) *ConfigMessageOverlay {
    return &ConfigMessageOverlay{bus: bus}
}

// Draw satisfies the overlay layer interface but currently does nothing.
func (c *ConfigMessageOverlay) Draw(screen *ebiten.Image) {}

// Optionally, if your overlay system expects Layout:
func (c *ConfigMessageOverlay) Layout(outW, outH int) (int, int) {
    return outW, outH
}

