package storage

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func FindFilePath(fileName string) ([]string, error) {
	var foundFiles []string
	dataDir, err := DataDir()
	if err != nil {
		return nil, err
	}

	// filepath.WalkDir traverses the directory tree rooted at rootDir, calling the walkFn for each file/directory.
	err = filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // returning nil allows walking to continue in other branches
		}

		// Check if the current entry is a regular file and its name contains the search term (case-insensitive)
		if !d.IsDir() && strings.Contains(strings.ToLower(d.Name()), strings.ToLower(fileName)) {
			foundFiles = append(foundFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return foundFiles, nil
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func DataDir() (string, error) {

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	base := filepath.Join(home, ".local", "share", "marginalia", "collections")

	if err := os.MkdirAll(base, 0755); err != nil {
		return "", err
	}

	return base, nil
}
