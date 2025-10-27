package cursor

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// drawBlock renders a filled rectangle cursor.
func (c *System) drawBlock(screen *ebiten.Image, cx, cy int) {
	ebitenutil.DrawRect(
		screen,
		float64(cx*c.cellW),
		float64(cy*c.cellH),
		float64(c.cellW),
		float64(c.cellH),
		c.style.Color,
	)
}

// drawUnderline renders a small underline cursor.
func (c *System) drawUnderline(screen *ebiten.Image, cx, cy int) {
	h := float64(c.cellH) / 6
	y := float64((cy+1)*c.cellH) - h
	ebitenutil.DrawRect(
		screen,
		float64(cx*c.cellW),
		y,
		float64(c.cellW),
		h,
		c.style.Color,
	)
}

