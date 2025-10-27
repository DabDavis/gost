package cursor

import "fmt"

// Describe returns a human-readable summary of cursor state.
func (c *System) Describe() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fmt.Sprintf("Cursor[%s, blink=%v, color=%v]",
		c.style.Shape, c.style.Blink, c.style.Color)
}

