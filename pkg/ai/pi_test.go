package ai_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestPiParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	projectDir := filepath.Join(home, "project")
	require.NoError(t, os.MkdirAll(projectDir, 0o755))

	readFile := filepath.Join(projectDir, "README.md")
	editFile := filepath.Join(projectDir, "main.go")

	sessionsDir := filepath.Join(home, ".pi", "agent", "sessions", "--tmp-project--")
	require.NoError(t, os.MkdirAll(sessionsDir, 0o755))

	transcriptPath := filepath.Join(sessionsDir, "2026-04-19T12-00-00-000Z_session-1.jsonl")
	noteFile := filepath.Join(projectDir, "notes.txt")
	transcript := strings.Join([]string{
		mustJSONLine(t, map[string]interface{}{
			"type":      "session",
			"version":   3,
			"id":        "session-1",
			"timestamp": "2026-04-19T12:00:00.000Z",
			"cwd":       projectDir,
		}),
		mustJSONLine(t, map[string]interface{}{
			"type":      "model_change",
			"id":        "a1",
			"parentId":  nil,
			"timestamp": "2026-04-19T12:00:00.010Z",
			"provider":  "anthropic",
			"modelId":   "claude-opus-4-5",
		}),
		mustJSONLine(t, map[string]interface{}{
			"type":      "message",
			"id":        "a2",
			"parentId":  "a1",
			"timestamp": "2026-04-19T12:00:01.000Z",
			"message": map[string]interface{}{
				"role": "user",
				"content": []map[string]interface{}{
					{"type": "text", "text": "Inspect the project and fix main.go"},
				},
				"timestamp": 1776600001000,
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"type":      "message",
			"id":        "a3",
			"parentId":  "a2",
			"timestamp": "2026-04-19T12:00:02.000Z",
			"message": map[string]interface{}{
				"role": "assistant",
				"content": []map[string]interface{}{
					{
						"type": "toolCall", "id": "tool-read", "name": "read",
						"arguments": map[string]interface{}{"path": "README.md"},
					},
				},
				"provider":  "anthropic",
				"model":     "claude-opus-4-5",
				"usage":     map[string]interface{}{"input": 10, "output": 20, "cacheRead": 3, "cacheWrite": 2},
				"timestamp": 1776600002000,
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"type":      "message",
			"id":        "a4",
			"parentId":  "a3",
			"timestamp": "2026-04-19T12:00:03.000Z",
			"message": map[string]interface{}{
				"role":       "toolResult",
				"toolCallId": "tool-read",
				"toolName":   "read",
				"content":    []map[string]interface{}{{"type": "text", "text": "# Project\n"}},
				"isError":    false,
				"timestamp":  1776600003000,
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"type":      "message",
			"id":        "a5",
			"parentId":  "a4",
			"timestamp": "2026-04-19T12:00:04.000Z",
			"message": map[string]interface{}{
				"role": "assistant",
				"content": []map[string]interface{}{
					{
						"type": "toolCall", "id": "tool-edit", "name": "edit",
						"arguments": map[string]interface{}{
							"filePath": "main.go",
							"newText":  "package main\n\nfunc main() {}\n",
						},
					},
				},
				"provider":  "anthropic",
				"model":     "claude-opus-4-5",
				"usage":     map[string]interface{}{"input": 7, "output": 11, "cacheRead": 0, "cacheWrite": 0},
				"timestamp": 1776600004000,
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"type":      "message",
			"id":        "a6",
			"parentId":  "a5",
			"timestamp": "2026-04-19T12:00:05.000Z",
			"message": map[string]interface{}{
				"role":       "toolResult",
				"toolCallId": "tool-edit",
				"toolName":   "edit",
				"content": []map[string]interface{}{
					{"type": "text", "text": "Successfully replaced text in " + editFile + "."},
				},
				"details":   map[string]interface{}{"diff": "-old\n+new\n+more"},
				"isError":   false,
				"timestamp": 1776600005000,
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"type":      "message",
			"id":        "a7",
			"parentId":  "a6",
			"timestamp": "2026-04-19T12:00:06.000Z",
			"message": map[string]interface{}{
				"role": "assistant",
				"content": []map[string]interface{}{
					{
						"type": "toolCall", "id": "tool-write", "name": "write",
						"arguments": map[string]interface{}{
							"path":    "notes.txt",
							"content": "alpha\nbeta\n",
						},
					},
				},
				"provider":  "anthropic",
				"model":     "claude-opus-4-5",
				"usage":     map[string]interface{}{"input": 5, "output": 9, "cacheRead": 0, "cacheWrite": 0},
				"timestamp": 1776600006000,
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"type":      "message",
			"id":        "a8",
			"parentId":  "a7",
			"timestamp": "2026-04-19T12:00:07.000Z",
			"message": map[string]interface{}{
				"role":       "toolResult",
				"toolCallId": "tool-write",
				"toolName":   "write",
				"content": []map[string]interface{}{
					{"type": "text", "text": "Successfully wrote 11 bytes to " + noteFile},
				},
				"isError":   false,
				"timestamp": 1776600007000,
			},
		}),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Pi{
		After:             time.Date(2026, 4, 19, 11, 59, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			editFile: heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 4)

	assert.Equal(t, "Pi 2026-04-19T12-00-00-000Z_session-1", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "session-1", got[0].AISession)
	assert.Equal(t, len([]rune("Inspect the project and fix main.go")), got[0].AIPromptLength)
	assert.Equal(t, projectDir, got[0].ProjectPathOverride)
	require.NotNil(t, got[0].IsWrite)
	assert.False(t, *got[0].IsWrite)

	assert.Equal(t, readFile, got[1].Entity)
	assert.Equal(t, heartbeat.FileType, got[1].EntityType)
	assert.Equal(t, "session-1", got[1].AISession)
	require.NotNil(t, got[1].AILineChanges)
	assert.Zero(t, *got[1].AILineChanges)
	assert.Equal(t, int64(15), got[1].AIInputTokens)
	assert.Equal(t, int64(20), got[1].AIOutputTokens)
	require.NotNil(t, got[1].IsWrite)
	assert.False(t, *got[1].IsWrite)
	assert.Contains(t, got[1].UserAgent, "Pi/claude-opus-4-5")

	assert.Equal(t, editFile, got[2].Entity)
	require.NotNil(t, got[2].AILineChanges)
	assert.Equal(t, 1, *got[2].AILineChanges)
	assert.Equal(t, int64(7), got[2].AIInputTokens)
	assert.Equal(t, int64(11), got[2].AIOutputTokens)
	require.NotNil(t, got[2].IsWrite)
	assert.True(t, *got[2].IsWrite)
	assert.Contains(t, got[2].UserAgent, "Pi/claude-opus-4-5")
	assert.Contains(t, got[2].UserAgent, "editor/1.2.3")

	assert.Equal(t, noteFile, got[3].Entity)
	require.NotNil(t, got[3].AILineChanges)
	assert.Equal(t, 3, *got[3].AILineChanges)
	assert.Equal(t, int64(5), got[3].AIInputTokens)
	assert.Equal(t, int64(9), got[3].AIOutputTokens)
	require.NotNil(t, got[3].IsWrite)
	assert.True(t, *got[3].IsWrite)
}

func TestPiParse_NoPiSessionsDir(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := ai.Pi{After: time.Now()}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func mustJSONLine(t *testing.T, v interface{}) string {
	t.Helper()

	data, err := json.Marshal(v)
	require.NoError(t, err)

	return string(data)
}
