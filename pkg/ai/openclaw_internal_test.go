package ai

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenClawParsesRealSessionUsageSchema(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	sessionsDir := filepath.Join(home, ".openclaw", "agents", "main", "sessions")
	require.NoError(t, os.MkdirAll(sessionsDir, 0o755))

	transcript := strings.Join([]string{
		`{"type":"session","id":"openclaw-session","timestamp":"2026-08-06T11:59:59Z","cwd":"/workspace/openclaw"}`,
		`{"type":"message","id":"user-1","parentId":null,"timestamp":"2026-08-06T12:00:00Z",` +
			`"message":{"role":"user","content":[{"type":"text","text":"Inspect the cache accounting"}]}}`,
		`{"type":"message","id":"assistant-1","parentId":"user-1","timestamp":"2026-08-06T12:00:01Z",` +
			`"message":{"role":"assistant","content":[{"type":"text","text":"Done"}],` +
			`"provider":"anthropic","model":"claude-sonnet-4",` +
			`"usage":{"input":10,"output":139,"cacheRead":13962,"cacheWrite":465,"totalTokens":14576}}}`,
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionsDir, "openclaw-session.jsonl"), []byte(transcript), 0o600,
	))

	got, err := (OpenClaw{After: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}).Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "OpenClaw openclaw-session", got[0].Entity)
	assert.Equal(t, len([]rune("Inspect the cache accounting")), got[0].AIPromptLength)
	assert.Equal(t, "/workspace/openclaw", got[0].ProjectPathOverride)

	assert.Equal(t, int64(475), got[1].AIInputTokens)
	assert.Equal(t, int64(13962), got[1].AICachedInputTokens)
	assert.Equal(t, int64(139), got[1].AIOutputTokens)
	assert.Contains(t, got[1].UserAgent, "sonnet/4")
}
