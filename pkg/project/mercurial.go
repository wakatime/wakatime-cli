package project

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/yookoala/realpath"
)

// Mercurial contains mercurial data.
type Mercurial struct {
	// Filepath contains the entity path.
	Filepath string
}

// Detect gets information about the mercurial project for a given file.
func (m Mercurial) Detect() (Result, bool, error) {
	log.Debugln("execute mercurial project detection")

	fp, err := realpath.Realpath(m.Filepath)
	if err != nil {
		return Result{}, false,
			Err(fmt.Sprintf("failed to get the real path: %s", err))
	}

	// Take only the directory
	if fileExists(fp) {
		fp = filepath.Dir(fp)
	}

	// Find for .hg folder
	hgDirectory, ok := FindFileOrDirectory(fp, "", ".hg")
	if !ok {
		return Result{}, false, nil
	}

	project := filepath.Base(filepath.Join(hgDirectory, ".."))

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
		Folder:  filepath.Dir(filepath.Join(hgDirectory, "..")),
	}, true, nil
}

func findHgBranch(fp string) (string, error) {
	p := filepath.Join(fp, "branch")
	if !fileExists(p) {
		return "default", nil
	}

	lines, err := readFile(p)
	if err != nil {
		return "", Err(fmt.Sprintf("failed while opening file %q: %s", fp, err))
	}

	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return "default", nil
}

// String returns its name.
func (m Mercurial) String() string {
	return "hg-detector"
}
