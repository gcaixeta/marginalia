package storage

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileItem represents a file with its metadata
type FileItem struct {
	Path           string    // Full path to the file
	Name           string    // File name
	Collection     string    // Collection name (parent directory)
	ModTime        time.Time // Last modification time
	Size           int64     // File size in bytes
}

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

// ListAllFiles returns all files in all collections with their metadata
func ListAllFiles() ([]FileItem, error) {
	dataDir, err := DataDir()
	if err != nil {
		return nil, err
	}

	var files []FileItem

	err = filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip entries with errors
		}

		// Only process regular files (not directories)
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return nil // Skip if we can't get info
			}

			// Get the collection name (parent directory name)
			parentDir := filepath.Dir(path)
			collectionName := filepath.Base(parentDir)

			files = append(files, FileItem{
				Path:       path,
				Name:       d.Name(),
				Collection: collectionName,
				ModTime:    info.ModTime(),
				Size:       info.Size(),
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// FindFiles returns files matching a search term with their metadata
func FindFiles(searchTerm string) ([]FileItem, error) {
	allFiles, err := ListAllFiles()
	if err != nil {
		return nil, err
	}

	if searchTerm == "" {
		return allFiles, nil
	}

	matchingFiles := []FileItem{}
	searchLower := strings.ToLower(searchTerm)

	for _, file := range allFiles {
		if strings.Contains(strings.ToLower(file.Name), searchLower) ||
			strings.Contains(strings.ToLower(file.Collection), searchLower) {
			matchingFiles = append(matchingFiles, file)
		}
	}

	return matchingFiles, nil
}
