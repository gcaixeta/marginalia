package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gcaixeta/marginalia/internal/slug"
	"github.com/gcaixeta/marginalia/internal/snippet"
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

	filename := slug.MdSlugWithTime(title)
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

	content := snippet.Default(title, collection)

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

func openInEditor(filePath string) {
	// Create a command to run vim with the specified file
	cmd := exec.Command("nvim", filePath)

	// Attach the command's standard input, output, and error streams
	// to the current process's standard streams. This allows the user
	// to interact with Vim directly in the terminal.
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Opening %s in NeoVim...\n", filePath)

	// Run the command and wait for it to complete
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error running NeoVim: %v\n", err)
	}

	fmt.Printf("Vim session for %s closed.\n", filePath)
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
		filePath, err := newFile(textType, title)
		if err != nil {
			fmt.Println("Error creating new file:", err)
		}
		openInEditor(filePath)
	case "list":
		listFiles(textType)
	case "remove":
		removeFile(textType)
	}
}
