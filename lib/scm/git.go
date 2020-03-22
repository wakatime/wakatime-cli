package scm

import (
	"log"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/wakatime/wakatime-cli/lib/utils"
)

// Git Information about the git project for a given file.
type Git struct {
	Entity      string
	ConfigItems map[string]string
}

var (
	gitProjectName *string
	gitBranchName  *string
)

// Process Process
func (s Git) Process() bool {
	return s.findGitConfigFile()
}

func (s Git) findGitConfigFile() bool {
	absPath, err := filepath.Abs(s.Entity)
	if err == nil {
		if utils.FileExists(absPath) {
			path_, filename := path.Split(absPath)
			if utils.FileExists(path.Join(path_, ".git", "config")) {
				*gitProjectName = filepath.Base(path_)
				gitBranchName = getBranch(path.Join(path_, ".git", "HEAD"))
				return true
			}

			if linkPath := getPathFromGitdirLinkFile(path_); linkPath != nil {
				//first check if this is a worktree
				if isWortree(*linkPath) {
					gitProjectName = getProjectFromWorktree(*linkPath)
					gitBranchName = getBranch(path.Join(*linkPath, "HEAD"))
					return true
				}

				//next check if this is a submodule
				if isSubmodulesSupportedForPath(path_, s.ConfigItems) {
					*gitProjectName = filepath.Base(path_)
					gitBranchName = getBranch(path.Join(*linkPath, "HEAD"))
					return true
				}
			}

			if len(filename) == 0 {
				return false
			}
		}
	}

	return s.findGitConfigFile()
}

// ProjectName ProjectName
func (s Git) ProjectName() *string {
	return gitProjectName
}

// BranchName BranchName
func (s Git) BranchName() *string {
	return gitBranchName
}

func getBranch(headFile string) *string {
	if len(headFile) > 0 {
		if lines, err := utils.ReadFile(headFile); err == nil {
			if utils.Isset(lines, 0) {
				return getBranchFromHeadFile(lines[0])
			}
		}
	}

	branch := "master"
	return &branch
}

func getBranchFromHeadFile(line string) *string {
	if strings.HasPrefix(strings.TrimSpace(line), "ref: ") {
		arr := strings.SplitN(line, "/", 2)
		return &arr[len(arr)-1]
	}
	return nil
}

func getPathFromGitdirLinkFile(path_ string) *string {
	link := path.Join(path_, ".git")
	if !utils.FileExists(link) {
		return nil
	}

	if lines, err := utils.ReadFile(link); err == nil {
		if !utils.Isset(lines, 0) {
			return getPathfromGitdirString(path_, lines[0])
		}
	}
	return nil
}

func getPathfromGitdirString(path_ string, line string) *string {
	if strings.HasPrefix(line, "gitdir: ") {
		subpath := strings.TrimSpace(line[len("gitdir: "):])
		if utils.FileExists(path.Join(path_, subpath, "HEAD")) {
			if absPath, err := filepath.Abs(path.Join(path_, subpath)); err == nil {
				return &absPath
			}
		}
	}
	return nil
}

func isWortree(linkPath string) bool {
	dir, _ := path.Split(linkPath)
	return filepath.Base(dir) == "worktrees"
}

func getProjectFromWorktree(linkPath string) *string {
	commondir := path.Join(linkPath, "commondir")
	if utils.FileExists(commondir) {
		if lines, err := utils.ReadFile(commondir); err == nil {
			if utils.Isset(lines, 0) {
				if gitdir, err := filepath.Abs(path.Join(linkPath, lines[0])); err == nil {
					if filepath.Base(gitdir) == ".git" {
						dir, _ := path.Split(gitdir)
						base := filepath.Base(dir)
						return &base
					}
				}
			}
		}
	}
	return nil
}

func isSubmodulesSupportedForPath(path_ string, configItems map[string]string) bool {
	if rawValue, ok := configItems["submodules_disabled"]; ok {
		if disabled, err := strconv.ParseBool(rawValue); err == nil {
			return !disabled
		}

		for _, pattern := range strings.Split(rawValue, "\n") {
			if len(strings.TrimSpace(pattern)) > 0 {
				re, err := regexp.Compile("(?i)" + pattern)
				if err != nil {
					log.Printf("Regex error (%s) for disable git submodules pattern: %s", err, pattern)
					continue
				}

				if re.MatchString(path_) {
					return false
				}
			}
		}
	}
	return true
}
