package storage

import (
	"os"
	"path/filepath"
)

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func DataDir() (string, error) {

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	base := filepath.Join(home, ".local", "share", "marginalia")

	if err := os.MkdirAll(base, 0755); err != nil {
		return "", err
	}

	return base, nil
}
