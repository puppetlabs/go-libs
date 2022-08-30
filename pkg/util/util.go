// Package util provides utility functions.
package util

import (
	"fmt"
	"os"
	"path/filepath"
)

const fileModeUserReadWriteOnly = 0o600

// FileCopy will attempt to copy an entire file from a source to a destination.
func FileCopy(src, dst string) error {
	input, err := os.ReadFile(filepath.Clean(src))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	err = os.WriteFile(dst, input, fileModeUserReadWriteOnly)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
