package processing

import (
	"fmt"
	"os"
)

func newWorkingDirectory() (string, error) {
	directory, err := os.MkdirTemp("", "gallery-image-*")
	if err != nil {
		return "", fmt.Errorf("create processing workspace: %w", err)
	}
	return directory, nil
}

func removeWorkingDirectory(directory string) {
	// The Lambda runtime owns /tmp, but each invocation must remove its own
	// inputs so warm workers do not accumulate image files between messages.
	_ = os.RemoveAll(directory)
}
