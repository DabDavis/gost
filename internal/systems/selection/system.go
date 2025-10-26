package selection

import (
    "strings"
    "image/color"

    "github.com/atotto/clipboard"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "gost/internal/components"
)

// System enables click-drag text selection and clipboard copy.
type System struct {
    buffer         *components.TermBuffer
    selecting      bool
    startX, startY int
    endX, endY     int
    cellW, cellH   int
}

// NewSystem creates a new selection handler.
func NewSystem(buffer *components.TermBuffer, cellW, cellH int) *System {
    return &System{
        buffer: buffer,
        cellW:  cellW,
        cellH:  cellH,
    }
}

// UpdateECS handles mouse drag and clipboard copy.
func (s *System) UpdateECS() {
    if s.buffer == nil {
        return
    }

    mx, my := ebiten.CursorPosition()
    cx := mx / s.cellW
    cy := my / s.cellH

    if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
        if !s.selecting {
            s.selecting = true
            s.startX, s.startY = cx, cy
            s.endX, s.endY = cx, cy
        } else {
            s.endX, s.endY = cx, cy
        }
    } else if s.selecting {
        // Mouse released → finalize copy
        s.selecting = false
        txt := s.extractText()
        if txt != "" {
            _ = clipboard.WriteAll(txt)
        }
    }
}

// DrawSelection renders the active selection highlight.
func (s *System) DrawSelection(screen *ebiten.Image) {
    if !s.selecting || s.buffer == nil {
        return
    }

    sx, sy := s.startX, s.startY
    ex, ey := s.endX, s.endY
    if sy > ey {
        sy, ey = ey, sy
    }
    if sx > ex {
        sx, ex = ex, sx
    }

    hlColor := color.RGBA{80, 120, 255, 120}

    for y := sy; y <= ey && y < s.buffer.Height; y++ {
        for x := sx; x <= ex && x < s.buffer.Width; x++ {
            x0 := float64(x * s.cellW)
            y0 := float64(y * s.cellH)
            ebitenutil.DrawRect(screen, x0, y0, float64(s.cellW), float64(s.cellH), hlColor)
        }
    }
}

// extractText returns the selected text as a string for clipboard copy.
func (s *System) extractText() string {
    if s.buffer == nil {
        return ""
    }

    sx, sy := s.startX, s.startY
    ex, ey := s.endX, s.endY

    // normalize coordinates
    if sy > ey {
        sy, ey = ey, sy
    }
    if sx > ex {
        sx, ex = ex, sx
    }

    var out strings.Builder
    for y := sy; y <= ey && y < s.buffer.Height; y++ {
        for x := sx; x <= ex && x < s.buffer.Width; x++ {
            out.WriteRune(s.buffer.Cells[y][x].Rune)
        }
        if y < ey {
            out.WriteByte('\n')
        }
    }
    return out.String()
}

