package ai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func dshJSONLine(t *testing.T, v interface{}) string {
	t.Helper()

	raw, err := json.Marshal(v)
	require.NoError(t, err)

	return string(raw)
}

func dshCompressTranscript(t *testing.T, dest string, lines []string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(dest), 0o755))

	var buf bytes.Buffer

	writer, err := zstd.NewWriter(&buf)
	require.NoError(t, err)

	for _, line := range lines {
		_, err := writer.Write([]byte(line + "\n"))
		require.NoError(t, err)
	}

	require.NoError(t, writer.Close())
	require.NoError(t, os.WriteFile(dest, buf.Bytes(), 0o644))
}

func TestDshParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	projectDir := filepath.Join(home, "project")
	readFile := filepath.Join(projectDir, "README.md")
	editFile := filepath.Join(projectDir, "main.go")

	transcriptPath := filepath.Join(
		home, ".dsh", "sessions",
		"--tmp-project--",
		"session-abc123",
		"session.jsonl.zstd",
	)

	// An out-of-window request whose usage must not leak into emitted heartbeats.
	outOfWindow := time.Date(2026, 4, 19, 11, 0, 0, 0, time.UTC)

	dshCompressTranscript(t, transcriptPath, []string{
		dshJSONLine(t, map[string]interface{}{
			"type":      "session",
			"version":   0,
			"id":        "session-abc123",
			"createdAt": 1776598800000,
			"cwd":       projectDir,
		}),
		dshJSONLine(t, map[string]interface{}{
			"type": "request/header",
			"time": outOfWindow.UnixMilli(),
			"data": map[string]interface{}{
				"header": map[string]interface{}{
					"config": map[string]interface{}{
						"provider": "deepseek",
						"model":    "deepseek-v4",
					},
				},
			},
		}),
		dshJSONLine(t, map[string]interface{}{
			"type": "assistant/message",
			"time": outOfWindow.UnixMilli(),
			"data": map[string]interface{}{
				"usage": map[string]interface{}{"inputTokens": 999999, "outputTokens": 888888, "cacheReadTokens": 777777},
				"message": map[string]interface{}{
					"role": "assistant",
					"content": []map[string]interface{}{
						{"type": "text", "text": "old summary"},
					},
					"source": map[string]interface{}{"kind": "model", "provider": "deepseek", "model": "deepseek-v4"},
				},
			},
		}),
		dshJSONLine(t, map[string]interface{}{
			"type": "assistant/chunk",
			"time": 1776600000000,
			"data": map[string]interface{}{
				"chunk": map[string]interface{}{
					"type":  "usage",
					"usage": map[string]interface{}{"inputTokens": 1, "outputTokens": 2},
				},
			},
		}),
		dshJSONLine(t, map[string]interface{}{
			"type": "user/message",
			"time": 1776600001000,
			"data": map[string]interface{}{
				"role": "user",
				"content": []map[string]interface{}{
					{"type": "text", "text": "Inspect the project and fix main.go"},
				},
			},
		}),
		dshJSONLine(t, map[string]interface{}{
			"type": "tool/call",
			"time": 1776600002000,
			"data": map[string]interface{}{
				"name":      "read",
				"arguments": `{"file_path":"` + readFile + `"}`,
			},
		}),
		dshJSONLine(t, map[string]interface{}{
			"type": "tool/call",
			"time": 1776600003000,
			"data": map[string]interface{}{
				"name":      "edit",
				"arguments": `{"file_path":"` + editFile + `","old_string":"old\nline\n","new_string":"new\nlines\nhere\n"}`,
			},
		}),
		dshJSONLine(t, map[string]interface{}{
			"type": "tool/call",
			"time": 1776600003500,
			"data": map[string]interface{}{
				"name":      "bash",
				"arguments": `{"command":"go test ./..."}`,
			},
		}),
		dshJSONLine(t, map[string]interface{}{
			"type": "tool/call",
			"time": 1776600003600,
			"data": map[string]interface{}{
				"name":      "grep",
				"arguments": `{"pattern":"TODO","path":"` + projectDir + `"}`,
			},
		}),
		dshJSONLine(t, map[string]interface{}{
			"type": "assistant/message",
			"time": 1776600004000,
			"data": map[string]interface{}{
				"usage": map[string]interface{}{"inputTokens": 100, "outputTokens": 50, "cacheReadTokens": 200},
				"message": map[string]interface{}{
					"role": "assistant",
					"content": []map[string]interface{}{
						{"type": "reasoning", "text": "thinking"},
						{"type": "text", "text": "Fixed the parser.\nAll tests pass."},
					},
					"source": map[string]interface{}{"kind": "model", "provider": "deepseek", "model": "deepseek-v4"},
				},
			},
		}),
	})

	parser := ai.Dsh{
		After:             time.Date(2026, 4, 19, 11, 59, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 4)

	assert.Equal(t, "DeepSeek Harness", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "abc123", got[0].AISession)
	assert.Equal(t, len([]rune("Inspect the project and fix main.go")), got[0].AIPromptLength)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[0].Category)
	assert.Equal(t, projectDir, got[0].ProjectPathOverride)
	require.NotNil(t, got[0].IsWrite)
	assert.False(t, *got[0].IsWrite)
	assert.Equal(t, float64(1776600001), got[0].Time)

	assert.Equal(t, readFile, got[1].Entity)
	assert.Equal(t, heartbeat.FileType, got[1].EntityType)
	assert.Equal(t, "abc123", got[1].AISession)
	require.NotNil(t, got[1].AILineChanges)
	assert.Zero(t, *got[1].AILineChanges)
	require.NotNil(t, got[1].IsWrite)
	assert.False(t, *got[1].IsWrite)

	assert.Equal(t, editFile, got[2].Entity)
	assert.Equal(t, heartbeat.FileType, got[2].EntityType)
	require.NotNil(t, got[2].AILineChanges)
	assert.Equal(t, 4, *got[2].AILineChanges)
	require.NotNil(t, got[2].IsWrite)
	assert.True(t, *got[2].IsWrite)

	// usage from the out-of-window assistant/message must be excluded
	assert.Equal(t, int64(100), got[3].AIInputTokens)
	assert.Equal(t, int64(200), got[3].AICachedInputTokens)
	assert.Equal(t, int64(50), got[3].AIOutputTokens)
	assert.Contains(t, got[3].UserAgent, "deepseek/")
	assert.True(t, strings.Index(got[3].UserAgent, "deepseek/") <
		strings.Index(got[3].UserAgent, "plugin/0.0.1"))
}

func TestDshParseSkipsOldTranscripts(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptPath := filepath.Join(
		home, ".dsh", "sessions",
		"--tmp-project--",
		"session-old456",
		"session.jsonl.zstd",
	)

	dshCompressTranscript(t, transcriptPath, []string{
		dshJSONLine(t, map[string]interface{}{
			"type": "session",
			"id":   "session-old456",
			"cwd":  home,
		}),
		dshJSONLine(t, map[string]interface{}{
			"type": "user/message",
			"time": 1776600001000,
			"data": map[string]interface{}{
				"role":    "user",
				"content": []map[string]interface{}{{"type": "text", "text": "old activity"}},
			},
		}),
	})

	old := time.Date(2026, 4, 19, 11, 0, 0, 0, time.UTC)
	require.NoError(t, os.Chtimes(transcriptPath, old, old))

	parser := ai.Dsh{After: time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestDshParseMissingDirectory(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	parser := ai.Dsh{After: time.Now().Add(-time.Minute)}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}
