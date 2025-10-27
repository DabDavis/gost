package overlay

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// -----------------------------------------------------------------------------
// MessageOverlay — base type for displaying transient HUD messages
// -----------------------------------------------------------------------------

type MessageOverlay struct {
	msg   string
	clr   color.Color
	alpha float64
	timer time.Time
}

// NewMessageOverlay returns a ready overlay.
func NewMessageOverlay() *MessageOverlay {
	return &MessageOverlay{}
}

// Show sets message text and color and restarts fade timer.
func (o *MessageOverlay) Show(text string, clr color.Color) {
	o.msg = text
	o.clr = clr
	o.alpha = 1.0
	o.timer = time.Now()
}

// UpdateECS handles fading logic (2s delay + fade out).
func (o *MessageOverlay) UpdateECS() {
	if o.msg == "" {
		return
	}
	elapsed := time.Since(o.timer).Seconds()
	if elapsed > 2.0 {
		o.alpha -= 0.05
		if o.alpha <= 0 {
			o.msg = ""
			o.alpha = 0
		}
	}
}

// Draw renders centered text near the bottom.
func (o *MessageOverlay) Draw(screen *ebiten.Image, width, height int) {
	if o.msg == "" {
		return
	}
	r, g, b, _ := o.clr.RGBA()
	clr := color.RGBA{
		uint8(r >> 8),
		uint8(g >> 8),
		uint8(b >> 8),
		uint8(o.alpha * 255),
	}
	textW := len(o.msg) * 8
	x := (width - textW) / 2
	y := height - 40
	text.Draw(screen, o.msg, basicfont.Face7x13, x, y, clr)
}

