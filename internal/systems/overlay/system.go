package overlay

import (
	"log"
	"time"

	"image/color"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"

	"gost/internal/events"
)

// System renders transient messages like "Copied ✓" or system notices.
type System struct {
	bus   *events.Bus
	msg   string
	alpha float64
	timer time.Time
}

// NewSystem subscribes to clipboard events and initializes HUD state.
func NewSystem(bus *events.Bus) *System {
	os := &System{bus: bus}

	sub := bus.Subscribe("clipboard_copy")
	go func() {
		for evt := range sub {
			txt, ok := evt.(string)
			if !ok {
				continue
			}
			log.Printf("[Overlay] Copied %d chars", len(txt))
			os.showMessage("Copied ✓")
		}
	}()

	return os
}

// showMessage displays a HUD message for a short duration.
func (s *System) showMessage(text string) {
	s.msg = text
	s.alpha = 1.0
	s.timer = time.Now()
}

// UpdateECS handles fade-out logic over time.
func (s *System) UpdateECS() {
	if s.msg == "" {
		return
	}
	elapsed := time.Since(s.timer).Seconds()
	if elapsed > 2.0 {
		s.alpha -= 0.05
		if s.alpha <= 0 {
			s.msg = ""
			s.alpha = 0
		}
	}
}

// DrawOverlay renders the current HUD message.
func (s *System) DrawOverlay(screen *ebiten.Image, width, height int) {
	if s.msg == "" {
		return
	}

	clr := color.RGBA{255, 255, 255, uint8(s.alpha * 255)}
	textW := len(s.msg) * 8
	x := (width - textW) / 2
	y := height - 20 // bottom area of the screen

	text.Draw(screen, s.msg, basicfont.Face7x13, x, y, clr)
}

