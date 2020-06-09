package project_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFile_Detect_FileExists(t *testing.T) {
	f := project.File{
		Filepath: "testdata/.wakatime-project",
	}

	result, detected, err := f.Detect()
	require.NoError(t, err)

	expected := project.Result{
		Project: "wakatime-cli",
		Branch:  "master",
	}

	assert.True(t, detected)
	assert.Equal(t, expected, result)
}

func TestFile_Detect_ParentFolderExists(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	dir := path.Join(tmpDir, "src", "otherfolder")

	err = os.MkdirAll(dir, os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(
		t,
		path.Join(wd, "testdata/.wakatime-project"),
		path.Join(tmpDir, ".wakatime-project"),
	)

	f := project.File{
		Filepath: dir,
	}

	result, detected, err := f.Detect()
	require.NoError(t, err)

	expected := project.Result{
		Project: "wakatime-cli",
		Branch:  "master",
	}

	assert.True(t, detected)
	assert.Equal(t, expected, result)
}

func TestFile_Detect_AnyFileFound(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "wakatime-project")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	f := project.File{
		Filepath: os.TempDir(),
	}

	result, detected, err := f.Detect()
	require.NoError(t, err)

	expected := project.Result{}

	assert.False(t, detected)
	assert.Equal(t, expected, result)
}

func TestFile_Detect_WrongPath(t *testing.T) {
	f := project.File{
		Filepath: "path/to/non-file",
	}

	_, detected, err := f.Detect()

	var pferr project.Err

	errMsg := fmt.Sprintf("error %q differs from the string set", err)

	assert.False(t, detected)
	assert.True(t, errors.As(err, &pferr))
	assert.Contains(
		t,
		err.Error(),
		"failed to get the real path",
		errMsg,
	)
}

func TestFile_String(t *testing.T) {
	f := project.File{}

	assert.Equal(t, "project-file-detector", f.String())
}

func copyFile(t *testing.T, source, destination string) {
	input, err := ioutil.ReadFile(source)
	require.NoError(t, err)

	err = ioutil.WriteFile(destination, input, 0600)
	require.NoError(t, err)
}
