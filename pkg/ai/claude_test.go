package ai_test

import (
	"context"
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

func TestClaudeParse(t *testing.T) {
	ctx := context.Background()
	expectedPrompt := "look for any bugs in the AI parsers. " +
		"Feel free to read local jsonl log files to make sure schemas are as expected"

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	require.NoError(t, os.WriteFile(
		filepath.Join(home, ".claude.json"),
		[]byte(`{"oauthAccount":{"organizationType":"claude_pro"}}`),
		0o644,
	))

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project", "subagents")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "agent-worker.jsonl")
	transcript := strings.Join([]string{
		"{\"timestamp\":\"2026-03-18T10:00:00Z\",\"version\":\"2.1.44\"," +
			"\"toolUseResult\":{\"filePath\":\"/tmp/skip.txt\"," +
			"\"structuredPatch\":[{\"oldLines\":1,\"newLines\":2}]},\"usage\":{\"total_tokens\":5}}",
		"not json",
		"{\"timestamp\":\"2026-03-18T11:30:00Z\"}",
		strings.Join([]string{
			`{"timestamp":"2026-03-18T11:45:00Z","sessionId":"claude-session","version":"2.1.45",`,
			`"cwd":"/tmp","isSidechain":false,"type":"user","message":{"role":"user","content":[`,
			`{"type":"text","text":"<ide_opened_file>ignore this harness text</ide_opened_file>"},`,
			`{"type":"text","text":"` + expectedPrompt + `"}`,
			`]}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-03-18T11:50:00Z","sessionId":"claude-session","version":"2.1.45",`,
			`"cwd":"/tmp","isSidechain":true,"type":"user","message":{"role":"user","content":[`,
			`{"type":"text","text":"this sidechain prompt should not be counted"}`,
			`]}}`,
		}, ""),
		"{\"timestamp\":\"2026-03-18T12:00:00Z\",\"sessionId\":\"claude-session\",\"version\":\"2.1.45\"," +
			"\"toolUseResult\":{\"filePath\":\"/tmp/edited.go\"," +
			"\"structuredPatch\":[{\"oldLines\":3,\"newLines\":5}," +
			"{\"oldLines\":4,\"newLines\":1}]},\"message\":{\"usage\":{\"input_tokens\":11,\"output_tokens\":12}}}",
		strings.Join([]string{
			`{"timestamp":"2026-03-18T12:15:00Z","sessionId":"claude-session",`,
			`"toolUseResult":"plain string result"}`,
		}, ""),
		"{\"timestamp\":\"2026-03-18T12:20:00Z\",\"sessionId\":\"claude-session\",\"toolUseResult\":{" +
			"\"agentId\":\"worker-1\",\"agentType\":\"worker\",\"content\":[{" +
			"\"type\":\"text\",\"text\":\"summary block\"}]},\"message\":{" +
			"\"usage\":{\"input_tokens\":4,\"output_tokens\":1}}}",
		"{\"timestamp\":\"2026-03-18T12:22:00Z\",\"sessionId\":\"claude-session\"," +
			"\"toolUseResult\":[{\"type\":\"text\",\"text\":\"mcp summary\"}]}",
		"{\"timestamp\":\"2026-03-18T12:25:00Z\",\"sessionId\":\"claude-session\",\"toolUseResult\":{" +
			"\"filePath\":\"/tmp/array.go\",\"content\":[{" +
			"\"type\":\"text\",\"text\":\"first\\nsecond\"},{" +
			"\"type\":\"text\",\"text\":\"third\"}]},\"message\":{" +
			"\"usage\":{\"input_tokens\":5,\"output_tokens\":3}}}",
		"{\"timestamp\":\"2026-03-18T12:30:00Z\",\"sessionId\":\"claude-session\",\"toolUseResult\":{" +
			"\"filePath\":\"/tmp/new.go\",\"content\":\"first\\nsecond\\nthird\"}," +
			"\"message\":{\"usage\":{\"input_tokens\":6,\"output_tokens\":3}}}",
		"{\"timestamp\":\"2026-03-18T13:00:00Z\",\"sessionId\":\"claude-session\",\"toolUseResult\":{" +
			"\"filePath\":\"/tmp/read.go\",\"content\":\"existing\",\"originalFile\":\"before\"}," +
			"\"message\":{\"usage\":{\"input_tokens\":2,\"output_tokens\":1}}}",
		"{\"timestamp\":\"2026-03-18T13:30:00Z\",\"sessionId\":\"claude-session\",\"toolUseResult\":{" +
			"\"filePath\":\"/tmp/empty.go\",\"structuredPatch\":[]}," +
			"\"message\":{\"usage\":{\"input_tokens\":1,\"output_tokens\":1}}}",
		"{\"timestamp\":\"2026-03-18T13:45:00Z\",\"sessionId\":\"claude-session\",\"toolUseResult\":{" +
			"\"filePath\":\"/tmp/same-lines.go\",\"oldString\":\"before\",\"newString\":\"after\"," +
			"\"originalFile\":\"before\",\"structuredPatch\":[{\"oldLines\":1,\"newLines\":1}]}," +
			"\"message\":{\"usage\":{\"input_tokens\":3,\"output_tokens\":2}}}",
		"{\"timestamp\":\"2026-03-18T14:00:00Z\",\"sessionId\":\"claude-session\",\"toolUseResult\":{" +
			"\"structuredPatch\":[{\"oldLines\":1,\"newLines\":1}]}," +
			"\"message\":{\"usage\":{\"input_tokens\":8,\"output_tokens\":2}}}",
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Claude{
		After:             time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			"/tmp/edited.go": heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 9)

	for _, heartbeat := range got {
		assert.Equal(t, "pro", heartbeat.AISubscriptionPlan)
	}

	assert.Equal(t, "Claude agent-worker", got[0].Entity)
	assert.Equal(t, "claude-session", got[0].AISession)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Nil(t, got[0].AILineChanges)
	assert.Equal(t, len([]rune(expectedPrompt)), got[0].AIPromptLength)
	assert.Equal(t, "/tmp", got[0].ProjectPathOverride)
	assert.Zero(t, got[0].AIInputTokens)
	assert.Zero(t, got[0].AIOutputTokens)
	require.NotNil(t, got[0].IsWrite)
	assert.False(t, *got[0].IsWrite)
	assert.Equal(t, float64(time.Date(2026, 3, 18, 11, 45, 0, 0, time.UTC).Unix()), got[0].Time)
	assert.Equal(
		t,
		"claude-code/2.1.45 plugin/0.0.1",
		got[0].UserAgent,
	)
	assert.Contains(t, got[0].UserAgent, "claude-code/2.1.45")

	assert.Equal(t, "/tmp/edited.go", got[1].Entity)
	assert.Equal(t, "claude-session", got[1].AISession)
	assert.Equal(t, heartbeat.FileType, got[1].EntityType)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[1].Category)
	require.NotNil(t, got[1].AILineChanges)
	assert.Equal(t, -1, *got[1].AILineChanges)
	assert.Zero(t, got[1].AIPromptLength)
	assert.Equal(t, int64(11), got[1].AIInputTokens)
	assert.Equal(t, int64(12), got[1].AIOutputTokens)
	require.NotNil(t, got[1].IsWrite)
	assert.True(t, *got[1].IsWrite)
	assert.Equal(t, float64(time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC).Unix()), got[1].Time)
	assert.Equal(
		t,
		"claude-code/2.1.45 "+heartbeat.UserAgent(ctx, "editor/1.2.3"),
		got[1].UserAgent,
	)
	assert.Contains(t, got[1].UserAgent, "claude-code/2.1.45")

	assert.Equal(t, "Claude agent-worker", got[2].Entity)
	assert.Equal(t, "claude-session", got[2].AISession)
	assert.Equal(t, heartbeat.AppType, got[2].EntityType)
	assert.Nil(t, got[2].AILineChanges)
	assert.Zero(t, got[2].AIPromptLength)
	assert.Equal(t, "/tmp", got[2].ProjectPathOverride)
	assert.Zero(t, got[2].AIInputTokens)
	assert.Zero(t, got[2].AIOutputTokens)
	require.NotNil(t, got[2].IsWrite)
	assert.False(t, *got[2].IsWrite)

	assert.Equal(t, "Claude agent-worker", got[3].Entity)
	assert.Equal(t, "claude-session", got[3].AISession)
	assert.Equal(t, heartbeat.AppType, got[3].EntityType)
	assert.Nil(t, got[3].AILineChanges)
	assert.Zero(t, got[3].AIPromptLength)
	assert.Equal(t, "/tmp", got[3].ProjectPathOverride)
	assert.Zero(t, got[3].AIInputTokens)
	assert.Zero(t, got[3].AIOutputTokens)
	require.NotNil(t, got[3].IsWrite)
	assert.False(t, *got[3].IsWrite)

	assert.Equal(t, "/tmp/array.go", got[4].Entity)
	assert.Equal(t, "claude-session", got[4].AISession)
	require.NotNil(t, got[4].AILineChanges)
	assert.Equal(t, 3, *got[4].AILineChanges)
	assert.Zero(t, got[4].AIPromptLength)
	assert.Equal(t, int64(5), got[4].AIInputTokens)
	assert.Equal(t, int64(3), got[4].AIOutputTokens)
	require.NotNil(t, got[4].IsWrite)
	assert.True(t, *got[4].IsWrite)

	assert.Equal(t, "/tmp/new.go", got[5].Entity)
	assert.Equal(t, "claude-session", got[5].AISession)
	require.NotNil(t, got[5].AILineChanges)
	assert.Equal(t, 3, *got[5].AILineChanges)
	assert.Zero(t, got[5].AIPromptLength)
	assert.Equal(t, int64(6), got[5].AIInputTokens)
	assert.Equal(t, int64(3), got[5].AIOutputTokens)

	assert.Equal(t, "/tmp/read.go", got[6].Entity)
	assert.Equal(t, "claude-session", got[6].AISession)
	require.NotNil(t, got[6].AILineChanges)
	assert.Equal(t, 0, *got[6].AILineChanges)
	assert.Zero(t, got[6].AIPromptLength)
	assert.Equal(t, int64(2), got[6].AIInputTokens)
	assert.Equal(t, int64(1), got[6].AIOutputTokens)

	assert.Equal(t, "/tmp/empty.go", got[7].Entity)
	assert.Equal(t, "claude-session", got[7].AISession)
	require.NotNil(t, got[7].AILineChanges)
	assert.Equal(t, 0, *got[7].AILineChanges)
	assert.Zero(t, got[7].AIPromptLength)
	assert.Equal(t, int64(1), got[7].AIInputTokens)
	assert.Equal(t, int64(1), got[7].AIOutputTokens)
	require.NotNil(t, got[7].IsWrite)
	assert.False(t, *got[7].IsWrite)

	assert.Equal(t, "/tmp/same-lines.go", got[8].Entity)
	assert.Equal(t, "claude-session", got[8].AISession)
	require.NotNil(t, got[8].AILineChanges)
	assert.Equal(t, 0, *got[8].AILineChanges)
	assert.Zero(t, got[8].AIPromptLength)
	assert.Equal(t, int64(3), got[8].AIInputTokens)
	assert.Equal(t, int64(2), got[8].AIOutputTokens)
	require.NotNil(t, got[8].IsWrite)
	assert.True(t, *got[8].IsWrite)
}

func TestClaudeParse_SubscriptionPlanFromConfig(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	require.NoError(t, os.WriteFile(
		filepath.Join(home, ".claude.json"),
		[]byte(`{"oauthAccount":{"organizationType":"claude_max_20x"}}`),
		0o644,
	))

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-03-18T11:45:00Z","sessionId":"claude-session","version":"2.1.45",`,
			`"cwd":"/tmp","isSidechain":false,"type":"user",`,
			`"message":{"role":"user","content":"please continue"}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	got, err := ai.Claude{
		After: time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC),
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, "max", got[0].AISubscriptionPlan)
}

func TestClaudeParse_SubscriptionPlanFromBackupConfig(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	backupDir := filepath.Join(home, ".claude", "backups")
	require.NoError(t, os.MkdirAll(backupDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(backupDir, ".claude.json.backup.1"),
		[]byte(`{"oauthAccount":{"organizationType":"claude_pro"}}`),
		0o644,
	))

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-03-18T11:45:00Z","sessionId":"claude-session","version":"2.1.45",`,
			`"cwd":"/tmp","isSidechain":false,"type":"user",`,
			`"message":{"role":"user","content":"please continue"}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	got, err := ai.Claude{
		After: time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC),
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, "pro", got[0].AISubscriptionPlan)
}

func TestClaudeParse_ServiceTierIsNotSubscriptionPlan(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-03-18T11:45:00Z","sessionId":"claude-session","version":"2.1.45",`,
			`"cwd":"/tmp","isSidechain":false,"type":"user",`,
			`"message":{"role":"user","content":"please continue"}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-03-18T11:46:00Z","sessionId":"claude-session","version":"2.1.45",`,
			`"type":"assistant","message":{"role":"assistant",`,
			`"usage":{"input_tokens":1,"output_tokens":1,"service_tier":"standard"}}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	got, err := ai.Claude{
		After: time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC),
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Empty(t, got[0].AISubscriptionPlan)
}

func TestClaudeParse_NewToolUseResultShapes(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-05-02T12:00:00Z","sessionId":"claude-session","version":"2.1.119",`,
			`"cwd":"/workspace","toolUseResult":{"taskId":"task-1","statusChange":"done",`,
			`"updatedFields":["status"],"verificationNudgeNeeded":false},`,
			`"message":{"usage":{"input_tokens":10,"output_tokens":5}}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-05-02T12:01:00Z","sessionId":"claude-session","version":"2.1.119",`,
			`"toolUseResult":{"interrupted":false,"isImage":false,"noOutputExpected":false,`,
			`"stdout":"first\nsecond","stderr":"warning"},`,
			`"message":{"usage":{"input_tokens":3,"output_tokens":2}}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-05-02T12:02:00Z","sessionId":"claude-session","version":"2.1.119",`,
			`"toolUseResult":{"durationSeconds":1,"query":"wakatime","results":[{"title":"ignore"}]},`,
			`"message":{"usage":{"input_tokens":9,"output_tokens":9}}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-05-02T12:03:00Z","sessionId":"claude-session","version":"2.1.119",`,
			`"toolUseResult":{"bytes":128,"codeText":"document.title\nlocation.href","durationMs":5,`,
			`"result":"ok"},"message":{"usage":{"input_tokens":4,"output_tokens":1}}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-05-02T12:04:00Z","sessionId":"claude-session","version":"2.1.119",`,
			`"toolUseResult":{"filePath":"/workspace/main.go","oldString":"package main\n",`,
			`"newString":"package main\n\nfunc main() {}\n","originalFile":"package main\n",`,
			`"replaceAll":false},"message":{"usage":{"input_tokens":5,"output_tokens":2}}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Claude{
		After:             time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 3)

	assert.Equal(t, "Claude session", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "claude-session", got[0].AISession)
	assert.Equal(t, "/workspace", got[0].ProjectPathOverride)
	assert.Equal(t, int64(3), got[0].AIInputTokens)
	assert.Equal(t, int64(2), got[0].AIOutputTokens)
	require.NotNil(t, got[0].IsWrite)
	assert.False(t, *got[0].IsWrite)

	assert.Equal(t, "Claude session", got[1].Entity)
	assert.Equal(t, heartbeat.AppType, got[1].EntityType)
	assert.Equal(t, int64(4), got[1].AIInputTokens)
	assert.Equal(t, int64(1), got[1].AIOutputTokens)

	assert.Equal(t, "/workspace/main.go", got[2].Entity)
	assert.Equal(t, heartbeat.FileType, got[2].EntityType)
	require.NotNil(t, got[2].AILineChanges)
	assert.Equal(t, 2, *got[2].AILineChanges)
	assert.Equal(t, int64(5), got[2].AIInputTokens)
	assert.Equal(t, int64(2), got[2].AIOutputTokens)
	require.NotNil(t, got[2].IsWrite)
	assert.True(t, *got[2].IsWrite)
}

func TestClaudeParse_FileHeartbeatUsesCwdProjectPathOverrideForOutOfTreeFile(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	projectDir := filepath.Join(home, "Projects", "web", "hppmonitor-backend")
	plansDir := filepath.Join(home, ".claude", "plans")
	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")

	require.NoError(t, os.MkdirAll(projectDir, 0o755))
	require.NoError(t, os.MkdirAll(plansDir, 0o755))
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	planPath := filepath.Join(plansDir, "make-a-resource-for-humble-kettle.md")
	require.NoError(t, os.WriteFile(planPath, []byte("plan"), 0o644))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-05-02T12:00:00Z","sessionId":"claude-session","version":"2.1.143",`,
			`"cwd":` + jsonString(projectDir) + `,"type":"user",`,
			`"message":{"role":"user","content":"please plan this change"}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-05-02T12:01:00Z","sessionId":"claude-session","version":"2.1.143",`,
			`"toolUseResult":{"filePath":` + jsonString(planPath) + `,"content":"updated plan"},`,
			`"message":{"usage":{"input_tokens":10,"output_tokens":5}}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	got, err := ai.Claude{
		After: time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, projectDir, got[0].ProjectPathOverride)

	assert.Equal(t, heartbeat.FileType, got[1].EntityType)
	assert.Equal(t, planPath, got[1].Entity)
	assert.Equal(t, projectDir, got[1].ProjectPathOverride)
}

func TestClaudeParse_NoClaudeProjectsDir(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := ai.Claude{After: time.Now()}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestClaudeParse_RetainsProjectFolderFromSkippedFileLine(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project", "subagents")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "agent-worker.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-03-18T11:00:00Z","sessionId":"claude-session","version":"2.1.45",`,
			`"toolUseResult":{"filePath":"/workspace/project/main.go","content":"one"}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-03-18T12:20:00Z","sessionId":"claude-session",`,
			`"toolUseResult":{"content":[{"type":"text","text":"summary block"}]}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Claude{
		After: time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC),
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, filepath.Dir("/workspace/project/main.go"), got[0].ProjectPathOverride)
}

func TestClaudeParse_DoesNotUseEditorUserAgentWithoutIDEContext(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-03-18T11:45:00Z","sessionId":"claude-session","version":"2.1.45",`,
			`"cwd":"/tmp","type":"user","message":{"role":"user",`,
			`"model":"claude-opus-4.1","effort":"medium","content":[`,
			`{"type":"text","text":"please update the file"}`,
			`]}}`,
		}, ""),
		"{\"timestamp\":\"2026-03-18T12:00:00Z\",\"sessionId\":\"claude-session\",\"version\":\"2.1.45\"," +
			"\"toolUseResult\":{\"filePath\":\"/tmp/edited.go\"," +
			"\"structuredPatch\":[{\"oldLines\":1,\"newLines\":2}]}}",
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Claude{
		After:             time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC),
		FallbackUserAgent: "nvim/0.11.0",
		UserAgents: map[string]string{
			"/tmp/edited.go": heartbeat.UserAgent(ctx, "nvim/0.11.0"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "opus/4.1-medium claude-code/2.1.45", got[0].UserAgent)
	assert.Equal(t, "opus/4.1-medium claude-code/2.1.45", got[1].UserAgent)
}

