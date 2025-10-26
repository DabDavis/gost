package input

import (
	"time"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
)

// handlePrintable sends A–Z, 0–9, space, and punctuation keys.
func handlePrintable(s *System, now time.Time) {
	for k := ebiten.KeyA; k <= ebiten.KeyZ; k++ {
		processCharKey(s, now, k, 'a'+rune(k-ebiten.KeyA))
	}
	for k := ebiten.Key0; k <= ebiten.Key9; k++ {
		processCharKey(s, now, k, '0'+rune(k-ebiten.Key0))
	}

	// punctuation
	for key, ch := range punctuationMap {
		processCharKey(s, now, key, ch)
	}

	// space
	processCharKey(s, now, ebiten.KeySpace, ' ')
}

// handleSpecial processes Enter, Tab, Backspace, arrows.
func handleSpecial(s *System, now time.Time) {
	processKey(s, now, ebiten.KeyEnter, []byte{'\n'})
	processKey(s, now, ebiten.KeyTab, []byte{'\t'})
	processKey(s, now, ebiten.KeyBackspace, []byte{0x7f})

	// arrow keys
	processKey(s, now, ebiten.KeyArrowUp, []byte("\x1b[A"))
	processKey(s, now, ebiten.KeyArrowDown, []byte("\x1b[B"))
	processKey(s, now, ebiten.KeyArrowRight, []byte("\x1b[C"))
	processKey(s, now, ebiten.KeyArrowLeft, []byte("\x1b[D"))
}

// processCharKey handles Shift/Ctrl/Alt modifiers.
func processCharKey(s *System, now time.Time, key ebiten.Key, ch rune) {
	shifted := ebiten.IsKeyPressed(ebiten.KeyShift)
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl)
	alt := ebiten.IsKeyPressed(ebiten.KeyAlt)

	if shifted && unicode.IsLetter(ch) {
		ch = unicode.ToUpper(ch)
	}
	if shifted {
		if mapped, ok := shiftPunct[ch]; ok {
			ch = mapped
		}
	}

	// Ctrl handling
	if ctrl && unicode.IsLetter(ch) {
		lower := unicode.ToLower(ch)
		ctrlCode := byte(lower-'a') + 1
		processKey(s, now, key, buildSeq(alt, ctrlCode))
		return
	}

	processKey(s, now, key, buildSeq(alt, byte(ch)))
}

// processKey applies repeat delay logic and writes to PTY.
// 🔹 Now also publishes "key_any_pressed" when the key is first pressed.
func processKey(s *System, now time.Time, key ebiten.Key, seq []byte) {
	ks, ok := s.keys[key]
	if !ok {
		ks = &keyState{}
		s.keys[key] = ks
	}

	if ebiten.IsKeyPressed(key) {
		if !ks.pressed {
			ks.pressed = true
			ks.next = now.Add(repeatDelay)
			WriteToPTY(seq)
			s.publishKeyAny() // notify renderer to reset scrollback
			return
		}
		if now.After(ks.next) {
			ks.next = now.Add(repeatRate)
			WriteToPTY(seq)
		}
	} else {
		ks.pressed = false
	}
}

