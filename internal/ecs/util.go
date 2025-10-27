package ecs

import (
	"fmt"
	"reflect"
)

// Describe returns a textual summary of all systems (for debug overlay or logs).
func (w *World) Describe() string {
	systems := w.SortedSystems()
	out := "ECS Systems:\n"
	for _, s := range systems {
		out += fmt.Sprintf("  [%02d] %s\n", s.Priority, s.TypeName)
	}
	return out
}

// HasSystem checks if a system of a given type is already registered.
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

