//go:build !darwin

package notify

// Send is a no-op on non-macOS platforms.
func Send(title, body, sound string) {}
