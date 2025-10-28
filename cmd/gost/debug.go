package main

import (
	"fmt"
	"sort"
	"reflect"
	"strings"

	"gost/internal/ecs"
)

// debugPrintSystems outputs a formatted ECS system list.
func debugPrintSystems(world *ecs.World) {
	count := world.Count()
	fmt.Println()
	fmt.Println("───────────────────────────────────────────────")
	fmt.Printf(" GoST ECS System Registry (%d systems)\n", count)
	fmt.Println("───────────────────────────────────────────────")

	systems := getSystemSummary(world)
	sort.SliceStable(systems, func(i, j int) bool {
		return systems[i].Priority < systems[j].Priority
	})

	for _, s := range systems {
		fmt.Printf(" [%02d]  %-16s  %s\n", s.Priority, s.Name, s.Type)
	}

	fmt.Println("───────────────────────────────────────────────")
	fmt.Println()
}

// getSystemSummary safely extracts names & priorities.
func getSystemSummary(world *ecs.World) []sysInfo {
	// we’ll reflect on world.systems; since it's private, we rely on ECS accessor helpers
	// If world had a Systems() accessor, use that; otherwise we replicate structure.

	type systemEntry struct {
		System   ecs.System
		Priority int
	}
	entries := []systemEntry{}

	// use reflection to reach internal systems slice
	worldVal := reflectValue(world)
	field := worldVal.FieldByName("systems")
	if !field.IsValid() {
		return nil
	}
	for i := 0; i < field.Len(); i++ {
		entry := field.Index(i)
		sys := entry.FieldByName("System").Interface().(ecs.System)
		prio := int(entry.FieldByName("Priority").Int())
		entries = append(entries, systemEntry{sys, prio})
	}

	result := []sysInfo{}
	for _, e := range entries {
		result = append(result, sysInfo{
			Name:     typeShortName(e.System),
			Type:     fmt.Sprintf("%T", e.System),
			Priority: e.Priority,
		})
	}
	return result
}

type sysInfo struct {
	Name     string
	Type     string
	Priority int
}

// typeShortName trims package path to last element.
func typeShortName(s interface{}) string {
	full := fmt.Sprintf("%T", s)
	if idx := strings.LastIndex(full, "."); idx != -1 {
		return full[idx+1:]
	}
	return full
}

// reflectValue returns reflect.Value safely (to bypass private fields)
func reflectValue(i interface{}) reflect.Value {
	val := reflect.ValueOf(i)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val
}

