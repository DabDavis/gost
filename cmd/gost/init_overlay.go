package main

import (
	"gost/internal/events"
	"gost/internal/systems/overlay"
)

func initOverlay(bus *events.Bus, s *GameSystems) {
	configOverlay := overlay.NewConfigMessageOverlay(bus)
	s.Overlay.AddLayer(s.Selection)
	s.Overlay.AddLayer(s.Cursor)
	s.Overlay.AddLayer(configOverlay)
}

