//go:build windows

package file_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/file"
)

func TestOpen(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	f, err := os.Open(tmpFile.Name())
	require.NoError(t, err)

	defer f.Close()

	err = os.Remove(tmpFile.Name())
	require.NoError(t, err)
}

func TestOpenNoLock(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	f, err := file.OpenNoLock(tmpFile.Name())
	require.NoError(t, err)

	defer f.Close()

	err = os.Remove(tmpFile.Name())
	require.NoError(t, err)
}
