package render

import "gost/internal/components"

// subscribeEvents wires up renderer to listen for term updates, scroll offsets, and keypress resets.
func (r *System) subscribeEvents() {
	// --- Terminal updates ---
	subTerm := r.bus.Subscribe("term_updated")
	go func() {
		for evt := range subTerm {
			if tb, ok := evt.(*components.TermBuffer); ok {
				r.mu.Lock()
				r.term = tb
				r.mu.Unlock()
			}
		}
	}()

	// --- Scroll offset changes ---
	subScroll := r.bus.Subscribe("scroll_offset_changed")
	go func() {
		for evt := range subScroll {
			if off, ok := evt.(int); ok {
				r.SetScrollOffset(off)
			}
		}
	}()

	// --- Any key input resets scrollback ---
	subInput := r.bus.Subscribe("key_any_pressed")
	go func() {
		for range subInput {
			r.ResetScrollOffset()
		}
	}()
}

// SetScrollOffset adjusts the visible scrollback offset.
func (r *System) SetScrollOffset(offset int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scrollOffset = offset
	r.scrollDirty = (offset > 0)
}

// ResetScrollOffset clears scrollback view and returns to live output.
func (r *System) ResetScrollOffset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.scrollDirty {
		r.scrollOffset = 0
		r.scrollDirty = false
	}
}

