package selection

import (
	"github.com/atotto/clipboard"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// cursorToCell converts the mouse position to terminal cell coordinates.
func (s *System) cursorToCell() (int, int) {
	mx, my := ebiten.CursorPosition()
	return mx / s.cellW, my / s.cellH
}

// UpdateSelectionInput handles both mouse and keyboard input for selection.
func (s *System) UpdateSelectionInput() {
	if s.buffer == nil {
		return
	}

	cx, cy := s.cursorToCell()
	mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	if mousePressed {
		s.updateSelection(cx, cy)
	} else if s.selecting && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		// Mouse released — keep selection persistent
		s.selecting = false
		s.bus.Publish("selection_changed", [4]int{s.startX, s.startY, s.endX, s.endY})
	}

	s.handleKeyboardShortcuts()
	s.handleKeyboardSelection()
}

// updateSelection begins or updates an active drag region.
func (s *System) updateSelection(cx, cy int) {
	if !s.selecting {
		s.selecting = true
		s.startX, s.startY = cx, cy
	}
	s.endX, s.endY = cx, cy

	if s.bus != nil {
		s.bus.Publish("selection_changed", [4]int{s.startX, s.startY, s.endX, s.endY})
	}
}

// handleKeyboardShortcuts processes Ctrl+Shift+C and Escape keys.
func (s *System) handleKeyboardShortcuts() {
	// Ctrl+Shift+C → Copy
	if ebiten.IsKeyPressed(ebiten.KeyControl) && ebiten.IsKeyPressed(ebiten.KeyShift) &&
		inpututil.IsKeyJustPressed(ebiten.KeyC) {
		txt := s.extractText()
		if txt != "" {
			_ = clipboard.WriteAll(txt)
			s.bus.Publish("clipboard_copy", txt)
		}
	}

	// Escape → Clear selection
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.clearSelection()
		s.bus.Publish("selection_changed", nil)
	}
}

// handleKeyboardSelection allows extending or shrinking the selection with Shift+Arrow.
func (s *System) handleKeyboardSelection() {
	if !ebiten.IsKeyPressed(ebiten.KeyShift) {
		return
	}

	// Fetch cursor as logical anchor point
	cx, cy := s.buffer.GetCursor()

	// Extend selection with arrow keys
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft):
		if cx > 0 {
			cx--
		}
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowRight):
		if cx < s.buffer.Width-1 {
			cx++
		}
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowUp):
		if cy > 0 {
			cy--
		}
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowDown):
		if cy < s.buffer.Height-1 {
			cy++
		}
	default:
		return
	}

	// Initialize or extend selection
	if !s.selecting {
		s.selecting = true
		s.startX, s.startY = cx, cy
	}
	s.endX, s.endY = cx, cy

	// Update visible cursor
	s.buffer.SetCursor(cx, cy)

	// Notify ECS
	if s.bus != nil {
		s.bus.Publish("selection_changed", [4]int{s.startX, s.startY, s.endX, s.endY})
	}
}

// clearSelection resets selection state.
func (s *System) clearSelection() {
	s.startX, s.startY, s.endX, s.endY = 0, 0, 0, 0
	s.selecting = false
}

