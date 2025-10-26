package selection

import "strings"

// extractText retrieves the selected text as a single string.
func (s *System) extractText() string {
	if s.buffer == nil {
		return ""
	}

	sx, sy, ex, ey := normalizeRect(s.startX, s.startY, s.endX, s.endY)
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

