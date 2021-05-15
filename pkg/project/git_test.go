package project_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/regex"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGit_Detect(t *testing.T) {
	fp, tearDown := setupTestGitBasic(t)
	defer tearDown()

	g := project.Git{
		Filepath: filepath.Join(fp, "wakatime-cli/src/pkg/file.go"),
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, filepath.Join(fp, "wakatime-cli"))
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "master",
		Folder:  result.Folder,
	}, result)
}

func TestGit_Detect_BranchWithSlash(t *testing.T) {
	fp, tearDown := setupTestGitBasicBranchWithSlash(t)
	defer tearDown()

	g := project.Git{
		Filepath: filepath.Join(fp, "wakatime-cli/src/pkg/file.go"),
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, filepath.Join(fp, "wakatime-cli"))
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "feature/detection",
		Folder:  result.Folder,
	}, result)
}

func TestGit_Detect_DetachedHead(t *testing.T) {
	fp, tearDown := setupTestGitBasicDetachedHead(t)
	defer tearDown()

	g := project.Git{
		Filepath: filepath.Join(fp, "wakatime-cli/src/pkg/file.go"),
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, filepath.Join(fp, "wakatime-cli"))
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "",
		Folder:  result.Folder,
	}, result)
}

func TestGit_Detect_GitConfigFile_File(t *testing.T) {
	fp, tearDown := setupTestGitFile(t)
	defer tearDown()

	tests := map[string]struct {
		Filepath string
		Project  string
	}{
		"main_repo": {
			Filepath: filepath.Join(fp, "wakatime-cli/src/pkg/file.go"),
			Project:  "wakatime-cli",
		},
		"relative_path": {
			Filepath: filepath.Join(fp, "feed/src/pkg/file.go"),
			Project:  "feed",
		},
		"absolute_path": {
			Filepath: filepath.Join(fp, "mobile/src/pkg/file.go"),
			Project:  "mobile",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			g := project.Git{
				Filepath: test.Filepath,
			}

			result, detected, err := g.Detect()
			require.NoError(t, err)

			assert.True(t, detected)
			assert.Contains(t, result.Folder, filepath.Join(fp, "wakatime-cli"))
			assert.Equal(t, project.Result{
				Project: test.Project,
				Branch:  "feature/list-elements",
				Folder:  result.Folder,
			}, result)
		})
	}
}

func TestGit_Detect_Worktree(t *testing.T) {
	fp, tearDown := setupTestGitWorktree(t)
	defer tearDown()

	g := project.Git{
		Filepath: filepath.Join(fp, "api/src/pkg/file.go"),
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, filepath.Join(fp, "wakatime-cli"))
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "feature/api",
		Folder:  result.Folder,
	}, result)
}

func TestGit_Detect_Submodule(t *testing.T) {
	fp, tearDown := setupTestGitSubmodule(t)
	defer tearDown()

	g := project.Git{
		Filepath:          filepath.Join(fp, "wakatime-cli/lib/billing/src/lib/lib.cpp"),
		SubmodulePatterns: []regex.Regex{regexp.MustCompile("not_matching")},
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, filepath.Join(fp, "wakatime-cli"))
	assert.Equal(t, project.Result{
		Project: "billing",
		Branch:  "master",
		Folder:  result.Folder,
	}, result)
}

func TestGit_Detect_SubmoduleDisabled(t *testing.T) {
	fp, tearDown := setupTestGitSubmodule(t)
	defer tearDown()

	g := project.Git{
		Filepath:          filepath.Join(fp, "wakatime-cli/lib/billing/src/lib/lib.cpp"),
		SubmodulePatterns: []regex.Regex{regexp.MustCompile(".*billing.*")},
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, filepath.Join(fp, "wakatime-cli"))
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "feature/billing",
		Folder:  result.Folder,
	}, result)
}

func setupTestGitBasic(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-git")
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.Create(filepath.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	err = os.Mkdir(filepath.Join(tmpDir, "wakatime-cli/.git"), os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(t, "testdata/git_basic/config", filepath.Join(tmpDir, "wakatime-cli/.git/config"))
	copyFile(t, "testdata/git_basic/HEAD", filepath.Join(tmpDir, "wakatime-cli/.git/HEAD"))

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}

func setupTestGitBasicBranchWithSlash(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-git")
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.Create(filepath.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	err = os.Mkdir(filepath.Join(tmpDir, "wakatime-cli/.git"), os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(t, "testdata/git_basic/config", filepath.Join(tmpDir, "wakatime-cli/.git/config"))
	copyFile(t, "testdata/git_basic/HEAD_WITH_SLASH", filepath.Join(tmpDir, "wakatime-cli/.git/HEAD"))

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}

func setupTestGitBasicDetachedHead(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-git")
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.Create(filepath.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	err = os.Mkdir(filepath.Join(tmpDir, "wakatime-cli/.git"), os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(t, "testdata/git_basic/config", filepath.Join(tmpDir, "wakatime-cli/.git/config"))
	copyFile(t, "testdata/git_basic/HEAD_DETACHED", filepath.Join(tmpDir, "wakatime-cli/.git/HEAD"))

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}

func setupTestGitFile(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-git")
	require.NoError(t, err)

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

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}

func setupTestGitWorktree(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-git")
	require.NoError(t, err)

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

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}

func setupTestGitSubmodule(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-git")
	require.NoError(t, err)

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
	copyFile(t, "testdata/git_basic/config", filepath.Join(tmpDir, "wakatime-cli/.git/modules/lib/billing/config"))
	copyFile(t, "testdata/git_submodule/HEAD2", filepath.Join(tmpDir, "wakatime-cli/.git/modules/lib/billing/HEAD"))
	copyFile(t, "testdata/git_basic/config", filepath.Join(tmpDir, "billing/.git/config"))
	copyFile(t, "testdata/git_submodule/HEAD2", filepath.Join(tmpDir, "billing/.git/HEAD"))
	copyFile(t, "testdata/git_submodule/git", filepath.Join(tmpDir, "wakatime-cli/lib/billing/.git"))

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}
