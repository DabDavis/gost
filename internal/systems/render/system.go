package render

import (
	"image/color"
	"log"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"

	"gost/internal/components"
	"gost/internal/ecs"
	"gost/internal/events"
	"gost/internal/systems/config"
)

// System draws the composed terminal view using cached backgrounds and font.
// It listens to configuration and scrollback events dynamically.
type System struct {
	mu sync.RWMutex

	bus        *events.Bus
	term       *components.TermBuffer
	scrollback *components.Scrollback
	viewport   *Viewport

	fontFace font.Face
	cellW, cellH int

	scrollOffset int // scrollback offset (updated by event listener)
	fgPalette    [8]colorRGBA
	bgPalette    [8]colorRGBA
	bgTiles      [8]*ebiten.Image
	bgColor      color.Color

	configSub <-chan events.Event // config change listener
	offsetSub <-chan events.Event // scroll offset listener
}

// NewSystem initializes the renderer and subscribes to bus events.
func NewSystem(bus *events.Bus) *System {
	r := &System{
		bus:       bus,
		fontFace:  basicfont.Face7x13,
		cellW:     7,
		cellH:     14,
		fgPalette: defaultFgPalette(),
		bgPalette: defaultBgPalette(),
		bgColor:   color.Black,
	}

	r.precacheTiles()
	r.subscribe()
	return r
}

// subscribe connects to bus events for config reloads and scroll offset.
func (r *System) subscribe() {
	if r.bus == nil {
		return
	}

	// --- Scroll offset updates ---
	r.offsetSub = r.bus.Subscribe("scroll_offset_changed")
	go func() {
		for evt := range r.offsetSub {
			if offset, ok := evt.(int); ok {
				r.mu.Lock()
				r.scrollOffset = offset
				r.mu.Unlock()
			}
		}
	}()

	// --- Config reloads ---
	r.configSub = r.bus.Subscribe("config_changed")
	go func() {
		for evt := range r.configSub {
			cfg, ok := evt.(*config.RootConfig)
			if !ok || cfg == nil {
				continue
			}
			log.Println("[Render] Applying new configuration...")
			r.applyConfig(cfg)
		}
	}()
}

// applyConfig updates renderer settings based on config data.
func (r *System) applyConfig(cfg *config.RootConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cfg.Renderer.CellWidth > 0 {
		r.cellW = cfg.Renderer.CellWidth
	}
	if cfg.Renderer.CellHeight > 0 {
		r.cellH = cfg.Renderer.CellHeight
	}

	// Apply custom theme if present
	if cfg.Theme.Name != "" {
		r.fgPalette = loadCustomPalette(cfg.Theme.Palette)
		r.bgPalette = loadCustomPalette(cfg.Theme.Palette)
		r.precacheTiles()
		log.Printf("[Render] Theme set to '%s'\n", cfg.Theme.Name)
	}

	// Optionally reload font here later (dynamic TTF loading)
}

// loadCustomPalette safely converts JSON [256][3]uint8 to base [8]colorRGBA.
// For now, truncates to first 8 entries.
func loadCustomPalette(src [256][3]uint8) [8]colorRGBA {
	var out [8]colorRGBA
	for i := 0; i < len(out) && i < len(src); i++ {
		out[i] = colorRGBA{src[i][0], src[i][1], src[i][2], 255}
	}
	return out
}

// AttachTerm links the renderer to a live TermBuffer.
func (r *System) AttachTerm(tb *components.TermBuffer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.term = tb
	if r.viewport != nil {
		r.viewport.term = tb
	}
}

// AttachScrollback links the renderer to a Scrollback buffer.
func (r *System) AttachScrollback(sb *components.Scrollback) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scrollback = sb
	if r.viewport != nil {
		r.viewport.scrollback = sb
	}
}

// Buffer returns the active TermBuffer.
func (r *System) Buffer() *components.TermBuffer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.term
}

// UpdateECS does nothing; rendering happens in Draw().
func (r *System) UpdateECS() {}

// Draw renders the composed view from TermBuffer + Scrollback.
func (r *System) Draw(screen *ebiten.Image) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.term == nil {
		screen.Fill(r.bgColor)
		return
	}

	// Initialize viewport lazily
	if r.viewport == nil {
		r.viewport = NewViewport(r.scrollback, r.term)
	}
	r.viewport.SetOffset(r.scrollOffset)

	screen.Fill(r.bgColor)

	lines := r.viewport.Compose(r.term.Height)
	if len(lines) == 0 {
		return
	}

	for y := 0; y < len(lines); y++ {
		row := lines[y]
		for x := 0; x < len(row); x++ {
			g := row[x]
			bg := r.bgTiles[g.Bg%8]
			fg := r.fgPalette[g.Fg%8].toColor()

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*r.cellW), float64(y*r.cellH))
			screen.DrawImage(bg, op)

			if g.Rune != 0 && g.Rune != ' ' {
				text.Draw(screen, string(g.Rune), r.fontFace,
					x*r.cellW, y*r.cellH+r.cellH-2, fg)
			}
		}
	}
}

// Layout defines logical render size based on terminal cell grid.
func (r *System) Layout(outW, outH int) (int, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.term != nil {
		return r.term.Width * r.cellW, r.term.Height * r.cellH
	}
	return 640, 384
}

// Interface check for ECS compliance.
var _ ecs.System = (*System)(nil)

