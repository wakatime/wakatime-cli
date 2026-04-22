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

func TestCodexParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	now := time.Now()
	transcriptDir := filepath.Join(home, ".codex", "sessions", now.Format("2006"), now.Format("01"), now.Format("02"))
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "rollout-2026-03-28T07-33-13-019d3438-39ae-7fb2-8526-d6c02ba3577c.jsonl")
	copyFile(t, "testdata/codex.jsonl", transcriptPath)

	parser := ai.Codex{
		After:             time.Date(2026, 3, 28, 04, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			"/home/user/projects/wakatime-cli/pkg/ai/claude.go": heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 3)

	assert.Equal(t, "Codex rollout-2026-03-28T07-33-13-019d3438-39ae-7fb2-8526-d6c02ba3577c", got[0].Entity)
	assert.Equal(t, "019d3438-39ae-7fb2-8526-d6c02ba3577c", got[0].AISession)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[0].Category)
	assert.Nil(t, got[0].AILineChanges)
	assert.Equal(t, len([]rune("Please implement the code as described by the comment.")), got[0].AIPromptLength)
	assert.Equal(t, "/root/wakatime-cli", got[0].ProjectPathOverride)
	require.NotNil(t, got[0].IsWrite)
	assert.False(t, *got[0].IsWrite)
	assert.Equal(t, float64(time.Date(2026, 3, 28, 11, 33, 14, 289, time.UTC).Unix()), got[0].Time)
	assert.Contains(t, got[0].UserAgent, "Codex/0.116.0-alpha.1")
	assert.True(
		t,
		strings.Index(got[0].UserAgent, "Codex/0.116.0-alpha.1") <
			strings.Index(got[0].UserAgent, "plugin/0.0.1"),
	)
	assert.Contains(t, got[0].UserAgent, "plugin/0.0.1")

	assert.Equal(t, "Codex rollout-2026-03-28T07-33-13-019d3438-39ae-7fb2-8526-d6c02ba3577c", got[1].Entity)
	assert.Equal(t, heartbeat.AppType, got[1].EntityType)
	assert.Equal(t, "019d3438-39ae-7fb2-8526-d6c02ba3577c", got[1].AISession)
	assert.Nil(t, got[1].AILineChanges)
	assert.Zero(t, got[1].AIPromptLength)
	assert.Equal(t, "/root/wakatime-cli", got[1].ProjectPathOverride)
	require.NotNil(t, got[1].IsWrite)
	assert.False(t, *got[1].IsWrite)
	assert.Zero(t, got[1].AIInputTokens)
	assert.Zero(t, got[1].AIOutputTokens)
	assert.Equal(t, float64(time.Date(2026, 3, 28, 11, 33, 18, 535000000, time.UTC).Unix()), got[1].Time)
	assert.Contains(t, got[1].UserAgent, "Codex/0.116.0-alpha.10")
	assert.True(
		t,
		strings.Index(got[1].UserAgent, "Codex/0.116.0-alpha.10") <
			strings.Index(got[1].UserAgent, "plugin/0.0.1"),
	)
	assert.Contains(t, got[1].UserAgent, "plugin/0.0.1")

	assert.Equal(t, "/home/user/projects/wakatime-cli/pkg/ai/claude.go", got[2].Entity)
	assert.Equal(t, "019d3438-39ae-7fb2-8526-d6c02ba3577c", got[2].AISession)
	assert.Equal(t, heartbeat.FileType, got[2].EntityType)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[2].Category)
	require.NotNil(t, got[2].AILineChanges)
	assert.Equal(t, 20, *got[2].AILineChanges)
	assert.Zero(t, got[2].AIPromptLength)
	require.NotNil(t, got[2].IsWrite)
	assert.True(t, *got[2].IsWrite)
	assert.Equal(t, float64(time.Date(2026, 3, 28, 11, 33, 34, 952, time.UTC).Unix()), got[2].Time)
	assert.Contains(t, got[2].UserAgent, "Codex/0.116.0-alpha.10")
	assert.True(
		t,
		strings.Index(got[2].UserAgent, "Codex/0.116.0-alpha.10") <
			strings.Index(got[2].UserAgent, "editor/1.2.3"),
	)
	assert.Contains(t, got[2].UserAgent, "editor/1.2.3")
}

