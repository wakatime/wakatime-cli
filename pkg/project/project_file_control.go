package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// Write saves wakatime project file.
func Write(folder, project string) error {
	err := os.WriteFile(filepath.Join(folder, defaultProjectFile), []byte(project+"\n"), 0600)
	if err != nil {
		return fmt.Errorf("failed to save wakatime project file: %s", err)
	}

	return nil
}
