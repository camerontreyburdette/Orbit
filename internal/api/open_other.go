//go:build !windows

package api

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenPath(targetPath string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", targetPath)
	default:
		command = exec.Command("xdg-open", targetPath)
	}

	if startError := command.Start(); startError != nil {
		return fmt.Errorf("failed to open path: %w", startError)
	}

	return nil
}
