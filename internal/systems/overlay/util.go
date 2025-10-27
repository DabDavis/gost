package overlay

// SafeIndex clamps index values safely for any slice operations.
func SafeIndex(i, max int) int {
	if i < 0 {
		return 0
	}
	if i >= max {
		return max - 1
	}
	return i
}

