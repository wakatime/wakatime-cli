package project_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMercurial_Detect(t *testing.T) {
	fp, tearDown := setupTestMercurial(t)
	defer tearDown()

	m := project.Mercurial{
		Filepath: path.Join(fp, "wakatime-cli/src/pkg/file.go"),
	}

	result, detected, err := m.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "billing",
	}, result)
}

func TestMercurial_Detect_BranchWithSlash(t *testing.T) {
	fp, tearDown := setupTestMercurialBranchWithSlash(t)
	defer tearDown()

	m := project.Mercurial{
		Filepath: path.Join(fp, "wakatime-cli/src/pkg/file.go"),
	}

	result, detected, err := m.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "feature/billing",
	}, result)
}

func TestMercurial_Detect_NoBranch(t *testing.T) {
	fp, tearDown := setupTestMercurialNoBranch(t)
	defer tearDown()

	m := project.Mercurial{
		Filepath: path.Join(fp, "wakatime-cli/src/pkg/file.go"),
	}

	result, detected, err := m.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "default",
	}, result)
}

func setupTestMercurial(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-hg")
	require.NoError(t, err)

	err = os.MkdirAll(path.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.Create(path.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	tmpFile.Close()

	err = os.Mkdir(path.Join(tmpDir, "wakatime-cli/.hg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(t, "testdata/hg/branch", path.Join(tmpDir, "wakatime-cli/.hg/branch"))

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}

func setupTestMercurialBranchWithSlash(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-hg")
	require.NoError(t, err)

	err = os.MkdirAll(path.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.Create(path.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	tmpFile.Close()

	err = os.Mkdir(path.Join(tmpDir, "wakatime-cli/.hg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(t, "testdata/hg/branch_with_slash", path.Join(tmpDir, "wakatime-cli/.hg/branch"))

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}

func setupTestMercurialNoBranch(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-hg")
	require.NoError(t, err)

	err = os.MkdirAll(path.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.Create(path.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	tmpFile.Close()

	err = os.Mkdir(path.Join(tmpDir, "wakatime-cli/.hg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}
