package selection

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// DrawSelection renders an active selection highlight overlay.
func (s *System) DrawSelection(screen *ebiten.Image) {
	if !s.selecting || s.buffer == nil {
		return
	}

	sx, sy, ex, ey := normalizeRect(s.startX, s.startY, s.endX, s.endY)
	hlColor := color.RGBA{80, 120, 255, 120}

	for y := sy; y <= ey && y < s.buffer.Height; y++ {
		for x := sx; x <= ex && x < s.buffer.Width; x++ {
			x0 := float64(x * s.cellW)
			y0 := float64(y * s.cellH)
			ebitenutil.DrawRect(screen, x0, y0, float64(s.cellW), float64(s.cellH), hlColor)
		}
	}
}

// normalizeRect ensures coordinates are ordered top-left → bottom-right.
func normalizeRect(x1, y1, x2, y2 int) (int, int, int, int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	return x1, y1, x2, y2
}

