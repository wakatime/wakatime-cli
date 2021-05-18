package project_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubversion_Detect(t *testing.T) {
	skipIfBinaryNotFound(t)

	fp, tearDown := setupTestSvn(t)
	defer tearDown()

	s := project.Subversion{
		Filepath: filepath.Join(fp, "wakatime-cli", "src", "pkg", "file.go"),
	}

	result, detected, err := s.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "trunk",
		Folder:  "file:///D:/temp/SVN/wakatime-cli",
	}, result)
}

func TestSubversion_Detect_Branch(t *testing.T) {
	skipIfBinaryNotFound(t)

	fp, tearDown := setupTestSvnBranch(t)
	defer tearDown()

	s := project.Subversion{
		Filepath: filepath.Join(fp, "wakatime-cli/src/pkg/file.go"),
	}

	result, detected, err := s.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "billing",
		Folder:  "file:///D:/temp/SVN/wakatime-cli",
	}, result)
}

func setupTestSvn(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-svn")
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.Create(filepath.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	copyDir(t, "testdata/svn", filepath.Join(tmpDir, "wakatime-cli/.svn"))

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}

func setupTestSvnBranch(t *testing.T) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-svn")
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.Create(filepath.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	copyDir(t, "testdata/svn_branch", filepath.Join(tmpDir, "wakatime-cli/.svn"))

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}

func copyDir(t *testing.T, src string, dst string) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	require.NoError(t, err)

	if !si.IsDir() {
		return
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}

	if err == nil {
		return
	}

	err = os.MkdirAll(dst, si.Mode())
	require.NoError(t, err)

	entries, err := ioutil.ReadDir(src)
	require.NoError(t, err)

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			copyDir(t, srcPath, dstPath)
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			copyFile(t, srcPath, dstPath)
		}
	}
}

func findSvnBinary() (string, bool) {
	locations := []string{
		"svn",
		"/usr/bin/svn",
		"/usr/local/bin/svn",
	}

	for _, loc := range locations {
		cmd := exec.Command(loc, "--version")

		err := cmd.Run()
		if err != nil {
			continue
		}

		return loc, true
	}

	return "", false
}

func skipIfBinaryNotFound(t *testing.T) {
	_, found := findSvnBinary()
	if !found {
		t.Skip("Skipping because svn binary is not installed in this machine.")
	}
}
