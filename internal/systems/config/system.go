package config

import (
	"log"
	"sync"

	"gost/internal/ecs"
	"gost/internal/events"
)

// System manages persistent configuration and reload/save events.
type System struct {
	bus    *events.Bus
	path   string
	data   *RootConfig
	mu     sync.RWMutex
	reload <-chan events.Event
	save   <-chan events.Event
}

// NewSystem loads the config and begins listening for reload/save events.
func NewSystem(bus *events.Bus, path string) *System {
	s := &System{bus: bus, path: path}

	// Initial load
	cfg, err := loadFromDisk(path)
	if err != nil {
		log.Printf("[Config] Error loading config (%v), using defaults.\n", err)
		cfg = DefaultConfig()
	}
	s.data = cfg

	// Subscriptions
	s.reload = bus.Subscribe("config_reload_requested")
	s.save = bus.Subscribe("config_save_requested")

	go s.listenReload()
	go s.listenSave()

	// Publish initial load event
	bus.Publish("config_loaded", s.Data())

	return s
}

// UpdateECS is a no-op; this system reacts via events.
func (s *System) UpdateECS() {}

// Data returns a safe copy of the current configuration.
func (s *System) Data() *RootConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := *s.data
	return &cp
}

// listenReload handles config_reload_requested.
func (s *System) listenReload() {
	for range s.reload {
		cfg, err := loadFromDisk(s.path)
		if err != nil {
			s.bus.Publish("config_reload_failed", err.Error())
			log.Println("[Config] reload failed:", err)
			continue
		}
		s.mu.Lock()
		s.data = cfg
		s.mu.Unlock()

		s.bus.Publish("config_changed", s.Data())
		log.Println("[Config] reloaded ✓")
	}
}

// listenSave handles config_save_requested.
func (s *System) listenSave() {
	for evt := range s.save {
		var cfg *RootConfig
		if c, ok := evt.(*RootConfig); ok {
			cfg = c
		} else {
			cfg = s.Data()
		}
		if err := saveToDisk(s.path, cfg); err != nil {
			s.bus.Publish("config_save_failed", err.Error())
			log.Println("[Config] save failed:", err)
			continue
		}
		s.bus.Publish("config_saved", s.Data())
		log.Println("[Config] saved ✓")
	}
}

// ECS compliance
var _ ecs.System = (*System)(nil)

