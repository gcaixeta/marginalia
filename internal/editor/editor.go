package editor

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func OpenInEditor(filePath string) {
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
