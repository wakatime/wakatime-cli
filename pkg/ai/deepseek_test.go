package ai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestDeepSeekParseCompressedTranscript(t *testing.T) {
	ctx := context.Background()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("DSH_HOME", "")

	projectDir := filepath.Join(home, "project")
	readFile := filepath.ToSlash(filepath.Join(projectDir, "README.md"))
	editFile := filepath.ToSlash(filepath.Join(projectDir, "main.go"))
	transcript := filepath.Join(home, ".dsh", "sessions", "--project--", "session-abc", "session.jsonl.zstd")
	start := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)

	dshWriteCompressedTranscript(t, transcript, []interface{}{
		map[string]interface{}{"type": "session", "id": "session-abc", "cwd": projectDir},
		dshEvent("user/message", 0, start, map[string]interface{}{
			"role": "user", "source": map[string]interface{}{"kind": "plugin"},
			"content": []map[string]interface{}{{"type": "text", "text": "injected context"}},
		}),
		dshEvent("user/message", 1, start.Add(time.Second), map[string]interface{}{
			"role": "user", "source": map[string]interface{}{"kind": "user"},
			"content": []map[string]interface{}{{"type": "text", "text": "Fix main.go"}},
		}),
		dshEvent("request/header", 2, start.Add(2*time.Second), map[string]interface{}{
			"header": map[string]interface{}{"config": map[string]interface{}{
				"provider": "deepseek", "model": "deepseek-v4",
			}},
		}),
		dshEvent("assistant/chunk", 3, start.Add(3*time.Second), map[string]interface{}{
			"turn": 1, "step": 1,
			"chunk": map[string]interface{}{"type": "usage", "usage": map[string]interface{}{
				"inputTokens": 10, "cacheReadTokens": 3, "cacheWriteTokens": 2, "outputTokens": 5,
			}},
		}),
		dshEvent("assistant/chunk", 4, start.Add(4*time.Second), map[string]interface{}{
			"turn": 1, "step": 1,
			"chunk": map[string]interface{}{"type": "finish", "reason": map[string]interface{}{"kind": "error"}},
		}),
		dshEvent("assistant/chunk", 5, start.Add(5*time.Second), map[string]interface{}{
			"turn": 1, "step": 1,
			"chunk": map[string]interface{}{"type": "usage", "usage": map[string]interface{}{
				"inputTokens": 20, "cacheReadTokens": 7, "cacheWriteTokens": 1, "outputTokens": 8,
			}},
		}),
		dshEvent("assistant/chunk", 6, start.Add(6*time.Second), map[string]interface{}{
			"turn": 1, "step": 1,
			"chunk": map[string]interface{}{"type": "finish", "reason": map[string]interface{}{"kind": "stop"}},
		}),
		dshEvent("assistant/message", 7, start.Add(7*time.Second), map[string]interface{}{
			"turn": 1, "step": 1,
			"usage": map[string]interface{}{
				"inputTokens": 20, "cacheReadTokens": 7, "cacheWriteTokens": 1, "outputTokens": 8,
			},
			"message": map[string]interface{}{
				"role": "assistant", "source": map[string]interface{}{
					"kind": "model", "provider": "deepseek", "model": "deepseek-v4",
				},
				"content": []map[string]interface{}{{"type": "text", "text": "I will update it."}},
			},
		}),
		dshEvent("tool/call", 8, start.Add(8*time.Second), map[string]interface{}{
			"callId": "read-1", "name": "read", "arguments": `{"file_path":"README.md"}`,
		}),
		dshToolResultEvent(9, start.Add(9*time.Second), "read-1", false),
		dshEvent("tool/call", 10, start.Add(10*time.Second), map[string]interface{}{
			"callId": "edit-1", "name": "edit",
			"arguments": `{"file_path":"main.go","old_string":"old\n","new_string":"new\nmore\n"}`,
		}),
		dshToolResultEvent(11, start.Add(11*time.Second), "edit-1", false),
		dshEvent("tool/call", 12, start.Add(12*time.Second), map[string]interface{}{
			"callId": "write-failed", "name": "write",
			"arguments": `{"file_path":"failed.go","content":"package failed\n"}`,
		}),
		dshToolResultEvent(13, start.Add(13*time.Second), "write-failed", true),
		dshEvent("tool/call", 14, start.Add(14*time.Second), map[string]interface{}{
			"callId": "grep-1", "name": "grep", "arguments": `{"path":".","pattern":"TODO"}`,
		}),
		dshToolResultEvent(15, start.Add(15*time.Second), "grep-1", false),
		dshEvent("tool/call", 16, start.Add(16*time.Second), map[string]interface{}{
			"callId": "create-1", "name": "str_replace_editor",
			"arguments": `{"command":"create","path":"created.go","file_text":"package created\n"}`,
		}),
		dshToolResultEvent(17, start.Add(17*time.Second), "create-1", false),
		dshEvent("tool/call", 18, start.Add(18*time.Second), map[string]interface{}{
			"callId": "view-dir", "name": "str_replace_editor",
			"arguments": `{"command":"view","path":"project"}`,
		}),
		dshToolResultEventWithText(
			19,
			start.Add(19*time.Second),
			"view-dir",
			false,
			"Here're the files and directories up to 2 levels deep in project, excluding hidden items",
		),
	})

	parser := ai.DeepSeek{
		After:             start.Add(-time.Minute),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			editFile: heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 5)

	assert.Equal(t, "DeepSeek Harness", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "session-abc", got[0].AISession)
	assert.Equal(t, len([]rune("Fix main.go")), got[0].AIPromptLength)
	assert.Equal(t, projectDir, got[0].ProjectPathOverride)
	assert.Equal(t, int64(12), got[0].AIInputTokens)
	assert.Equal(t, int64(3), got[0].AICachedInputTokens)
	assert.Equal(t, int64(5), got[0].AIOutputTokens)
	assert.Contains(t, got[0].UserAgent, "deepseek/4")

	assert.Equal(t, "DeepSeek Harness", got[1].Entity)
	assert.Equal(t, int64(21), got[1].AIInputTokens)
	assert.Equal(t, int64(7), got[1].AICachedInputTokens)
	assert.Equal(t, int64(8), got[1].AIOutputTokens)
	assert.Contains(t, got[1].UserAgent, "deepseek/4")

	assert.Equal(t, readFile, got[2].Entity)
	assert.Equal(t, heartbeat.FileType, got[2].EntityType)
	require.NotNil(t, got[2].IsWrite)
	assert.False(t, *got[2].IsWrite)

	assert.Equal(t, editFile, got[3].Entity)
	require.NotNil(t, got[3].AILineChanges)
	assert.Equal(t, 1, *got[3].AILineChanges)
	require.NotNil(t, got[3].IsWrite)
	assert.True(t, *got[3].IsWrite)
	assert.Contains(t, got[3].UserAgent, "editor/1.2.3")

	assert.Equal(t, filepath.ToSlash(filepath.Join(projectDir, "created.go")), got[4].Entity)
	require.NotNil(t, got[4].AILineChanges)
	assert.Equal(t, 2, *got[4].AILineChanges)
	require.NotNil(t, got[4].IsWrite)
	assert.True(t, *got[4].IsWrite)
}

