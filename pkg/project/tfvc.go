package project

import (
	"context"
	"path/filepath"
	"runtime"
)

// Tfvc contains tfvc data.
type Tfvc struct {
	// Filepath contains the entity path.
	Filepath string
}

// Detect gets information about the tfvc project for a given file.
func (t Tfvc) Detect(ctx context.Context) (Result, bool, error) {
	var fp string

	// Take only the directory
	if fileOrDirExists(t.Filepath) {
		fp = filepath.Dir(t.Filepath)
	}

	tfFolderName := ".tf"
	if runtime.GOOS == "windows" {
		tfFolderName = "$tf"
	}

	// Find for tf/properties.tf1 file
	tfDirectory, found := FindFileOrDirectory(ctx, fp, filepath.Join(tfFolderName, "properties.tf1"))
	if !found {
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
