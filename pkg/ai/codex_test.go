package ai_test

import (
	"context"
	"os"
	"path/filepath"
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

	assert.Equal(t, "Codex", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[0].Category)
	assert.Nil(t, got[0].AILineChanges)
	require.NotNil(t, got[0].IsWrite)
	assert.False(t, *got[0].IsWrite)
	assert.Equal(t, float64(time.Date(2026, 3, 28, 11, 33, 14, 289, time.UTC).Unix()), got[0].Time)
	assert.Contains(t, got[0].UserAgent, heartbeat.UserAgent(ctx, "plugin/0.0.1"))
	assert.Contains(t, got[0].UserAgent, "Codex/0.116.0-alpha.1")

	assert.Equal(t, "Codex", got[1].Entity)
	assert.Equal(t, heartbeat.AppType, got[1].EntityType)
	assert.Nil(t, got[1].AILineChanges)
	require.NotNil(t, got[1].IsWrite)
	assert.False(t, *got[1].IsWrite)
	assert.Equal(t, float64(time.Date(2026, 3, 28, 11, 33, 18, 535, time.UTC).Unix()), got[1].Time)
	assert.Contains(t, got[1].UserAgent, heartbeat.UserAgent(ctx, "plugin/0.0.1"))
	assert.Contains(t, got[1].UserAgent, "Codex/0.116.0-alpha.1")

	assert.Equal(t, "/home/user/projects/wakatime-cli/pkg/ai/claude.go", got[2].Entity)
	assert.Equal(t, heartbeat.FileType, got[2].EntityType)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[2].Category)
	require.NotNil(t, got[2].AILineChanges)
	assert.Equal(t, 20, *got[2].AILineChanges)
	require.NotNil(t, got[2].IsWrite)
	assert.True(t, *got[2].IsWrite)
	assert.Equal(t, float64(time.Date(2026, 3, 28, 11, 33, 34, 952, time.UTC).Unix()), got[2].Time)
	assert.Contains(t, got[2].UserAgent, heartbeat.UserAgent(ctx, "editor/1.2.3"))
	assert.Contains(t, got[2].UserAgent, "Codex/0.116.0-alpha.1")
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

func copyFile(t *testing.T, source, destination string) {
	input, err := os.ReadFile(source)
	require.NoError(t, err)

	err = os.WriteFile(destination, input, 0600)
	require.NoError(t, err)
}
