package file_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/file"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFind(t *testing.T) {
	tmpDir := t.TempDir()

	dir := filepath.Join(tmpDir, "src", "folder-1", "folder-2", "folder-3")

	err := os.MkdirAll(dir, os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.OpenFile(filepath.Join(dir, "some-entity.go"), os.O_RDONLY|os.O_CREATE, 0700)
	require.NoError(t, err)

	tmpProjectFile, err := os.OpenFile(filepath.Join(tmpDir, "src", ".wakatime"), os.O_RDONLY|os.O_CREATE, 0700)
	require.NoError(t, err)

	defer func() {
		tmpFile.Close()
		tmpProjectFile.Close()
	}()

	fp, ok := file.Find(t.Context(), dir, ".wakatime")

	assert.True(t, ok)
	assert.Equal(t, filepath.Join(tmpDir, "src", ".wakatime"), fp)
}

func TestFind_EmptyInput(t *testing.T) {
	tests := map[string]struct {
		directory string
		filename  string
	}{
		"empty directory": {
			directory: "",
			filename:  "some-entity.go",
		},
		"empty filename": {
			directory: "some-directory",
			filename:  "",
		},
		"both empty": {
			directory: "",
			filename:  "",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			fp, ok := file.Find(t.Context(), test.directory, test.filename)

			assert.False(t, ok)
			assert.Empty(t, fp)
		})
	}
}

func TestExists(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-file-*.go")
	require.NoError(t, err)

	found := file.Exists(tmpFile.Name())

	assert.True(t, found)
}

func TestExists_NotExists(t *testing.T) {
	found := file.Exists("/nonfile")

	assert.False(t, found)
}

func TestExists_SkipDir(t *testing.T) {
	found := file.Exists(t.TempDir())

	assert.False(t, found)
}
