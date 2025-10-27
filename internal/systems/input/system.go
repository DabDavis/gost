package input

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"gost/internal/events"
)

// -----------------------------------------------------------------------------
// Input System — ECS-compatible keyboard & mouse handler
// -----------------------------------------------------------------------------

type System struct {
	bus  *events.Bus
	keys map[ebiten.Key]*keyState
}

type keyState struct {
	pressed bool
	next    time.Time
}

const (
	repeatDelay  = 400 * time.Millisecond
	repeatRate   = 35 * time.Millisecond
	saveCooldown = 1 * time.Second
	exitCooldown = 2 * time.Second
)

var (
	lastSave time.Time
	lastExit time.Time
)

// NewSystem creates a new input system.
func NewSystem(bus *events.Bus) *System {
	return &System{
		bus:  bus,
		keys: make(map[ebiten.Key]*keyState),
	}
}

// -----------------------------------------------------------------------------
// ECS Update Loop
// -----------------------------------------------------------------------------

// UpdateECS polls input each frame and emits appropriate ECS events.
func (s *System) UpdateECS() {
	now := time.Now()

	// 1. Regular key input → PTY
	handlePrintable(s, now)
	handleSpecial(s, now)

	// 2. Scrollback input (line + page scroll)
	s.handleScrollback(now)

	// 3. Mouse scroll events
	s.handleMouseScroll(now)

	// 4. Global hotkeys (save/reload/exit)
	s.handleGlobalHotkeys(now)
}

// -----------------------------------------------------------------------------
// Scrollback & Hotkeys
// -----------------------------------------------------------------------------

// handleScrollback publishes line or page scroll events.
func (s *System) handleScrollback(now time.Time) {
	pageUp := ebiten.IsKeyPressed(ebiten.KeyPageUp)
	pageDown := ebiten.IsKeyPressed(ebiten.KeyPageDown)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)

	if shift {
		// Shift + PageUp/PageDown → full-page scroll
		handleBusKey(s, now, ebiten.KeyPageUp, pageUp, "scroll_page_up")
		handleBusKey(s, now, ebiten.KeyPageDown, pageDown, "scroll_page_down")
	} else {
		// Regular PageUp/PageDown → line scroll
		handleBusKey(s, now, ebiten.KeyPageUp, pageUp, "scroll_up")
		handleBusKey(s, now, ebiten.KeyPageDown, pageDown, "scroll_down")
	}
}

// handleGlobalHotkeys detects Ctrl+S (save), Ctrl+R (reload), and Shift+Alt+C (exit).
func (s *System) handleGlobalHotkeys(now time.Time) {
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl)
	alt := ebiten.IsKeyPressed(ebiten.KeyAlt)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)

	// Ctrl + S → Save configuration
	if ctrl && ebiten.IsKeyPressed(ebiten.KeyS) {
		if now.Sub(lastSave) > saveCooldown {
			lastSave = now
			s.bus.Publish("config_save_requested", nil)
		}
	}

	// Ctrl + R → Reload configuration
	if ctrl && ebiten.IsKeyPressed(ebiten.KeyR) {
		s.bus.Publish("config_reload_requested", nil)
	}

	// Shift + Alt + C → graceful exit
	if shift && alt && ebiten.IsKeyPressed(ebiten.KeyC) {
		if now.Sub(lastExit) > exitCooldown {
			lastExit = now
			s.bus.Publish("system_exit", nil)
		}
	}
}

// -----------------------------------------------------------------------------
// Debounce Helpers
// -----------------------------------------------------------------------------

// handleBusKey manages debounce for bus-published control keys.
func handleBusKey(s *System, now time.Time, key ebiten.Key, pressed bool, topic string) {
	ks, ok := s.keys[key]
	if !ok {
		ks = &keyState{}
		s.keys[key] = ks
	}

	if pressed {
		if !ks.pressed {
			ks.pressed = true
			ks.next = now.Add(repeatDelay)
			s.bus.Publish(topic, nil)
			return
		}
		if now.After(ks.next) {
			ks.next = now.Add(repeatRate)
			s.bus.Publish(topic, nil)
		}
	} else {
		ks.pressed = false
	}
}

// publishKeyAny notifies renderer to reset scrollback when *any* key is pressed.
func (s *System) publishKeyAny() {
	s.bus.Publish("key_any_pressed", nil)
}

