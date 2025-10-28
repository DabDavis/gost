package input

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"gost/internal/events"
)

// -----------------------------------------------------------------------------
// Unified Input System
// -----------------------------------------------------------------------------

type System struct {
	bus  *events.Bus
	keys map[ebiten.Key]*keyState

	isSelecting bool // mouse drag active
	lastX, lastY int // last cursor position for selection
}

type keyState struct {
	pressed bool
	next    time.Time
}

// --- Key timing constants ---
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

// --- PTY writer callback (set externally) ---
var WriteToPTY = func(b []byte) {}

// -----------------------------------------------------------------------------
// Constructor
// -----------------------------------------------------------------------------

func NewSystem(bus *events.Bus) *System {
	return &System{
		bus:  bus,
		keys: make(map[ebiten.Key]*keyState),
	}
}

// -----------------------------------------------------------------------------
// ECS Loop
// -----------------------------------------------------------------------------

func (s *System) UpdateECS() {
	now := time.Now()

	s.handlePrintable(now)
	s.handleSpecial(now)
	s.handleScrollback(now)
	s.handleMouseScroll(now)
	s.handleGlobalHotkeys(now)
	s.handleSelection()
}

// -----------------------------------------------------------------------------
// Keyboard Input
// -----------------------------------------------------------------------------

func (s *System) handlePrintable(now time.Time) {
	for k := ebiten.KeyA; k <= ebiten.KeyZ; k++ {
		if ebiten.IsKeyPressed(k) {
			s.publishKeyAny()
			alt := ebiten.IsKeyPressed(ebiten.KeyAlt)
			b := byte('a' + (k - ebiten.KeyA))
			if ebiten.IsKeyPressed(ebiten.KeyShift) {
				b -= 32 // uppercase
			}
			WriteToPTY(buildSeq(alt, b))
		}
	}

	for k := ebiten.Key0; k <= ebiten.Key9; k++ {
		if ebiten.IsKeyPressed(k) {
			s.publishKeyAny()
			WriteToPTY(buildSeq(ebiten.IsKeyPressed(ebiten.KeyAlt), byte('0'+(k-ebiten.Key0))))
		}
	}
}

func (s *System) handleSpecial(now time.Time) {
	keySeqs := map[ebiten.Key][]byte{
		ebiten.KeyEnter:      {'\r'},
		ebiten.KeyBackspace:  {0x7f},
		ebiten.KeyTab:        {'\t'},
		ebiten.KeyEscape:     {0x1b},
		ebiten.KeyArrowUp:    []byte{0x1b, '[', 'A'},
		ebiten.KeyArrowDown:  []byte{0x1b, '[', 'B'},
		ebiten.KeyArrowRight: []byte{0x1b, '[', 'C'},
		ebiten.KeyArrowLeft:  []byte{0x1b, '[', 'D'},
	}

	for k, seq := range keySeqs {
		if ebiten.IsKeyPressed(k) {
			s.publishKeyAny()
			WriteToPTY(seq)
		}
	}
}

// -----------------------------------------------------------------------------
// Scrollback and Hotkeys
// -----------------------------------------------------------------------------

func (s *System) handleScrollback(now time.Time) {
	pageUp := ebiten.IsKeyPressed(ebiten.KeyPageUp)
	pageDown := ebiten.IsKeyPressed(ebiten.KeyPageDown)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)

	if shift {
		handleBusKey(s, now, ebiten.KeyPageUp, pageUp, "scroll_page_up")
		handleBusKey(s, now, ebiten.KeyPageDown, pageDown, "scroll_page_down")
	} else {
		handleBusKey(s, now, ebiten.KeyPageUp, pageUp, "scroll_up")
		handleBusKey(s, now, ebiten.KeyPageDown, pageDown, "scroll_down")
	}
}

func (s *System) handleGlobalHotkeys(now time.Time) {
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl)
	alt := ebiten.IsKeyPressed(ebiten.KeyAlt)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)

	if ctrl && ebiten.IsKeyPressed(ebiten.KeyS) && now.Sub(lastSave) > saveCooldown {
		lastSave = now
		s.bus.Publish("config_save_requested", nil)
	}

	if ctrl && ebiten.IsKeyPressed(ebiten.KeyR) {
		s.bus.Publish("config_reload_requested", nil)
	}

	if shift && alt && ebiten.IsKeyPressed(ebiten.KeyC) && now.Sub(lastExit) > exitCooldown {
		lastExit = now
		s.bus.Publish("system_exit", nil)
	}
}

// -----------------------------------------------------------------------------
// Mouse Input + Selection Integration
// -----------------------------------------------------------------------------

func (s *System) handleMouseScroll(now time.Time) {
	_, dy := ebiten.Wheel()
	if dy == 0 {
		return
	}
	if dy > 0 {
		s.bus.Publish("scroll_up", nil)
	} else {
		s.bus.Publish("scroll_down", nil)
	}
	s.publishKeyAny()
}

// handleSelection manages mouse drag selection (start, update, end).
func (s *System) handleSelection() {
	x, y := ebiten.CursorPosition()
	leftPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	if leftPressed && !s.isSelecting {
		s.isSelecting = true
		s.lastX, s.lastY = x, y
		s.bus.Publish("selection_start", map[string]int{"x": x, "y": y})
		return
	}

	if leftPressed && s.isSelecting {
		if x != s.lastX || y != s.lastY {
			s.lastX, s.lastY = x, y
			s.bus.Publish("selection_update", map[string]int{"x": x, "y": y})
		}
		return
	}

	if !leftPressed && s.isSelecting {
		s.isSelecting = false
		s.bus.Publish("selection_end", map[string]int{"x": x, "y": y})
	}
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

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

func buildSeq(alt bool, b byte) []byte {
	if alt {
		return append([]byte{0x1b}, b)
	}
	return []byte{b}
}

func (s *System) publishKeyAny() {
	s.bus.Publish("key_any_pressed", nil)
}