func TestDeepSeekParseUncompressedTranscriptFromConfiguredHome(t *testing.T) {
	ctx := context.Background()
	home := t.TempDir()
	dshHome := filepath.Join(home, "configured-dsh")
	t.Setenv("HOME", filepath.Join(home, "unused-home"))
	t.Setenv("USERPROFILE", filepath.Join(home, "unused-home"))
	t.Setenv("DSH_HOME", dshHome)

	start := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)
	transcript := filepath.Join(dshHome, "sessions", "--project--", "session-raw", "session.jsonl")
	dshWriteRawTranscript(t, transcript, []interface{}{
		map[string]interface{}{"type": "session", "id": "session-raw", "cwd": home},
		dshEvent("user/message", 0, start, map[string]interface{}{
			"role": "user", "source": map[string]interface{}{"kind": "user"},
			"content": []map[string]interface{}{{"type": "text", "text": "Raw transcript prompt"}},
		}),
		dshEvent("assistant/message", 1, start.Add(time.Second), map[string]interface{}{
			"turn": 1, "step": 1,
			"usage": map[string]interface{}{
				"inputTokens": 4, "cacheReadTokens": 2, "cacheWriteTokens": 1, "outputTokens": 3,
			},
			"message": map[string]interface{}{
				"role": "assistant", "source": map[string]interface{}{
					"kind": "model", "provider": "deepseek", "model": "deepseek-v4",
				},
				"content": []map[string]interface{}{{"type": "text", "text": "Raw response"}},
			},
		}),
	})

	got, err := (ai.DeepSeek{After: start.Add(-time.Second)}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "session-raw", got[0].AISession)
	assert.Equal(t, len([]rune("Raw transcript prompt")), got[0].AIPromptLength)
	assert.Equal(t, int64(5), got[1].AIInputTokens)
	assert.Equal(t, int64(2), got[1].AICachedInputTokens)
	assert.Equal(t, int64(3), got[1].AIOutputTokens)
}

