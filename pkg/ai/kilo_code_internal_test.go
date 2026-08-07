package ai

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKiloCodeParsesClineStyleTasks(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, "data"))

	paths, err := kiloCodePaths(t.Context())
	require.NoError(t, err)
	require.NotEmpty(t, paths.roots)
	taskDir := filepath.Join(paths.roots[0], "task-1")
	require.NoError(t, os.MkdirAll(taskDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(taskDir, "ui_messages.json"), []byte(
		`[{"ts":1786017600000,"say":"api_req_started",`+
			`"text":"{\"request\":\"Build it\",\"tokensIn\":12,\"tokensOut\":4}"}]`,
	), 0o600))

	got, err := (KiloCode{After: time.Date(2026, 8, 6, 11, 0, 0, 0, time.UTC)}).Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "KiloCode task-1", got[0].Entity)
	assert.Equal(t, int64(12), got[0].AIInputTokens)
	assert.Equal(t, int64(4), got[0].AIOutputTokens)
	assert.Equal(t, len([]rune("Build it")), got[0].AIPromptLength)
}
