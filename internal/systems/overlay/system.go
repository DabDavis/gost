package overlay

import (
    "image/color"
    "sync"
    "time"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/text"
    "golang.org/x/image/font"
    "golang.org/x/image/font/basicfont"

    "gost/internal/events"
    "gost/internal/ecs"
)

// -----------------------------------------------------------------------------
// Base Overlay Interfaces
// -----------------------------------------------------------------------------

type Drawable interface {
    Draw(*ebiten.Image)
}

type Message struct {
    Text      string
    Color     color.Color
    CreatedAt time.Time
    Duration  time.Duration
}

func (m *Message) Expired() bool {
    return time.Since(m.CreatedAt) > m.Duration
}

// -----------------------------------------------------------------------------
// Overlay System — manages layered drawing and message fadeouts
// -----------------------------------------------------------------------------

type System struct {
    mu     sync.RWMutex
    layers []Drawable
    msgs   []*Message
    font   font.Face
    bus    *events.Bus
}

// NewSystem creates an empty overlay compositor.
func NewSystem() *System {
    return &System{
        layers: make([]Drawable, 0, 8),
        msgs:   make([]*Message, 0, 8),
        font:   basicfont.Face7x13,
    }
}

// -----------------------------------------------------------------------------
// Layer Management
// -----------------------------------------------------------------------------

func (o *System) AddLayer(d Drawable) {
    o.mu.Lock()
    defer o.mu.Unlock()
    o.layers = append(o.layers, d)
}

func (o *System) RemoveLayer(target Drawable) {
    o.mu.Lock()
    defer o.mu.Unlock()
    for i, l := range o.layers {
        if l == target {
            o.layers = append(o.layers[:i], o.layers[i+1:]...)
            break
        }
    }
}

// -----------------------------------------------------------------------------
// Message Handling
// -----------------------------------------------------------------------------

func (o *System) Post(text string, clr color.Color, dur time.Duration) {
    o.mu.Lock()
    defer o.mu.Unlock()
    o.msgs = append(o.msgs, &Message{
        Text:      text,
        Color:     clr,
        CreatedAt: time.Now(),
        Duration:  dur,
    })
}

func (o *System) purgeExpired() {
    o.mu.Lock()
    defer o.mu.Unlock()
    now := time.Now()
    filtered := o.msgs[:0]
    for _, m := range o.msgs {
        if now.Sub(m.CreatedAt) < m.Duration {
            filtered = append(filtered, m)
        }
    }
    o.msgs = filtered
}

// -----------------------------------------------------------------------------
// ECS Integration
// -----------------------------------------------------------------------------

func (o *System) UpdateECS() {
    o.mu.RLock()
    for _, l := range o.layers {
        if sys, ok := l.(ecs.System); ok {
            sys.UpdateECS()
        }
    }
    o.mu.RUnlock()
    o.purgeExpired()
}

func (o *System) Draw(screen *ebiten.Image) {
    o.mu.RLock()
    defer o.mu.RUnlock()

    // Draw all layers in z-order
    for _, layer := range o.layers {
        layer.Draw(screen)
    }

    // Draw active messages (top-left corner)
    y := 16
    for _, m := range o.msgs {
        text.Draw(screen, m.Text, o.font, 10, y, m.Color)
        y += 14
    }
}

// Compile-time ECS interface check
var _ ecs.System = (*System)(nil)