func TestDeepSeekParseCompressedTranscriptWithIncompleteTail(t *testing.T) {
	ctx := context.Background()
	home := t.TempDir()
	t.Setenv("DSH_HOME", home)

	start := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)
	transcript := filepath.Join(home, "sessions", "--project--", "session-truncated", "session.jsonl.zstd")
	dshWriteCompressedTranscript(t, transcript, []interface{}{
		map[string]interface{}{"type": "session", "id": "session-truncated", "cwd": home},
		dshEvent("user/message", 0, start, map[string]interface{}{
			"role": "user", "source": map[string]interface{}{"kind": "user"},
			"content": []map[string]interface{}{{"type": "text", "text": "Complete prompt"}},
		}),
		dshEvent("user/message", 1, start.Add(time.Second), map[string]interface{}{
			"role": "user", "source": map[string]interface{}{"kind": "user"},
			"content": []map[string]interface{}{{"type": "text", "text": "Incomplete prompt"}},
		}),
	})

	info, err := os.Stat(transcript)
	require.NoError(t, err)
	require.NoError(t, os.Truncate(transcript, info.Size()-5))

	got, err := (ai.DeepSeek{After: start.Add(-time.Second)}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, len([]rune("Complete prompt")), got[0].AIPromptLength)
}

func TestDeepSeekParseSkipsForkSeedPrefix(t *testing.T) {
	ctx := context.Background()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("DSH_HOME", "")

	start := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)
	transcript := filepath.Join(home, ".dsh", "sessions", "--project--", "session-child", "session.jsonl.zstd")
	dshWriteCompressedTranscript(t, transcript, []interface{}{
		map[string]interface{}{
			"type": "session", "id": "session-child", "cwd": home,
			"parentSession": "session-parent", "seedLength": 3,
		},
		dshEvent("user/message", 0, start, map[string]interface{}{
			"role": "user", "source": map[string]interface{}{"kind": "user"},
			"content": []map[string]interface{}{{"type": "text", "text": "Inherited parent prompt"}},
		}),
		dshEvent("tool/call", 1, start.Add(time.Second), map[string]interface{}{
			"callId": "inherited-read", "name": "read", "arguments": `{"file_path":"parent.go"}`,
		}),
		dshToolResultEvent(2, start.Add(2*time.Second), "inherited-read", false),
		dshEvent("user/message", 3, start.Add(3*time.Second), map[string]interface{}{
			"role": "user", "source": map[string]interface{}{"kind": "user"},
			"content": []map[string]interface{}{{"type": "text", "text": "Child prompt"}},
		}),
	})

	got, err := (ai.DeepSeek{After: start.Add(-time.Second)}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "session-child", got[0].AISession)
	assert.Equal(t, len([]rune("Child prompt")), got[0].AIPromptLength)
}

