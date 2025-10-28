package util

import "log"

// clipboardBuffer acts as an in-memory clipboard fallback.
var clipboardBuffer string

// SetClipboardString saves text to clipboard (or buffer if unsupported).
func SetClipboardString(s string) {
	clipboardBuffer = s
	log.Printf("[Clipboard] Copied %d bytes (fallback buffer)", len(s))
}

// ClipboardString returns the last copied value.
func ClipboardString() string {
	return clipboardBuffer
}

