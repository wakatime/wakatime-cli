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

func TestKimiParsesWireUsage(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("KIMI_SHARE_DIR", filepath.Join(home, ".kimi-test"))

	sessionDir := filepath.Join(home, ".kimi-test", "sessions", "project-hash", "session-1")
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	transcript := strings.Join([]string{
		`{"timestamp":1776162400,"message":{"type":"TurnBegin","payload":{"user_input":"Add status endpoint"}}}`,
		`{"timestamp":1776162403,"message":{"type":"StatusUpdate","payload":{"token_usage":{` +
			`"input_other":100,"input_cache_read":25,"output":40}}}}`,
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "wire.jsonl"), []byte(transcript), 0o600))

	got, err := (Kimi{After: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)}).Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "Kimi session-1", got[1].Entity)
	assert.Equal(t, int64(100), got[1].AIInputTokens)
	assert.Equal(t, int64(25), got[1].AICachedInputTokens)
	assert.Equal(t, int64(40), got[1].AIOutputTokens)
}
