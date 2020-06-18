package project_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGit_Detect_GitConfigFile_Directory(t *testing.T) {
	fp, tearDown := setupTestGitBasic(t)
	defer tearDown()

	g := project.Git{
		Filepath: path.Join(fp, "wakatime-cli/src/pkg/file.go"),
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "feature/detection",
	}, result)
}

func TestGit_Detect_GitConfigFile_File(t *testing.T) {
	fp, tearDown := setupTestGitFile(t)
	defer tearDown()

	tests := map[string]struct {
		Filepath string
	}{
		"main_repo": {
			Filepath: path.Join(fp, "wakatime-cli/src/pkg/file.go"),
		},
		"relative_path": {
			Filepath: path.Join(fp, "feed/src/pkg/file.go"),
		},
		"absolute_pasth": {
			Filepath: path.Join(fp, "mobile/src/pkg/file.go"),
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
			assert.Equal(t, project.Result{
				Project: "wakatime-cli",
				Branch:  "feature/list-elements",
			}, result)
		})
	}
}

func TestGit_Detect_Worktree(t *testing.T) {
	fp, tearDown := setupTestGiWorktree(t)
	defer tearDown()

	g := project.Git{
		Filepath: path.Join(fp, "api/src/pkg/file.go"),
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "feature/api",
	}, result)
}

func TestGit_Detect_Submodule(t *testing.T) {
	fp, tearDown := setupTestGiSubmodule(t)
	defer tearDown()

	g := project.Git{
		Filepath:          path.Join(fp, "wakatime-cli/lib/billing/src/lib/lib.cpp"),
		SubmodulePatterns: []*regexp.Regexp{regexp.MustCompile("not_matching")},
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Equal(t, project.Result{
		Project: "billing",
		Branch:  "master",
	}, result)
}

func TestGit_Detect_SubmoduleDisabled(t *testing.T) {
	fp, tearDown := setupTestGiSubmodule(t)
	defer tearDown()

	g := project.Git{
		Filepath:          path.Join(fp, "wakatime-cli/lib/billing/src/lib/lib.cpp"),
		SubmodulePatterns: []*regexp.Regexp{regexp.MustCompile(".*billing.*")},
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "feature/billing",
	}, result)
}

