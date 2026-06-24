package ai

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClinePathHelpers(t *testing.T) {
	assert.True(t, clineUsesPosixPaths("/workspace/project"))
	assert.False(t, clineUsesPosixPaths(`C:\workspace\project`))
	assert.False(t, clineUsesPosixPaths("relative/path"))

	assert.True(t, clineIsAbsPath("/workspace/project/main.go"))

	if filepath.Separator == '\\' {
		assert.Equal(t, `/workspace/project/main.go`, clineJoinPath("/workspace/project", "main.go"))
		assert.Equal(t, filepath.Join(`C:\workspace\project`, "main.go"), clineJoinPath(`C:\workspace\project`, "main.go"))
	} else {
		assert.Equal(t, "/workspace/project/main.go", clineJoinPath("/workspace/project", "main.go"))
	}
}

func TestCodexDecodeHexEscape(t *testing.T) {
	value, consumed, ok := codexDecodeHexEscape("41 rest", 2)
	assert.True(t, ok)
	assert.Equal(t, 'A', value)
	assert.Equal(t, 2, consumed)

	value, consumed, ok = codexDecodeHexEscape("03bb", 4)
	assert.True(t, ok)
	assert.Equal(t, rune(0x03bb), value)
	assert.Equal(t, 4, consumed)

	_, _, ok = codexDecodeHexEscape("4", 2)
	assert.False(t, ok)

	_, _, ok = codexDecodeHexEscape("zz", 2)
	assert.False(t, ok)
}

func TestCopilotURIHelpers(t *testing.T) {
	c := Copilot{}

	assert.Empty(t, c.pathToFileURI(""))
	assert.Equal(t, "file:///tmp/main.go", c.pathToFileURI("/tmp/main.go"))

	path, err := c.fileURIToPath("file:///tmp/main.go")
	assert.NoError(t, err)
	assert.Equal(t, "/tmp/main.go", path)

	_, err = c.fileURIToPath("%")
	assert.Error(t, err)
}

func TestCopilotResponseHasActivity(t *testing.T) {
	c := Copilot{}

	assert.True(t, c.responseHasActivity([]copilotResponseItem{{Kind: "thinking"}}))
	assert.True(t, c.responseHasActivity([]copilotResponseItem{{Kind: "inlineReference"}}))
	assert.False(t, c.responseHasActivity([]copilotResponseItem{{Kind: "ignored"}}))
	assert.False(t, c.responseHasActivity(nil))
}

func TestWindsurfLineChangesFromDiff(t *testing.T) {
	diff := "--- a/main.go\n+++ b/main.go\n-old\n+new\n+extra\n context"

	assert.Equal(t, 1, Windsurf{}.lineChangesFromDiff(diff))
}

func TestGeminiPromptHeartbeatsFromLogs(t *testing.T) {
	root := t.TempDir()
	transcript := filepath.Join(root, "session", "transcript.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(transcript), 0700))

	logs := []geminiPromptLog{
		{
			SessionID: "session-1",
			Type:      "user",
			Message:   "write a test",
			Timestamp: "2026-01-02T03:04:05Z",
		},
		{
			SessionID: "session-1",
			Type:      "model",
			Message:   "ignored",
			Timestamp: "2026-01-02T03:05:05Z",
		},
		{
			SessionID: "other-session",
			Type:      "user",
			Message:   "ignored",
			Timestamp: "2026-01-02T03:06:05Z",
		},
	}

	contents, err := json.Marshal(logs)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root, "logs.json"), contents, 0600))

	heartbeats, err := Gemini{
		After:             time.Date(2026, 1, 2, 3, 0, 0, 0, time.UTC),
		FallbackUserAgent: "fallback/1.0.0",
	}.promptHeartbeatsFromLogs(transcript, "Gemini session-1", "session-1", "/workspace/project")
	require.NoError(t, err)
	require.Len(t, heartbeats, 1)

	assert.Equal(t, "Gemini session-1", heartbeats[0].Entity)
	assert.Equal(t, heartbeat.AppType, heartbeats[0].EntityType)
	assert.Equal(t, "session-1", heartbeats[0].AISession)
	assert.Equal(t, len([]rune("write a test")), heartbeats[0].AIPromptLength)
	assert.Equal(t, "/workspace/project", heartbeats[0].ProjectPathOverride)
	assert.Equal(t, "fallback/1.0.0", heartbeats[0].UserAgent)
}

func TestGeminiPromptHeartbeatsFromLogsMissingFile(t *testing.T) {
	heartbeats, err := Gemini{}.promptHeartbeatsFromLogs(
		filepath.Join(t.TempDir(), "session", "transcript.json"),
		"Gemini session-1",
		"session-1",
		"",
	)

	require.NoError(t, err)
	assert.Nil(t, heartbeats)
}

func TestQoderPromptHeartbeat(t *testing.T) {
	timestamp := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	h := Qoder{FallbackUserAgent: "fallback/1.0.0"}.qoderPromptHeartbeat(qoderPrompt{
		SessionID:  "session-1",
		ProjectURI: "/workspace/project",
		Timestamp:  timestamp,
		Length:     42,
	})

	assert.Equal(t, "Qoder session-1", h.Entity)
	assert.Equal(t, heartbeat.AppType, h.EntityType)
	assert.Equal(t, "session-1", h.AISession)
	assert.Equal(t, 42, h.AIPromptLength)
	assert.Equal(t, "/workspace/project", h.ProjectPathOverride)
	assert.Equal(t, float64(timestamp.UnixMilli())/1000, h.Time)
	assert.Equal(t, "fallback/1.0.0", h.UserAgent)
}

func TestQoderToolResultHelpers(t *testing.T) {
	assert.Equal(t, "/tmp/main.go", qoderToolResult{
		Parameters: qoderToolParameters{FilePath: "/tmp/main.go", Path: "/tmp/fallback.go"},
	}.parameterFilePath())
	assert.Equal(t, "/tmp/fallback.go", qoderToolResult{
		Parameters: qoderToolParameters{Path: "/tmp/fallback.go"},
	}.parameterFilePath())

	assert.Equal(t, 2, qoderToolFile{
		DiffInfo: qoderDiffInfo{Add: 3, Delete: 1},
	}.lineChanges())
	assert.Equal(t, -1, qoderToolFile{
		LastDiffInfo: qoderDiffInfo{Add: 1, Delete: 2},
	}.lineChanges())
}

func TestQwenCodeResolveDir(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")

	resolved, err := qwenCodeResolveDir("~", home)
	require.NoError(t, err)
	assert.Equal(t, filepath.Clean(home), resolved)

	resolved, err = qwenCodeResolveDir("~/runtime", home)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "runtime"), resolved)

	resolved, err = qwenCodeResolveDir(`~\runtime`, home)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "runtime"), resolved)
}

func TestSyncLockClockNow(t *testing.T) {
	assert.False(t, syncLockClock{}.Now().IsZero())
}
