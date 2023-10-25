package project

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/log"
)

// File contains file data.
type File struct {
	Filepath string
}

// Detect get information from a .wakatime-project file about the project for
// a given file. First line of .wakatime-project sets the project
// name. Second line sets the current branch name.
func (f File) Detect() (Result, bool, error) {
	fp, ok := FindFileOrDirectory(f.Filepath, WakaTimeProjectFile)
	if !ok {
		return Result{}, false, nil
	}

	log.Debugf("wakatime project file found at: %s", fp)

	lines, err := ReadFile(fp, 2)
	if err != nil {
		return Result{}, false, fmt.Errorf("error reading file: %s", err)
	}

	result := Result{
		Folder: filepath.Dir(fp),
	}

	if len(lines) > 0 {
		result.Project = strings.TrimSpace(lines[0])
	}

	if len(lines) > 1 {
		result.Branch = strings.TrimSpace(lines[1])
	}

	return result, true, nil
}

// fileExists checks if a file or directory exist.
func fileExists(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}

// ReadFile reads a file until max number of lines and return an array of lines.
func ReadFile(fp string, max int) ([]string, error) {
	if fp == "" {
		return nil, errors.New("filepath cannot be empty")
	}

	file, err := os.Open(fp) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed while opening file %q: %s", fp, err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Debugf("failed to close file '%s': %s", file.Name(), err)
		}
	}()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var (
		lines []string
		i     = 0
	)

	for scanner.Scan() {
		i++

		if i > max {
			break
		}

		lines = append(lines, scanner.Text())
	}

	return lines, nil
}

// ID returns its id.
func (File) ID() DetectorID {
	return FileDetector
}
