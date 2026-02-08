// Package browser provides utilities for opening files in the default browser.
package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Open opens the file at the given absolute path in the user's default browser.
// It launches the browser process in the background without waiting for it to exit.
//
// Supported platforms:
//   - darwin (macOS): uses the 'open' command
//   - linux: uses the 'xdg-open' command
//
// Returns an error if the platform is unsupported or if the browser command fails to start.
func Open(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