func setupTestGitBasic(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-git")
	require.NoError(t, err)

	err = os.MkdirAll(path.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.Create(path.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	tmpFile.Close()

	err = os.Mkdir(path.Join(tmpDir, "wakatime-cli/.git"), os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(t, "testdata/git_basic/config", path.Join(tmpDir, "wakatime-cli/.git/config"))
	copyFile(t, "testdata/git_basic/HEAD", path.Join(tmpDir, "wakatime-cli/.git/HEAD"))

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}

func setupTestGitFile(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-git")
	require.NoError(t, err)

	// Create directories
	err = os.MkdirAll(path.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	err = os.MkdirAll(path.Join(tmpDir, "feed/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	err = os.MkdirAll(path.Join(tmpDir, "mobile/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	// Create fake files
	tmpFile, err := os.Create(path.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	tmpFile.Close()

	tmpFile, err = os.Create(path.Join(tmpDir, "feed/src/pkg/file.go"))
	require.NoError(t, err)

	tmpFile.Close()

	tmpFile, err = os.Create(path.Join(tmpDir, "mobile/src/pkg/file.go"))
	require.NoError(t, err)

	tmpFile.Close()

	// Setup basic git
	err = os.Mkdir(path.Join(tmpDir, "wakatime-cli/.git"), os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(t, "testdata/git_basic/config", path.Join(tmpDir, "wakatime-cli/.git/config"))
	copyFile(t, "testdata/git_file/HEAD", path.Join(tmpDir, "wakatime-cli/.git/HEAD"))

	// Setup git file (relative)
	copyFile(t, "testdata/git_file/git_relative", path.Join(tmpDir, "feed/.git"))

	//Setup git file (absolute)
	tmpFile, err = os.Create(path.Join(tmpDir, "mobile/.git"))
	require.NoError(t, err)

	defer tmpFile.Close()

	_, err = tmpFile.WriteString(fmt.Sprintf("gitdir: %s/wakatime-cli/.git", tmpDir))
	require.NoError(t, err)

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}

func setupTestGiWorktree(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-git")
	require.NoError(t, err)

	// Create directories
	err = os.MkdirAll(path.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	err = os.MkdirAll(path.Join(tmpDir, "wakatime-cli/.git/worktrees/api"), os.FileMode(int(0700)))
	require.NoError(t, err)

	err = os.MkdirAll(path.Join(tmpDir, "api/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	// Create fake files
	tmpFile, err := os.Create(path.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	tmpFile.Close()

	tmpFile, err = os.Create(path.Join(tmpDir, "api/src/pkg/file.go"))
	require.NoError(t, err)

	tmpFile.Close()

	// Setup basic git
	copyFile(t, "testdata/git_basic/config", path.Join(tmpDir, "wakatime-cli/.git/config"))
	copyFile(t, "testdata/git_worktree/HEAD", path.Join(tmpDir, "wakatime-cli/.git/HEAD"))

	// Setup git worktree
	copyFile(t, "testdata/git_worktree/HEAD2", path.Join(tmpDir, "wakatime-cli/.git/worktrees/api/HEAD"))
	copyFile(t, "testdata/git_worktree/commondir", path.Join(tmpDir, "wakatime-cli/.git/worktrees/api/commondir"))

	tmpFile, err = os.Create(path.Join(tmpDir, "api/.git"))
	require.NoError(t, err)

	defer tmpFile.Close()

	_, err = tmpFile.WriteString(fmt.Sprintf("gitdir: %s/wakatime-cli/.git/worktrees/api", tmpDir))
	require.NoError(t, err)

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}

func setupTestGiSubmodule(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-git")
	require.NoError(t, err)

	// Create directories
	err = os.MkdirAll(path.Join(tmpDir, "wakatime-cli/.git/modules/lib/billing"), os.FileMode(int(0700)))
	require.NoError(t, err)

	err = os.MkdirAll(path.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	err = os.MkdirAll(path.Join(tmpDir, "wakatime-cli/lib/billing/src/lib"), os.FileMode(int(0700)))
	require.NoError(t, err)

	err = os.MkdirAll(path.Join(tmpDir, "billing/.git"), os.FileMode(int(0700)))
	require.NoError(t, err)

	err = os.MkdirAll(path.Join(tmpDir, "billing/src/lib"), os.FileMode(int(0700)))
	require.NoError(t, err)

	// Create fake files
	tmpFile, err := os.Create(path.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	tmpFile.Close()

	tmpFile, err = os.Create(path.Join(tmpDir, "wakatime-cli/lib/billing/src/lib/lib.cpp"))
	require.NoError(t, err)

	tmpFile.Close()

	tmpFile, err = os.Create(path.Join(tmpDir, "billing/src/lib/lib.cpp"))
	require.NoError(t, err)

	tmpFile.Close()

	// Setup basic git
	copyFile(t, "testdata/git_basic/config", path.Join(tmpDir, "wakatime-cli/.git/config"))
	copyFile(t, "testdata/git_submodule/HEAD", path.Join(tmpDir, "wakatime-cli/.git/HEAD"))

	// Setup git submodule
	copyFile(t, "testdata/git_basic/config", path.Join(tmpDir, "wakatime-cli/.git/modules/lib/billing/config"))
	copyFile(t, "testdata/git_submodule/HEAD2", path.Join(tmpDir, "wakatime-cli/.git/modules/lib/billing/HEAD"))
	copyFile(t, "testdata/git_basic/config", path.Join(tmpDir, "billing/.git/config"))
	copyFile(t, "testdata/git_submodule/HEAD2", path.Join(tmpDir, "billing/.git/HEAD"))
	copyFile(t, "testdata/git_submodule/git", path.Join(tmpDir, "wakatime-cli/lib/billing/.git"))

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}
