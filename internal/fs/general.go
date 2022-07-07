package fs

import (
	"errors"
	"os"
)

func existsAndPassesTest(name string, test func(os.FileInfo) bool) (bool, error) {
	info, exists, err := stat(name)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	return test(info), nil
}

func stat(name string) (os.FileInfo, bool, error) {
	info, err := os.Stat(name)
	if err == nil {
		return info, true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	return nil, false, err
}
