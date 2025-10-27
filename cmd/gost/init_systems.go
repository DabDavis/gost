package main

import (
	"gost/internal/components"
	"gost/internal/events"

	"gost/internal/systems/config"
	"gost/internal/systems/cursor"
	"gost/internal/systems/input"
	"gost/internal/systems/overlay"
	"gost/internal/systems/parser"
	"gost/internal/systems/pty"
	"gost/internal/systems/render"
	"gost/internal/systems/scrollback"
	"gost/internal/systems/selection"
	"gost/internal/systems/hotreload"
	"gost/internal/ecs"
)

type GameSystems struct {
	Config     *config.System
	HotReload  *hotreload.System
	Render     *render.System
	Cursor     *cursor.System
	Input      *input.System
	Selection  *selection.System
	Scrollback *scrollback.System
	Overlay    *overlay.System
	Parser     *parser.System
	PTY        *pty.System
}

// initSystems wires all core ECS systems.
func initSystems(bus *events.Bus, world *ecs.World) *GameSystems {
	// Components
	term := components.NewTermBuffer(80, 24)
	term.AttachBus(bus)
	sb := components.NewScrollback(1000)

	// Config
	cfg := config.NewSystem(bus, "config.json")
	hr := hotreload.NewSystem(bus, "config.json")

	// Core
	renderSys := render.NewSystem(bus)
	renderSys.AttachTerm(term)
	renderSys.AttachScrollback(sb)

	cursorSys := cursor.NewSystem(bus, 7, 14)
	inputSys := input.NewSystem(bus)
	selectionSys := selection.NewSystem(renderSys.Buffer(), 7, 14, bus)
	scrollbackSys := scrollback.NewSystem(bus, term, sb)
	parserSys := parser.NewSystem(bus, term)
	ptySys := pty.NewSystem(bus)

	overlaySys := overlay.NewSystem()

	return &GameSystems{
		Config:     cfg,
		HotReload:  hr,
		Render:     renderSys,
		Cursor:     cursorSys,
		Input:      inputSys,
		Selection:  selectionSys,
		Scrollback: scrollbackSys,
		Overlay:    overlaySys,
		Parser:     parserSys,
		PTY:        ptySys,
	}
}

// registerSystems attaches systems in priority order.
func registerSystems(world *ecs.World, s *GameSystems) {
	world.AddSystem(s.Config, ecs.PriorityConfig)
	world.AddSystem(s.HotReload, ecs.PriorityHotReload)
	world.AddSystem(s.Input, ecs.PriorityInput)
	world.AddSystem(s.PTY, ecs.PriorityPTY)
	world.AddSystem(s.Parser, ecs.PriorityParser)
	world.AddSystem(s.Scrollback, ecs.PriorityScrollback)
	world.AddSystem(s.Render, ecs.PriorityRender)
	world.AddSystem(s.Selection, ecs.PrioritySelection)
	world.AddSystem(s.Cursor, ecs.PriorityCursor)
	world.AddSystem(s.Overlay, ecs.PriorityOverlay)
}

