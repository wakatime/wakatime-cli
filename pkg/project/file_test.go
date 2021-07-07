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
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	dir := filepath.Join(tmpDir, "src", "otherfolder")

	err = os.MkdirAll(dir, os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(
		t,
		"testdata/.wakatime-project",
		filepath.Join(tmpDir, ".wakatime-project"),
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

func TestFile_Detect_InvalidPath(t *testing.T) {
	f := project.File{
		Filepath: "path/to/non-file",
	}

	_, detected, err := f.Detect()
	require.NoError(t, err)

	assert.False(t, detected)
}

func TestFile_String(t *testing.T) {
	f := project.File{}

	assert.Equal(t, "project-file-detector", f.String())
}

func TestFindFileOrDirectory(t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	dir := filepath.Join(tmpDir, "src", "otherfolder")

	err = os.MkdirAll(dir, os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(
		t,
		"testdata/.wakatime-project",
		filepath.Join(tmpDir, ".wakatime-project"),
	)

	tests := map[string]struct {
		Filepath string
		FileDir  string
		Filename string
		Expected string
	}{
		"filename": {
			Filepath: dir,
			Filename: ".wakatime-project",
			Expected: filepath.Join(tmpDir, ".wakatime-project"),
		},
		"directory": {
			Filepath: dir,
			Filename: "src",
			Expected: filepath.Join(tmpDir, "src"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			fp, ok := project.FindFileOrDirectory(test.Filepath, test.FileDir, test.Filename)
			require.True(t, ok)

			assert.Equal(t, test.Expected, fp)
		})
	}
}

func copyFile(t *testing.T, source, destination string) {
	input, err := ioutil.ReadFile(source)
	require.NoError(t, err)

	err = ioutil.WriteFile(destination, input, 0600)
	require.NoError(t, err)
}
