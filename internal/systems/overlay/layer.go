package overlay

import "github.com/hajimehoshi/ebiten/v2"

// Drawable defines the minimal contract for any overlay layer.
type Drawable interface {
	Draw(screen *ebiten.Image)
}

// SizedDrawable optionally supports window size awareness.
type SizedDrawable interface {
	DrawWithSize(screen *ebiten.Image, width, height int)
}

// UpdatableDrawable optionally supports per-frame logic (e.g., fading text).
type UpdatableDrawable interface {
	Update()
}

