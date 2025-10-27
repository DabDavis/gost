package ecs

import (
	"sort"
	"sync"
)

// World coordinates all ECS systems and manages their update order.
type World struct {
	mu      sync.RWMutex
	systems []systemEntry
	sorted  bool
}

// NewWorld initializes an ECS World with small preallocation.
func NewWorld() *World {
	return &World{
		systems: make([]systemEntry, 0, 8),
	}
}

// AddSystem registers a system with an explicit priority.
// Lower priorities execute earlier per frame.
func (w *World) AddSystem(s System, priority int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.systems = append(w.systems, systemEntry{
		System:   s,
		Priority: priority,
	})
	w.sorted = false
}

// RemoveSystem unregisters a specific system.
func (w *World) RemoveSystem(target System) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i, entry := range w.systems {
		if entry.System == target {
			w.systems = append(w.systems[:i], w.systems[i+1:]...)
			w.sorted = false
			break
		}
	}
}

// Update runs all systems once per frame (~60Hz typical).
// Systems execute in ascending priority order.
func (w *World) Update() {
	w.mu.Lock()
	if !w.sorted {
		sort.SliceStable(w.systems, func(i, j int) bool {
			return w.systems[i].Priority < w.systems[j].Priority
		})
		w.sorted = true
	}
	w.mu.Unlock()

	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, entry := range w.systems {
		entry.System.UpdateECS()
	}
}

// Count returns the number of active ECS systems.
func (w *World) Count() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.systems)
}

