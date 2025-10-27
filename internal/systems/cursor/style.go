package cursor

import (
	"image/color"
	"time"
)

// cursorStyle defines cursor shape and blink behavior.
type cursorStyle struct {
	Color     color.Color
	Shape     string // "block" or "underline"
	Blink     bool
	BlinkRate time.Duration
}

// defaultCursorStyle returns default cursor appearance.
func defaultCursorStyle() cursorStyle {
	return cursorStyle{
		Color:     color.RGBA{200, 200, 200, 200},
		Shape:     "block",
		Blink:     false,
		BlinkRate: 500 * time.Millisecond,
	}
}

