package project_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/git"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/regex"
	"github.com/wakatime/wakatime-cli/pkg/windows"

	"github.com/gandarez/go-realpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGit_Detect(t *testing.T) {
	dir := setupTestGitBasic(t)
	fp := filepath.Join(dir, "wakatime-cli/src/pkg/file.go")

	gc := git.New(fp)
	gc.GitCmd = func(args ...string) (string, error) {
		assert.Equal(t, args, []string{"-C", filepath.Dir(fp), "diff", "--numstat", fp})

		return "12       5       pkg/file.go", nil
	}

	g := project.Git{
		CountLinesChanged: true,
		Filepath:          filepath.Join(fp, "wakatime-cli/src/pkg/file.go"),
		GitClient:         gc,
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, filepath.Join(dir, "wakatime-cli"))
	assert.Equal(t, project.Result{
		Project:      "wakatime-cli",
		Branch:       "master",
		Folder:       result.Folder,
		LinesAdded:   heartbeat.PointerTo(12),
		LinesRemoved: heartbeat.PointerTo(5),
	}, result)
}

func TestGit_Detect_BranchWithSlash(t *testing.T) {
	dir := setupTestGitBasicBranchWithSlash(t)

	fp := filepath.Join(dir, "wakatime-cli/src/pkg/file.go")

	gc := git.New(fp)
	gc.GitCmd = func(args ...string) (string, error) {
		return "", nil
	}

	g := project.Git{
		Filepath:  fp,
		GitClient: gc,
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, filepath.Join(dir, "wakatime-cli"))
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "feature/detection",
		Folder:  result.Folder,
	}, result)
}

func TestGit_Detect_DetachedHead(t *testing.T) {
	dir := setupTestGitBasicDetachedHead(t)

	fp := filepath.Join(dir, "wakatime-cli/src/pkg/file.go")

	gc := git.New(fp)
	gc.GitCmd = func(args ...string) (string, error) {
		return "", nil
	}

	g := project.Git{
		Filepath:  fp,
		GitClient: gc,
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, filepath.Join(dir, "wakatime-cli"))
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "",
		Folder:  result.Folder,
	}, result)
}

func TestGit_Detect_GitConfigFile_File(t *testing.T) {
	dir := setupTestGitFile(t)

	tests := map[string]struct {
		Filepath string
		Project  string
	}{
		"main repo": {
			Filepath: filepath.Join(dir, "wakatime-cli/src/pkg/file.go"),
			Project:  "wakatime-cli",
		},
		"relative path": {
			Filepath: filepath.Join(dir, "feed/src/pkg/file.go"),
			Project:  "feed",
		},
		"absolute path": {
			Filepath: filepath.Join(dir, "mobile/src/pkg/file.go"),
			Project:  "mobile",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gc := git.New(test.Filepath)
			gc.GitCmd = func(args ...string) (string, error) {
				return "", nil
			}

			g := project.Git{
				Filepath:  test.Filepath,
				GitClient: gc,
			}

			result, detected, err := g.Detect()
			require.NoError(t, err)

			assert.True(t, detected)
			assert.Contains(t, result.Folder, filepath.Join(dir, "wakatime-cli"))
			assert.Equal(t, project.Result{
				Project: test.Project,
				Branch:  "feature/list-elements",
				Folder:  result.Folder,
			}, result)
		})
	}
}

func TestGit_Detect_Worktree(t *testing.T) {
	dir := setupTestGitWorktree(t)

	fp := filepath.Join(dir, "api/src/pkg/file.go")

	gc := git.New(fp)
	gc.GitCmd = func(args ...string) (string, error) {
		return "", nil
	}

	g := project.Git{
		Filepath:  fp,
		GitClient: gc,
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, filepath.Join(dir, "wakatime-cli"))
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "feature/api",
		Folder:  result.Folder,
	}, result)
}

