package overlay

import (
	"log"
	"time"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"

	"gost/internal/events"
)

// Message represents a transient on-screen notice.
type Message struct {
	Text   string
	Alpha  float64
	Timer  time.Time
	Expire float64 // lifetime in seconds
}

// System renders transient messages like “Copied ✓” or warnings.
type System struct {
	bus      *events.Bus
	messages []Message
}

// NewSystem subscribes to events and initializes the overlay HUD.
func NewSystem(bus *events.Bus) *System {
	os := &System{bus: bus}

	// Clipboard copy message
	subClip := bus.Subscribe("clipboard_copy")
	go func() {
		for evt := range subClip {
			txt, ok := evt.(string)
			if !ok {
				continue
			}
			log.Printf("[Overlay] Copied %d chars", len(txt))
			os.AddMessage("Copied ✓")
		}
	}()

	// Generic system message event
	subMsg := bus.Subscribe("overlay_message")
	go func() {
		for evt := range subMsg {
			if msg, ok := evt.(string); ok {
				os.AddMessage(msg)
			}
		}
	}()

	return os
}

// AddMessage queues a new transient HUD message.
func (s *System) AddMessage(text string) {
	s.messages = append(s.messages, Message{
		Text:   text,
		Alpha:  1.0,
		Timer:  time.Now(),
		Expire: 3.0,
	})
}

// UpdateECS handles fade-out logic for all active messages.
func (s *System) UpdateECS() {
	now := time.Now()
	newList := s.messages[:0]

	for _, m := range s.messages {
		elapsed := now.Sub(m.Timer).Seconds()
		if elapsed < m.Expire {
			// Start fading out after 2s
			if elapsed > m.Expire-1.0 {
				m.Alpha = 1.0 - (elapsed-(m.Expire-1.0))
			}
			newList = append(newList, m)
		}
	}
	s.messages = newList
}

// DrawOverlay renders all active messages near the bottom of the screen.
func (s *System) DrawOverlay(screen *ebiten.Image, width, height int) {
	if len(s.messages) == 0 {
		return
	}

	y := height - 20
	for i := len(s.messages) - 1; i >= 0; i-- {
		m := s.messages[i]
		clr := color.RGBA{255, 255, 255, uint8(m.Alpha * 255)}
		textW := len(m.Text) * 8
		x := (width - textW) / 2
		text.Draw(screen, m.Text, basicfont.Face7x13, x, y, clr)
		y -= 16 // stack upwards
	}
}

