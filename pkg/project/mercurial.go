package project

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/log"
)

// Mercurial contains mercurial data.
type Mercurial struct {
	// Filepath contains the entity path.
	Filepath string
}

// Detect gets information about the mercurial project for a given file.
func (m Mercurial) Detect() (Result, bool, error) {
	var fp string

	// Take only the directory
	if fileExists(m.Filepath) {
		fp = filepath.Dir(m.Filepath)
	}

	// Find for .hg folder
	hgDirectory, ok := FindFileOrDirectory(fp, ".hg")
	if !ok {
		return Result{}, false, nil
	}

	project := filepath.Base(filepath.Dir(hgDirectory))

	branch, err := findHgBranch(hgDirectory)
	if err != nil {
		log.Errorf(
			"error finding for branch name from %q: %s",
			hgDirectory,
			err,
		)
	}

	return Result{
		Project: project,
		Branch:  branch,
		Folder:  filepath.Dir(filepath.Dir(hgDirectory)),
	}, true, nil
}

func findHgBranch(fp string) (string, error) {
	p := filepath.Join(fp, "branch")
	if !fileExists(p) {
		return "default", nil
	}

	lines, err := ReadFile(p, 1)
	if err != nil {
		return "", fmt.Errorf("failed while opening file %q: %s", fp, err)
	}

	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return "default", nil
}

// ID returns its id.
func (Mercurial) ID() DetectorID {
	return MercurialDetector
}
