package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// colorRGBA is a lightweight struct for compact palette storage.
type colorRGBA struct {
	r, g, b, a uint8
}

func (c colorRGBA) toColor() color.Color {
	return color.RGBA{c.r, c.g, c.b, c.a}
}

// ----------------------------------------------------------------------------
// Palette definitions
// ----------------------------------------------------------------------------

// defaultFgPalette returns the standard 8-color ANSI foreground palette.
// These correspond to the base colors 0–7.
func defaultFgPalette() [8]colorRGBA {
	return [8]colorRGBA{
		{0, 0, 0, 255},       // black
		{205, 49, 49, 255},   // red
		{13, 188, 121, 255},  // green
		{229, 229, 16, 255},  // yellow
		{36, 114, 200, 255},  // blue
		{188, 63, 188, 255},  // magenta
		{17, 168, 205, 255},  // cyan
		{229, 229, 229, 255}, // white (bright gray)
	}
}

// defaultBgPalette returns the matching 8-color ANSI background palette.
func defaultBgPalette() [8]colorRGBA {
	return [8]colorRGBA{
		{0, 0, 0, 255},    // black
		{60, 0, 0, 255},   // dark red
		{0, 60, 0, 255},   // dark green
		{60, 60, 0, 255},  // dark yellow
		{0, 0, 60, 255},   // dark blue
		{60, 0, 60, 255},  // dark magenta
		{0, 60, 60, 255},  // dark cyan
		{80, 80, 80, 255}, // dark gray
	}
}

// brightPalette extends the base 8 colors with their brighter 8 equivalents.
// This can be swapped in dynamically when SGR bright codes (90–97, 100–107) are used.
func brightPalette() [16]colorRGBA {
	return [16]colorRGBA{
		{0, 0, 0, 255},       // 0 black
		{205, 49, 49, 255},   // 1 red
		{13, 188, 121, 255},  // 2 green
		{229, 229, 16, 255},  // 3 yellow
		{36, 114, 200, 255},  // 4 blue
		{188, 63, 188, 255},  // 5 magenta
		{17, 168, 205, 255},  // 6 cyan
		{229, 229, 229, 255}, // 7 white

		// Bright (8–15)
		{102, 102, 102, 255}, // bright black (gray)
		{241, 76, 76, 255},   // bright red
		{35, 209, 139, 255},  // bright green
		{245, 245, 67, 255},  // bright yellow
		{59, 142, 234, 255},  // bright blue
		{214, 112, 214, 255}, // bright magenta
		{41, 184, 219, 255},  // bright cyan
		{255, 255, 255, 255}, // bright white
	}
}

// ----------------------------------------------------------------------------
// Tile precomputation
// ----------------------------------------------------------------------------

// precacheTiles generates solid background tiles for all palette colors.
// These are drawn behind glyphs for each cell.
func (r *System) precacheTiles() {
	for i := range r.bgPalette {
		tile := ebiten.NewImage(r.cellW, r.cellH)
		tile.Fill(r.bgPalette[i].toColor())
		r.bgTiles[i] = tile
	}
}

// ----------------------------------------------------------------------------
// Future expansion hooks
// ----------------------------------------------------------------------------

// make256Color converts a 256-color ANSI index to a colorRGBA.
// This enables future support for xterm 256-color mode (SGR 38;5;n / 48;5;n).
func make256Color(index int) colorRGBA {
	switch {
	case index < 16:
		// First 16 entries are from brightPalette
		p := brightPalette()
		return p[index%16]
	case index >= 16 && index < 232:
		// 6×6×6 color cube
		i := index - 16
		r := uint8((i / 36) * 51)
		g := uint8(((i / 6) % 6) * 51)
		b := uint8((i % 6) * 51)
		return colorRGBA{r, g, b, 255}
	default:
		// Grayscale ramp (232–255)
		level := uint8(8 + (index-232)*10)
		return colorRGBA{level, level, level, 255}
	}
}

