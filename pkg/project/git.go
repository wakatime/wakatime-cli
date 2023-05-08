package project

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/regex"
)

// Git contains git data.
type Git struct {
	// Filepath contains the entity path.
	Filepath string
	// ProjectFromGitRemote when enabled uses the git remote as the project name instead of local git folder.
	ProjectFromGitRemote bool
	// SubmoduleDisabledPatterns will be matched against the submodule path and if matching, will skip it.
	SubmoduleDisabledPatterns []regex.Regex
	// SubmoduleProjectMapPatterns will be matched against the submodule path and if matching, will use the project map.
	SubmoduleProjectMapPatterns []MapPattern
}

// Detect gets information about the git project for a given file.
// It tries to return a project and branch name.
func (g Git) Detect() (Result, bool, error) {
	fp := g.Filepath

	// Take only the directory
	if fileExists(fp) {
		fp = filepath.Dir(fp)
	}

	// Find for submodule takes priority if enabled
	gitdirSubmodule, ok, err := findSubmodule(fp, g.SubmoduleDisabledPatterns)
	if err != nil {
		return Result{}, false, fmt.Errorf("failed to find submodule: %s", err)
	}

	if ok {
		project := projectOrRemote(filepath.Base(gitdirSubmodule), g.ProjectFromGitRemote, gitdirSubmodule)

		// If submodule has a project map, then use it.
		if result, ok := matchPattern(gitdirSubmodule, g.SubmoduleProjectMapPatterns); ok {
			project = result
		}

		branch, err := findGitBranch(filepath.Join(gitdirSubmodule, "HEAD"))
		if err != nil {
			log.Errorf(
				"error finding branch from %q: %s",
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

	// Find for .git file or directory
	dotGit, ok := FindFileOrDirectory(fp, ".git")
	if !ok {
		return Result{}, false, nil
	}

	// Find for gitdir path
	gitdir, err := findGitdir(dotGit)
	if err != nil {
		return Result{}, false, fmt.Errorf("error finding gitdir: %s", err)
	}

	// Commonly .git file is present when it's a worktree
	// Find for commondir file
	commondir, ok, err := findCommondir(gitdir)
	if err != nil {
		return Result{}, false, fmt.Errorf("error finding commondir: %s", err)
	}

	// we found a commondir file so this is a worktree
	if ok {
		project := projectOrRemote(filepath.Base(filepath.Dir(commondir)), g.ProjectFromGitRemote, commondir)

		branch, err := findGitBranch(filepath.Join(gitdir, "HEAD"))
		if err != nil {
			log.Errorf(
				"error finding branch from %q: %s",
				filepath.Join(filepath.Dir(dotGit), "HEAD"),
				err,
			)
		}

		return Result{
			Project: project,
			Branch:  branch,
			Folder:  filepath.Dir(commondir),
		}, true, nil
	}

	// Otherwise it's only a plain .git file and not a submodule
	if gitdir != "" && !strings.Contains(gitdir, "modules") {
		project := projectOrRemote(filepath.Base(filepath.Join(dotGit, "..")), g.ProjectFromGitRemote, gitdir)

		branch, err := findGitBranch(filepath.Join(gitdir, "HEAD"))
		if err != nil {
			log.Errorf(
				"error finding branch from %q: %s",
				filepath.Join(filepath.Dir(gitdir), "HEAD"),
				err,
			)
		}

		return Result{
			Project: project,
			Branch:  branch,
			Folder:  filepath.Join(gitdir, ".."),
		}, true, nil
	}

	// Find for .git/config file
	gitConfigFile, ok := FindFileOrDirectory(fp, filepath.Join(".git", "config"))

	if ok {
		gitDir := filepath.Dir(gitConfigFile)
		projectDir := filepath.Join(gitDir, "..")

		branch, err := findGitBranch(filepath.Join(gitDir, "HEAD"))
		if err != nil {
			log.Errorf(
				"error finding branch from %q: %s",
				filepath.Join(gitDir, "HEAD"),
				err,
			)
		}

		project := projectOrRemote(filepath.Base(projectDir), g.ProjectFromGitRemote, gitDir)

		return Result{
			Project: project,
			Branch:  branch,
			Folder:  projectDir,
		}, true, nil
	}

	return Result{}, false, nil
}

func findSubmodule(fp string, patterns []regex.Regex) (string, bool, error) {
	if !shouldTakeSubmodule(fp, patterns) {
		return "", false, nil
	}

	gitConfigFile, ok := FindFileOrDirectory(fp, ".git")
	if !ok {
		return "", false, nil
	}

	gitdir, err := findGitdir(gitConfigFile)
	if err != nil {
		return "", false,
			fmt.Errorf("error finding gitdir for submodule: %s", err)
	}

	if strings.Contains(gitdir, "modules") {
		return gitdir, true, nil
	}

	return "", false, nil
}

// shouldTakeSubmodule checks a filepath against the passed in regex patterns to determine,
// if submodule filepath should be taken.
func shouldTakeSubmodule(fp string, patterns []regex.Regex) bool {
	for _, p := range patterns {
		if p.MatchString(fp) {
			return false
		}
	}

	return true
}

func findGitdir(fp string) (string, error) {
	lines, err := ReadFile(fp, 1)
	if err != nil {
		return "", fmt.Errorf("failed while opening file %q: %s", fp, err)
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
	lines, err := ReadFile(filepath.Join(fp, "commondir"), 1)
	if err != nil {
		return "", false,
			fmt.Errorf("failed while opening file %q: %s", fp, err)
	}

	if len(lines) == 0 {
		return "", false, nil
	}

	gitdir, err := filepath.Abs(filepath.Join(fp, lines[0]))
	if err != nil {
		return "", false,
			fmt.Errorf("failed to get absolute path: %s", err)
	}

	if filepath.Base(gitdir) == ".git" {
		return gitdir, true, nil
	}

	return "", false, nil
}

func projectOrRemote(projectName string, projectFromGitRemote bool, dotGitFolder string) string {
	if !projectFromGitRemote {
		return projectName
	}

	configFile := filepath.Join(dotGitFolder, "config")

	remote, err := findGitRemote(configFile)
	if err != nil {
		log.Errorf("error finding git remote from %q: %s", configFile, err)

		return projectName
	}

	if remote != "" {
		return remote
	}

	return projectName
}

func findGitBranch(fp string) (string, error) {
	if !fileExists(fp) {
		return "master", nil
	}

	lines, err := ReadFile(fp, 1)
	if err != nil {
		return "", fmt.Errorf("failed while opening file %q: %s", fp, err)
	}

	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "ref: ") {
		return strings.TrimSpace(strings.SplitN(lines[0], "/", 3)[2]), nil
	}

	return "", nil
}

func findGitRemote(fp string) (string, error) {
	if !fileExists(fp) {
		return "", nil
	}

	lines, err := ReadFile(fp, 1000)
	if err != nil {
		return "", fmt.Errorf("failed while opening file %q: %s", fp, err)
	}

	for i, line := range lines {
		if strings.Trim(line, "\n\r\t") != "[remote \"origin\"]" {
			continue
		}

		if i >= len(lines) {
			continue
		}

		for _, subline := range lines[i+1:] {
			if strings.HasPrefix(subline, "[") {
				break
			}

			if strings.HasPrefix(strings.TrimSpace(subline), "url = ") {
				remote := strings.Trim(subline, "\n\r\t")

				parts := strings.SplitN(remote, "=", 2)
				if len(parts) != 2 {
					return "", fmt.Errorf("invalid origin url from %q: %s", fp, subline)
				}

				remote = parts[1]

				parts = strings.SplitN(remote, ":", 2)
				if len(parts) != 2 {
					return "", fmt.Errorf("invalid origin url from %q: %s", fp, subline)
				}

				return strings.TrimSpace(strings.TrimSuffix(parts[1], ".git")), nil
			}
		}
	}

	return "", nil
}

// ID returns its id.
func (Git) ID() DetectorID {
	return GitDetector
}
