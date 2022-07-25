// Package util provides utility functions.
package util

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

const fileModeUserReadWriteOnly = 0o600

// FileCopy will attempt to copy an entire file from a source to a destination.
func FileCopy(src, dst string) error {
	input, err := ioutil.ReadFile(filepath.Clean(src))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	err = ioutil.WriteFile(dst, input, fileModeUserReadWriteOnly)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
