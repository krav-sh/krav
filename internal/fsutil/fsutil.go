package fsutil

import (
	"os"
	"path/filepath"
)

func IsAccessibleDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err // permission denied or other error
	}
	return info.IsDir(), nil
}

func IsAccessibleFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err // permission denied or other error
	}
	return info.Mode().IsRegular(), nil
}

func JoinBasePath(base string, elem ...string) string {
	return filepath.Join(append([]string{base}, elem...)...)
}
