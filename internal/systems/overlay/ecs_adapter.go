package overlay

import "gost/internal/ecs"

// ECSOverlayAdapter allows a Drawable to also behave like an ECS system.
type ECSOverlayAdapter struct {
	layer Drawable
}

// NewECSOverlayAdapter wraps a Drawable so it can register in ECS world.
func NewECSOverlayAdapter(layer Drawable) ecs.System {
	return &ECSOverlayAdapter{layer: layer}
}

// UpdateECS delegates to Update() if available.
func (a *ECSOverlayAdapter) UpdateECS() {
	if up, ok := a.layer.(UpdatableDrawable); ok {
		up.Update()
	}
}

