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

func TestContinueParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	continueDir := filepath.Join(home, ".continue")
	devDataDir := filepath.Join(continueDir, "dev_data", "0.2.0")
	sessionsDir := filepath.Join(continueDir, "sessions")

	require.NoError(t, os.MkdirAll(devDataDir, 0o755))
	require.NoError(t, os.MkdirAll(sessionsDir, 0o755))

	workspace := filepath.Join(home, "wakatime-cli")
	readPath := filepath.Join(workspace, "README.md")
	editPath := filepath.Join(workspace, "pkg", "ai", "continue.go")
	sessionID := "5832a36f-ea52-4eb4-b8f7-16ed2bf063dd"

	writeJSON(t, filepath.Join(sessionsDir, "sessions.json"), []map[string]any{
		{
			"sessionId":          sessionID,
			"workspaceDirectory": "file://" + workspace,
		},
	})

	writeJSONL(t, filepath.Join(devDataDir, "tokensGenerated.jsonl"), []map[string]any{
		{
			"timestamp":       "2026-05-01T19:45:22.290Z",
			"eventName":       "tokensGenerated",
			"model":           "gpt-5.2",
			"provider":        "openai",
			"promptTokens":    151,
			"generatedTokens": 36,
		},
	})
	writeJSONL(t, filepath.Join(devDataDir, "chatInteraction.jsonl"), []map[string]any{
		{
			"timestamp":     "2026-05-01T19:45:22.287Z",
			"userAgent":     "Visual Studio Code/1.118.1 (Continue/1.2.22)",
			"eventName":     "chatInteraction",
			"prompt":        "<system>\nskip\n</system>\n<user>\nadd yourself to the readme in this project\n\n",
			"modelName":     "gpt-5.2",
			"modelProvider": "openai",
			"sessionId":     sessionID,
		},
	})
	writeJSONL(t, filepath.Join(devDataDir, "toolUsage.jsonl"), []map[string]any{
		{
			"timestamp":    "2026-05-01T19:45:26.107Z",
			"userAgent":    "Visual Studio Code/1.118.1 (Continue/1.2.22)",
			"eventName":    "toolUsage",
			"functionName": "read_file",
			"toolCallArgs": `{"filepath":"README.md"}`,
			"accepted":     true,
			"succeeded":    true,
		},
	})
	writeJSONL(t, filepath.Join(devDataDir, "editOutcome.jsonl"), []map[string]any{
		{
			"timestamp":     "2026-05-01T19:45:48.207Z",
			"userAgent":     "Visual Studio Code/1.118.1 (Continue/1.2.22)",
			"eventName":     "editOutcome",
			"prompt":        "add yourself to the readme in this project",
			"modelName":     "gpt-5.2",
			"modelProvider": "openai",
			"accepted":      true,
			"lineChange":    2,
			"filepath":      "file://" + editPath,
		},
	})

	parser := ai.Continue{
		After:             time.Date(2026, 5, 1, 19, 45, 22, 281000000, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			editPath: heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 3)

	assert.Equal(t, "Continue 5832a36f-ea52-4eb4-b8f7-16ed2bf063dd", got[0].Entity)
	assert.Equal(t, sessionID, got[0].AISession)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[0].Category)
	assert.Nil(t, got[0].AILineChanges)
	assert.Equal(t, len([]rune("add yourself to the readme in this project")), got[0].AIPromptLength)
	assert.EqualValues(t, 151, got[0].AIInputTokens)
	assert.EqualValues(t, 36, got[0].AIOutputTokens)
	assert.Equal(t, workspace, got[0].ProjectPathOverride)
	require.NotNil(t, got[0].IsWrite)
	assert.False(t, *got[0].IsWrite)
	assert.Contains(t, got[0].UserAgent, "gpt/5.2")
	assert.True(t, strings.Index(got[0].UserAgent, "gpt/5.2") <
		strings.Index(got[0].UserAgent, "Visual Studio Code/1.118.1"))
	assert.Contains(t, got[0].UserAgent, "Visual Studio Code/1.118.1 (Continue/1.2.22)")

	assert.Equal(t, readPath, got[1].Entity)
	assert.Equal(t, sessionID, got[1].AISession)
	assert.Equal(t, heartbeat.FileType, got[1].EntityType)
	require.NotNil(t, got[1].AILineChanges)
	assert.Zero(t, *got[1].AILineChanges)
	require.NotNil(t, got[1].IsWrite)
	assert.False(t, *got[1].IsWrite)

	assert.Equal(t, editPath, got[2].Entity)
	assert.Equal(t, sessionID, got[2].AISession)
	assert.Equal(t, heartbeat.FileType, got[2].EntityType)
	require.NotNil(t, got[2].AILineChanges)
	assert.Equal(t, 2, *got[2].AILineChanges)
	require.NotNil(t, got[2].IsWrite)
	assert.True(t, *got[2].IsWrite)
	assert.Contains(t, got[2].UserAgent, "gpt/5.2")
	assert.True(t, strings.Index(got[2].UserAgent, "gpt/5.2") <
		strings.Index(got[2].UserAgent, "editor/1.2.3"))
	assert.Contains(t, got[2].UserAgent, "editor/1.2.3")
}

func TestContinueParse_NoContinueDir(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := ai.Continue{After: time.Now()}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()

	data, err := json.Marshal(value)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o600))
}

func writeJSONL(t *testing.T, path string, rows []map[string]any) {
	t.Helper()

	var lines []string

	for _, row := range rows {
		data, err := json.Marshal(row)
		require.NoError(t, err)

		lines = append(lines, string(data))
	}

	require.NoError(t, os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600))
}
