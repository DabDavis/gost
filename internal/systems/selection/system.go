package selection

import (
	"log"
	"strings"
	"sync"

	"gost/internal/components"
	"gost/internal/events"
	"gost/internal/util"
)

// -----------------------------------------------------------------------------
// Selection System
// -----------------------------------------------------------------------------

type System struct {
	buffer *components.TermBuffer
	bus    *events.Bus
	mu     sync.RWMutex

	selecting bool
	startX, startY int
	endX, endY     int
	cellW, cellH   int
}

// NewSystem initializes a new selection handler with ECS bus linkage.
func NewSystem(buffer *components.TermBuffer, cellW, cellH int, bus *events.Bus) *System {
	s := &System{
		buffer: buffer,
		bus:    bus,
		cellW:  cellW,
		cellH:  cellH,
	}
	s.subscribeEvents()
	return s
}

// -----------------------------------------------------------------------------
// Event Subscriptions
// -----------------------------------------------------------------------------

func (s *System) subscribeEvents() {
	if s.bus == nil {
		return
	}

	startSub := s.bus.Subscribe("selection_start")
	updateSub := s.bus.Subscribe("selection_update")
	endSub := s.bus.Subscribe("selection_end")
	clearSub := s.bus.Subscribe("selection_clear")
	copySub := s.bus.Subscribe("selection_copy")

	go func() {
		for evt := range startSub {
			if pos, ok := evt.(map[string]int); ok {
				s.BeginSelection(pos["x"], pos["y"])
			}
		}
	}()
	go func() {
		for evt := range updateSub {
			if pos, ok := evt.(map[string]int); ok {
				s.UpdateSelection(pos["x"], pos["y"])
			}
		}
	}()
	go func() {
		for evt := range endSub {
			if pos, ok := evt.(map[string]int); ok {
				s.EndSelection(pos["x"], pos["y"])
			}
		}
	}()
	go func() {
		for range clearSub {
			s.Clear()
		}
	}()
	go func() {
		for range copySub {
			s.CopyToClipboard()
		}
	}()
}

// -----------------------------------------------------------------------------
// Selection Logic
// -----------------------------------------------------------------------------

func (s *System) BeginSelection(px, py int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selecting = true
	s.startX, s.startY = s.pixelToCell(px, py)
	s.endX, s.endY = s.startX, s.startY
	s.bus.Publish("selection_changed", s.Bounds())
}

func (s *System) UpdateSelection(px, py int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.selecting {
		return
	}
	s.endX, s.endY = s.pixelToCell(px, py)
	s.bus.Publish("selection_changed", s.Bounds())
}

func (s *System) EndSelection(px, py int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selecting = true // keep visible until cleared manually
	s.endX, s.endY = s.pixelToCell(px, py)
	s.bus.Publish("selection_finished", s.Bounds())
}

func (s *System) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selecting = false
	s.startX, s.startY, s.endX, s.endY = 0, 0, 0, 0
	s.bus.Publish("selection_cleared", nil)
}

// -----------------------------------------------------------------------------
// Copy Functionality (uses util clipboard shim)
// -----------------------------------------------------------------------------

func (s *System) CopyToClipboard() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.buffer == nil || !s.selecting {
		return
	}

	b := s.Bounds()
	var sb strings.Builder
	for y := b["y1"]; y <= b["y2"] && y < s.buffer.Height; y++ {
		for x := b["x1"]; x <= b["x2"] && x < s.buffer.Width; x++ {
			g := s.buffer.GetRune(x, y)
			sb.WriteRune(g.Rune)
		}
		if y < b["y2"] {
			sb.WriteByte('\n')
		}
	}
	text := sb.String()
	if text == "" {
		return
	}

	util.SetClipboardString(text)
	log.Printf("[Selection] Copied %d characters", len(text))
	s.bus.Publish("selection_copied", text)
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func (s *System) Bounds() map[string]int {
	sx, ex := s.startX, s.endX
	sy, ey := s.startY, s.endY
	if sx > ex {
		sx, ex = ex, sx
	}
	if sy > ey {
		sy, ey = ey, sy
	}
	return map[string]int{"x1": sx, "y1": sy, "x2": ex, "y2": ey}
}

func (s *System) pixelToCell(px, py int) (int, int) {
	return px / s.cellW, py / s.cellH
}

// -----------------------------------------------------------------------------
// ECS Hook
// -----------------------------------------------------------------------------

func (s *System) UpdateECS() {}

