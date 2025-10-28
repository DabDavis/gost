package render

import (
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"

	"gost/internal/components"
	"gost/internal/ecs"
	"gost/internal/events"
)

// -----------------------------------------------------------------------------
// System definition
// -----------------------------------------------------------------------------

// System draws the composed terminal view using cached backgrounds and font.
// It also handles scroll offset syncing with the Scrollback system.
type System struct {
	mu sync.RWMutex

	bus        *events.Bus
	term       *components.TermBuffer
	scrollback *components.Scrollback
	viewport   *Viewport

	fontFace font.Face
	cellW, cellH int

	scrollOffset int
	fgPalette    [8]colorRGBA
	bgPalette    [8]colorRGBA
	bgTiles      [8]*ebiten.Image
	bgColor      color.Color

	offsetSub <-chan events.Event // scroll offset listener
}

// NewSystem initializes the renderer and prepares background tiles.
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
	return r
}

// -----------------------------------------------------------------------------
// Attach and initialization
// -----------------------------------------------------------------------------

func (r *System) AttachTerm(term *components.TermBuffer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.term = term
	r.tryInitViewport()
}

func (r *System) AttachScrollback(sb *components.Scrollback) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scrollback = sb
	r.tryInitViewport()
}

func (r *System) tryInitViewport() {
	if r.viewport == nil && r.term != nil && r.scrollback != nil && r.bus != nil {
		r.viewport = NewViewport(r.scrollback, r.term, r.bus)
		r.subscribeScrollOffset()
	}
}

// -----------------------------------------------------------------------------
// Event subscription
// -----------------------------------------------------------------------------

func (r *System) subscribeScrollOffset() {
	if r.bus == nil {
		return
	}
	r.offsetSub = r.bus.Subscribe("scroll_offset_changed")
	go func() {
		for evt := range r.offsetSub {
			if offset, ok := evt.(int); ok {
				r.mu.Lock()
				r.scrollOffset = offset
				if r.viewport != nil {
					r.viewport.SetOffset(offset)
				}
				r.mu.Unlock()
			}
		}
	}()
}

// -----------------------------------------------------------------------------
// ECS integration
// -----------------------------------------------------------------------------

func (r *System) UpdateECS() {
	if r.viewport != nil {
		r.viewport.HandleScrollInput()
	}
}

func (r *System) Layout(outW, outH int) (int, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.term != nil {
		return r.term.Width * r.cellW, r.term.Height * r.cellH
	}
	return 640, 384
}

func (r *System) Buffer() *components.TermBuffer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.term
}

// -----------------------------------------------------------------------------
// Rendering
// -----------------------------------------------------------------------------

func (r *System) Draw(screen *ebiten.Image) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	screen.Fill(r.bgColor)
	if r.term == nil {
		return
	}

	lines := r.composeVisibleLines()
	for y := 0; y < len(lines); y++ {
		row := lines[y]
		for x := 0; x < r.term.Width && x < len(row); x++ {
			g := row[x]

			bgColor := r.resolveColor(g.Bg, false)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*r.cellW), float64(y*r.cellH))

			if g.Bg < len(r.bgTiles) && r.bgTiles[g.Bg] != nil {
				screen.DrawImage(r.bgTiles[g.Bg], op)
			} else {
				tile := ebiten.NewImage(r.cellW, r.cellH)
				tile.Fill(bgColor)
				screen.DrawImage(tile, op)
			}

			if g.Rune == 0 || g.Rune == ' ' {
				continue
			}
			fgColor := r.resolveColor(g.Fg, true)
			text.Draw(screen, string(g.Rune), r.fontFace,
				x*r.cellW, y*r.cellH+r.cellH-2, fgColor)
		}
	}
}

func (r *System) composeVisibleLines() [][]components.Glyph {
	var lines [][]components.Glyph
	if r.scrollback != nil && r.scrollOffset > 0 {
		sbLines := r.scrollback.GetVisibleLines(r.scrollOffset, r.term.Height)
		lines = append(lines, sbLines...)
	}
	lines = append(lines, r.term.Cells...)
	if len(lines) > r.term.Height {
		lines = lines[len(lines)-r.term.Height:]
	}
	return lines
}

