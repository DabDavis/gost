package ecs

// Priority constants define execution order for ECS systems.
// Lower numbers run earlier per frame; higher numbers render later.
const (
	// --- Configuration & Hot Reload ---
	PriorityConfig     = 5
	PriorityHotReload  = 6

	// --- Core Input & I/O ---
	PriorityInput      = 10
	PriorityPTY        = 20
	PriorityParser     = 30
	PriorityScrollback = 40

	// --- Rendering ---
	PriorityRender     = 50
	PrioritySelection  = 60
	PriorityCursor     = 70
	PriorityOverlay    = 80

	// --- Optional Extensions ---
	PriorityUIBase     = 100
	PriorityDebugTools = 200
)

