package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"gost/internal/components"
)

// Draw renders scrollback + terminal buffer using cached tiles and font.
func (r *System) Draw(screen *ebiten.Image) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	screen.Fill(color.Black)
	if r.term == nil {
		return
	}

	lines := r.mergeVisibleLines()

	for y := 0; y < len(lines); y++ {
		for x := 0; x < r.term.Width; x++ {
			if x >= len(lines[y]) {
				continue
			}
			g := lines[y][x]
			bgIndex := g.Bg % 8
			fgIndex := g.Fg % 8

			// Background tile
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*r.cellW), float64(y*r.cellH))
			screen.DrawImage(r.bgTiles[bgIndex], op)

			// Foreground glyph
			if g.Rune == 0 || g.Rune == ' ' {
				continue
			}
			fg := r.fgPalette[fgIndex].toColor()
			text.Draw(screen, string(g.Rune), r.fontFace,
				x*r.cellW, y*r.cellH+r.cellH-2, fg)
		}
	}
}

// mergeVisibleLines builds combined scrollback + live view.
func (r *System) mergeVisibleLines() [][]components.Glyph {
	lines := [][]components.Glyph{}
	if r.scrollback != nil && r.scrollOffset > 0 {
		sbLines := r.scrollback.GetVisibleLines(r.scrollOffset, r.term.Height)
		lines = append(lines, sbLines...)
	}
	lines = append(lines, r.term.Cells...)
	if len(lines) > r.term.Height {
		lines = lines[len(lines)-r.term.Height:]
	}
	return lines
}

