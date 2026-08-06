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

func TestRooCodeParse(t *testing.T) {
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
		"RooVeterinaryInc.roo-cline",
		"tasks",
		"1736395424460",
	)
	require.NoError(t, os.MkdirAll(taskDir, 0o755))

	copyFile(t, "testdata/roo_ui_messages.json", filepath.Join(taskDir, "ui_messages.json"))

	parser := ai.RooCode{
		After:             time.Date(2025, 2, 19, 17, 20, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			"/workspace/project/main.go": heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 3)

	assert.Equal(t, "Roo Code 1736395424460", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "1736395424460", got[0].AISession)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[0].Category)
	assert.Equal(t, "/workspace/project", got[0].ProjectPathOverride)
	assert.Equal(t, len([]rune("Refactor this function")), got[0].AIPromptLength)
	assert.EqualValues(t, 121, got[0].AIInputTokens)
	assert.EqualValues(t, 2, got[0].AICachedInputTokens)
	assert.EqualValues(t, 30, got[0].AIOutputTokens)
	require.NotNil(t, got[0].IsWrite)
	assert.False(t, *got[0].IsWrite)
	assert.Equal(t, "plugin/0.0.1", got[0].UserAgent)

	assert.Equal(t, "/workspace/project/main.go", got[1].Entity)
	assert.Equal(t, heartbeat.FileType, got[1].EntityType)
	assert.Equal(t, "1736395424460", got[1].AISession)
	require.NotNil(t, got[1].AILineChanges)
	assert.Equal(t, 1, *got[1].AILineChanges)
	require.NotNil(t, got[1].IsWrite)
	assert.True(t, *got[1].IsWrite)
	assert.NotContains(t, got[1].UserAgent, "Roo Code/")
	assert.Contains(t, got[1].UserAgent, "editor/1.2.3")

	assert.Equal(t, "Roo Code 1736395424460", got[2].Entity)
	assert.Equal(t, heartbeat.AppType, got[2].EntityType)
	assert.Zero(t, got[2].AIPromptLength)
	assert.EqualValues(t, 180, got[2].AIInputTokens)
	assert.EqualValues(t, 40, got[2].AIOutputTokens)
	assert.Equal(t, "/workspace/project", got[2].ProjectPathOverride)
}

func TestRooCodeParse_InclusiveInputTokens(t *testing.T) {
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
		"RooVeterinaryInc.roo-cline",
		"tasks",
		"1762000000000",
	)
	require.NoError(t, os.MkdirAll(taskDir, 0o755))

	request, err := json.Marshal(map[string]any{
		"request":     "<task>Update the parser</task>\n# Current Working Directory (/workspace/project)",
		"tokensIn":    120,
		"tokensOut":   30,
		"cacheWrites": 5,
		"cacheReads":  40,
		"apiProtocol": "anthropic",
	})
	require.NoError(t, err)

	messages, err := json.Marshal([]map[string]any{
		{
			"ts":   int64(1762000000000),
			"type": "say",
			"say":  "api_req_started",
			"text": string(request),
		},
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(taskDir, "ui_messages.json"), messages, 0o600))

	got, err := (ai.RooCode{
		After:             time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.EqualValues(t, 80, got[0].AIInputTokens)
	assert.EqualValues(t, 40, got[0].AICachedInputTokens)
	assert.EqualValues(t, 30, got[0].AIOutputTokens)
}

func TestRooCodeParse_NoTasksDir(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := ai.RooCode{After: time.Now()}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestRooCodeParse_SkipsFailedTool(t *testing.T) {
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
		"RooVeterinaryInc.roo-cline",
		"tasks",
		"1736395424461",
	)
	require.NoError(t, os.MkdirAll(taskDir, 0o755))

	textJSON := func(v any) *string {
		contents, err := json.Marshal(v)
		require.NoError(t, err)

		value := string(contents)

		return &value
	}

	sayAPIReqStarted := "api_req_started"
	askTool := "tool"
	uiMessages := []map[string]any{
		{
			"ts":   int64(1740000001000),
			"type": "say",
			"say":  &sayAPIReqStarted,
			"text": textJSON(map[string]any{
				"request":   "<task>Fix the file</task>\n# Current Working Directory (/workspace/project)",
				"tokensIn":  90,
				"tokensOut": 20,
			}),
		},
		{
			"ts":   int64(1740000002000),
			"type": "ask",
			"ask":  &askTool,
			"text": textJSON(map[string]any{
				"tool": "appliedDiff",
				"path": "main.go",
				"diff": "<<<<<<< SEARCH\nold\n=======\nnew\n>>>>>>> REPLACE",
			}),
		},
		{
			"ts":   int64(1740000003000),
			"type": "say",
			"say":  &sayAPIReqStarted,
			"text": textJSON(map[string]any{
				"request":   "[apply_diff for 'main.go'] Result:\n\nUnable to apply diff to file",
				"tokensIn":  120,
				"tokensOut": 30,
			}),
		},
	}

	contents, err := json.Marshal(uiMessages)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(taskDir, "ui_messages.json"), contents, 0o600))

	got, err := (ai.RooCode{
		After:             time.Date(2025, 2, 19, 17, 20, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, heartbeat.AppType, got[1].EntityType)
}
