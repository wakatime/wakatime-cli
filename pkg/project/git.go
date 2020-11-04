package project

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/yookoala/realpath"
)

// nolint
var driveLetterRegex = regexp.MustCompile(`^[a-zA-Z]:\\$`)

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
			Err(fmt.Sprintf("failed to get the real path: %s", err))
	}

	// Take only the directory
	if fileExists(fp) {
		fp = filepath.Dir(fp)
	}

	// Find for submodule takes priority if enabled
	gitdirSubmodule, ok, err := findSubmodule(fp, g.SubmodulePatterns)
	if err != nil {
		return Result{}, false,
			Err(fmt.Sprintf("failed to validate submodule: %s", err))
	}

	if ok {
		project := filepath.Base(gitdirSubmodule)

		branch, err := findGitBranch(filepath.Join(gitdirSubmodule, "HEAD"))
		if err != nil {
			jww.ERROR.Printf(
				"error finding for branch name from %q: %s",
				filepath.Join(filepath.Dir(gitdirSubmodule), "HEAD"),
				err,
			)
		}

		return Result{
			Project: project,
			Branch:  branch,
			Folder:  filepath.Dir(gitdirSubmodule),
		}, true, nil
	}

	// Find for .git/config file
	gitConfigFile, ok := findGitConfigFile(fp, ".git", "config")

	if ok {
		project := filepath.Base(filepath.Join(gitConfigFile, ".."))

		branch, err := findGitBranch(filepath.Join(gitConfigFile, "HEAD"))
		if err != nil {
			jww.ERROR.Printf(
				"error finding for branch name from %q: %s",
				filepath.Join(filepath.Dir(gitConfigFile), "HEAD"),
				err,
			)
		}

		return Result{
			Project: project,
			Branch:  branch,
			Folder:  filepath.Join(gitConfigFile, ".."),
		}, true, nil
	}

	// Find for .git file
	gitConfigFile, ok = findGitConfigFile(fp, "", ".git")
	if !ok {
		return Result{}, false, nil
	}

	// Find for gitdir path
	gitdir, err := findGitdir(filepath.Join(gitConfigFile, ".git"))
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
		project := filepath.Base(filepath.Dir(commondir))

		branch, err := findGitBranch(filepath.Join(gitdir, "HEAD"))
		if err != nil {
			jww.ERROR.Printf(
				"error finding for branch name from %q: %s",
				filepath.Join(filepath.Dir(gitConfigFile), "HEAD"),
				err,
			)
		}

		return Result{
			Project: project,
			Branch:  branch,
			Folder:  filepath.Dir(commondir),
		}, true, nil
	}

	if gitdir != "" {
		// Otherwise it's only a plain .git file
		project := filepath.Base(gitConfigFile)

		branch, err := findGitBranch(filepath.Join(gitdir, "HEAD"))
		if err != nil {
			jww.ERROR.Printf(
				"error finding for branch name from %q: %s",
				filepath.Join(filepath.Dir(gitdir), "HEAD"),
				err,
			)
		}

		return Result{
			Project: project,
			Branch:  branch,
			Folder:  gitdir,
		}, true, nil
	}

	return Result{}, false, nil
}

func findGitConfigFile(fp string, directory string, match string) (string, bool) {
	if fileExists(filepath.Join(fp, directory, match)) {
		return filepath.Join(fp, directory), true
	}

	dir := filepath.Clean(filepath.Join(fp, ".."))
	if dir == "/" || driveLetterRegex.MatchString(dir) {
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

	gitdir, err := findGitdir(filepath.Join(gitConfigFile, ".git"))
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
		return "", Err(fmt.Sprintf("failed while opening file %q: %s", fp, err))
	}

	if len(lines) > 0 && strings.HasPrefix(lines[0], "gitdir: ") {
		if arr := strings.Split(lines[0], "gitdir: "); len(arr) > 1 {
			return resolveGitdir(filepath.Join(fp, ".."), arr[1])
		}
	}

	return "", nil
}

func resolveGitdir(fp string, gitdir string) (string, error) {
	subPath := strings.TrimSpace(gitdir)
	if !filepath.IsAbs(subPath) {
		subPath = filepath.Join(fp, subPath)
	}

	if fileExists(filepath.Join(subPath, "HEAD")) {
		return subPath, nil
	}

	return "", nil
}

func findCommondir(fp string) (string, bool, error) {
	if fp == "" {
		return "", false, nil
	}

	if filepath.Base(filepath.Dir(fp)) != "worktrees" {
		return "", false, nil
	}

	if fileExists(filepath.Join(fp, "commondir")) {
		return resolveCommondir(fp)
	}

	return "", false, nil
}

func resolveCommondir(fp string) (string, bool, error) {
	lines, err := readFile(filepath.Join(fp, "commondir"))
	if err != nil {
		return "", false,
			Err(fmt.Sprintf("failed while opening file %q: %s", fp, err))
	}

	if len(lines) == 0 {
		return "", false, nil
	}

	gitdir, err := filepath.Abs(filepath.Join(fp, lines[0]))
	if err != nil {
		return "", false,
			Err(fmt.Sprintf("failed to get absolute path: %s", err))
	}

	if filepath.Base(gitdir) == ".git" {
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
		return "", Err(fmt.Sprintf("failed while opening file %q: %s", fp, err))
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