func TestDeepSeekParseSkipsOldTranscript(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("DSH_HOME", "")

	start := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)
	transcript := filepath.Join(home, ".dsh", "sessions", "--project--", "session-old", "session.jsonl")
	dshWriteRawTranscript(t, transcript, []interface{}{
		map[string]interface{}{"type": "session", "id": "session-old", "cwd": home},
		dshEvent("user/message", 0, start, map[string]interface{}{
			"role": "user", "source": map[string]interface{}{"kind": "user"},
			"content": []map[string]interface{}{{"type": "text", "text": "Old prompt"}},
		}),
	})
	require.NoError(t, os.Chtimes(transcript, start, start))

	got, err := (ai.DeepSeek{After: start.Add(time.Minute)}).Parse(context.Background())
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestDeepSeekParseUsageOnlyAssistantMessage(t *testing.T) {
	ctx := context.Background()
	home := t.TempDir()
	t.Setenv("DSH_HOME", home)

	start := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)
	transcript := filepath.Join(home, "sessions", "--project--", "session-usage-only", "session.jsonl")
	dshWriteRawTranscript(t, transcript, []interface{}{
		map[string]interface{}{"type": "session", "id": "session-usage-only", "cwd": home},
		dshEvent("user/message", 0, start.Add(-time.Second), map[string]interface{}{
			"role": "user", "source": map[string]interface{}{"kind": "user"},
			"content": []map[string]interface{}{{"type": "text", "text": "Old prompt"}},
		}),
		dshEvent("assistant/message", 1, start.Add(time.Second), map[string]interface{}{
			"turn": 1, "step": 1,
			"usage": map[string]interface{}{
				"inputTokens": 11, "cacheReadTokens": 7, "cacheWriteTokens": 3, "outputTokens": 5,
			},
			"message": map[string]interface{}{
				"role": "assistant", "content": []interface{}{},
				"source": map[string]interface{}{
					"kind": "model", "provider": "deepseek", "model": "deepseek-v4",
				},
			},
		}),
	})

	got, err := (ai.DeepSeek{After: start}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, int64(14), got[0].AIInputTokens)
	assert.Equal(t, int64(7), got[0].AICachedInputTokens)
	assert.Equal(t, int64(5), got[0].AIOutputTokens)
	assert.Contains(t, got[0].UserAgent, "deepseek/4")
}

func TestDeepSeekParseLegacyToolResultShapes(t *testing.T) {
	ctx := context.Background()
	home := t.TempDir()
	t.Setenv("DSH_HOME", home)

	projectDir := filepath.Join(home, "project")
	readFile := filepath.ToSlash(filepath.Join(projectDir, "README.md"))
	start := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)
	transcript := filepath.Join(home, "sessions", "--project--", "session-legacy-tools", "session.jsonl")
	dshWriteRawTranscript(t, transcript, []interface{}{
		map[string]interface{}{"type": "session", "id": "session-legacy-tools", "cwd": projectDir},
		dshEvent("tool/call", 0, start, map[string]interface{}{
			"callId": "failed-write", "name": "write",
			"arguments": `{"file_path":"failed.go","content":"package failed\n"}`,
		}),
		dshEvent("tool/result", 1, start.Add(time.Second), map[string]interface{}{
			"callId": "failed-write", "isError": true,
			"content": []map[string]interface{}{{"type": "text", "text": "write failed"}},
		}),
		dshEvent("tool/call", 2, start.Add(2*time.Second), map[string]interface{}{
			"callId": "directory-view", "name": "str_replace_editor",
			"arguments": `{"command":"view","path":"virtual-directory"}`,
		}),
		dshEvent("tool/result", 3, start.Add(3*time.Second), map[string]interface{}{
			"callId": "directory-view", "isError": false,
			"content": []map[string]interface{}{{
				"type": "text",
				"text": "Here're the files and directories up to 2 levels deep in virtual-directory",
			}},
		}),
		dshEvent("tool/call", 4, start.Add(4*time.Second), map[string]interface{}{
			"callId": "successful-read", "name": "read", "arguments": `{"file_path":"README.md"}`,
		}),
		dshEvent("tool/result", 5, start.Add(5*time.Second), map[string]interface{}{
			"callId": "successful-read", "isError": false,
			"content": []map[string]interface{}{{"type": "text", "text": "contents"}},
		}),
	})

	got, err := (ai.DeepSeek{After: start}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, readFile, got[0].Entity)
	require.NotNil(t, got[0].IsWrite)
	assert.False(t, *got[0].IsWrite)
}

