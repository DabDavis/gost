package main

import (
	"gost/internal/ecs"
	"gost/internal/events"
)

// initECSWorld builds the ECS world and returns (world, bus, gameSystem).
func initECSWorld() (*ecs.World, *events.Bus, *GameSystem) {
	bus := events.NewBus()
	world := ecs.NewWorld()

	// Build systems (delegated to helpers)
	systems := initSystems(bus, world)

	// Build overlay layers
	initOverlay(bus, systems)

	// Register all systems in correct order
	registerSystems(world, systems)

	// Subscribe to exit event
	exitSub := bus.Subscribe("system_exit")

	game := &GameSystem{
		world:        world,
		bus:          bus,
		renderSys:    systems.Render,
		cursorSys:    systems.Cursor,
		selectionSys: systems.Selection,
		overlaySys:   systems.Overlay,
		exitSub:      exitSub,
	}

	return world, bus, game
}

