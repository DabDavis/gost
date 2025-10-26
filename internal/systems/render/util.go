package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type colorRGBA struct{ r, g, b, a uint8 }

func (c colorRGBA) toColor() color.Color { return color.RGBA{c.r, c.g, c.b, c.a} }

func defaultFgPalette() [8]colorRGBA {
	return [8]colorRGBA{
		{0, 0, 0, 255}, {205, 49, 49, 255}, {13, 188, 121, 255},
		{229, 229, 16, 255}, {36, 114, 200, 255}, {188, 63, 188, 255},
		{17, 168, 205, 255}, {229, 229, 229, 255},
	}
}

func defaultBgPalette() [8]colorRGBA {
	return [8]colorRGBA{
		{0, 0, 0, 255}, {60, 0, 0, 255}, {0, 60, 0, 255}, {60, 60, 0, 255},
		{0, 0, 60, 255}, {60, 0, 60, 255}, {0, 60, 60, 255}, {80, 80, 80, 255},
	}
}

func (r *System) precacheTiles() {
	for i := range r.bgPalette {
		tile := ebiten.NewImage(r.cellW, r.cellH)
		tile.Fill(r.bgPalette[i].toColor())
		r.bgTiles[i] = tile
	}
}

