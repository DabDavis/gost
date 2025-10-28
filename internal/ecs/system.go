package ecs

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
)

// -----------------------------------------------------------------------------
// System Interface
// -----------------------------------------------------------------------------

// System defines the minimal interface all ECS systems must implement.
type System interface {
	UpdateECS()
}

// -----------------------------------------------------------------------------
// ECS Priority Constants
// -----------------------------------------------------------------------------

const (
	PriorityConfig     = 10
	PriorityHotReload  = 20
	PriorityInput      = 30
	PriorityPTY        = 40
	PriorityParser     = 50
	PriorityScrollback = 60
	PriorityRender     = 70
	PrioritySelection  = 80
	PriorityCursor     = 90
	PriorityOverlay    = 100
)

// -----------------------------------------------------------------------------
// Internal Types
// -----------------------------------------------------------------------------

type systemEntry struct {
	System   System
	Priority int
	TypeName string
}

// -----------------------------------------------------------------------------
// ECS World — Registry & Execution Manager
// -----------------------------------------------------------------------------

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

	typeName := reflect.TypeOf(s).String()
	w.systems = append(w.systems, systemEntry{
		System:   s,
		Priority: priority,
		TypeName: typeName,
	})
	w.sorted = false
}

// RemoveSystem unregisters a specific system instance.
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

// -----------------------------------------------------------------------------
// Introspection & Utilities
// -----------------------------------------------------------------------------

// Count returns the number of active ECS systems.
func (w *World) Count() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.systems)
}

// HasSystem checks if a system of the same type is already registered.
func (w *World) HasSystem(target System) bool {
	t1 := reflect.TypeOf(target)
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, e := range w.systems {
		if reflect.TypeOf(e.System) == t1 {
			return true
		}
	}
	return false
}

// SortedSystems returns the internal list sorted by priority.
func (w *World) SortedSystems() []systemEntry {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.sorted {
		sort.SliceStable(w.systems, func(i, j int) bool {
			return w.systems[i].Priority < w.systems[j].Priority
		})
		w.sorted = true
	}
	return append([]systemEntry(nil), w.systems...)
}

// Describe returns a formatted overview of all registered systems.
func (w *World) Describe() string {
	systems := w.SortedSystems()
	out := "ECS Systems:\n"
	for _, s := range systems {
		out += fmt.Sprintf("  [%02d] %s\n", s.Priority, s.TypeName)
	}
	return out
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// MustAdd ensures a system is only added once; logs if duplicated.
func (w *World) MustAdd(s System, priority int) {
	if w.HasSystem(s) {
		fmt.Printf("[ECS] Skipped duplicate: %T\n", s)
		return
	}
	w.AddSystem(s, priority)
}

// Reset clears all registered systems.
func (w *World) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.systems = w.systems[:0]
	w.sorted = false
}

