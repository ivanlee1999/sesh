//go:build darwin

package notify

import (
	"os/exec"
	"strings"
)

// Send displays a native macOS notification using osascript.
func Send(title, body, sound string) {
	script := `display notification "` + escape(body) + `" with title "` + escape(title) + `" sound name "` + escape(sound) + `"`
	exec.Command("osascript", "-e", script).Run() //nolint:errcheck
}

// escape escapes backslashes and double quotes for AppleScript string literals.
func escape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
