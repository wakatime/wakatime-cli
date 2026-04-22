package ai_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestGeminiParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	projectDir := filepath.Join(home, "wakatime-cli")
	require.NoError(t, os.MkdirAll(projectDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".gemini"), 0o755))

	projectSlugDir := filepath.Join(home, ".gemini", "tmp", "wakatime-cli")
	sessionDir := filepath.Join(projectSlugDir, "chats")
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(projectSlugDir, ".project_root"),
		[]byte(projectDir),
		0o644,
	))

	logs := []map[string]any{
		{
			"sessionId": "gem-session-1",
			"type":      "user",
			"message":   "fallback prompt",
			"timestamp": "2026-04-21T11:59:59Z",
		},
	}
	logContents, err := json.Marshal(logs)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(
		filepath.Join(projectSlugDir, "logs.json"),
		logContents,
		0o644,
	))

	mainFile := filepath.Join(projectDir, "pkg", "ai", "gemini.go")
	notesFile := filepath.Join(projectDir, "notes.txt")

	session := map[string]any{
		"sessionId":   "gem-session-1",
		"projectHash": "hash",
		"startTime":   "2026-04-21T12:00:00Z",
		"lastUpdated": "2026-04-21T12:00:07Z",
		"messages": []map[string]any{
			{
				"id":        "m1",
				"type":      "user",
				"timestamp": "2026-04-21T12:00:01Z",
				"content": []map[string]any{
					{"text": "look for bugs and fix any you find"},
				},
			},
			{
				"id":        "m2",
				"type":      "gemini",
				"timestamp": "2026-04-21T12:00:02Z",
				"content":   "I will inspect the parser files first.",
				"tokens": map[string]any{
					"input":  100,
					"output": 20,
					"cached": 0,
					"tool":   0,
					"total":  120,
				},
				"model": "gemini-3-flash-preview",
				"toolCalls": []map[string]any{
					{
						"id":        "list-1",
						"name":      "list_directory",
						"args":      map[string]any{"dir_path": "pkg/ai"},
						"status":    "success",
						"timestamp": "2026-04-21T12:00:02Z",
					},
					{
						"id":        "read-1",
						"name":      "read_file",
						"args":      map[string]any{"file_path": "pkg/ai/ai.go"},
						"status":    "success",
						"timestamp": "2026-04-21T12:00:02Z",
					},
				},
			},
			{
				"id":        "m3",
				"type":      "gemini",
				"timestamp": "2026-04-21T12:00:03Z",
				"content":   "I found one issue and patched it.",
				"tokens": map[string]any{
					"input":  160,
					"output": 34,
					"cached": 30,
					"tool":   0,
					"total":  194,
				},
				"model": "gemini-3-flash-preview",
				"toolCalls": []map[string]any{
					{
						"id":     "replace-1",
						"name":   "replace",
						"status": "success",
						"args": map[string]any{
							"file_path":  "pkg/ai/gemini.go",
							"old_string": "old line",
							"new_string": "new line\nextra line",
						},
						"timestamp": "2026-04-21T12:00:03Z",
						"resultDisplay": map[string]any{
							"filePath": mainFile,
							"diffStat": map[string]any{
								"model_added_lines":   3,
								"model_removed_lines": 1,
							},
						},
					},
				},
			},
			{
				"id":        "m4",
				"type":      "gemini",
				"timestamp": "2026-04-21T12:00:04Z",
				"content":   "This replace was cancelled.",
				"tokens": map[string]any{
					"input":  170,
					"output": 40,
					"cached": 30,
					"tool":   0,
					"total":  210,
				},
				"model": "gemini-3-flash-preview",
				"toolCalls": []map[string]any{
					{
						"id":     "replace-2",
						"name":   "replace",
						"status": "cancelled",
						"args": map[string]any{
							"file_path":  "pkg/ai/ai.go",
							"old_string": "before",
							"new_string": "after",
						},
						"timestamp": "2026-04-21T12:00:04Z",
						"resultDisplay": map[string]any{
							"filePath": filepath.Join(projectDir, "pkg", "ai", "ai.go"),
						},
					},
				},
			},
			{
				"id":        "m5",
				"type":      "gemini",
				"timestamp": "2026-04-21T12:00:05Z",
				"content":   "I also wrote a notes file.",
				"tokens": map[string]any{
					"input":  190,
					"output": 45,
					"cached": 30,
					"tool":   0,
					"total":  235,
				},
				"model": "gemini-3-flash-preview",
				"toolCalls": []map[string]any{
					{
						"id":     "write-1",
						"name":   "write_file",
						"status": "success",
						"args": map[string]any{
							"file_path": "notes.txt",
							"content":   "alpha\nbeta",
						},
						"timestamp":     "2026-04-21T12:00:05Z",
						"resultDisplay": "",
					},
				},
			},
		},
	}

	sessionContents, err := json.Marshal(session)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionDir, "session-2026-04-21T12-00-00-test.json"),
		sessionContents,
		0o644,
	))

	parser := ai.Gemini{
		After:             time.Date(2026, 4, 21, 11, 59, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			mainFile: heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 7)

	assert.Equal(t, "Gemini gem-session-1", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "gem-session-1", got[0].AISession)
	assert.Equal(t, projectDir, got[0].ProjectPathOverride)
	assert.Equal(t, len([]rune("look for bugs and fix any you find")), got[0].AIPromptLength)
	assert.Zero(t, got[0].AIInputTokens)
	assert.Zero(t, got[0].AIOutputTokens)
	assert.Contains(t, got[0].UserAgent, "Gemini")

	assert.Equal(t, "Gemini gem-session-1", got[1].Entity)
	assert.Equal(t, heartbeat.AppType, got[1].EntityType)
	assert.Zero(t, got[1].AIPromptLength)
	assert.EqualValues(t, 100, got[1].AIInputTokens)
	assert.EqualValues(t, 20, got[1].AIOutputTokens)
	assert.Equal(t, projectDir, got[1].ProjectPathOverride)
	assert.Contains(t, got[1].UserAgent, "Gemini/gemini-3-flash-preview")

	assert.Equal(t, "Gemini gem-session-1", got[2].Entity)
	assert.Equal(t, heartbeat.AppType, got[2].EntityType)
	assert.Zero(t, got[2].AIPromptLength)
	assert.EqualValues(t, 60, got[2].AIInputTokens)
	assert.EqualValues(t, 14, got[2].AIOutputTokens)

	assert.Equal(t, mainFile, got[3].Entity)
	assert.Equal(t, heartbeat.FileType, got[3].EntityType)
	require.NotNil(t, got[3].AILineChanges)
	assert.Equal(t, 2, *got[3].AILineChanges)
	require.NotNil(t, got[3].IsWrite)
	assert.True(t, *got[3].IsWrite)
	assert.Zero(t, got[3].AIInputTokens)
	assert.Zero(t, got[3].AIOutputTokens)
	assert.Contains(t, got[3].UserAgent, "Gemini/gemini-3-flash-preview")
	assert.Contains(t, got[3].UserAgent, "editor/1.2.3")

	assert.Equal(t, "Gemini gem-session-1", got[4].Entity)
	assert.Equal(t, heartbeat.AppType, got[4].EntityType)
	assert.Zero(t, got[4].AIPromptLength)
	assert.EqualValues(t, 10, got[4].AIInputTokens)
	assert.EqualValues(t, 6, got[4].AIOutputTokens)

	assert.Equal(t, "Gemini gem-session-1", got[5].Entity)
	assert.Equal(t, heartbeat.AppType, got[5].EntityType)
	assert.Zero(t, got[5].AIPromptLength)
	assert.EqualValues(t, 20, got[5].AIInputTokens)
	assert.EqualValues(t, 5, got[5].AIOutputTokens)

	assert.Equal(t, notesFile, got[6].Entity)
	assert.Equal(t, heartbeat.FileType, got[6].EntityType)
	require.NotNil(t, got[6].AILineChanges)
	assert.Equal(t, 2, *got[6].AILineChanges)
	require.NotNil(t, got[6].IsWrite)
	assert.True(t, *got[6].IsWrite)
	assert.Zero(t, got[6].AIInputTokens)
	assert.Zero(t, got[6].AIOutputTokens)
}

func TestGeminiParse_NoGeminiTmpDir(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := ai.Gemini{After: time.Now()}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}
