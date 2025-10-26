package ecs

import "sync"

// System represents a modular ECS system with a single UpdateECS tick method.
type System interface {
	UpdateECS()
}

// World is the central ECS container that manages systems and their updates.
type World struct {
	systems []System
	mu      sync.RWMutex
}

// NewWorld creates an empty ECS world instance.
func NewWorld() *World {
	return &World{
		systems: make([]System, 0, 8), // small preallocation
	}
}

// AddSystem registers a new ECS system in the world.
func (w *World) AddSystem(s System) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.systems = append(w.systems, s)
}

// RemoveSystem removes a system from the ECS world (optional utility).
func (w *World) RemoveSystem(target System) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for i, s := range w.systems {
		if s == target {
			w.systems = append(w.systems[:i], w.systems[i+1:]...)
			break
		}
	}
}

// Update ticks all registered systems once per frame (~60Hz typical).
func (w *World) Update() {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for _, s := range w.systems {
		s.UpdateECS()
	}
}

// Count returns the number of active systems in the ECS world.
func (w *World) Count() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.systems)
}

