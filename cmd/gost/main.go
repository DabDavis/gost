package main

import (
	"fmt"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"gost/internal/ecs"
	"gost/internal/events"

	"gost/internal/systems/config"
	"gost/internal/systems/cursor"
	"gost/internal/systems/hotreload"
	"gost/internal/systems/input"
	"gost/internal/systems/overlay"
	"gost/internal/systems/parser"
	"gost/internal/systems/pty"
	"gost/internal/systems/render"
	"gost/internal/systems/scrollback"
	"gost/internal/systems/selection"

	"gost/internal/components"
)

// -----------------------------------------------------------------------------
// GameSystems bundles all ECS systems for clarity and reuse.
// -----------------------------------------------------------------------------
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

// -----------------------------------------------------------------------------
// initSystems: Construct and wire up all ECS subsystems.
// -----------------------------------------------------------------------------
func initSystems(bus *events.Bus, world *ecs.World) *GameSystems {
	// Components
	term := components.NewTermBuffer(80, 24)
	term.AttachBus(bus)
	sb := components.NewScrollback(1000)

	// Config
	cfg := config.NewSystem(bus, "config.json")
	hr := hotreload.NewSystem(bus, "config.json")

	// Core rendering and interaction systems
	renderSys := render.NewSystem(bus)
	renderSys.AttachTerm(term)
	renderSys.AttachScrollback(sb)

	cursorSys := cursor.NewSystem(bus, 7, 14)
	cursorSys.AttachTerm(term)

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

// -----------------------------------------------------------------------------
// registerSystems: Register ECS systems in strict priority order.
// -----------------------------------------------------------------------------
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

// -----------------------------------------------------------------------------
// Game: Ebiten integration — drives ECS update loop and rendering.
// -----------------------------------------------------------------------------
type Game struct {
	world   *ecs.World
	systems *GameSystems
	bus     *events.Bus
	started time.Time
}

// StartGame initializes ECS, systems, and starts Ebiten loop.
func StartGame() error {
	bus := events.NewBus()
	world := ecs.NewWorld()

	systems := initSystems(bus, world)
	registerSystems(world, systems)

	game := &Game{
		world:   world,
		systems: systems,
		bus:     bus,
		started: time.Now(),
	}

	ebiten.SetWindowTitle("GoST — Modular ECS Terminal Emulator")
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(960, 540)

	return ebiten.RunGame(game)
}

// -----------------------------------------------------------------------------
// Ebiten API compliance
// -----------------------------------------------------------------------------

func (g *Game) Update() error {
	g.world.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.systems.Render != nil {
		g.systems.Render.Draw(screen)
	}

	if g.systems.Overlay != nil {
		g.systems.Overlay.Draw(screen)
	}
}

func (g *Game) Layout(outW, outH int) (int, int) {
	if g.systems.Render != nil {
		return g.systems.Render.Layout(outW, outH)
	}
	return 640, 384
}

// -----------------------------------------------------------------------------
// Debug utilities
// -----------------------------------------------------------------------------

func debugPrintSystems(world *ecs.World) {
	fmt.Println(world.Describe())
	fmt.Printf("[Debug] %d systems active.\n", world.Count())
}

// -----------------------------------------------------------------------------
// Main entry point
// -----------------------------------------------------------------------------

func main() {
	log.Println("GoST — modular ECS terminal emulator starting...")
	if err := StartGame(); err != nil {
		log.Fatal(err)
	}
}

