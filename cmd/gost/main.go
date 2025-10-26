package main

import (
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"gost/internal/ecs"
	"gost/internal/events"
	"gost/internal/components"

	// ECS Systems
	"gost/internal/systems/input"
	"gost/internal/systems/parser"
	"gost/internal/systems/pty"
	"gost/internal/systems/render"
	"gost/internal/systems/cursor"
	"gost/internal/systems/selection"
	"gost/internal/systems/scrollback"
	"gost/internal/systems/overlay"
)

// GameSystem is the top-level Ebiten adapter wrapping the ECS World.
type GameSystem struct {
	world        *ecs.World
	bus          *events.Bus
	renderSys    *render.System
	cursorSys    *cursor.System
	selectionSys *selection.System
	overlaySys   *overlay.System
	exitSub      <-chan events.Event
}

// Update advances ECS systems and listens for shutdown events.
func (g *GameSystem) Update() error {
	g.world.Update()

	select {
	case <-g.exitSub:
		log.Println("[GoST] Received system_exit event — shutting down gracefully.")
		os.Exit(0)
	default:
	}

	return nil
}

// Draw renders terminal layers, overlays, and visual elements.
func (g *GameSystem) Draw(screen *ebiten.Image) {
	// --- Base terminal rendering ---
	g.renderSys.Draw(screen)

	// --- Text selection overlay ---
	if g.selectionSys != nil {
		g.selectionSys.DrawSelection(screen)
	}

	// --- Cursor overlay ---
	if g.cursorSys != nil && g.renderSys != nil && g.renderSys.Buffer() != nil {
		buf := g.renderSys.Buffer()
		g.cursorSys.DrawCursor(screen, buf.CursorX, buf.CursorY)
	}

	// --- Transient HUD overlay ---
	if g.overlaySys != nil {
		w, h := ebiten.WindowSize()
		g.overlaySys.DrawOverlay(screen, w, h)
	}
}

// Layout defines logical pixel size based on terminal cell geometry.
func (g *GameSystem) Layout(outW, outH int) (int, int) {
	if g.renderSys != nil {
		return g.renderSys.Layout(outW, outH)
	}
	return 640, 384
}

func main() {
	log.Println("GoST — modular ECS terminal emulator starting...")

	// --- Core ECS setup ---
	bus := events.NewBus()
	world := ecs.NewWorld()

	// --- Terminal components ---
	termBuffer := components.NewTermBuffer(80, 24)
	termBuffer.AttachBus(bus)
	scrollbackBuf := components.NewScrollback(1000)

	// --- Modular systems ---
	ptySys := pty.NewSystem(bus)
	parserSys := parser.NewSystem(bus, termBuffer)
	renderSys := render.NewSystem(bus)
	renderSys.AttachScrollback(scrollbackBuf)
	renderSys.AttachTerm(termBuffer)
	inputSys := input.NewSystem(bus)
	cursorSys := cursor.NewSystem(bus, 7, 14)
	selectionSys := selection.NewSystem(renderSys.Buffer(), 7, 14, bus)
	scrollbackSys := scrollback.NewSystem(bus, termBuffer, scrollbackBuf)
	overlaySys := overlay.NewSystem(bus)

	// --- Register systems with explicit priorities ---
	world.AddSystem(inputSys, ecs.PriorityInput)
	world.AddSystem(ptySys, ecs.PriorityPTY)
	world.AddSystem(parserSys, ecs.PriorityParser)
	world.AddSystem(scrollbackSys, ecs.PriorityScrollback)
	world.AddSystem(renderSys, ecs.PriorityRender)
	world.AddSystem(cursorSys, ecs.PriorityCursor)
	world.AddSystem(selectionSys, ecs.PrioritySelection)
	world.AddSystem(overlaySys, ecs.PriorityOverlay)

	// --- Exit signal subscription ---
	exitSub := bus.Subscribe("system_exit")

	// --- Wrap ECS world for Ebiten ---
	game := &GameSystem{
		world:        world,
		bus:          bus,
		renderSys:    renderSys,
		cursorSys:    cursorSys,
		selectionSys: selectionSys,
		overlaySys:   overlaySys,
		exitSub:      exitSub,
	}

	// --- Ebiten runtime configuration ---
	ebiten.SetWindowTitle("GoST Terminal")
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(800, 480)
	ebiten.SetMaxTPS(60)

	// --- Run main loop ---
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

