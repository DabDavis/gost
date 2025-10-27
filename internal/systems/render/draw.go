package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"

	"gost/internal/components"
)

// Draw renders scrollback + terminal buffer using cached tiles and font.
// It automatically handles 8-, 16-, and 256-color ANSI modes.
func (r *System) Draw(screen *ebiten.Image) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// --- Clear background ---
	screen.Fill(color.Black)
	if r.term == nil {
		return
	}

	// Combine visible history + live buffer
	lines := r.mergeVisibleLines()

	for y := 0; y < len(lines); y++ {
		row := lines[y]
		for x := 0; x < r.term.Width && x < len(row); x++ {
			g := row[x]

			// --- Resolve background color ---
			bgColor := r.resolveColor(g.Bg, false)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*r.cellW), float64(y*r.cellH))

			// cached tile for palette colors, draw solid rect for 256+
			if g.Bg < len(r.bgTiles) && r.bgTiles[g.Bg] != nil {
				screen.DrawImage(r.bgTiles[g.Bg], op)
			} else {
				tile := ebiten.NewImage(r.cellW, r.cellH)
				tile.Fill(bgColor)
				screen.DrawImage(tile, op)
			}

			// --- Foreground glyph ---
			if g.Rune == 0 || g.Rune == ' ' {
				continue
			}
			fgColor := r.resolveColor(g.Fg, true)
			text.Draw(screen, string(g.Rune), r.fontFace,
				x*r.cellW, y*r.cellH+r.cellH-2, fgColor)
		}
	}
}

// mergeVisibleLines builds the combined scrollback + live buffer slice.
func (r *System) mergeVisibleLines() [][]components.Glyph {
	var lines [][]components.Glyph

	// If user scrolled up, prepend scrollback lines
	if r.scrollback != nil && r.scrollOffset > 0 {
		sbLines := r.scrollback.GetVisibleLines(r.scrollOffset, r.term.Height)
		lines = append(lines, sbLines...)
	}

	// Append live terminal contents
	lines = append(lines, r.term.Cells...)

	// Clamp to visible height
	if len(lines) > r.term.Height {
		lines = lines[len(lines)-r.term.Height:]
	}
	return lines
}

// resolveColor picks the appropriate color for a given ANSI index.
// It supports base palettes (0–15) and extended 256-color indexes.
func (r *System) resolveColor(idx int, isForeground bool) color.Color {
	// Fast path: classic 0–15
	if idx >= 0 && idx < 8 {
		if isForeground {
			return r.fgPalette[idx].toColor()
		}
		return r.bgPalette[idx].toColor()
	}
	if idx >= 8 && idx < 16 {
		// bright palette reuse
		c := brightPalette()[idx]
		return c.toColor()
	}

	// Extended 256-color lookup
	c256 := make256Color(idx)
	return c256.toColor()
}

