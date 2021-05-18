package project_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTfvc_Detect(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping because OS is windows.")
	}

	fp, tearDown := setupTestTfvc(t, ".tf")
	defer tearDown()

	s := project.Tfvc{
		Filepath: filepath.Join(fp, "wakatime-cli", "src", "pkg", "file.go"),
	}

	result, detected, err := s.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "",
		Folder:  result.Folder,
	}, result)
}

func TestTfvc_Detect_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping because OS is not windows.")
	}

	fp, tearDown := setupTestTfvc(t, "$tf")
	defer tearDown()

	s := project.Tfvc{
		Filepath: filepath.Join(fp, "wakatime-cli", "src", "pkg", "file.go"),
	}

	result, detected, err := s.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "",
		Folder:  result.Folder,
	}, result)
}

func setupTestTfvc(t *testing.T, tfFolderName string) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-tfvc")
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	err = os.Mkdir(filepath.Join(tmpDir, fmt.Sprintf("wakatime-cli/%s", tfFolderName)), os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.Create(filepath.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	tmpPropertiesFile, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("wakatime-cli/%s/properties.tf1", tfFolderName)))
	require.NoError(t, err)

	defer tmpPropertiesFile.Close()

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}
