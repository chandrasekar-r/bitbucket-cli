// Package browser provides cross-platform URL opening.
package browser

import (
	"os/exec"
	"runtime"
)

// Open opens url in the system's default browser.
// Returns an error if the browser command fails; the process is not waited on
// (the browser opens asynchronously).
func Open(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default: // linux and others
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
