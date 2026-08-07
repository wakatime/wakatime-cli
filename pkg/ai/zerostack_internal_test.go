package ai

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZerostackParse(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	dataDir := filepath.Join(home, "zerostack-data")
	t.Setenv("ZS_DATA_DIR", dataDir)

	sessionsDir := filepath.Join(dataDir, "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sessionsDir, "session.json"), []byte(`{
		"id":"zs-1","updated_at":"2026-08-06T12:00:00Z","working_dir":"/workspace/zerostack",
		"model":"deepseek/deepseek-v4-pro","total_input_tokens":80,"total_output_tokens":24,
		"messages":[{"role":"user","content":"Refactor this package"}]
	}`), 0o600))

	got, err := (Zerostack{After: time.Date(2026, 8, 6, 11, 30, 0, 0, time.UTC)}).Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "Zerostack zs-1", got[0].Entity)
	assert.Equal(t, "zs-1", got[0].AISession)
	assert.Equal(t, "/workspace/zerostack", got[0].ProjectPathOverride)
	assert.Equal(t, int64(80), got[0].AIInputTokens)
	assert.Equal(t, int64(24), got[0].AIOutputTokens)
	assert.Equal(t, len([]rune("Refactor this package")), got[0].AIPromptLength)
}
