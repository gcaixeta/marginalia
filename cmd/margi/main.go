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

func editFile(title, editorCmd string, sync *storage.GitSync) {
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
		if sync != nil {
			if err := sync.CommitAndPush("edit: " + title); err != nil {
				fmt.Printf("Warning: git sync failed: %v\n", err)
			}
		}
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
	if sync != nil {
		if err := sync.CommitAndPush("edit: " + title); err != nil {
			fmt.Printf("Warning: git sync failed: %v\n", err)
		}
	}
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
			return "", fmt.Errorf("A file with this name exists in collection %s!", collection)
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

func deleteFile(searchTerm string, sync *storage.GitSync) {
	selectedFile, err := ui.RunDeletePicker(searchTerm)
	if err != nil {
		fmt.Printf("Operação cancelada: %v\n", err)
		return
	}

	if selectedFile == nil {
		fmt.Println("Nenhum arquivo selecionado.")
		return
	}

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

	err = os.Remove(selectedFile.Path)
	if err != nil {
		fmt.Printf("Erro ao excluir arquivo: %v\n", err)
		return
	}

	fmt.Printf("✓ Arquivo excluído com sucesso: %s/%s\n", selectedFile.Collection, selectedFile.Name)

	if sync != nil {
		if err := sync.CommitAndPush("rm: " + selectedFile.Collection + "/" + selectedFile.Name); err != nil {
			fmt.Printf("Warning: git sync failed: %v\n", err)
		}
	}
}

func runNew(editorCmd string, sync *storage.GitSync) {
	var collectionName, title string

	if len(os.Args) == 3 {
		title = os.Args[2]
		selectedCollection, err := ui.RunPicker()
		if err != nil {
			fmt.Printf("Operação cancelada: %v\n", err)
			return
		}
		collectionName = selectedCollection
	} else if len(os.Args) >= 4 {
		collectionName = os.Args[2]
		title = os.Args[3]
	} else {
		fmt.Fprintln(os.Stderr, "Usage: margi new [title] ou margi new [collection] [title]")
		return
	}

	filePath, err := newFile(collectionName, title)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating new file: %v\n", err)
		return
	}
	editor.OpenInEditor(filePath, editorCmd)
	if sync != nil {
		if err := sync.CommitAndPush("add: " + title); err != nil {
			fmt.Printf("Warning: git sync failed: %v\n", err)
		}
	}
}

func runSync(sync *storage.GitSync) {
	if sync == nil {
		fmt.Fprintln(os.Stderr, "no backup configured")
		return
	}
	if err := sync.Synchronize(); err != nil {
		fmt.Fprintf(os.Stderr, "sync error: %v\n", err)
		os.Exit(1)
	}
	if err := sync.CommitAndPush("sync"); err != nil {
		fmt.Fprintf(os.Stderr, "sync error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
		config.Save(cfg)
	}

	editorCmd := editor.ResolveEditor(cfg.Editor)

	sync, err := storage.NewGitSync(&cfg.Backup)
	if err != nil {
		fmt.Printf("Warning: could not initialize git sync: %v\n", err)
	}
	if sync != nil {
		if err := sync.Synchronize(); err != nil {
			fmt.Printf("Warning: git sync failed: %v\n", err)
		}
	}

	if len(os.Args) < 2 {
		selected, err := ui.RunBrowsePicker()
		if err != nil || selected == nil {
			return
		}
		editor.OpenInEditor(selected.Path, editorCmd)
		if sync != nil {
			if err := sync.CommitAndPush("edit: " + selected.Collection + "/" + selected.Name); err != nil {
				fmt.Printf("Warning: git sync failed: %v\n", err)
			}
		}
		return
	}

	cmds := map[string]func(){
		"new": func() { runNew(editorCmd, sync) },
		"edit": func() {
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "Usage: margi edit [search_term]")
				return
			}
			editFile(os.Args[2], editorCmd, sync)
		},
		"rm": func() {
			searchTerm := ""
			if len(os.Args) >= 3 {
				searchTerm = os.Args[2]
			}
		}
	case "edit":
		if len(os.Args) < 3 {
			fmt.Println("Usage: margi edit [search_term]")
			return
		}
		title := os.Args[2]
		editFile(title, editorCmd, sync)
	case "rm":
		searchTerm := ""
		if len(os.Args) >= 3 {
			searchTerm = os.Args[2]
		}
		deleteFile(searchTerm, sync)
	case "list":
		if len(os.Args) < 3 {
			fmt.Println("Usage: margi list [collection]")
			return
		}
		textType := os.Args[2]
		listFiles(textType)
	case "collections":
		listCollections()
	case "sync":
		if sync == nil {
			fmt.Println("Git sync is not configured. Set backup.provider = \"git\" in your config.")
			return
		}
		if err := sync.Synchronize(); err != nil {
			fmt.Printf("Warning: git sync failed: %v\n", err)
		}
		if err := sync.CommitAndPush("sync"); err != nil {
			fmt.Printf("Warning: git sync failed: %v\n", err)
		}
	default:
		fmt.Printf("Unknown action: %s\n", action)
			deleteFile(searchTerm, sync)
		},
		"list": func() {
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "Usage: margi list [collection]")
				return
			}
			listFiles(os.Args[2])
		},
		"collections": listCollections,
		"sync":        func() { runSync(sync) },
	}

	fn, ok := cmds[os.Args[1]]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown action: %s\n", os.Args[1])
		os.Exit(1)
	}
	fn()
}
