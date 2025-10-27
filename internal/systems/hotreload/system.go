package hotreload

import (
	"log"
	"os"
	"sync"
	"time"

	"gost/internal/events"
	"gost/internal/ecs"
)

// System monitors a configuration file for changes and emits reload events.
type System struct {
	bus      *events.Bus
	path     string
	lastMod  time.Time
	mu       sync.RWMutex
	interval time.Duration
	enabled  bool
	stopCh   chan struct{}
}

// NewSystem creates a hotreload watcher that periodically checks for file changes.
func NewSystem(bus *events.Bus, path string) *System {
	s := &System{
		bus:      bus,
		path:     path,
		interval: 2 * time.Second, // reasonable default check interval
		enabled:  true,
		stopCh:   make(chan struct{}),
	}
	go s.watchLoop()
	return s
}

// UpdateECS is a no-op (the watcher runs asynchronously).
func (s *System) UpdateECS() {}

// watchLoop periodically polls the target file for modification time changes.
func (s *System) watchLoop() {
	log.Printf("[HotReload] Watching %s for changes...\n", s.path)

	for {
		select {
		case <-time.After(s.interval):
			if !s.enabled {
				continue
			}
			s.checkFile()
		case <-s.stopCh:
			log.Println("[HotReload] Stopped watcher.")
			return
		}
	}
}

// checkFile compares modification time and triggers reload if changed.
func (s *System) checkFile() {
	info, err := os.Stat(s.path)
	if err != nil {
		// If file disappears temporarily, ignore.
		return
	}

	modTime := info.ModTime()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lastMod.IsZero() {
		s.lastMod = modTime
		return
	}

	if modTime.After(s.lastMod) {
		s.lastMod = modTime
		log.Printf("[HotReload] Detected change in %s — requesting reload...\n", s.path)
		s.bus.Publish("config_reload_requested", s.path)
	}
}

// Enable allows dynamic enabling/disabling of the watcher.
func (s *System) Enable(state bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = state
}

// Stop cleanly terminates the background watcher goroutine.
func (s *System) Stop() {
	close(s.stopCh)
}

// Interface check for ECS compliance.
var _ ecs.System = (*System)(nil)

