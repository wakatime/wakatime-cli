package project

import (
	"path/filepath"
	"runtime"
)

// Tfvc contains tfvc data.
type Tfvc struct {
	// Filepath contains the entity path.
	Filepath string
}

// Detect gets information about the tfvc project for a given file.
func (t Tfvc) Detect() (Result, bool, error) {
	var fp string

	// Take only the directory
	if fileExists(t.Filepath) {
		fp = filepath.Dir(t.Filepath)
	}

	tfFolderName := ".tf"
	if runtime.GOOS == "windows" {
		tfFolderName = "$tf"
	}

	// Find for tf/properties.tf1 file
	tfDirectory, ok := FindFileOrDirectory(fp, filepath.Join(tfFolderName, "properties.tf1"))
	if !ok {
		return Result{}, false, nil
	}

	project := filepath.Base(filepath.Join(tfDirectory, "..", ".."))

	return Result{
		Project: project,
		Folder:  filepath.Dir(filepath.Join(tfDirectory, "..", "..")),
	}, true, nil
}

// ID returns its id.
func (Tfvc) ID() DetectorID {
	return TfvcDetector
}
