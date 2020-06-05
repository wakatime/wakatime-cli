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

func TestFileDetect_FileExists(t *testing.T) {
	f := project.File{
		Filepath: "testdata/.wakatime-project",
	}

	result, detected, err := f.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Equal(t, "wakatime-cli", result.Project)
	assert.Equal(t, "master", result.Branch)
}

func TestFileDetect_AnyFileFound(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "wakatime-project")
	require.NoError(t, err)

	dir, _ := path.Split(tmpFile.Name())

	defer os.Remove(tmpFile.Name())

	f := project.File{
		Filepath: dir,
	}

	_, detected, err := f.Detect()
	require.NoError(t, err)

	assert.False(t, detected)
}

func TestFileDetect_WrongPath(t *testing.T) {
	f := project.File{
		Filepath: "path/to/non-file",
	}

	_, detected, err := f.Detect()

	var pferr project.ErrProject

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
