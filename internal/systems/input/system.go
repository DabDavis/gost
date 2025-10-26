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

// keyState stores state for key repeat tracking.
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

// UpdateECS polls keyboard state and sends input events to PTY.
func (s *System) UpdateECS() {
    now := time.Now()
    handlePrintable(s, now)
    handleSpecial(s, now)
}

