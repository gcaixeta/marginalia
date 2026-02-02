package snippet

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

func Default(title, collection string) string {
	return fmt.Sprintf(`# %s

## Created: %s
## Collection: %s

`, title, time.Now().Format(time.RFC3339), collection)
}

func ReadSnippet(title, collection string) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Error while getting default user config dir: %v", err)
	}

	snippetPath := filepath.Join(configDir, "marginalia", "collections", collection+".md")

	content, err := os.ReadFile(snippetPath)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("snippet").Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("error parsing snippet template: %w", err)
	}

	data := struct {
		Title      string
		Collection string
		Date       string
	}{
		Title:      title,
		Collection: collection,
		Date:       time.Now().Format(time.RFC3339),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing snippet template: %w", err)
	}

	return buf.String(), nil
}
