package project

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

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
	fp, ok, err := findProjectFile(f.Filepath)
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

// findProjectFile find for .wakatime-project file in the given path.
func findProjectFile(fp string) (string, bool, error) {
	fp, err := realpath.Realpath(fp)
	if err != nil {
		return "", false, Err(fmt.Errorf("failed to get the real path: %w", err).Error())
	}

	dir, _ := path.Split(fp)
	if fileExists(path.Join(dir, defaultProjectFile)) {
		fp = path.Join(dir, defaultProjectFile)
		return fp, true, nil
	}

	dir, file := path.Split(fp)
	if len(file) == 0 {
		return "", false, nil
	}

	return findProjectFile(dir)
}

// fileExists checks if a file exist and is not a directory.
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
