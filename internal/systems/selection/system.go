package selection

import (
	"gost/internal/components"
	"gost/internal/events"
)

// System handles click-drag text selection and clipboard copy.
type System struct {
	buffer    *components.TermBuffer
	bus       *events.Bus
	selecting bool
	startX, startY int
	endX, endY     int
	cellW, cellH   int
}

// NewSystem initializes a new selection handler with ECS bus linkage.
func NewSystem(buffer *components.TermBuffer, cellW, cellH int, bus *events.Bus) *System {
	return &System{
		buffer: buffer,
		bus:    bus,
		cellW:  cellW,
		cellH:  cellH,
	}
}

// UpdateECS runs selection input logic each tick.
func (s *System) UpdateECS() {
	s.UpdateSelectionInput()
}