func TestGit_Detect_WorktreeGitRemote(t *testing.T) {
	dir := setupTestGitWorktree(t)

	fp := filepath.Join(dir, "api/src/pkg/file.go")

	gc := git.New(fp)
	gc.GitCmd = func(args ...string) (string, error) {
		return "", nil
	}

	g := project.Git{
		Filepath:             fp,
		GitClient:            gc,
		ProjectFromGitRemote: true,
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, filepath.Join(dir, "wakatime-cli"))
	assert.Equal(t, project.Result{
		Project: "wakatime/wakatime-cli",
		Branch:  "feature/api",
		Folder:  result.Folder,
	}, result)
}

func TestGit_Detect_Submodule(t *testing.T) {
	dir := setupTestGitSubmodule(t)

	fp := filepath.Join(dir, "wakatime-cli/lib/billing/src/lib/lib.cpp")

	gc := git.New(fp)
	gc.GitCmd = func(args ...string) (string, error) {
		return "", nil
	}

	g := project.Git{
		Filepath:                  fp,
		GitClient:                 gc,
		SubmoduleDisabledPatterns: []regex.Regex{regexp.MustCompile("not_matching")},
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, filepath.Join(dir, "wakatime-cli"))
	assert.Equal(t, project.Result{
		Project: "billing",
		Branch:  "master",
		Folder:  result.Folder,
	}, result)
}

func TestGit_Detect_SubmoduleDisabled(t *testing.T) {
	dir := setupTestGitSubmodule(t)

	fp := filepath.Join(dir, "wakatime-cli/lib/billing/src/lib/lib.cpp")

	gc := git.New(fp)
	gc.GitCmd = func(args ...string) (string, error) {
		return "", nil
	}

	g := project.Git{
		Filepath:                  fp,
		GitClient:                 gc,
		SubmoduleDisabledPatterns: []regex.Regex{regexp.MustCompile(".*billing.*")},
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, filepath.Join(dir, "wakatime-cli"))
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "feature/billing",
		Folder:  result.Folder,
	}, result)
}

func TestGit_Detect_SubmoduleProjectMap_NotMatch(t *testing.T) {
	dir := setupTestGitSubmodule(t)

	fp := filepath.Join(dir, "wakatime-cli/lib/billing/src/lib/lib.cpp")

	gc := git.New(fp)
	gc.GitCmd = func(args ...string) (string, error) {
		return "", nil
	}

	g := project.Git{
		Filepath:  fp,
		GitClient: gc,
		SubmoduleProjectMapPatterns: []project.MapPattern{
			{
				Name:  "my-project-1",
				Regex: regexp.MustCompile(formatRegex("not_matching")),
			},
		},
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, filepath.Join(dir, "wakatime-cli"))
	assert.Equal(t, project.Result{
		Project: "billing",
		Branch:  "master",
		Folder:  result.Folder,
	}, result)
}

func TestGit_Detect_SubmoduleProjectMap(t *testing.T) {
	dir := setupTestGitSubmodule(t)

	fp := filepath.Join(dir, "wakatime-cli/lib/billing/src/lib/lib.cpp")

	gc := git.New(fp)
	gc.GitCmd = func(args ...string) (string, error) {
		return "", nil
	}

	g := project.Git{
		Filepath:  fp,
		GitClient: gc,
		SubmoduleProjectMapPatterns: []project.MapPattern{
			{
				Name:  "my-project-1",
				Regex: regexp.MustCompile(formatRegex(".*billing.*")),
			},
		},
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, filepath.Join(dir, "wakatime-cli"))
	assert.Equal(t, project.Result{
		Project: "my-project-1",
		Branch:  "master",
		Folder:  result.Folder,
	}, result)
}

func TestGit_Detect_SubmoduleGitRemote(t *testing.T) {
	dir := setupTestGitSubmodule(t)

	fp := filepath.Join(dir, "wakatime-cli/lib/billing/src/lib/lib.cpp")

	gc := git.New(fp)
	gc.GitCmd = func(args ...string) (string, error) {
		return "", nil
	}

	g := project.Git{
		Filepath:                  fp,
		GitClient:                 gc,
		ProjectFromGitRemote:      true,
		SubmoduleDisabledPatterns: []regex.Regex{regexp.MustCompile("not_matching")},
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, filepath.Join(dir, "wakatime-cli"))
	assert.Equal(t, project.Result{
		Project: "wakatime/billing",
		Branch:  "master",
		Folder:  result.Folder,
	}, result)
}

