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
	assert.Equal(t, "/root/wakatime-cli", got[1].ProjectPathOverride)
	require.NotNil(t, got[1].IsWrite)
	assert.False(t, *got[1].IsWrite)
	assert.Zero(t, got[1].AIInputTokens)
	assert.Equal(t, int64(12), got[1].AIOutputTokens)
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

func copyFile(t *testing.T, source, destination string) {
	input, err := os.ReadFile(source)
	require.NoError(t, err)

	err = os.WriteFile(destination, input, 0600)
	require.NoError(t, err)
}
