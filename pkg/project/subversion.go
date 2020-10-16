package project

import (
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/yookoala/realpath"
)

// Subversion contains svn data.
type Subversion struct {
	// Filepath contains the entity path.
	Filepath string
}

// Detect gets information about the svn project for a given file.
func (s Subversion) Detect() (Result, bool, error) {
	binary, ok := findSvnBinary()
	if !ok {
		return Result{}, false, Err("svn binary not found")
	}

	fp, err := realpath.Realpath(s.Filepath)
	if err != nil {
		return Result{}, false, Err(fmt.Errorf("failed to get the real path: %w", err).Error())
	}

	// Take only the directory
	if fileExists(fp) {
		fp = path.Dir(fp)
	}

	// Find for .svn/wc.db file
	svnConfigFile, ok := findSvnConfigFile(fp, ".svn", "wc.db")
	if !ok {
		return Result{}, false, nil
	}

	info, ok, err := svnInfo(path.Join(svnConfigFile, ".."), binary)
	if err != nil {
		return Result{}, false, Err(fmt.Errorf("failed to get svn info: %w", err).Error())
	}

	if ok {
		return Result{
			Project: resolveSvnInfo(info, "Repository Root"),
			Branch:  resolveSvnInfo(info, "URL"),
			Folder:  readSvnInfo(info, "Repository Root"),
		}, true, nil
	}

	return Result{}, false, nil
}

func findSvnConfigFile(fp string, directory string, match string) (string, bool) {
	if fileExists(path.Join(fp, directory, match)) {
		return path.Join(fp, directory), true
	}

	dir := filepath.Clean(path.Join(fp, ".."))
	if dir == "/" {
		return "", false
	}

	return findSvnConfigFile(dir, directory, match)
}

func svnInfo(fp string, binary string) (map[string]string, bool, error) {
	if runtime.GOOS == "darwin" && !hasXcodeTools() {
		return nil, false, nil
	}

	cmd := exec.Command(binary, "info", fp)
	out, err := cmd.Output()

	if err != nil {
		return nil, false, Err(fmt.Sprintf("error getting svn info: %s", err))
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

func findSvnBinary() (string, bool) {
	locations := []string{
		"svn",
		"/usr/bin/svn",
		"/usr/local/bin/svn",
	}

	for _, loc := range locations {
		cmd := exec.Command(loc, "--version")

		err := cmd.Run()
		if err != nil {
			jww.ERROR.Printf("failed while calling %s --version: %s", loc, err)
			continue
		}

		return loc, true
	}

	return "", false
}

func hasXcodeTools() bool {
	cmd := exec.Command("/usr/bin/xcode-select", "-p")

	err := cmd.Run()

	return err == nil
}

func resolveSvnInfo(info map[string]string, key string) string {
	if val, ok := info[key]; ok {
		parts := strings.Split(val, "/")
		last := parts[len(parts)-1]
		parts2 := strings.Split(last, "\\")
		last2 := parts2[len(parts2)-1]

		return last2
	}

	return ""
}

func readSvnInfo(info map[string]string, key string) string {
	if val, ok := info[key]; ok {
		return val
	}

	return ""
}

// String returns its name.
func (s Subversion) String() string {
	return "svn-detector"
}
