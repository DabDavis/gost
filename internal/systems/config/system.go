package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"gost/internal/ecs"
	"gost/internal/events"
)

// -----------------------------------------------------------------------------
// Root Configuration Schema
// -----------------------------------------------------------------------------

// RootConfig represents the persistent user configuration file.
type RootConfig struct {
	Version     int           `json:"version"`
	ShellPath   string        `json:"shell_path"`
	FontFamily  string        `json:"font_family"`
	FontSize    int           `json:"font_size"`
	Theme       ThemeConfig   `json:"theme"`
	KeyBindings []KeyBinding  `json:"key_bindings,omitempty"`
}

// ThemeConfig defines terminal foreground/background color preferences.
type ThemeConfig struct {
	Foreground string `json:"foreground"`
	Background string `json:"background"`
	Cursor     string `json:"cursor"`
	Selection  string `json:"selection"`
}

// KeyBinding describes a single custom key → action mapping.
type KeyBinding struct {
	Key     string `json:"key"`
	Action  string `json:"action"`
	Control bool   `json:"ctrl,omitempty"`
	Shift   bool   `json:"shift,omitempty"`
	Alt     bool   `json:"alt,omitempty"`
}

// -----------------------------------------------------------------------------
// Default Configuration Factory
// -----------------------------------------------------------------------------

func DefaultConfig() *RootConfig {
	return &RootConfig{
		Version:    1,
		ShellPath:  defaultShell(),
		FontFamily: "monospace",
		FontSize:   14,
		Theme: ThemeConfig{
			Foreground: "#E5E5E5",
			Background: "#000000",
			Cursor:     "#FFFFFF",
			Selection:  "#4444FF",
		},
		KeyBindings: []KeyBinding{
			{Key: "C", Action: "copy", Control: true, Shift: true},
			{Key: "Q", Action: "clear_selection", Control: true, Shift: true},
			{Key: "R", Action: "reload_config", Control: true, Shift: true},
		},
	}
}

// -----------------------------------------------------------------------------
// File I/O (load/save)
// -----------------------------------------------------------------------------

// loadFromDisk reads a JSON config file or returns defaults.
func loadFromDisk(path string) (*RootConfig, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		log.Printf("[Config] %s not found — creating default.\n", path)
		cfg := DefaultConfig()
		if err := saveToDisk(path, cfg); err != nil {
			return cfg, err
		}
		return cfg, nil
	} else if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var cfg RootConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// saveToDisk writes the configuration back to disk in JSON.
func saveToDisk(path string, cfg *RootConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

// -----------------------------------------------------------------------------
// Config System (ECS-Integrated)
// -----------------------------------------------------------------------------

// System manages persistent configuration and reload/save events.
type System struct {
	bus    *events.Bus
	path   string
	data   *RootConfig
	mu     sync.RWMutex
	reload <-chan events.Event
	save   <-chan events.Event
}

// NewSystem loads config and subscribes to reload/save requests.
func NewSystem(bus *events.Bus, path string) *System {
	s := &System{bus: bus, path: path}

	cfg, err := loadFromDisk(path)
	if err != nil {
		log.Printf("[Config] Error loading config (%v), using defaults.\n", err)
		cfg = DefaultConfig()
	}
	s.data = cfg

	s.reload = bus.Subscribe("config_reload_requested")
	s.save = bus.Subscribe("config_save_requested")

	go s.listenReload()
	go s.listenSave()

	bus.Publish("config_loaded", s.Data())
	return s
}

// UpdateECS is a no-op (event-driven).
func (s *System) UpdateECS() {}

// Data returns a thread-safe copy of the configuration.
func (s *System) Data() *RootConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := *s.data
	return &cp
}

// -----------------------------------------------------------------------------
// Event Listeners
// -----------------------------------------------------------------------------

// listenReload handles "config_reload_requested".
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

// listenSave handles "config_save_requested".
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

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func defaultShell() string {
	if sh := os.Getenv("SHELL"); sh != "" {
		return sh
	}
	return "/bin/bash"
}

// ECS compliance
var _ ecs.System = (*System)(nil)

