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
		`{"type":"session","id":"session-1","timestamp":"2026-08-06T11:59:59Z","cwd":"/workspace/project"}`,
		`{"type":"message","timestamp":"2026-08-06T12:00:00Z",` +
			`"message":{"role":"user","content":[{"type":"text","text":"Implement the parser"}]}}`,
		`{"type":"message","timestamp":"2026-08-06T12:00:01Z",` +
			`"message":{"role":"assistant","model":"claude-sonnet-4","content":[{"type":"text","text":"Working on it"}],` +
			`"usage":{"input_tokens":120,"cache_creation_input_tokens":5,` +
			`"cache_read_input_tokens":20,"output_tokens":30}}}`,
		`{"timestamp":"2026-08-06T12:01:00Z","sessionId":"session-1",` +
			`"tool":{"tool_name":"edit_file","path":"pkg/main.go","diff":"--- a\n+++ b\n-old\n+new"}}`,
		`{"timestamp":"2026-08-06T11:00:00Z","sessionId":"old","message":{"role":"user","content":"old"}}`,
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(sessionsDir, "session-1.jsonl"), []byte(transcript), 0o600))

	got, err := (Droid{After: time.Date(2026, 8, 6, 11, 30, 0, 0, time.UTC)}).Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 4)
	assert.Equal(t, "Droid session-1", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, len([]rune("Implement the parser")), got[0].AIPromptLength)
	assert.Equal(t, int64(125), got[1].AIInputTokens)
	assert.Equal(t, int64(20), got[1].AICachedInputTokens)
	assert.Equal(t, int64(30), got[1].AIOutputTokens)
	assert.Contains(t, got[0].UserAgent, "Droid")
	assert.Equal(t, heartbeat.AppType, got[2].EntityType)
	assert.Equal(t, filepath.ToSlash(filepath.Join("/workspace/project", "pkg/main.go")), got[3].Entity)
	assert.Equal(t, heartbeat.FileType, got[3].EntityType)
	assert.Equal(t, heartbeat.PointerTo(true), got[3].IsWrite)
	assert.Equal(t, heartbeat.PointerTo(2), got[3].AILineChanges)
}
