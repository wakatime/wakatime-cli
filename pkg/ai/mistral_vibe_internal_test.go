package ai

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMistralVibeParsesSessionMetadata(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("VIBE_HOME", filepath.Join(home, ".vibe-test"))

	sessionDir := filepath.Join(home, ".vibe-test", "logs", "session", "session-1")
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "messages.jsonl"), nil, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "meta.json"), []byte(`{
		"session_id":"session-1","end_time":"2026-05-11T10:05:00Z",
		"environment":{"working_directory":"/workspace/vibe"},
		"stats":{"session_prompt_tokens":2000,"session_completion_tokens":3000,"session_cached_tokens":400},
		"config":{"active_model":"mistral-medium-3.5"},"title":"Implement Vibe support"
	}`), 0o600))

	got, err := (MistralVibe{After: time.Date(2026, 5, 11, 10, 0, 0, 0, time.UTC)}).Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "Mistral Vibe session-1", got[0].Entity)
	assert.Equal(t, int64(1600), got[0].AIInputTokens)
	assert.Equal(t, int64(400), got[0].AICachedInputTokens)
	assert.Equal(t, int64(3000), got[0].AIOutputTokens)
	assert.Equal(t, "/workspace/vibe", got[0].ProjectPathOverride)
}
