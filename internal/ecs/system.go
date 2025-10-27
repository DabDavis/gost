package ecs

// System defines a modular ECS component that can be updated every frame.
// Systems should be self-contained and thread-safe.
type System interface {
	UpdateECS()
}

// systemEntry is an internal metadata record pairing a system and its priority.
type systemEntry struct {
	System   System
	Priority int
}

