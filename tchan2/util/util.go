package util

import (
	"os"

	"github.com/pkg/errors"
)

// DirExists checks for the existence of a directory in the filesystem.
func DirExists(dir string) (bool, error) {
	if stat, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	} else if stat.IsDir() {
		return true, nil
	} else {
		return false, errors.Errorf("file exists and is not a directory: %s", dir)
	}
}

// FileExists checks the existence of a file in the filesystem.
func FileExists(filePath string) (bool, error) {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}
