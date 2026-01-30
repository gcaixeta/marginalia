package snippet

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

func Default(title, collection string) string {
	return fmt.Sprintf(`# %s

		## Created: %s
		## Collection: %s

		`, title, time.Now().Format(time.RFC3339), collection)

}

func ReadSnippet(collection string) (snippet string, err error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Error while getting default user config dir: %v", err)
	}

	collectionsDir := path.Join(configDir, "marginalia", "collections")

	return os.ReadFile(collectionsDir + collection + ".md")
}
