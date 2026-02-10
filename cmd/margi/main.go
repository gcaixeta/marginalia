package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gcaixeta/marginalia/internal/collection"
	"github.com/gcaixeta/marginalia/internal/slug"
	"github.com/gcaixeta/marginalia/internal/snippet"
	"github.com/gcaixeta/marginalia/internal/storage"
	"github.com/gcaixeta/marginalia/internal/ui"
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

func listCollections() {
	collections, err := collection.ListCollections()
	if err != nil {
		fmt.Printf("Error listing collections: %v\n", err)
		return
	}

	if len(collections) == 0 {
		fmt.Println("Nenhuma collection encontrada.")
		fmt.Println("Crie uma nova collection com: margi new [título]")
		return
	}

	fmt.Println("Collections disponíveis:")
	fmt.Println()
	for _, c := range collections {
		plural := ""
		if c.FileCount != 1 {
			plural = "s"
		}
		fmt.Printf("  • %s (%d nota%s)\n", c.Name, c.FileCount, plural)
	}
}

func deleteFile(searchTerm string) {
	// Run the visual delete picker with optional initial filter
	selectedFile, err := ui.RunDeletePicker(searchTerm)
	if err != nil {
		fmt.Printf("Operação cancelada: %v\n", err)
		return
	}

	if selectedFile == nil {
		fmt.Println("Nenhum arquivo selecionado.")
		return
	}

	// Show interactive confirmation dialog
	dataDir, _ := storage.DataDir()
	confirmed, err := ui.RunConfirmDialog(selectedFile.Path, dataDir)
	if err != nil {
		fmt.Printf("Erro ao mostrar diálogo de confirmação: %v\n", err)
		return
	}

	if !confirmed {
		fmt.Println("Exclusão cancelada.")
		return
	}

	// Delete the file
	err = os.Remove(selectedFile.Path)
	if err != nil {
		fmt.Printf("Erro ao excluir arquivo: %v\n", err)
		return
	}

	fmt.Printf("✓ Arquivo excluído com sucesso: %s/%s\n", selectedFile.Collection, selectedFile.Name)
}

func removeCollection(collectionName string) {
	// TODO: Implement collection removal functionality
	fmt.Println("Funcionalidade de remoção de collection ainda não implementada.")
	fmt.Printf("Collection: %s\n", collectionName)
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
		fmt.Println("  new [title] - criar nova nota (mostra seletor de collection)")
		fmt.Println("  new [collection] [title] - criar nova nota em collection específica")
		fmt.Println("  edit [search_term] - editar nota existente")
		fmt.Println("  rm [search_term] - excluir nota existente (interface visual com confirmação)")
		fmt.Println("  collections - listar todas as collections")
		fmt.Println("  list [collection] - listar arquivos de uma collection")
		fmt.Println("  remove [collection] - remover collection")
		return
	}

	action := os.Args[1]

	switch action {
	case "new":
		var collectionName, title string
		
		if len(os.Args) == 3 {
			// Only title provided, show collection picker
			title = os.Args[2]
			selectedCollection, err := ui.RunPicker()
			if err != nil {
				fmt.Printf("Operação cancelada: %v\n", err)
				return
			}
			collectionName = selectedCollection
		} else if len(os.Args) >= 4 {
			// Collection and title provided
			collectionName = os.Args[2]
			title = os.Args[3]
		} else {
			fmt.Println("Usage: margi new [title] ou margi new [collection] [title]")
			return
		}
		
		filePath, err := newFile(collectionName, title)
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
	case "rm":
		// Search term is optional - if not provided, show all files
		searchTerm := ""
		if len(os.Args) >= 3 {
			searchTerm = os.Args[2]
		}
		deleteFile(searchTerm)
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
		collectionName := os.Args[2]
		removeCollection(collectionName)
	case "collections":
		listCollections()
	default:
		fmt.Printf("Unknown action: %s\n", action)
	}
}
