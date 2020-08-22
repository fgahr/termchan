package util

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

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
