package scrollback

import "gost/internal/components"

// listenEvents runs continuously, reacting to bus events.
func (s *System) listenEvents() {
	for {
		select {
		case evt := <-s.termScrollSub:
			if line, ok := evt.([]components.Glyph); ok {
				s.scrollback.PushLine(line)
			}

		case <-s.scrollUpSub:
			s.adjustOffset(+s.scrollStep)

		case <-s.scrollDownSub:
			s.adjustOffset(-s.scrollStep)

		case <-s.scrollPageUpSub:
			if s.buffer != nil {
				s.adjustOffset(+s.buffer.Height)
			}

		case <-s.scrollPageDnSub:
			if s.buffer != nil {
				s.adjustOffset(-s.buffer.Height)
			}
		}
	}
}

