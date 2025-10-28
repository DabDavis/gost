package hotreload

import (
	"log"
	"os"
	"sync"
	"time"

	"gost/internal/ecs"
	"gost/internal/events"
)

// -----------------------------------------------------------------------------
// HotReload System
// -----------------------------------------------------------------------------

// System monitors a file (typically config.json) for modifications and
// triggers an event-based reload when the file changes.
type System struct {
	bus      *events.Bus
	path     string
	lastMod  time.Time
	mu       sync.RWMutex
	interval time.Duration
	enabled  bool
	stopCh   chan struct{}
	started  bool
}

// NewSystem creates and launches a hot-reload watcher.
func NewSystem(bus *events.Bus, path string) *System {
	s := &System{
		bus:      bus,
		path:     path,
		interval: 2 * time.Second,
		enabled:  true,
		stopCh:   make(chan struct{}),
	}
	return s
}

// UpdateECS starts the watcher loop once; ECS calls this once per tick.
func (s *System) UpdateECS() {
	if !s.started {
		s.started = true
		go s.watchLoop()
	}
}

// -----------------------------------------------------------------------------
// Watch Loop
// -----------------------------------------------------------------------------

func (s *System) watchLoop() {
	log.Printf("[HotReload] Watching %s for changes every %v…", s.path, s.interval)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if s.enabled {
				s.checkFile()
			}
		case <-s.stopCh:
			log.Println("[HotReload] Stopped watcher.")
			return
		}
	}
}

// -----------------------------------------------------------------------------
// Core Logic
// -----------------------------------------------------------------------------

func (s *System) checkFile() {
	info, err := os.Stat(s.path)
	if err != nil {
		// Missing file or permission issue — ignore.
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
		log.Printf("[HotReload] Detected change in %s — publishing reload request.", s.path)
		s.bus.Publish("config_reload_requested", s.path)
	}
}

// -----------------------------------------------------------------------------
// Controls
// -----------------------------------------------------------------------------

// Enable dynamically enables/disables watching.
func (s *System) Enable(state bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = state
}

// Stop gracefully terminates the background watcher.
func (s *System) Stop() {
	select {
	case <-s.stopCh:
		// already closed
	default:
		close(s.stopCh)
	}
}

// -----------------------------------------------------------------------------
// ECS Compliance
// -----------------------------------------------------------------------------

var _ ecs.System = (*System)(nil)

