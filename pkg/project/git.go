package project

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/yookoala/realpath"
)

// Git contains git data.
type Git struct {
	// Filepath conaints the entity path.
	Filepath string
	// SubmodulePatterns will be matched against the submodule path and if matching, will skip it.
	SubmodulePatterns []*regexp.Regexp
}

// Detect gets information about the git project for a given file.
// It tries to return a project and branch name.
func (g Git) Detect() (Result, bool, error) {
	fp, err := realpath.Realpath(g.Filepath)
	if err != nil {
		return Result{}, false,
			Err(fmt.Errorf("failed to get the real path: %w", err).Error())
	}

	// Take only the directory
	if fileExists(fp) {
		fp = path.Dir(fp)
	}

	// Find for submodule takes priority if enabled
	gitdirSubmodule, ok, err := findSubmodule(fp, g.SubmodulePatterns)
	if err != nil {
		return Result{}, false,
			Err(fmt.Sprintf("failed to validate submodule: %s", err))
	}

	result := Result{}

	if ok {
		result.Project = path.Base(gitdirSubmodule)

		branch, err := findGitBranch(path.Join(gitdirSubmodule, "HEAD"))
		if err != nil {
			jww.ERROR.Printf(
				"error finding for branch name from %q: %s",
				path.Join(path.Dir(gitdirSubmodule), "HEAD"),
				err,
			)
		}

		result.Branch = branch

		return result, true, nil
	}

	// Find for .git/config file
	gitConfigFile, ok := findGitConfigFile(fp, ".git", "config")

	if ok {
		result.Project = path.Base(path.Join(gitConfigFile, ".."))

		branch, err := findGitBranch(path.Join(gitConfigFile, "HEAD"))
		if err != nil {
			jww.ERROR.Printf(
				"error finding for branch name from %q: %s",
				path.Join(path.Dir(gitConfigFile), "HEAD"),
				err,
			)
		}

		result.Branch = branch

		return result, true, nil
	}

	// Find for .git file
	gitConfigFile, ok = findGitConfigFile(fp, "", ".git")
	if !ok {
		return Result{}, false, nil
	}

	// Find for gitdir path
	gitdir, err := findGitdir(path.Join(gitConfigFile, ".git"))
	if err != nil {
		return Result{}, false,
			Err(fmt.Sprintf("error finding gitdir: %s", err))
	}

	// Commonly .git file is present when it's a worktree
	// Find for commondir file
	commondir, ok, err := findCommondir(gitdir)
	if err != nil {
		return Result{}, false,
			Err(fmt.Sprintf("error finding commondir: %s", err))
	}

	if ok {
		result.Project = path.Base(path.Dir(commondir))

		branch, err := findGitBranch(path.Join(gitdir, "HEAD"))
		if err != nil {
			jww.ERROR.Printf(
				"error finding for branch name from %q: %s",
				path.Join(path.Dir(gitConfigFile), "HEAD"),
				err,
			)
		}

		result.Branch = branch

		return result, true, nil
	}

	if gitdir != "" {
		// Otherwise it's only a plain .git file
		result.Project = path.Base(gitConfigFile)

		branch, err := findGitBranch(path.Join(gitdir, "HEAD"))
		if err != nil {
			jww.ERROR.Printf(
				"error finding for branch name from %q: %s",
				path.Join(path.Dir(gitdir), "HEAD"),
				err,
			)
		}

		result.Branch = branch

		return result, true, nil
	}

	return Result{}, false, nil
}

func findGitConfigFile(fp string, directory string, match string) (string, bool) {
	if fileExists(path.Join(fp, directory, match)) {
		return path.Join(fp, directory), true
	}

	dir := filepath.Clean(path.Join(fp, ".."))
	if dir == "/" {
		return "", false
	}

	return findGitConfigFile(dir, directory, match)
}

func findSubmodule(fp string, patterns []*regexp.Regexp) (string, bool, error) {
	if !shouldTakeSubmodule(fp, patterns) {
		return "", false, nil
	}

	gitConfigFile, ok := findGitConfigFile(fp, "", ".git")
	if !ok {
		return "", false, nil
	}

	gitdir, err := findGitdir(path.Join(gitConfigFile, ".git"))
	if err != nil {
		return "", false,
			Err(fmt.Sprintf("error finding gitdir for submodule: %s", err))
	}

	if strings.Contains(gitdir, "modules") {
		return gitdir, true, nil
	}

	return "", false, nil
}

// shouldTakeSubmodule checks a filepath against the passed in regex patterns to determine,
// if submodule filepath should be taken.
func shouldTakeSubmodule(fp string, patterns []*regexp.Regexp) bool {
	for _, p := range patterns {
		if p.MatchString(fp) {
			return false
		}
	}

	return true
}

func findGitdir(fp string) (string, error) {
	lines, err := readFile(fp)
	if err != nil {
		return "", Err(fmt.Errorf("failed while opening file %q: %w", fp, err).Error())
	}

	if len(lines) > 0 && strings.HasPrefix(lines[0], "gitdir: ") {
		if arr := strings.Split(lines[0], "gitdir: "); len(arr) > 1 {
			return resolveGitdir(path.Join(fp, ".."), arr[1])
		}
	}

	return "", nil
}

func resolveGitdir(fp string, gitdir string) (string, error) {
	subPath := strings.TrimSpace(gitdir)
	if !filepath.IsAbs(subPath) {
		subPath = path.Join(fp, subPath)
	}

	if fileExists(path.Join(subPath, "HEAD")) {
		return subPath, nil
	}

	return "", nil
}

func findCommondir(fp string) (string, bool, error) {
	if fp == "" {
		return "", false, nil
	}

	if path.Base(path.Dir(fp)) != "worktrees" {
		return "", false, nil
	}

	if fileExists(path.Join(fp, "commondir")) {
		return resolveCommondir(fp)
	}

	return "", false, nil
}

func resolveCommondir(fp string) (string, bool, error) {
	lines, err := readFile(path.Join(fp, "commondir"))
	if err != nil {
		return "", false,
			Err(fmt.Errorf("failed while opening file %q: %w", fp, err).Error())
	}

	if len(lines) == 0 {
		return "", false, nil
	}

	gitdir, err := filepath.Abs(path.Join(fp, lines[0]))
	if err != nil {
		return "", false,
			Err(fmt.Errorf("failed to get absolute path: %w", err).Error())
	}

	if path.Base(gitdir) == ".git" {
		return gitdir, true, nil
	}

	return "", false, nil
}

func findGitBranch(fp string) (string, error) {
	if !fileExists(fp) {
		return "master", nil
	}

	lines, err := readFile(fp)
	if err != nil {
		return "", Err(fmt.Errorf("failed while opening file %q: %w", fp, err).Error())
	}

	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "ref: ") {
		return strings.TrimSpace(strings.SplitN(lines[0], "/", 3)[2]), nil
	}

	return "", nil
}

// String returns its name.
func (g Git) String() string {
	return "git-detector"
}
