package overlay

import (
	"fmt"
	"image/color"
	"log"

	"gost/internal/events"
)

// -----------------------------------------------------------------------------
// ConfigFeedbackOverlay — listens for config save/reload events.
// -----------------------------------------------------------------------------

type ConfigFeedbackOverlay struct {
	*MessageOverlay
	bus *events.Bus
}

// NewConfigFeedbackOverlay attaches to the ECS bus.
func NewConfigFeedbackOverlay(bus *events.Bus) *ConfigFeedbackOverlay {
	o := &ConfigFeedbackOverlay{
		MessageOverlay: NewMessageOverlay(),
		bus:            bus,
	}

	eventsMap := map[string]struct {
		msg string
		clr color.Color
	}{
		"config_changed":       {"Configuration Reloaded ✓", color.RGBA{0, 255, 0, 255}},
		"config_reload_failed": {"Config Reload Failed ✗", color.RGBA{255, 80, 80, 255}},
		"config_saved":         {"Settings Saved ✓", color.RGBA{0, 200, 255, 255}},
		"config_save_failed":   {"Settings Save Failed ✗", color.RGBA{255, 120, 0, 255}},
	}

	for topic, info := range eventsMap {
		sub := bus.Subscribe(topic)
		go func(event string, info struct {
			msg string
			clr color.Color
		}) {
			for evt := range sub {
				txt := info.msg
				if s, ok := evt.(string); ok && s != "" {
					txt += ": " + s
				}
				// Append cursor info on config reload
				if event == "config_changed" {
					if cfg, ok := evt.(interface {
						GetCursorShape() string
						GetCursorBlink() bool
						GetCursorBlinkRate() int
					}); ok {
						style := cfg.GetCursorShape()
						blink := "off"
						if cfg.GetCursorBlink() {
							blink = fmt.Sprintf("on (%dms)", cfg.GetCursorBlinkRate())
						}
						txt = fmt.Sprintf("Config Reloaded ✓ — cursor: %s, blink %s", style, blink)
					}
				}
				log.Printf("[Overlay] %s → %s", event, txt)
				o.Show(txt, info.clr)
			}
		}(topic, info)
	}

	return o
}

