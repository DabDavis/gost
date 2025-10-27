package overlay

import (
	"fmt"
	"image/color"
	"log"

	"gost/internal/events"
)

// -----------------------------------------------------------------------------
// PTYFeedbackOverlay — reacts to PTY lifecycle events (restart, errors)
// -----------------------------------------------------------------------------

type PTYFeedbackOverlay struct {
	*MessageOverlay
	bus *events.Bus
}

func NewPTYFeedbackOverlay(bus *events.Bus) *PTYFeedbackOverlay {
	o := &PTYFeedbackOverlay{
		MessageOverlay: NewMessageOverlay(),
		bus:            bus,
	}

	topics := map[string]struct {
		msg string
		clr color.Color
	}{
		"pty_restarted":      {"Shell restarted ✓", color.RGBA{0, 255, 0, 255}},
		"pty_restart_failed": {"Shell restart failed ✗", color.RGBA{255, 80, 80, 255}},
	}

	for event, info := range topics {
		sub := bus.Subscribe(event)
		go func(ev string, info struct {
			msg string
			clr color.Color
		}) {
			for evt := range sub {
				txt := info.msg
				if s, ok := evt.(string); ok && s != "" {
					txt += ": " + s
				}
				log.Printf("[Overlay] %s → %s", ev, txt)
				o.Show(fmt.Sprintf("%s", txt), info.clr)
			}
		}(event, info)
	}
	return o
}