func TestDeepSeekParseAssistantProvenanceModel(t *testing.T) {
	ctx := context.Background()
	home := t.TempDir()
	t.Setenv("DSH_HOME", home)

	start := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)
	transcript := filepath.Join(home, "sessions", "--project--", "session-provenance", "session.jsonl")
	dshWriteRawTranscript(t, transcript, []interface{}{
		map[string]interface{}{"type": "session", "id": "session-provenance", "cwd": home},
		dshEvent("user/message", 0, start, map[string]interface{}{
			"role": "user", "source": map[string]interface{}{"kind": "user"},
			"content": []map[string]interface{}{{"type": "text", "text": "Use the configured model"}},
		}),
		dshEvent("assistant/message", 1, start.Add(time.Second), map[string]interface{}{
			"turn": 1, "step": 1, "role": "assistant",
			"provenance": map[string]interface{}{
				"kind": "model", "provider": "deepseek", "model": "deepseek-v4",
			},
			"content": []map[string]interface{}{{"type": "text", "text": "Done"}},
		}),
	})

	got, err := (ai.DeepSeek{After: start}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Contains(t, got[0].UserAgent, "deepseek/4")
	assert.Contains(t, got[1].UserAgent, "deepseek/4")
}

func dshEvent(eventType string, seq int, timestamp time.Time, data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type": eventType,
		"seq":  seq,
		"time": timestamp.UnixMilli(),
		"data": data,
	}
}

func dshToolResultEvent(seq int, timestamp time.Time, callID string, failed bool) map[string]interface{} {
	return dshToolResultEventWithText(seq, timestamp, callID, failed, "result")
}

func dshToolResultEventWithText(
	seq int,
	timestamp time.Time,
	callID string,
	failed bool,
	text string,
) map[string]interface{} {
	data := map[string]interface{}{
		"message": map[string]interface{}{
			"role": "user", "source": map[string]interface{}{"kind": "tool", "callId": callID},
			"content": []map[string]interface{}{{
				"type": "tool-result", "toolCallId": callID, "isError": failed,
				"content": []map[string]interface{}{{"type": "text", "text": text}},
			}},
		},
	}
	if failed {
		data["error"] = map[string]interface{}{"name": "ToolError", "code": "FAILED"}
	}

	return dshEvent("tool/result", seq, timestamp, data)
}

func dshWriteCompressedTranscript(t *testing.T, destination string, rows []interface{}) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(destination), 0o755))

	var compressed bytes.Buffer
	for _, row := range rows {
		writer, err := zstd.NewWriter(&compressed)
		require.NoError(t, err)

		_, err = writer.Write(append(dshJSON(t, row), '\n'))
		require.NoError(t, err)
		require.NoError(t, writer.Close())
	}

	require.NoError(t, os.WriteFile(destination, compressed.Bytes(), 0o644))
}

func dshWriteRawTranscript(t *testing.T, destination string, rows []interface{}) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(destination), 0o755))

	var raw bytes.Buffer
	for _, row := range rows {
		raw.Write(dshJSON(t, row))
		raw.WriteByte('\n')
	}

	require.NoError(t, os.WriteFile(destination, raw.Bytes(), 0o644))
}

func dshJSON(t *testing.T, value interface{}) []byte {
	t.Helper()

	raw, err := json.Marshal(value)
	require.NoError(t, err)

	return raw
}
