package project_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMercurial_Detect(t *testing.T) {
	fp, tearDown := setupTestMercurial(t)
	defer tearDown()

	m := project.Mercurial{
		Filepath: filepath.Join(fp, "wakatime-cli/src/pkg/file.go"),
	}

	result, detected, err := m.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, fp)
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "billing",
		Folder:  result.Folder,
	}, result)
}

func TestMercurial_Detect_BranchWithSlash(t *testing.T) {
	fp, tearDown := setupTestMercurialBranchWithSlash(t)
	defer tearDown()

	m := project.Mercurial{
		Filepath: filepath.Join(fp, "wakatime-cli/src/pkg/file.go"),
	}

	result, detected, err := m.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, fp)
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "feature/billing",
		Folder:  result.Folder,
	}, result)
}

func TestMercurial_Detect_NoBranch(t *testing.T) {
	fp, tearDown := setupTestMercurialNoBranch(t)
	defer tearDown()

	m := project.Mercurial{
		Filepath: filepath.Join(fp, "wakatime-cli/src/pkg/file.go"),
	}

	result, detected, err := m.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Contains(t, result.Folder, fp)
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "default",
		Folder:  result.Folder,
	}, result)
}

func setupTestMercurial(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-hg")
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.Create(filepath.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	err = os.Mkdir(filepath.Join(tmpDir, "wakatime-cli/.hg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(t, "testdata/hg/branch", filepath.Join(tmpDir, "wakatime-cli/.hg/branch"))

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}

func setupTestMercurialBranchWithSlash(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-hg")
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.Create(filepath.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	err = os.Mkdir(filepath.Join(tmpDir, "wakatime-cli/.hg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(t, "testdata/hg/branch_with_slash", filepath.Join(tmpDir, "wakatime-cli/.hg/branch"))

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}

func setupTestMercurialNoBranch(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-hg")
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.Create(filepath.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	err = os.Mkdir(filepath.Join(tmpDir, "wakatime-cli/.hg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}
