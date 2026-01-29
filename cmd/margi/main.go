package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gcaixeta/marginalia/internal/slug"
	"github.com/gcaixeta/marginalia/internal/storage"
)

func newFile(collection, title string) (string, error) {

	dataDir, err := storage.DataDir()
	if err != nil {
		return "", err
	}
	collectionPath := filepath.Join(dataDir, collection)

	if err := storage.EnsureDir(collectionPath); err != nil {
		return "", err
	}

	filename := slug.SlugWithTime(title) + ".md"
	filePath := filepath.Join(collectionPath, filename)

	f, err := os.OpenFile(
		filePath,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		0644,
	)

	if err != nil {
		if os.IsExist(err) {
			return "", fmt.Errorf("A file with this name existis in collection %s!", collection)
		}

		return "", err
	}

	defer f.Close()

	content := fmt.Sprintf(`# %s

		## Created: %s
		## Collection: %s

		---

		`, title, time.Now().Format(time.RFC3339), collection)

	if _, err := f.WriteString(content); err != nil {
		return "", fmt.Errorf("Error writing %s to %s", content, filePath)
	}

	return filePath, nil
}

func listFiles(textGenre string) {
	fmt.Println("list files of type", textGenre)
}

func removeFile(textGenre string) {
	fmt.Println("remove file of type", textGenre)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage example: margi new note")
		return
	}

	action := os.Args[1]
	textType := os.Args[2]
	title := os.Args[3]

	switch action {
	case "new":
		_, err := newFile(textType, title)
		if err != nil {
			fmt.Println("Error creating new file:", err)
		}
	case "list":
		listFiles(textType)
	case "remove":
		removeFile(textType)
	}
}
