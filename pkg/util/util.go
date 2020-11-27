package util

import (
	"io/ioutil"
	"path/filepath"
)

//FileCopy will attempt to copy an entire file from a source to a destination.
func FileCopy(src, dst string) error {
	input, err := ioutil.ReadFile(filepath.Clean(src))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dst, input, 0600)
	if err != nil {
		return err
	}
	return nil
}
