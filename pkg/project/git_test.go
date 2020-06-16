package project_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGit_Detect_GitConfigFile_Directory(t *testing.T) {
	fp, tearDown := setupTestGitFolder(t, "basic")
	defer tearDown()

	g := project.Git{
		Filepath: path.Join(fp, "wakatime-cli/src/pkg/project.go"),
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
	fp, tearDown := setupTestGitFolder(t, "git_file")
	defer tearDown()

	tests := map[string]struct {
		Filepath string
	}{
		"relative_path": {
			Filepath: path.Join(fp, "otherproject/src/pkg/project.go"),
		},
		"absolute_pasth": {
			Filepath: path.Join(fp, "someproject/src/pkg/project.go"),
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
				Branch:  "feature/detection",
			}, result)
		})
	}
}

func TestGit_Detect_Worktree(t *testing.T) {
	fp, tearDown := setupTestGitFolder(t, "worktree")
	defer tearDown()

	g := project.Git{
		Filepath: path.Join(fp, "project_api/src/pkg/project.go"),
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
	fp, tearDown := setupTestGitFolder(t, "submodule")
	defer tearDown()

	g := project.Git{
		Filepath:          path.Join(fp, "wakatime-cli/lib/module_a/src/lib/lib.cpp"),
		SubmodulePatterns: []*regexp.Regexp{regexp.MustCompile("not_matching")},
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Equal(t, project.Result{
		Project: "module_a",
		Branch:  "master",
	}, result)
}

func TestGit_Detect_SubmoduleDisabled(t *testing.T) {
	fp, tearDown := setupTestGitFolder(t, "submodule")
	defer tearDown()

	g := project.Git{
		Filepath:          path.Join(fp, "wakatime-cli/lib/module_a/src/lib/lib.cpp"),
		SubmodulePatterns: []*regexp.Regexp{regexp.MustCompile(".*module_a.*")},
	}

	result, detected, err := g.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "bugfix/log",
	}, result)
}

func setupTestGitFolder(t *testing.T, args ...string) (fp string, tearDown func()) {
	dir, err := ioutil.TempDir(os.TempDir(), "wakatime-git")
	require.NoError(t, err)

	args = append([]string{dir}, args...)
	cmd := exec.Command("testdata/setup_git.sh", args...)

	err = cmd.Run()
	require.NoError(t, err)

	out, err := exec.Command("ls", "-R", dir).CombinedOutput()
	require.NoError(t, err)

	fmt.Println(string(out))

	return dir, func() { os.RemoveAll(dir) }
}
