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

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		"{\"timestamp\":\"2026-03-18T10:00:00Z\",\"version\":\"2.1.44\"," +
			"\"toolUseResult\":{\"filePath\":\"/tmp/skip.txt\"," +
			"\"structuredPatch\":[{\"oldLines\":1,\"newLines\":2}]}}",
		"not json",
		"{\"timestamp\":\"2026-03-18T11:30:00Z\"}",
		"{\"timestamp\":\"2026-03-18T12:00:00Z\",\"version\":\"2.1.45\"," +
			"\"toolUseResult\":{\"filePath\":\"/tmp/edited.go\"," +
			"\"structuredPatch\":[{\"oldLines\":3,\"newLines\":5},{\"oldLines\":4,\"newLines\":1}]}}",
		"{\"timestamp\":\"2026-03-18T12:15:00Z\",\"toolUseResult\":\"plain string result\"}",
		"{\"timestamp\":\"2026-03-18T12:20:00Z\",\"toolUseResult\":{" +
			"\"agentId\":\"worker-1\",\"agentType\":\"worker\",\"content\":[{" +
			"\"type\":\"text\",\"text\":\"summary block\"}]}}",
		"{\"timestamp\":\"2026-03-18T12:25:00Z\",\"toolUseResult\":{" +
			"\"filePath\":\"/tmp/array.go\",\"content\":[{" +
			"\"type\":\"text\",\"text\":\"first\\nsecond\"},{\"type\":\"text\",\"text\":\"third\"}]}}",
		"{\"timestamp\":\"2026-03-18T12:30:00Z\",\"toolUseResult\":{" +
			"\"filePath\":\"/tmp/new.go\",\"content\":\"first\\nsecond\\nthird\"}}",
		"{\"timestamp\":\"2026-03-18T13:00:00Z\",\"toolUseResult\":{" +
			"\"filePath\":\"/tmp/read.go\",\"content\":\"existing\",\"originalFile\":\"before\"}}",
		"{\"timestamp\":\"2026-03-18T13:30:00Z\",\"toolUseResult\":{" +
			"\"filePath\":\"/tmp/empty.go\",\"structuredPatch\":[]}}",
		"{\"timestamp\":\"2026-03-18T14:00:00Z\",\"toolUseResult\":{" +
			"\"structuredPatch\":[{\"oldLines\":1,\"newLines\":1}]}}",
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
	require.Len(t, got, 7)

	assert.Equal(t, "/tmp/edited.go", got[0].Entity)
	assert.Equal(t, heartbeat.FileType, got[0].EntityType)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[0].Category)
	require.NotNil(t, got[0].AILineChanges)
	assert.Equal(t, -1, *got[0].AILineChanges)
	require.NotNil(t, got[0].IsWrite)
	assert.True(t, *got[0].IsWrite)
	assert.Equal(t, float64(time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC).Unix()), got[0].Time)
	assert.Equal(
		t,
		"ClaudeCode/2.1.45 "+heartbeat.UserAgent(ctx, "editor/1.2.3"),
		got[0].UserAgent,
	)
	assert.Contains(t, got[0].UserAgent, "ClaudeCode/2.1.45")

	assert.Equal(t, "Claude session", got[1].Entity)
	assert.Equal(t, heartbeat.AppType, got[1].EntityType)
	assert.Nil(t, got[1].AILineChanges)
	assert.Equal(t, filepath.Dir("/tmp/edited.go"), got[1].ProjectPathOverride)
	require.NotNil(t, got[1].IsWrite)
	assert.False(t, *got[1].IsWrite)
	assert.Contains(t, got[1].UserAgent, "plugin/0.0.1")
	assert.Contains(t, got[1].UserAgent, "ClaudeCode/2.1.45")

	assert.Equal(t, "Claude session", got[2].Entity)
	assert.Equal(t, heartbeat.AppType, got[2].EntityType)
	assert.Nil(t, got[2].AILineChanges)
	assert.Equal(t, filepath.Dir("/tmp/array.go"), got[2].ProjectPathOverride)
	require.NotNil(t, got[2].IsWrite)
	assert.False(t, *got[2].IsWrite)

	assert.Equal(t, "/tmp/array.go", got[3].Entity)
	require.NotNil(t, got[3].AILineChanges)
	assert.Equal(t, 3, *got[3].AILineChanges)
	require.NotNil(t, got[3].IsWrite)
	assert.True(t, *got[3].IsWrite)
	assert.Contains(t, got[3].UserAgent, "plugin/0.0.1")
	assert.Contains(t, got[3].UserAgent, "ClaudeCode/2.1.45")

	assert.Equal(t, "/tmp/new.go", got[4].Entity)
	require.NotNil(t, got[4].AILineChanges)
	assert.Equal(t, 3, *got[4].AILineChanges)

	assert.Equal(t, "/tmp/read.go", got[5].Entity)
	require.NotNil(t, got[5].AILineChanges)
	assert.Equal(t, 0, *got[5].AILineChanges)

	assert.Equal(t, "/tmp/empty.go", got[6].Entity)
	require.NotNil(t, got[6].AILineChanges)
	assert.Equal(t, 0, *got[6].AILineChanges)
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

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-03-18T11:00:00Z","version":"2.1.45",`,
			`"toolUseResult":{"filePath":"/workspace/project/main.go","content":"one"}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-03-18T12:20:00Z",`,
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