func TestCodexParse_ParsesTranscriptFromPreviousDayFolder(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".codex", "sessions", "2026", "03", "27")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "rollout-2026-03-28T07-33-13-019d3438-39ae-7fb2-8526-d6c02ba3577c.jsonl")
	copyFile(t, "testdata/codex.jsonl", transcriptPath)

	now := time.Now()
	require.NoError(t, os.Chtimes(transcriptPath, now, now))

	parser := ai.Codex{
		After: time.Date(2026, 3, 28, 11, 0, 0, 0, time.UTC),
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 3)
	assert.Equal(t, "/home/user/projects/wakatime-cli/pkg/ai/claude.go", got[2].Entity)
}

func TestCodexParse_NoCodexSessionsDir(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := ai.Codex{After: time.Now()}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestCodexParse_RetainsCwdFromSkippedSessionMetaLine(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	now := time.Now()
	transcriptDir := filepath.Join(home, ".codex", "sessions", now.Format("2006"), now.Format("01"), now.Format("02"))
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-03-28T11:00:00Z","type":"session_meta",`,
			`"payload":{"cwd":"/workspace/project","cli_version":"0.116.0-alpha.10"}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-03-28T11:33:18Z","type":"message",`,
			`"payload":{"type":"message","role":"assistant",`,
			`"content":[{"type":"output_text","text":"I am on it"}]}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Codex{
		After: time.Date(2026, 3, 28, 11, 30, 0, 0, time.UTC),
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "/workspace/project", got[0].ProjectPathOverride)
}

func TestCodexParse_SkipsHarnessInputWhenCalculatingPromptLength(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	now := time.Now()
	transcriptDir := filepath.Join(home, ".codex", "sessions", now.Format("2006"), now.Format("01"), now.Format("02"))
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-04-15T04:14:05Z","type":"session_meta",`,
			`"payload":{"cwd":"/workspace/project","cli_version":"0.119.0-alpha.28"}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-04-15T04:14:05Z","type":"response_item",`,
			`"payload":{"type":"message","role":"user","content":[`,
			`{"type":"input_text","text":"<environment_context>\n  <cwd>/workspace/project</cwd>\n</environment_context>"}`,
			`]}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-04-15T04:14:06Z","type":"response_item",`,
			`"payload":{"type":"message","role":"user","content":[`,
			`{"type":"input_text","text":"Add a unit test for this line"}`,
			`]}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Codex{
		After: time.Date(2026, 4, 15, 4, 0, 0, 0, time.UTC),
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, len([]rune("Add a unit test for this line")), got[0].AIPromptLength)
	assert.Equal(t, "/workspace/project", got[0].ProjectPathOverride)
}

func TestCodexParse_StripsBundledHarnessPrefixBeforeCountingPromptLength(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	now := time.Now()
	transcriptDir := filepath.Join(home, ".codex", "sessions", now.Format("2006"), now.Format("01"), now.Format("02"))
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	userPrompt := strings.Join([]string{
		"<environment_context>",
		"  <cwd>/Users/user/git/vim-wakatime</cwd>",
		"</environment_context>",
		"# Context from my IDE setup:",
		"",
		"## Active file: lua/wakatime/init.lua",
		"",
		"## Open tabs:",
		"- init.lua: lua/wakatime/init.lua",
		"",
		"## My request for Codex:",
		"look for any divergences in the new Lua plugin vs the old VimL plugin, " +
			"or any bugs in the Lua plugin since it's not as tested as the older plugin.",
		"",
	}, "\n")
	expectedPrompt := "look for any divergences in the new Lua plugin vs the old VimL plugin, " +
		"or any bugs in the Lua plugin since it's not as tested as the older plugin."
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-04-21T10:57:40Z","type":"session_meta",`,
			`"payload":{"cwd":"/Users/user/git/vim-wakatime","cli_version":"0.120.0-alpha.1"}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-04-21T10:57:40Z","type":"response_item",`,
			`"payload":{"type":"message","role":"developer","content":[`,
			`{"type":"input_text","text":"<permissions instructions>\nblocked\n</permissions instructions>"}`,
			`]}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-04-21T10:57:40Z","type":"response_item",`,
			`"payload":{"type":"message","role":"user","content":[`,
			`{"type":"input_text","text":` + jsonString(userPrompt) + `}`,
			`]}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Codex{
		After: time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC),
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, len([]rune(expectedPrompt)), got[0].AIPromptLength)
	assert.Equal(t, "/Users/user/git/vim-wakatime", got[0].ProjectPathOverride)
}

func TestCodexParse_StripsVSCodePrefixFromUserMessage(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	now := time.Now()
	transcriptDir := filepath.Join(home, ".codex", "sessions", now.Format("2006"), now.Format("01"), now.Format("02"))
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-04-16T00:02:25Z","type":"session_meta",`,
			`"payload":{"cwd":"/workspace/project","cli_version":"0.119.0-alpha.28"}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-04-16T00:02:25Z","type":"event_msg",`,
			`"payload":{"type":"user_message","message":"# Context from my IDE setup:\n\n` +
				`## Active file: wakatime-cli/static/css/index.less\n\n## Open tabs:\n` +
				`- index.less: wakatime-cli/static/css/index.less\n\n## My request for Codex:\n` +
				`remove scrolling horizontally\n"}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Codex{
		After: time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC),
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, len([]rune("remove scrolling horizontally")), got[0].AIPromptLength)
	assert.Equal(t, "/workspace/project", got[0].ProjectPathOverride)
}

func TestCodexParse_RolloutFixtureIncludesExpectedHeartbeatAttributes(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	now := time.Now()
	transcriptDir := filepath.Join(home, ".codex", "sessions", now.Format("2026"), now.Format("04"), now.Format("15"))
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcript := "rollout-2026-04-15T18-46-36-019d9353-333c-7c41-a909-26f5c6221a5a.jsonl"
	transcriptPath := filepath.Join(transcriptDir, transcript)
	copyFile(t, filepath.Join("testdata", transcript), transcriptPath)

	parser := ai.Codex{
		After:             time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}

	heartbeats, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, heartbeats, 9)

	i := 0
	assert.EqualValues(t, 1776297745, heartbeats[i].Time)
	assert.Equal(t, heartbeat.AppType, heartbeats[i].EntityType)
	assert.Equal(t, "Codex rollout-2026-04-15T18-46-36-019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.False(t, *heartbeats[i].IsWrite)
	assert.Zero(t, heartbeats[i].AIInputTokens)
	assert.Zero(t, heartbeats[i].AIOutputTokens)
	assert.Equal(t, 29, heartbeats[i].AIPromptLength)
	assert.Nil(t, heartbeats[i].AILineChanges)
	assert.Equal(t, "/Users/user/git/wakatime-cli", heartbeats[i].ProjectPathOverride)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)

	// 1
	i++
	assert.EqualValues(t, 1776297755, heartbeats[i].Time)
	assert.Equal(t, heartbeat.AppType, heartbeats[i].EntityType)
	assert.Equal(t, "Codex rollout-2026-04-15T18-46-36-019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.False(t, *heartbeats[i].IsWrite)
	assert.EqualValues(t, 1, heartbeats[i].AIInputTokens)
	assert.EqualValues(t, 2, heartbeats[i].AIOutputTokens)
	assert.Zero(t, heartbeats[i].AIPromptLength)
	assert.Nil(t, heartbeats[i].AILineChanges)
	assert.Equal(t, "/Users/user/git/wakatime-cli", heartbeats[i].ProjectPathOverride)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)

	// 2
	i++
	assert.EqualValues(t, 1776297759, heartbeats[i].Time)
	assert.Equal(t, heartbeat.AppType, heartbeats[i].EntityType)
	assert.Equal(t, "Codex rollout-2026-04-15T18-46-36-019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.False(t, *heartbeats[i].IsWrite)
	assert.EqualValues(t, 105133, heartbeats[i].AIInputTokens)
	assert.EqualValues(t, 477, heartbeats[i].AIOutputTokens)
	assert.Zero(t, heartbeats[i].AIPromptLength)
	assert.Nil(t, heartbeats[i].AILineChanges)
	assert.Equal(t, "/Users/user/git/wakatime-cli", heartbeats[i].ProjectPathOverride)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)

	// 3
	i++
	assert.EqualValues(t, 1776297781, heartbeats[i].Time)
	assert.Equal(t, heartbeat.AppType, heartbeats[i].EntityType)
	assert.Equal(t, "Codex rollout-2026-04-15T18-46-36-019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.False(t, *heartbeats[i].IsWrite)
	assert.EqualValues(t, 217197, heartbeats[i].AIInputTokens)
	assert.EqualValues(t, 599, heartbeats[i].AIOutputTokens)
	assert.Zero(t, heartbeats[i].AIPromptLength)
	assert.Nil(t, heartbeats[i].AILineChanges)
	assert.Equal(t, "/Users/user/git/wakatime-cli", heartbeats[i].ProjectPathOverride)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)

	// 4
	i++
	assert.EqualValues(t, 1776297784, heartbeats[i].Time)
	assert.Equal(t, heartbeat.FileType, heartbeats[i].EntityType)
	assert.Equal(t, "/Users/user/git/wakatime-cli/templates/index.html", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.True(t, *heartbeats[i].IsWrite)
	assert.Zero(t, heartbeats[i].AIInputTokens)
	assert.Zero(t, heartbeats[i].AIOutputTokens)
	assert.Zero(t, heartbeats[i].AIPromptLength)
	assert.EqualValues(t, -1, *heartbeats[i].AILineChanges)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)

	// 5
	i++
	assert.EqualValues(t, 1776297787, heartbeats[i].Time)
	assert.Equal(t, heartbeat.AppType, heartbeats[i].EntityType)
	assert.Equal(t, "Codex rollout-2026-04-15T18-46-36-019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.False(t, *heartbeats[i].IsWrite)
	assert.EqualValues(t, 109309, heartbeats[i].AIInputTokens)
	assert.EqualValues(t, 843, heartbeats[i].AIOutputTokens)
	assert.Zero(t, heartbeats[i].AIPromptLength)
	assert.Nil(t, heartbeats[i].AILineChanges)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)

	// 6
	i++
	assert.EqualValues(t, 1776297787, heartbeats[i].Time)
	assert.Equal(t, heartbeat.AppType, heartbeats[i].EntityType)
	assert.Equal(t, "Codex rollout-2026-04-15T18-46-36-019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.False(t, *heartbeats[i].IsWrite)
	assert.Zero(t, heartbeats[i].AIInputTokens)
	assert.Zero(t, heartbeats[i].AIOutputTokens)
	assert.Equal(t, 10, heartbeats[i].AIPromptLength)
	assert.Nil(t, heartbeats[i].AILineChanges)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)

	// 7
	i++
	assert.EqualValues(t, 1776297789, heartbeats[i].Time)
	assert.Equal(t, heartbeat.FileType, heartbeats[i].EntityType)
	assert.Equal(t, "/Users/user/git/wakatime-cli/static/css/index.less", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.True(t, *heartbeats[i].IsWrite)
	assert.Zero(t, heartbeats[i].AIInputTokens)
	assert.Zero(t, heartbeats[i].AIOutputTokens)
	assert.Zero(t, heartbeats[i].AIPromptLength)
	assert.Equal(t, 23, *heartbeats[i].AILineChanges)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)

	// 8
	i++
	assert.EqualValues(t, 1776297792, heartbeats[i].Time)
	assert.Equal(t, heartbeat.AppType, heartbeats[i].EntityType)
	assert.Equal(t, "Codex rollout-2026-04-15T18-46-36-019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.False(t, *heartbeats[i].IsWrite)
	assert.EqualValues(t, 110200, heartbeats[i].AIInputTokens)
	assert.EqualValues(t, 223, heartbeats[i].AIOutputTokens)
	assert.Zero(t, heartbeats[i].AIPromptLength)
	assert.Nil(t, heartbeats[i].AILineChanges)
	assert.Equal(t, "/Users/user/git/wakatime-cli", heartbeats[i].ProjectPathOverride)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)
}

func copyFile(t *testing.T, source, destination string) {
	input, err := os.ReadFile(source)
	require.NoError(t, err)

	err = os.WriteFile(destination, input, 0600)
	require.NoError(t, err)
}

func jsonString(value string) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	return string(encoded)
}
