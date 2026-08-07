package ai

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestDroidParse(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	factoryDir := filepath.Join(home, "factory-data")
	t.Setenv("FACTORY_DIR", factoryDir)
	sessionsDir := filepath.Join(factoryDir, "sessions", "project")
	require.NoError(t, os.MkdirAll(sessionsDir, 0o755))

	transcript := strings.Join([]string{
		`{"timestamp":"2026-08-06T12:00:00Z","sessionId":"session-1","model":"claude-sonnet-4",` +
			`"cwd":"/workspace/project","message":{"role":"user","content":"Implement the parser"},` +
			`"usage":{"input_tokens":120,"cache_read_tokens":20,"output_tokens":30}}`,
		`{"timestamp":"2026-08-06T12:01:00Z","sessionId":"session-1",` +
			`"tool":{"tool_name":"edit_file","path":"pkg/main.go","diff":"--- a\n+++ b\n-old\n+new"}}`,
		`{"timestamp":"2026-08-06T11:00:00Z","sessionId":"old","message":{"role":"user","content":"old"}}`,
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(sessionsDir, "session.jsonl"), []byte(transcript), 0o600))

	got, err := (Droid{After: time.Date(2026, 8, 6, 11, 30, 0, 0, time.UTC)}).Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 3)
	assert.Equal(t, "Droid session-1", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, int64(120), got[0].AIInputTokens)
	assert.Equal(t, int64(20), got[0].AICachedInputTokens)
	assert.Equal(t, int64(30), got[0].AIOutputTokens)
	assert.Equal(t, len([]rune("Implement the parser")), got[0].AIPromptLength)
	assert.Contains(t, got[0].UserAgent, "Droid")
	assert.Equal(t, heartbeat.AppType, got[1].EntityType)
	assert.Equal(t, filepath.ToSlash(filepath.Join("/workspace/project", "pkg/main.go")), got[2].Entity)
	assert.Equal(t, heartbeat.FileType, got[2].EntityType)
	assert.Equal(t, heartbeat.PointerTo(true), got[2].IsWrite)
	assert.Equal(t, heartbeat.PointerTo(2), got[2].AILineChanges)
}