// -----------------------------------------------------------------------------
// Viewport (merged from viewport.go)
// -----------------------------------------------------------------------------

type Viewport struct {
	scrollback *components.Scrollback
	term       *components.TermBuffer
	bus        *events.Bus
	offset     int
}

func NewViewport(sb *components.Scrollback, term *components.TermBuffer, bus *events.Bus) *Viewport {
	return &Viewport{scrollback: sb, term: term, bus: bus}
}

func (v *Viewport) SetOffset(offset int) { v.offset = offset }
func (v *Viewport) Offset() int          { return v.offset }

func (v *Viewport) HandleScrollInput() {
	if v.bus == nil {
		return
	}
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		if wheelY > 0 {
			v.bus.Publish("scroll_up", nil)
		} else {
			v.bus.Publish("scroll_down", nil)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
		v.bus.Publish("scroll_up", nil)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
		v.bus.Publish("scroll_down", nil)
	}

	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyEnd) {
		v.bus.Publish("scroll_reset_request", nil)
	}
}

// -----------------------------------------------------------------------------
// Color utilities (merged from util.go & draw.go)
// -----------------------------------------------------------------------------

type colorRGBA struct{ r, g, b, a uint8 }

func (c colorRGBA) toColor() color.Color { return color.RGBA{c.r, c.g, c.b, c.a} }

func defaultFgPalette() [8]colorRGBA {
	return [8]colorRGBA{
		{0, 0, 0, 255}, {205, 49, 49, 255}, {13, 188, 121, 255},
		{229, 229, 16, 255}, {36, 114, 200, 255}, {188, 63, 188, 255},
		{17, 168, 205, 255}, {229, 229, 229, 255},
	}
}

func defaultBgPalette() [8]colorRGBA {
	return [8]colorRGBA{
		{0, 0, 0, 255}, {60, 0, 0, 255}, {0, 60, 0, 255},
		{60, 60, 0, 255}, {0, 0, 60, 255}, {60, 0, 60, 255},
		{0, 60, 60, 255}, {80, 80, 80, 255},
	}
}

func brightPalette() [16]colorRGBA {
	return [16]colorRGBA{
		{0, 0, 0, 255}, {205, 49, 49, 255}, {13, 188, 121, 255},
		{229, 229, 16, 255}, {36, 114, 200, 255}, {188, 63, 188, 255},
		{17, 168, 205, 255}, {229, 229, 229, 255},
		{102, 102, 102, 255}, {241, 76, 76, 255}, {35, 209, 139, 255},
		{245, 245, 67, 255}, {59, 142, 234, 255}, {214, 112, 214, 255},
		{41, 184, 219, 255}, {255, 255, 255, 255},
	}
}

func (r *System) precacheTiles() {
	for i := range r.bgPalette {
		tile := ebiten.NewImage(r.cellW, r.cellH)
		tile.Fill(r.bgPalette[i].toColor())
		r.bgTiles[i] = tile
	}
}

func make256Color(index int) colorRGBA {
	switch {
	case index < 16:
		return brightPalette()[index%16]
	case index >= 16 && index < 232:
		i := index - 16
		r := uint8((i / 36) * 51)
		g := uint8(((i / 6) % 6) * 51)
		b := uint8((i % 6) * 51)
		return colorRGBA{r, g, b, 255}
	default:
		level := uint8(8 + (index-232)*10)
		return colorRGBA{level, level, level, 255}
	}
}

func (r *System) resolveColor(idx int, isForeground bool) color.Color {
	if idx >= 0 && idx < 8 {
		if isForeground {
			return r.fgPalette[idx].toColor()
		}
		return r.bgPalette[idx].toColor()
	}
	if idx >= 8 && idx < 16 {
		return brightPalette()[idx].toColor()
	}
	return make256Color(idx).toColor()
}

// -----------------------------------------------------------------------------
// ECS compliance
// -----------------------------------------------------------------------------

var _ ecs.System = (*System)(nil)

