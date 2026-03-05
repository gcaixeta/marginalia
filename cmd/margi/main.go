package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gcaixeta/marginalia/internal/collection"
	"github.com/gcaixeta/marginalia/internal/config"
	"github.com/gcaixeta/marginalia/internal/editor"
	"github.com/gcaixeta/marginalia/internal/slug"
	"github.com/gcaixeta/marginalia/internal/snippet"
	"github.com/gcaixeta/marginalia/internal/storage"
	"github.com/gcaixeta/marginalia/internal/ui"
)

func editFile(title, editorCmd string) {
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
		editor.OpenInEditor(files[0], editorCmd)
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

	editor.OpenInEditor(files[choice-1], editorCmd)
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

func main() {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
		config.Save(cfg)
	}

	editorCmd := editor.ResolveEditor(cfg.Editor)

	storage.Synchronize()

	if len(os.Args) < 2 {
		selected, err := ui.RunBrowsePicker()
		if err != nil || selected == nil {
			return
		}
		editor.OpenInEditor(selected.Path, editorCmd)
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
		editor.OpenInEditor(filePath, editorCmd)
	case "edit":
		if len(os.Args) < 3 {
			fmt.Println("Usage: margi edit [search_term]")
			return
		}
		title := os.Args[2]
		editFile(title, editorCmd)
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
	case "collections":
		listCollections()
	default:
		fmt.Printf("Unknown action: %s\n", action)
	}
}
