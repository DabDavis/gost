package input

// buildSeq prefixes with ESC for Alt+ combos.
func buildSeq(alt bool, b byte) []byte {
    if alt {
        return append([]byte{0x1b}, b)
    }
    return []byte{b}
}

// WriteToPTY is provided by the PTY system.
var WriteToPTY = func(b []byte) {
    // this will be linked dynamically by pty.go
}

