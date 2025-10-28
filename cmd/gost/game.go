package main

import (
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"gost/internal/ecs"
    	"gost/internal/events"
)

// GameSystem implements Ebiten’s Game interface.
type GameSystem struct {
	world        *ecs.World
	bus          *events.Bus
	renderSys    interface{ Draw(*ebiten.Image); Layout(int, int) (int, int) }
	cursorSys    interface{ Draw(*ebiten.Image) }
	selectionSys interface{ Draw(*ebiten.Image) }
	overlaySys   interface{ Draw(*ebiten.Image) }
	exitSub      <-chan events.Event
}

// Update advances ECS world each tick.
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

// Draw renders systems in proper order.
func (g *GameSystem) Draw(screen *ebiten.Image) {
	g.renderSys.Draw(screen)
	g.selectionSys.Draw(screen)
	g.cursorSys.Draw(screen)
	g.overlaySys.Draw(screen)
}

// Layout delegates to render system.
func (g *GameSystem) Layout(outW, outH int) (int, int) {
	if g.renderSys != nil {
		return g.renderSys.Layout(outW, outH)
	}
	return 640, 384
}

// StartGame builds ECS world and launches Ebiten loop.
func StartGame() error {
    _, _, game := initECSWorld() // ✅ ignore world and bus
    ebiten.SetWindowTitle("GoST Terminal")
    ebiten.SetWindowResizable(true)
    ebiten.SetWindowSize(800, 480)
    ebiten.SetMaxTPS(60)
    return ebiten.RunGame(game)
}