func TestClaudeParse_SupportsClaudeFable5Model(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-07-06T11:45:00Z","sessionId":"claude-session","version":"2.5.0",`,
			`"cwd":"/tmp","type":"user","message":{"role":"user",`,
			`"model":"claude-fable-5","effort":"high","content":[`,
			`{"type":"text","text":"please refactor this module"}`,
			`]}}`,
		}, ""),
		"{\"timestamp\":\"2026-07-06T12:00:00Z\",\"sessionId\":\"claude-session\",\"version\":\"2.5.0\"," +
			"\"toolUseResult\":{\"filePath\":\"/tmp/edited.go\"," +
			"\"structuredPatch\":[{\"oldLines\":1,\"newLines\":2}]}}",
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Claude{
		After: time.Date(2026, 7, 6, 11, 0, 0, 0, time.UTC),
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "fable/5-high claude-code/2.5.0", got[0].UserAgent)
	assert.Equal(t, "fable/5-high claude-code/2.5.0", got[1].UserAgent)
}

func TestClaudeParse_PreservesClaudeCodePluginUserAgent(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-03-18T11:45:00Z","sessionId":"claude-session","version":"2.1.148",`,
			`"cwd":"/tmp","isSidechain":false,"type":"user",`,
			`"message":{"role":"user","content":"please continue"}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	got, err := ai.Claude{
		After:             time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC),
		FallbackUserAgent: "claude-code/2.1.148 claude-code-wakatime/3.1.6",
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(
		t,
		"claude-code/2.1.148 claude-code-wakatime/3.1.6",
		got[0].UserAgent,
	)
	assert.Equal(
		t,
		heartbeat.UserAgent(ctx, "claude-code/2.1.148 claude-code-wakatime/3.1.6"),
		heartbeat.UserAgent(ctx, got[0].UserAgent),
	)
}

func TestClaudeParse_UserMessageContentAsPlainString(t *testing.T) {
	ctx := context.Background()
	expectedPrompt := "yes continue"

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-03-18T11:45:00Z","sessionId":"claude-session","version":"2.1.45",`,
			`"cwd":"/tmp","isSidechain":false,"type":"user",`,
			`"message":{"role":"user","content":"` + expectedPrompt + `"}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Claude{
		After:             time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "claude-session", got[0].AISession)
	assert.Equal(t, len([]rune(expectedPrompt)), got[0].AIPromptLength)
}

func TestClaudeParse_UserMessageContentArrayWithUnknownBlocks(t *testing.T) {
	ctx := context.Background()
	expectedPrompt := "please continue"

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-03-18T11:45:00Z","sessionId":"claude-session","version":"2.1.45",`,
			`"cwd":"/tmp","isSidechain":false,"type":"user","message":{"role":"user","content":[`,
			`{"type":"tool_result","content":[{"type":"text","text":"ignore nested result"}]},`,
			`{"type":"text","text":{"unexpected":"shape"}},`,
			`{"type":"text","text":"` + expectedPrompt + `"}`,
			`]}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Claude{
		After:             time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "claude-session", got[0].AISession)
	assert.Equal(t, len([]rune(expectedPrompt)), got[0].AIPromptLength)
}

func TestClaudeParse_UserMessageContentAsPlainStringWithIDEContext(t *testing.T) {
	ctx := context.Background()
	expectedPrompt := "please continue from the selected file"

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	content := `<ide_opened_file><path>/tmp/main.go</path></ide_opened_file>\n\n` + expectedPrompt
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-03-18T11:45:00Z","sessionId":"claude-session","version":"2.1.45",`,
			`"cwd":"/tmp","isSidechain":false,"type":"user",`,
			`"message":{"role":"user","content":"` + content + `"}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Claude{
		After:             time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "claude-session", got[0].AISession)
	assert.Equal(t, len([]rune(expectedPrompt)), got[0].AIPromptLength)
	assert.Equal(t, "claude-code/2.1.45 plugin/0.0.1", got[0].UserAgent)
}

func TestClaudeParse_UserMessageContentAsPlainStringSystemReminder(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	content := `<system-reminder>\nThis is metadata.\n\nDo not count this as a user prompt.\n</system-reminder>`
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-03-18T11:45:00Z","sessionId":"claude-session","version":"2.1.45",`,
			`"cwd":"/tmp","isSidechain":false,"type":"user",`,
			`"message":{"role":"user","content":"` + content + `"}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Claude{
		After:             time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}
