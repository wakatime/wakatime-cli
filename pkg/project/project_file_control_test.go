package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "wakatime-git")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	err = project.Write(tmpDir, "billing")
	require.NoError(t, err)

	actual, err := os.ReadFile(filepath.Join(tmpDir, ".wakatime-project"))
	require.NoError(t, err)

	assert.Equal(t, string([]byte("billing\n")), string(actual))
}
