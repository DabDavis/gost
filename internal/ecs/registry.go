package ecs

import "reflect"
import "sort"

// SystemInfo describes one registered system for introspection.
type SystemInfo struct {
	TypeName string
	Priority int
}

// Systems returns a snapshot of registered systems and priorities.
func (w *World) Systems() []SystemInfo {
	w.mu.RLock()
	defer w.mu.RUnlock()

	out := make([]SystemInfo, len(w.systems))
	for i, s := range w.systems {
		t := reflect.TypeOf(s.System)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		out[i] = SystemInfo{
			TypeName: t.Name(),
			Priority: s.Priority,
		}
	}
	return out
}

// SortedSystems returns systems ordered by execution priority.
func (w *World) SortedSystems() []SystemInfo {
	systems := w.Systems()
	sort.SliceStable(systems, func(i, j int) bool {
		return systems[i].Priority < systems[j].Priority
	})
	return systems
}

