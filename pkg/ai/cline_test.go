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

func TestClineParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	taskDir := filepath.Join(
		home,
		".config",
		"Code",
		"User",
		"globalStorage",
		"saoudrizwan.claude-dev",
		"tasks",
		"1740000000000",
	)
	require.NoError(t, os.MkdirAll(taskDir, 0o755))

	textJSON := func(v any) *string {
		contents, err := json.Marshal(v)
		require.NoError(t, err)

		value := string(contents)

		return &value
	}

	sayAPIReqStarted := "api_req_started"
	sayTool := "tool"

	uiMessages := []map[string]any{
		{
			"ts":   int64(1740000001000),
			"type": "say",
			"say":  &sayAPIReqStarted,
			"text": textJSON(map[string]any{
				"request":   "<task>Fix the auth bug</task>\n# Current Working Directory (/workspace/project)",
				"tokensIn":  90,
				"tokensOut": 20,
			}),
		},
		{
			"ts":   int64(1740000002000),
			"type": "say",
			"say":  &sayTool,
			"text": textJSON(map[string]any{
				"tool": "editedExistingFile",
				"path": "/workspace/project/main.go",
				"diff": "<<<<<<< SEARCH\nold\n=======\nnew\n>>>>>>> REPLACE",
			}),
		},
		{
			"ts":   int64(1740000003000),
			"type": "say",
			"say":  &sayTool,
			"text": textJSON(map[string]any{
				"tool":    "newFileCreated",
				"path":    "/workspace/project/new.go",
				"content": "one\ntwo",
			}),
		},
		{
			"ts":   int64(1740000004000),
			"type": "say",
			"say":  &sayTool,
			"text": textJSON(map[string]any{
				"tool": "readFile",
				"path": "/workspace/project/read.go",
			}),
		},
	}

	contents, err := json.Marshal(uiMessages)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(taskDir, "ui_messages.json"), contents, 0o644))

	parser := ai.Cline{
		After:             time.Date(2025, 2, 19, 17, 19, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			"/workspace/project/main.go": heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 4)

	assert.Equal(t, "Cline 1740000000000", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "1740000000000", got[0].AISession)
	assert.Equal(t, "/workspace/project", got[0].ProjectPathOverride)
	assert.Equal(t, len([]rune("Fix the auth bug")), got[0].AIPromptLength)
	assert.EqualValues(t, 90, got[0].AIInputTokens)
	assert.EqualValues(t, 20, got[0].AIOutputTokens)
	assert.Contains(t, got[0].UserAgent, "Cline")

	assert.Equal(t, "/workspace/project/main.go", got[1].Entity)
	require.NotNil(t, got[1].AILineChanges)
	assert.Equal(t, 0, *got[1].AILineChanges)
	require.NotNil(t, got[1].IsWrite)
	assert.True(t, *got[1].IsWrite)
	assert.Contains(t, got[1].UserAgent, "Cline")
	assert.Contains(t, got[1].UserAgent, "editor/1.2.3")

	assert.Equal(t, "/workspace/project/new.go", got[2].Entity)
	require.NotNil(t, got[2].AILineChanges)
	assert.Equal(t, 2, *got[2].AILineChanges)
	require.NotNil(t, got[2].IsWrite)
	assert.True(t, *got[2].IsWrite)

	assert.Equal(t, "/workspace/project/read.go", got[3].Entity)
	require.NotNil(t, got[3].AILineChanges)
	assert.Equal(t, 0, *got[3].AILineChanges)
	require.NotNil(t, got[3].IsWrite)
	assert.False(t, *got[3].IsWrite)
}

func TestClineParse_NoTasksDir(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := ai.Cline{After: time.Now()}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}
