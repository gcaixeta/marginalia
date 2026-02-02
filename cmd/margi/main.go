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

func editFile(title string) {
	files, err := storage.FindFilePath(title)
	if err != nil {
		fmt.Printf("Error searching for files: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Printf("No files found matching: %s\n", title)
		return
	}

	if len(files) == 1 {
		openInEditor(files[0])
		return
	}

	fmt.Println("Multiple files found. Please choose one:")
	dataDir, _ := storage.DataDir()
	for i, file := range files {
		relPath, err := filepath.Rel(dataDir, file)
		if err != nil {
			relPath = file
		}
		fmt.Printf("[%d] %s\n", i+1, relPath)
	}

	var choice int
	fmt.Print("Enter the number of the file to edit: ")
	_, err = fmt.Scanf("%d", &choice)
	if err != nil || choice < 1 || choice > len(files) {
		fmt.Println("Invalid selection.")
		return
	}

	openInEditor(files[choice-1])
}

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

	content, err := snippet.ReadSnippet(title, collection)
	if err != nil {
		content = snippet.Default(title, collection)
	}

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
	if len(os.Args) < 2 {
		fmt.Println("Usage: margi [action] [arguments]")
		fmt.Println("Actions:")
		fmt.Println("  new [collection] [title]")
		fmt.Println("  edit [search_term]")
		fmt.Println("  list [collection]")
		fmt.Println("  remove [collection]")
		return
	}

	action := os.Args[1]

	switch action {
	case "new":
		if len(os.Args) < 4 {
			fmt.Println("Usage: margi new [collection] [title]")
			return
		}
		textType := os.Args[2]
		title := os.Args[3]
		filePath, err := newFile(textType, title)
		if err != nil {
			fmt.Println("Error creating new file:", err)
			return
		}
		openInEditor(filePath)
	case "edit":
		if len(os.Args) < 3 {
			fmt.Println("Usage: margi edit [search_term]")
			return
		}
		title := os.Args[2]
		editFile(title)
	case "list":
		if len(os.Args) < 3 {
			fmt.Println("Usage: margi list [collection]")
			return
		}
		textType := os.Args[2]
		listFiles(textType)
	case "remove":
		if len(os.Args) < 3 {
			fmt.Println("Usage: margi remove [collection]")
			return
		}
		textType := os.Args[2]
		removeFile(textType)
	default:
		fmt.Printf("Unknown action: %s\n", action)
	}
}
