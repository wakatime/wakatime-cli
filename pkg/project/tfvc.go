package project

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/yookoala/realpath"
)

// Tfvc contains tfvc data.
type Tfvc struct {
	// Filepath contains the entity path.
	Filepath string
}

// Detect gets information about the tfvc project for a given file.
func (t Tfvc) Detect() (Result, bool, error) {
	fp, err := realpath.Realpath(t.Filepath)
	if err != nil {
		return Result{}, false,
			Err(fmt.Errorf("failed to get the real path: %w", err).Error())
	}

	// Take only the directory
	if fileExists(fp) {
		fp = filepath.Dir(fp)
	}

	tfFolderName := ".tf"
	if runtime.GOOS == "windows" {
		tfFolderName = "$tf"
	}

	// Find for tf/properties.tf1 file
	tfDirectory, ok := findFileOrDirectory(fp, tfFolderName, "properties.tf1")
	if !ok {
		return Result{}, false, nil
	}

	project := filepath.Base(filepath.Join(tfDirectory, "../.."))

	return Result{
		Project: project,
		Folder:  filepath.Dir(filepath.Join(tfDirectory, "../..")),
	}, true, nil
}

// String returns its name.
func (t Tfvc) String() string {
	return "tfvc-detector"
}
