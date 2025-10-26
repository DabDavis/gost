package main

import (
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"gost/internal/ecs"
	"gost/internal/events"
	"gost/internal/components"

	// Modular ECS systems
	"gost/internal/systems/input"
	"gost/internal/systems/parser"
	"gost/internal/systems/pty"
	"gost/internal/systems/render"
	"gost/internal/systems/cursor"
	"gost/internal/systems/selection"
	"gost/internal/systems/scrollback"
)

// GameSystem wraps the ECS world and delegates rendering to Ebiten.
type GameSystem struct {
	world        *ecs.World
	bus          *events.Bus
	renderSys    *render.System
	cursorSys    *cursor.System
	selectionSys *selection.System
	exitSub      <-chan events.Event
}

// Update runs every tick (~60 Hz) and advances all ECS systems.
func (g *GameSystem) Update() error {
	g.world.Update()

	// Check for "system_exit" event
	select {
	case <-g.exitSub:
		log.Println("[GoST] Received system_exit event — shutting down gracefully.")
		os.Exit(0)
	default:
	}

	return nil
}

// Draw runs each frame and delegates to render + overlays.
func (g *GameSystem) Draw(screen *ebiten.Image) {
	// --- Render terminal grid ---
	g.renderSys.Draw(screen)

	// --- Draw text selection overlay (if any) ---
	if g.selectionSys != nil {
		g.selectionSys.DrawSelection(screen)
	}

	// --- Draw static cursor ---
	if g.cursorSys != nil && g.renderSys != nil && g.renderSys.Buffer() != nil {
		buf := g.renderSys.Buffer()
		g.cursorSys.DrawCursor(screen, buf.CursorX, buf.CursorY)
	}
}

// Layout defines logical pixel size based on terminal cell grid.
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

	// --- Core terminal components ---
	termBuffer := components.NewTermBuffer(80, 24)
	termBuffer.AttachBus(bus)
	scrollbackBuf := components.NewScrollback(1000)

	// --- Instantiate modular systems ---
	ptySys := pty.NewSystem(bus)
	parserSys := parser.NewSystem(bus)
	renderSys := render.NewSystem(bus)
	renderSys.AttachScrollback(scrollbackBuf) // link scrollback to renderer
	inputSys := input.NewSystem(bus)
	cursorSys := cursor.NewSystem(bus, 7, 14)
	selectionSys := selection.NewSystem(renderSys.Buffer(), 7, 14)
	scrollbackSys := scrollback.NewSystem(bus, termBuffer, scrollbackBuf)

	// --- Register all systems with the ECS world ---
	world.AddSystem(ptySys)
	world.AddSystem(parserSys)
	world.AddSystem(renderSys)
	world.AddSystem(inputSys)
	world.AddSystem(cursorSys)
	world.AddSystem(selectionSys)
	world.AddSystem(scrollbackSys)

	// --- Subscribe to exit signal ---
	exitSub := bus.Subscribe("system_exit")

	// --- Wrap ECS in Ebiten adapter ---
	game := &GameSystem{
		world:        world,
		bus:          bus,
		renderSys:    renderSys,
		cursorSys:    cursorSys,
		selectionSys: selectionSys,
		exitSub:      exitSub,
	}

	// --- Ebiten setup ---
	ebiten.SetWindowTitle("GoST Terminal")
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(800, 480)
	ebiten.SetMaxTPS(60)

	// --- Run game loop ---
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