func setupTestGitReal(t *testing.T) string {
	tmpDir := t.TempDir()

	tmpDir, err := realpath.Realpath(tmpDir)
	require.NoError(t, err)

	if runtime.GOOS == "windows" {
		tmpDir = windows.FormatFilePath(tmpDir)
	}

	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(t, "testdata/git_real/file.go", filepath.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))

	// setup a real git repo
	err = exec.Command("git", "-C", filepath.Join(tmpDir, "wakatime-cli"), "init", "-b", "master").Run()
	require.NoError(t, err)

	return tmpDir
}

func setupTestGitBasic(t *testing.T) string {
	tmpDir := t.TempDir()

	tmpDir, err := realpath.Realpath(tmpDir)
	require.NoError(t, err)

	if runtime.GOOS == "windows" {
		tmpDir = windows.FormatFilePath(tmpDir)
	}

	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.Create(filepath.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	err = os.Mkdir(filepath.Join(tmpDir, "wakatime-cli/.git"), os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(t, "testdata/git_basic/config", filepath.Join(tmpDir, "wakatime-cli/.git/config"))
	copyFile(t, "testdata/git_basic/HEAD", filepath.Join(tmpDir, "wakatime-cli/.git/HEAD"))

	return tmpDir
}

func setupTestGitBasicBranchWithSlash(t *testing.T) string {
	tmpDir := t.TempDir()

	tmpDir, err := realpath.Realpath(tmpDir)
	require.NoError(t, err)

	if runtime.GOOS == "windows" {
		tmpDir = windows.FormatFilePath(tmpDir)
	}

	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.Create(filepath.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	err = os.Mkdir(filepath.Join(tmpDir, "wakatime-cli/.git"), os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(t, "testdata/git_basic/config", filepath.Join(tmpDir, "wakatime-cli/.git/config"))
	copyFile(t, "testdata/git_basic/HEAD_WITH_SLASH", filepath.Join(tmpDir, "wakatime-cli/.git/HEAD"))

	return tmpDir
}

func setupTestGitBasicDetachedHead(t *testing.T) string {
	tmpDir := t.TempDir()

	tmpDir, err := realpath.Realpath(tmpDir)
	require.NoError(t, err)

	if runtime.GOOS == "windows" {
		tmpDir = windows.FormatFilePath(tmpDir)
	}

	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.Create(filepath.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	err = os.Mkdir(filepath.Join(tmpDir, "wakatime-cli/.git"), os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(t, "testdata/git_basic/config", filepath.Join(tmpDir, "wakatime-cli/.git/config"))
	copyFile(t, "testdata/git_basic/HEAD_DETACHED", filepath.Join(tmpDir, "wakatime-cli/.git/HEAD"))

	return tmpDir
}

func setupTestGitFile(t *testing.T) string {
	tmpDir := t.TempDir()

	tmpDir, err := realpath.Realpath(tmpDir)
	require.NoError(t, err)

	if runtime.GOOS == "windows" {
		tmpDir = windows.FormatFilePath(tmpDir)
	}

	// Create directories
	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "feed/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "mobile/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	// Create fake files
	tmpFile, err := os.Create(filepath.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	tmpFile, err = os.Create(filepath.Join(tmpDir, "feed/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	tmpFile, err = os.Create(filepath.Join(tmpDir, "mobile/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	// Setup basic git
	err = os.Mkdir(filepath.Join(tmpDir, "wakatime-cli/.git"), os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(t, "testdata/git_basic/config", filepath.Join(tmpDir, "wakatime-cli/.git/config"))
	copyFile(t, "testdata/git_file/HEAD", filepath.Join(tmpDir, "wakatime-cli/.git/HEAD"))

	// Setup git file (relative)
	copyFile(t, "testdata/git_file/git_relative", filepath.Join(tmpDir, "feed/.git"))

	// Setup git file (absolute)
	tmpFile, err = os.Create(filepath.Join(tmpDir, "mobile/.git"))
	require.NoError(t, err)

	defer tmpFile.Close()

	gitdir := filepath.Join(tmpDir, "wakatime-cli", ".git")

	_, err = tmpFile.WriteString(fmt.Sprintf("gitdir: %s", gitdir))
	require.NoError(t, err)

	return tmpDir
}

func setupTestGitWorktree(t *testing.T) string {
	tmpDir := t.TempDir()

	tmpDir, err := realpath.Realpath(tmpDir)
	require.NoError(t, err)

	if runtime.GOOS == "windows" {
		tmpDir = windows.FormatFilePath(tmpDir)
	}

	// Create directories
	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/.git/worktrees/api"), os.FileMode(int(0700)))
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "api/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	// Create fake files
	tmpFile, err := os.Create(filepath.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	tmpFile, err = os.Create(filepath.Join(tmpDir, "api/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	// Setup basic git
	copyFile(t, "testdata/git_basic/config", filepath.Join(tmpDir, "wakatime-cli/.git/config"))
	copyFile(t, "testdata/git_worktree/HEAD", filepath.Join(tmpDir, "wakatime-cli/.git/HEAD"))

	// Setup git worktree
	copyFile(t, "testdata/git_worktree/HEAD2", filepath.Join(tmpDir, "wakatime-cli/.git/worktrees/api/HEAD"))
	copyFile(t, "testdata/git_worktree/commondir", filepath.Join(tmpDir, "wakatime-cli/.git/worktrees/api/commondir"))

	tmpFile, err = os.Create(filepath.Join(tmpDir, "api/.git"))
	require.NoError(t, err)

	defer tmpFile.Close()

	_, err = tmpFile.WriteString(fmt.Sprintf("gitdir: %s/wakatime-cli/.git/worktrees/api", tmpDir))
	require.NoError(t, err)

	return tmpDir
}

func setupTestGitSubmodule(t *testing.T) string {
	tmpDir := t.TempDir()

	tmpDir, err := realpath.Realpath(tmpDir)
	require.NoError(t, err)

	if runtime.GOOS == "windows" {
		tmpDir = windows.FormatFilePath(tmpDir)
	}

	// Create directories
	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/.git/modules/lib/billing"), os.FileMode(int(0700)))
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/lib/billing/src/lib"), os.FileMode(int(0700)))
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "billing/.git"), os.FileMode(int(0700)))
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "billing/src/lib"), os.FileMode(int(0700)))
	require.NoError(t, err)

	// Create fake files
	tmpFile, err := os.Create(filepath.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	tmpFile, err = os.Create(filepath.Join(tmpDir, "wakatime-cli/lib/billing/src/lib/lib.cpp"))
	require.NoError(t, err)

	defer tmpFile.Close()

	tmpFile, err = os.Create(filepath.Join(tmpDir, "billing/src/lib/lib.cpp"))
	require.NoError(t, err)

	defer tmpFile.Close()

	// Setup basic git
	copyFile(t, "testdata/git_basic/config", filepath.Join(tmpDir, "wakatime-cli/.git/config"))
	copyFile(t, "testdata/git_submodule/HEAD", filepath.Join(tmpDir, "wakatime-cli/.git/HEAD"))

	// Setup git submodule
	copyFile(t, "testdata/git_submodule/config", filepath.Join(tmpDir, "wakatime-cli/.git/modules/lib/billing/config"))
	copyFile(t, "testdata/git_submodule/HEAD2", filepath.Join(tmpDir, "wakatime-cli/.git/modules/lib/billing/HEAD"))
	copyFile(t, "testdata/git_basic/config", filepath.Join(tmpDir, "billing/.git/config"))
	copyFile(t, "testdata/git_submodule/HEAD2", filepath.Join(tmpDir, "billing/.git/HEAD"))
	copyFile(t, "testdata/git_submodule/git", filepath.Join(tmpDir, "wakatime-cli/lib/billing/.git"))

	return tmpDir
}
