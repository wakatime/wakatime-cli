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
		After: time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC),
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 4)

	assert.Equal(t, "/tmp/edited.go", got[0].Entity)
	assert.Equal(t, heartbeat.FileType, got[0].EntityType)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[0].Category)
	require.NotNil(t, got[0].AILineChanges)
	assert.Equal(t, -1, *got[0].AILineChanges)
	require.NotNil(t, got[0].IsWrite)
	assert.True(t, *got[0].IsWrite)
	assert.Equal(t, float64(time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC).Unix()), got[0].Time)
	assert.Contains(t, got[0].UserAgent, "Claude Code/2.1.45")
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
