package project

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type svnCommands struct {
	version         func(string) error
	info            func(string, string) ([]byte, error)
	xcodeToolsExist func() bool
}

// Subversion contains svn data.
type Subversion struct {
	// Filepath contains the entity path.
	Filepath string
}

// Detect gets information about the svn project for a given file.
func (s Subversion) Detect(ctx context.Context) (Result, bool, error) {
	return s.detect(ctx, svnCommands{
		version:         svnVersion,
		info:            svnInfoOutput,
		xcodeToolsExist: hasXcodeTools,
	})
}

func (s Subversion) detect(ctx context.Context, commands svnCommands) (Result, bool, error) {
	binary, ok := findSvnBinary(commands)
	if !ok {
		return Result{}, false, nil
	}

	var fp string

	// Take only the directory
	if fileOrDirExists(s.Filepath) {
		fp = filepath.Dir(s.Filepath)
	}

	// Find for .svn/wc.db file
	svnConfigFile, found := FindFileOrDirectory(ctx, fp, filepath.Join(".svn", "wc.db"))
	if !found {
		return Result{}, false, nil
	}

	info, ok, err := svnInfo(filepath.Join(svnConfigFile, "..", ".."), binary, commands)
	if err != nil {
		return Result{}, false, fmt.Errorf("failed to get svn info: %s", err)
	}

	if !ok {
		return Result{}, false, nil
	}

	return Result{
		Project: resolveSvnInfo(info, "Repository Root"),
		Branch:  resolveSvnInfo(info, "URL"),
		Folder:  strings.ReplaceAll(info["Repository Root"], "\r", ""),
	}, true, nil
}

func svnVersion(loc string) error {
	return exec.Command(loc, "--version").Run() //nolint:gosec
}

func svnInfoOutput(binary, fp string) ([]byte, error) {
	return exec.Command(binary, "info", fp).Output() //nolint:gosec
}

func svnInfo(fp string, binary string, commands svnCommands) (map[string]string, bool, error) {
	if runtime.GOOS == "darwin" && !commands.xcodeToolsExist() {
		return nil, false, nil
	}

	out, err := commands.info(binary, fp)
	if err != nil {
		return nil, false, fmt.Errorf("error getting svn info: %s", err)
	}

	result := map[string]string{}

	for _, line := range strings.Split(string(out), "\n") {
		item := strings.Split(line, ": ")
		if len(item) == 2 {
			result[item[0]] = item[1]
		}
	}

	return result, true, nil
}

func findSvnBinary(commands svnCommands) (string, bool) {
	locations := []string{
		"svn",
		"/usr/bin/svn",
		"/usr/local/bin/svn",
	}

	for _, loc := range locations {
		if err := commands.version(loc); err != nil {
			continue
		}

		return loc, true
	}

	return "", false
}

func hasXcodeTools() bool {
	cmd := exec.Command("/usr/bin/xcode-select", "-p")

	return cmd.Run() == nil
}

func resolveSvnInfo(info map[string]string, key string) string {
	if val, ok := info[key]; ok {
		parts := strings.Split(val, "/")
		last := parts[len(parts)-1]
		parts2 := strings.Split(last, "\\")
		last2 := parts2[len(parts2)-1]

		return strings.ReplaceAll(last2, "\r", "")
	}

	return ""
}

// ID returns its id.
func (Subversion) ID() DetectorID {
	return SubversionDetector
}
