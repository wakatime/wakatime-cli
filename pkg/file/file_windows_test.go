//go:build windows

package file_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/file"

	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	// name := tmpFile.Name()

	err = tmpFile.Close()
	require.NoError(t, err)

	f, err := os.Open(tmpFile.Name())
	require.NoError(t, err)

	defer f.Close()

	err = os.Remove(tmpFile.Name())
	require.Error(t, err)
}

func TestOpenNoLock(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	// name := tmpFile.Name()

	err = tmpFile.Close()
	require.NoError(t, err)

	f, err := file.OpenNoLock(tmpFile.Name())
	require.NoError(t, err)

	defer f.Close()

	err = os.Remove(tmpFile.Name())
	require.NoError(t, err)
}
