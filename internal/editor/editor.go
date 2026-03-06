package editor

import (
	"log"
	"os"
	"os/exec"
)

// ResolveEditor returns the editor to use: config value → $VISUAL → $EDITOR → "vi"
func ResolveEditor(configured string) string {
	if configured != "" {
		return configured
	}
	if v := os.Getenv("VISUAL"); v != "" {
		return v
	}
	if v := os.Getenv("EDITOR"); v != "" {
		return v
	}
	return "vi"
}

func OpenInEditor(filePath, editorCmd string) {
	cmd := exec.Command(editorCmd, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error running %s: %v\n", editorCmd, err)
	}
}
