package ai_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestKiroParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	root := kiroStorageRoot(t, home)
	createKiroSession(t, root)
	createKiroExecutionLogs(t, root)

	got, err := ai.Kiro{
		After:             time.UnixMilli(1777312050000),
		FallbackUserAgent: "plugin/0.0.1",
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 8)

	projectPath := "/Users/user/git/wakatime-cli"
	authorsPath := projectPath + "/AUTHORS"
	readmePath := projectPath + "/README.md"
	usagePath := projectPath + "/USAGE.md"

	firstPrompt := "Add your name to the AUTHORS file in this repo"
	secondPrompt := "Now make two changes to the README.md in different lines, " +
		"and also modify the USAGE.md with info about how to use Kiro"

	assert.Equal(t, "Kiro c2618220-b591-4431-8f14-fcd7ae3e6f56", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "c2618220-b591-4431-8f14-fcd7ae3e6f56", got[0].AISession)
	assert.Equal(t, utf8.RuneCountInString(firstPrompt), got[0].AIPromptLength)
	assert.Equal(t, projectPath, got[0].ProjectPathOverride)
	assert.Contains(t, got[0].UserAgent, "Kiro")

	assert.Equal(t, authorsPath, got[1].Entity)
	assert.Equal(t, heartbeat.FileType, got[1].EntityType)
	assert.Equal(t, heartbeat.PointerTo(false), got[1].IsWrite)
	assert.Nil(t, got[1].AILineChanges)

	assert.Equal(t, authorsPath, got[2].Entity)
	assert.Equal(t, heartbeat.PointerTo(true), got[2].IsWrite)
	require.NotNil(t, got[2].AILineChanges)
	assert.Equal(t, 0, *got[2].AILineChanges)

	assert.Equal(t, "Kiro c2618220-b591-4431-8f14-fcd7ae3e6f56", got[3].Entity)
	assert.Equal(t, utf8.RuneCountInString(secondPrompt), got[3].AIPromptLength)

	assert.Equal(t, readmePath, got[4].Entity)
	assert.Equal(t, heartbeat.PointerTo(false), got[4].IsWrite)
	assert.Equal(t, usagePath, got[5].Entity)
	assert.Equal(t, heartbeat.PointerTo(false), got[5].IsWrite)

	assert.Equal(t, readmePath, got[6].Entity)
	assert.Equal(t, heartbeat.PointerTo(true), got[6].IsWrite)
	require.NotNil(t, got[6].AILineChanges)
	assert.Equal(t, 1, *got[6].AILineChanges)

	assert.Equal(t, usagePath, got[7].Entity)
	assert.Equal(t, heartbeat.PointerTo(true), got[7].IsWrite)
	require.NotNil(t, got[7].AILineChanges)
	assert.Equal(t, 2, *got[7].AILineChanges)
}

func TestKiroParse_NoStorage(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := ai.Kiro{After: time.Now()}.Parse(context.Background())
	require.NoError(t, err)
	assert.Empty(t, got)
}

func kiroStorageRoot(t *testing.T, home string) string {
	t.Helper()

	root := filepath.Join(
		home,
		"Library",
		"Application Support",
		"Kiro",
		"User",
		"globalStorage",
		"kiro.kiroagent",
	)
	require.NoError(t, os.MkdirAll(root, 0o755))

	return root
}

func createKiroSession(t *testing.T, root string) {
	t.Helper()

	sessionDir := filepath.Join(root, "workspace-sessions", "workspace-hash")
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	writeKiroJSON(t, filepath.Join(sessionDir, "c2618220-b591-4431-8f14-fcd7ae3e6f56.json"), map[string]any{
		"sessionId":          "c2618220-b591-4431-8f14-fcd7ae3e6f56",
		"workspacePath":      "/Users/user/git/wakatime-cli",
		"workspaceDirectory": "/Users/user/git/wakatime-cli",
		"history": []map[string]any{
			kiroUserMessage("Add your name to the AUTHORS file in this repo"),
			kiroAssistantMessage("4490c49d-f38f-4d24-907e-d14daa4224cc"),
			kiroUserMessage(
				"Now make two changes to the README.md in different lines, " +
					"and also modify the USAGE.md with info about how to use Kiro",
			),
			kiroAssistantMessage("f1164130-26b5-43e9-88bd-eebfcb527260"),
		},
	})
}

func createKiroExecutionLogs(t *testing.T, root string) {
	t.Helper()

	executionDir := filepath.Join(root, "project-hash", "session-hash")
	require.NoError(t, os.MkdirAll(executionDir, 0o755))

	writeKiroJSON(t, filepath.Join(executionDir, "first"), map[string]any{
		"executionId":   "4490c49d-f38f-4d24-907e-d14daa4224cc",
		"chatSessionId": "c2618220-b591-4431-8f14-fcd7ae3e6f56",
		"startTime":     int64(1777312055796),
		"actions": []map[string]any{
			kiroReadFilesAction(1777312061066, "AUTHORS"),
			kiroReplaceAction(
				1777312065416,
				"AUTHORS",
				"Patches and Suggestions\n-----------------------\n\n\n",
				"Patches and Suggestions\n-----------------------\n\n- Kiro <kiro@kiro.dev>\n",
			),
		},
	})

	writeKiroJSON(t, filepath.Join(executionDir, "second"), map[string]any{
		"executionId":   "f1164130-26b5-43e9-88bd-eebfcb527260",
		"chatSessionId": "c2618220-b591-4431-8f14-fcd7ae3e6f56",
		"startTime":     int64(1777312102986),
		"actions": []map[string]any{
			kiroReadFilesAction(1777312105884, "README.md", "USAGE.md"),
			kiroReplaceAction(1777312110408, "README.md", "line one\n", "line one\nline two\n"),
			kiroReplaceAction(1777312117829, "USAGE.md", "intro\n", "intro\nline two\nline three\n"),
		},
	})
}

func kiroUserMessage(text string) map[string]any {
	return map[string]any{
		"message": map[string]any{
			"role": "user",
			"content": []map[string]string{
				{
					"type": "text",
					"text": text,
				},
			},
		},
	}
}

func kiroAssistantMessage(executionID string) map[string]any {
	return map[string]any{
		"executionId": executionID,
		"message": map[string]any{
			"role":    "assistant",
			"content": "On it.",
		},
	}
}

func kiroReadFilesAction(emittedAt int64, paths ...string) map[string]any {
	files := make([]map[string]string, 0, len(paths))
	for _, path := range paths {
		files = append(files, map[string]string{"path": path})
	}

	return map[string]any{
		"actionId":    "read-files",
		"actionType":  "readFiles",
		"actionState": "Accepted",
		"emittedAt":   emittedAt,
		"input":       map[string]any{"files": files},
	}
}

func kiroReplaceAction(emittedAt int64, filePath string, original string, modified string) map[string]any {
	return map[string]any{
		"actionId":    "replace-" + filePath,
		"actionType":  "replace",
		"actionState": "Accepted",
		"emittedAt":   emittedAt,
		"input": map[string]any{
			"file":            filePath,
			"local":           "file:///Users/user/git/wakatime-cli/" + filePath,
			"originalContent": original,
			"modifiedContent": modified,
		},
	}
}

func writeKiroJSON(t *testing.T, path string, value any) {
	t.Helper()

	raw, err := json.Marshal(value)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, raw, 0o600))
}
