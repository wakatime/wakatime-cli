package project

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/yookoala/realpath"
)

const defaultProjectFile = ".wakatime-project"

// File contains file data.
type File struct {
	Filepath string
}

// Detect get information from a .wakatime-project file about the project for
// a given file. First line of .wakatime-project sets the project
// name. Second line sets the current branch name.
func (f File) Detect() (Result, bool, error) {
	fp, ok, err := FindFile(f.Filepath)
	if err != nil {
		return Result{}, false,
			Err(fmt.Sprintf("error finding project file: %s", err))
	} else if !ok {
		return Result{}, false, nil
	}

	lines, err := readFile(fp)
	if err != nil {
		return Result{}, false,
			Err(fmt.Sprintf("error reading file: %s", err))
	}

	result := Result{}

	if len(lines) > 0 {
		result.Project = strings.TrimSpace(lines[0])
	}

	if len(lines) > 1 {
		result.Branch = strings.TrimSpace(lines[1])
	}

	return result, true, nil
}

// FindFile find for .wakatime-project file in the given path.
func FindFile(fp string) (string, bool, error) {
	i := 0
	for i < maxRecursiveIteration {
		resolved, err := realpath.Realpath(fp)
		if err != nil {
			return "", false, Err(fmt.Sprintf("failed to get the real path for file %q: %s", fp, err))
		}

		fp = resolved

		dir, _ := filepath.Split(fp)
		if fileExists(filepath.Join(dir, defaultProjectFile)) {
			fp = filepath.Join(dir, defaultProjectFile)
			return fp, true, nil
		}

		dir, file := filepath.Split(fp)
		if len(file) == 0 {
			return "", false, nil
		}

		fp = dir
		i++
	}

	log.Warnf("didn't find .wakatime-project file after %d iterations", maxRecursiveIteration)

	return "", false, nil
}

// fileExists checks if a file or directory exist.
func fileExists(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}

// readFile reads a file and return an array of lines.
func readFile(fp string) ([]string, error) {
	if fp == "" {
		return nil, Err("filepath cannot be empty")
	}

	file, err := os.Open(fp)
	if err != nil {
		return nil, Err(fmt.Errorf("failed while opening file %q: %w", fp, err).Error())
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, nil
}

// String returns its name.
func (f File) String() string {
	return "project-file-detector"
}
