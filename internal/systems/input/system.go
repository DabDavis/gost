package input

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"gost/internal/events"
)

// System is the ECS-compatible keyboard input system.
type System struct {
	bus  *events.Bus
	keys map[ebiten.Key]*keyState
}

type keyState struct {
	pressed bool
	next    time.Time
}

const (
	repeatDelay = 400 * time.Millisecond
	repeatRate  = 35 * time.Millisecond
)

// NewSystem creates a new input handler.
func NewSystem(bus *events.Bus) *System {
	return &System{
		bus:  bus,
		keys: make(map[ebiten.Key]*keyState),
	}
}

// UpdateECS polls keyboard state and sends input events to PTY or event bus.
func (s *System) UpdateECS() {
	now := time.Now()

	// 1️⃣ Regular key input
	handlePrintable(s, now)
	handleSpecial(s, now)

	// 2️⃣ Scrollback keys
	s.handleScrollback(now)
	
	// 3️⃣ Scrollback via mouse wheel 🖱️
	s.handleMouseScroll(now)
}

// handleScrollback publishes PageUp/PageDown events for scrollback.
func (s *System) handleScrollback(now time.Time) {
	pageUp := ebiten.IsKeyPressed(ebiten.KeyPageUp)
	pageDown := ebiten.IsKeyPressed(ebiten.KeyPageDown)

	handleBusKey(s, now, ebiten.KeyPageUp, pageUp, "scroll_up")
	handleBusKey(s, now, ebiten.KeyPageDown, pageDown, "scroll_down")
}

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
// It’s called from handlePrintable() and handleSpecial() when a key is first pressed.
func (s *System) publishKeyAny() {
	s.bus.Publish("key_any_pressed", nil)
}

