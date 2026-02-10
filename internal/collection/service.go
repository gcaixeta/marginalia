package collection

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/gcaixeta/marginalia/internal/storage"
)

// Collection represents a collection of notes with metadata
type Collection struct {
	Name      string
	FileCount int
	Path      string
}

// ListCollections returns all available collections with their file counts
func ListCollections() ([]Collection, error) {
	dataDir, err := storage.DataDir()
	if err != nil {
		return nil, err
	}

	var collections []Collection

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			collectionPath := filepath.Join(dataDir, entry.Name())
			fileCount, err := countFilesInDir(collectionPath)
			if err != nil {
				// If we can't count files, just set to 0
				fileCount = 0
			}

			collections = append(collections, Collection{
				Name:      entry.Name(),
				FileCount: fileCount,
				Path:      collectionPath,
			})
		}
	}

	// Sort collections alphabetically by name
	sort.Slice(collections, func(i, j int) bool {
		return collections[i].Name < collections[j].Name
	})

	return collections, nil
}

// GetCollectionStats returns the number of files in a specific collection
func GetCollectionStats(name string) (int, error) {
	dataDir, err := storage.DataDir()
	if err != nil {
		return 0, err
	}

	collectionPath := filepath.Join(dataDir, name)
	return countFilesInDir(collectionPath)
}

// CreateCollection creates a new collection directory
func CreateCollection(name string) error {
	dataDir, err := storage.DataDir()
	if err != nil {
		return err
	}

	collectionPath := filepath.Join(dataDir, name)
	return storage.EnsureDir(collectionPath)
}

// CollectionExists checks if a collection already exists
func CollectionExists(name string) bool {
	dataDir, err := storage.DataDir()
	if err != nil {
		return false
	}

	collectionPath := filepath.Join(dataDir, name)
	info, err := os.Stat(collectionPath)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// countFilesInDir counts the number of files (not directories) in a directory
func countFilesInDir(dirPath string) (int, error) {
	count := 0

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip entries that cause errors
		}

		// Only count files, not directories
		if !d.IsDir() {
			count++
		}

		return nil
	})

	return count, err
}
