package file_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/file"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadHead(t *testing.T) {
	tmpDir := t.TempDir()

	f, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer f.Close()

	// create a temporary file smaller than MaxFileSizeSupported
	const size = 2 * 1024 * 1024 // 2MB

	data := make([]byte, 1024*1024) // 1MB buffer

	for i := 0; i < size/len(data); i++ {
		_, err := f.Write(data)
		require.NoError(t, err)
	}

	head, err := file.ReadHead(t.Context(), f.Name(), 1*1024*1024)
	require.NoError(t, err)

	assert.Len(t, head, 1*1024*1024) // should read only 1MB
}

func TestReadHead_MaxFileSizeSupported(t *testing.T) {
	tmpDir := t.TempDir()

	f, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer f.Close()

	// create a temporary file bigger than MaxFileSizeSupported
	const size = 6 * 1024 * 1024 // 6MB

	data := make([]byte, 1024*1024) // 1MB buffer

	for i := 0; i < size/len(data); i++ {
		_, err := f.Write(data)
		require.NoError(t, err)
	}

	head, err := file.ReadHead(t.Context(), f.Name(), -1)
	require.NoError(t, err)

	assert.Len(t, head, file.MaxFileSizeSupported) // should read only 5MB
}

func TestReadHead_NonFile(t *testing.T) {
	_, err := file.ReadHead(t.Context(), "non-file", 1*1024*1024)

	assert.ErrorContains(t, err, "failed to open file \"non-file\"")
}

func TestReadLines(t *testing.T) {
	tmpDir := t.TempDir()

	f, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer f.Close()

	// Write n lines of random strings
	for range 2000 {
		line := "Go\n"
		_, err := f.WriteString(line)
		require.NoError(t, err)
	}

	lines, err := file.ReadLines(t.Context(), f.Name(), 1000)
	require.NoError(t, err)

	assert.Len(t, lines, 1000) // should read only 1000 lines
}

func TestReadLines_ZeroLines(t *testing.T) {
	tmpDir := t.TempDir()

	f, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer f.Close()

	// Write n lines of random strings
	for range 2000 {
		line := "Go\n"
		_, err := f.WriteString(line)
		require.NoError(t, err)
	}

	lines, err := file.ReadLines(t.Context(), f.Name(), 0)
	require.NoError(t, err)

	assert.Empty(t, lines)
}

func TestReadLines_NegativeLines(t *testing.T) {
	tmpDir := t.TempDir()

	f, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer f.Close()

	// Write n lines of random strings
	for range 2000 {
		line := "Go\n"
		_, err := f.WriteString(line)
		require.NoError(t, err)
	}

	lines, err := file.ReadLines(t.Context(), f.Name(), -2)
	require.NoError(t, err)

	assert.Empty(t, lines)
}

func TestReadLines_NonFile(t *testing.T) {
	_, err := file.ReadLines(t.Context(), "non-file", 1*1024*1024)

	assert.ErrorContains(t, err, "failed to open file \"non-file\"")
}

func TestCountLines(t *testing.T) {
	tmpDir := t.TempDir()

	f, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer f.Close()

	// Write n lines of random strings
	for range 2000 {
		line := "Go\n"
		_, err := f.WriteString(line)
		require.NoError(t, err)
	}

	lines, err := file.CountLines(t.Context(), f.Name())
	require.NoError(t, err)

	assert.Equal(t, lines, 2000)
}

func TestReadLines_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	f, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer f.Close()

	lines, err := file.CountLines(t.Context(), f.Name())
	require.NoError(t, err)

	assert.Empty(t, lines)
}

func TestCountLines_NonFile(t *testing.T) {
	lines, err := file.CountLines(t.Context(), "non-file")

	assert.Empty(t, lines)
	assert.ErrorContains(t, err, "failed to open file \"non-file\"")
}
