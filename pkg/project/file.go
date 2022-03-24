package project

import (
	"bufio"
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
	log.Debugln("execute file project detection")

	fp, ok := FindFileOrDirectory(f.Filepath, WakaTimeProjectFile)
	if !ok {
		return Result{}, false, nil
	}

	log.Debugf("wakatime project file found at: %s", fp)

	lines, err := readFile(fp, 2)
	if err != nil {
		return Result{}, false, Err(fmt.Sprintf("error reading file: %s", err))
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

// readFile reads a file until max number of lines and return an array of lines.
func readFile(fp string, max int) ([]string, error) {
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

// String returns its name.
func (File) String() string {
	return "project-file-detector"
}
