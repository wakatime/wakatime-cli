package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/gandarez/go-realpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFile_Detect_FileExists(t *testing.T) {
	tmpDir, err := realpath.Realpath(t.TempDir())
	require.NoError(t, err)

	copyFile(
		t,
		"testdata/wakatime-project",
		filepath.Join(tmpDir, ".wakatime-project"),
	)

	f := project.File{
		Filepath: filepath.Join(tmpDir, ".wakatime-project"),
	}

	result, detected, err := f.Detect()
	require.NoError(t, err)

	expected := project.Result{
		Branch:  "master",
		Folder:  tmpDir,
		Project: "wakatime-cli",
	}

	assert.True(t, detected)
	assert.Equal(t, expected, result)
}

func TestFile_Detect_ParentFolderExists(t *testing.T) {
	tmpDir, err := realpath.Realpath(t.TempDir())
	require.NoError(t, err)

	dir := filepath.Join(tmpDir, "src", "otherfolder")

	err = os.MkdirAll(dir, os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(
		t,
		"testdata/wakatime-project",
		filepath.Join(tmpDir, ".wakatime-project"),
	)

	f := project.File{
		Filepath: dir,
	}

	result, detected, err := f.Detect()
	require.NoError(t, err)

	expected := project.Result{
		Branch:  "master",
		Folder:  tmpDir,
		Project: "wakatime-cli",
	}

	assert.True(t, detected)
	assert.Equal(t, expected, result)
}

func TestFile_Detect_NoFileFound(t *testing.T) {
	tmpDir := t.TempDir()

	tmpFile, err := os.CreateTemp(tmpDir, "wakatime-project")
	require.NoError(t, err)

	defer tmpFile.Close()

	f := project.File{
		Filepath: tmpDir,
	}

	result, detected, err := f.Detect()
	require.NoError(t, err)

	expected := project.Result{}

	assert.False(t, detected)
	assert.Equal(t, expected, result)
}

func TestFile_Detect_InvalidPath(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "non-valid-file")
	require.NoError(t, err)

	defer tmpFile.Close()

	f := project.File{
		Filepath: tmpFile.Name(),
	}

	_, detected, err := f.Detect()
	require.NoError(t, err)

	assert.False(t, detected)
}

func TestFindFileOrDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	dir := filepath.Join(tmpDir, "src", "otherfolder")

	err := os.MkdirAll(dir, os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(
		t,
		"testdata/wakatime-project",
		filepath.Join(tmpDir, ".wakatime-project"),
	)

	tests := map[string]struct {
		Filepath string
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
			fp, ok := project.FindFileOrDirectory(test.Filepath, test.Filename)
			require.True(t, ok)

			assert.Equal(t, test.Expected, fp)
		})
	}
}

func TestFile_ID(t *testing.T) {
	f := project.File{}

	assert.Equal(t, project.FileDetector, f.ID())
}

func copyFile(t *testing.T, source, destination string) {
	input, err := os.ReadFile(source)
	require.NoError(t, err)

	err = os.WriteFile(destination, input, 0600)
	require.NoError(t, err)
}
