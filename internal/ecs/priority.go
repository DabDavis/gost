package ecs

// Priority constants define default system update ordering for GoST.
// Lower numbers run first each frame; higher numbers render or overlay last.
const (
	// --- Input & Core ---
	PriorityInput      = 10 // keyboard/mouse → emits events
	PriorityPTY        = 20 // terminal IO (read/write)
	PriorityParser     = 30 // parses PTY → updates TermBuffer
	PriorityScrollback = 40 // manages scrollback & history

	// --- Rendering Layers ---
	PriorityRender     = 50 // draws terminal grid
	PriorityCursor     = 60 // draws text cursor
	PrioritySelection  = 70 // draws selection region
	PriorityOverlay    = 90 // transient overlays, HUD messages, etc.

	// --- Reserved range for future use ---
	PriorityUIBase     = 100 // for UI toolkits, windows, etc.
	PriorityDebugTools = 200 // for developer overlays, profiling
)

