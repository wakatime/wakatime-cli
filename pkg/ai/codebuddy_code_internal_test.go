package ai

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestCodeBuddyCodeParse(t *testing.T) {
	configDir, projectDir, sessionsDir := codeBuddyCodeTestDirs(t)
	startedAt := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)

	transcript := codeBuddyCodeJSONLines(t,
		map[string]any{
			"id":        "user-1",
			"timestamp": startedAt.UnixMilli(),
			"type":      "message",
			"role":      "user",
			"cwd":       projectDir,
			"content":   []map[string]any{{"type": "input_text", "text": "Implement the parser"}},
		},
		map[string]any{
			"id":        "assistant-1",
			"timestamp": startedAt.Add(time.Second).UnixMilli(),
			"type":      "message",
			"role":      "assistant",
			"message": map[string]any{
				"usage": map[string]any{
					"input_tokens":                120,
					"cache_creation_input_tokens": 5,
					"cache_read_input_tokens":     20,
					"output_tokens":               30,
				},
			},
			"providerData": map[string]any{
				"model": "hunyuan-2.0-instruct",
				"usage": map[string]any{
					"inputTokens":               120,
					"outputTokens":              30,
					"inputTokensDetails":        []map[string]any{{"cached_tokens": 20}},
					"prompt_cache_write_tokens": 5,
				},
			},
			"content": []map[string]any{{"type": "output_text", "text": "Working on it"}},
		},
		map[string]any{
			"id":           "internal-1",
			"timestamp":    startedAt.Add(2 * time.Second).UnixMilli(),
			"type":         "message",
			"role":         "user",
			"content":      "internal compact request",
			"providerData": map[string]any{"isCompactInternal": true},
		},
		codeBuddyCodeToolCallRecord(t, "failed-write", startedAt.Add(3*time.Second), projectDir, "Write", map[string]any{
			"file_path": filepath.Join("pkg", "failed.go"),
			"content":   "package failed",
		}),
		codeBuddyCodeToolResult("failed-write", startedAt.Add(4*time.Second), "incomplete", map[string]any{
			"error": "permission denied",
		}),
		codeBuddyCodeToolCallRecord(t, "successful-edit", startedAt.Add(5*time.Second), projectDir, "Edit", map[string]any{
			"file_path":   filepath.Join("pkg", "main.go"),
			"old_string":  "func main() {}",
			"new_string":  "func main() {\n\tprintln()\n}",
			"replace_all": false,
		}),
		codeBuddyCodeToolResult("successful-edit", startedAt.Add(6*time.Second), "completed", map[string]any{
			"content": "Successfully edited file",
			"title":   "Made 1 replacement",
		}),
	)

	codeBuddyCodeWriteTranscript(t, filepath.Join(sessionsDir, "session-1.jsonl"), transcript, startedAt)
	codeBuddyCodeWriteStaleTranscript(t, sessionsDir, startedAt)

	got, err := (CodeBuddyCode{
		After:             startedAt.Add(-time.Minute),
		FallbackUserAgent: "plugin/0.0.1",
	}).Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 4)

	assert.Equal(t, "CodeBuddy Code session-1", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "session-1", got[0].AISession)
	assert.Equal(t, len([]rune("Implement the parser")), got[0].AIPromptLength)
	assert.Equal(t, projectDir, got[0].ProjectPathOverride)
	assert.Contains(t, got[0].UserAgent, "hunyuan/2.0-instruct")

	assert.Equal(t, int64(100), got[1].AIInputTokens)
	assert.Equal(t, int64(20), got[1].AICachedInputTokens)
	assert.Equal(t, int64(30), got[1].AIOutputTokens)
	assert.Contains(t, got[1].UserAgent, "hunyuan/2.0-instruct")

	assert.Equal(t, heartbeat.AppType, got[2].EntityType)
	assert.Equal(t, "session-1", got[2].AISession)

	expectedEntity := filepath.ToSlash(filepath.Join(projectDir, "pkg", "main.go"))
	assert.Equal(t, expectedEntity, got[3].Entity)
	assert.Equal(t, heartbeat.FileType, got[3].EntityType)
	assert.Equal(t, heartbeat.PointerTo(true), got[3].IsWrite)
	assert.Equal(t, heartbeat.PointerTo(2), got[3].AILineChanges)
	assert.Equal(t, "session-1", got[3].AISession)

	assert.Equal(t, configDir, os.Getenv("CODEBUDDY_CONFIG_DIR"))
}

func TestCodeBuddyCodeParseToolOutcomes(t *testing.T) {
	_, projectDir, sessionsDir := codeBuddyCodeTestDirs(t)
	startedAt := time.Date(2026, 8, 28, 13, 0, 0, 0, time.UTC)

	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "directory"), 0o755))

	transcript := codeBuddyCodeJSONLines(t,
		codeBuddyCodeToolCallRecord(t, "failed", startedAt, projectDir, "Read", map[string]any{
			"file_path": "failed.go",
		}),
		codeBuddyCodeToolResult("failed", startedAt.Add(time.Second), "incomplete", map[string]any{
			"error": "file does not exist",
		}),
		codeBuddyCodeToolCallRecord(t, "structured-error", startedAt.Add(2*time.Second), projectDir, "Write", map[string]any{
			"file_path": "structured-error.go",
			"content":   "package main",
		}),
		map[string]any{
			"id":        "structured-error-result",
			"callId":    "structured-error",
			"timestamp": startedAt.Add(3 * time.Second).UnixMilli(),
			"type":      "function_call_result",
			"status":    "completed",
			"output":    `{"isError":true,"content":[{"type":"text","text":"denied"}]}`,
		},
		codeBuddyCodeToolCallRecord(t, "directory", startedAt.Add(4*time.Second), projectDir, "Read", map[string]any{
			"file_path": "directory",
		}),
		codeBuddyCodeToolResult("directory", startedAt.Add(5*time.Second), "completed", map[string]any{
			"content":  "directory listing",
			"renderer": map[string]any{"type": "list"},
		}),
		codeBuddyCodeToolCallRecord(t, "success", startedAt.Add(6*time.Second), projectDir, "Read", map[string]any{
			"file_path": "main.go",
		}),
		codeBuddyCodeToolResult("success", startedAt.Add(7*time.Second), "completed", map[string]any{
			"content": "package main",
		}),
	)

	codeBuddyCodeWriteTranscript(t, filepath.Join(sessionsDir, "outcomes.jsonl"), transcript, startedAt)

	got, err := (CodeBuddyCode{After: startedAt.Add(-time.Minute)}).Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, filepath.ToSlash(filepath.Join(projectDir, "main.go")), got[1].Entity)
	assert.Equal(t, heartbeat.FileType, got[1].EntityType)
	assert.Equal(t, heartbeat.PointerTo(false), got[1].IsWrite)
}

func TestCodeBuddyCodeParseSubagentFileActivityOnly(t *testing.T) {
	_, projectDir, sessionsDir := codeBuddyCodeTestDirs(t)
	startedAt := time.Date(2026, 8, 28, 14, 0, 0, 0, time.UTC)
	subagentsDir := filepath.Join(sessionsDir, "parent-session", "subagents")
	require.NoError(t, os.MkdirAll(subagentsDir, 0o755))

	transcript := codeBuddyCodeJSONLines(t,
		map[string]any{
			"id":        "usage-only",
			"timestamp": startedAt.UnixMilli(),
			"type":      "message",
			"role":      "assistant",
			"message": map[string]any{
				"usage": map[string]any{"input_tokens": 9999, "output_tokens": 9999},
			},
			"providerData": map[string]any{"model": "hunyuan-t1"},
		},
		codeBuddyCodeToolCallRecord(t, "subagent-edit", startedAt.Add(time.Second), projectDir, "Edit", map[string]any{
			"file_path":  "subagent.go",
			"old_string": "old",
			"new_string": "new\nline",
		}),
		codeBuddyCodeToolResult("subagent-edit", startedAt.Add(2*time.Second), "completed", map[string]any{
			"content": "Successfully edited file",
		}),
	)

	codeBuddyCodeWriteTranscript(t, filepath.Join(subagentsDir, "agent-1.jsonl"), transcript, startedAt)

	got, err := (CodeBuddyCode{
		After:             startedAt.Add(-time.Minute),
		FallbackUserAgent: "plugin/0.0.1",
	}).Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, filepath.ToSlash(filepath.Join(projectDir, "subagent.go")), got[0].Entity)
	assert.Equal(t, heartbeat.FileType, got[0].EntityType)
	assert.Equal(t, "parent-session/agent-1", got[0].AISession)
	assert.Equal(t, heartbeat.PointerTo(1), got[0].AILineChanges)
	assert.Zero(t, got[0].AIInputTokens)
	assert.Zero(t, got[0].AIOutputTokens)
	assert.Contains(t, got[0].UserAgent, "t/1")
}

func TestCodeBuddyCodeInternalPrompts(t *testing.T) {
	startedAt := time.Date(2026, 8, 28, 15, 0, 0, 0, time.UTC)
	values := []any{
		map[string]any{
			"timestamp": startedAt.UnixMilli(), "type": "message", "role": "user", "content": "meta",
			"providerData": map[string]any{"isMeta": true},
		},
		map[string]any{
			"timestamp": startedAt.Add(time.Second).UnixMilli(), "type": "message", "role": "user", "content": "skip",
			"providerData": map[string]any{"skipRun": true},
		},
		map[string]any{
			"timestamp": startedAt.Add(2 * time.Second).UnixMilli(), "type": "message", "role": "user", "content": "compact",
			"providerData": map[string]any{"agent": "compact"},
		},
		map[string]any{
			"timestamp": startedAt.Add(2500 * time.Millisecond).UnixMilli(),
			"type":      "custom-title",
			"title":     "generated title",
		},
		map[string]any{
			"timestamp": startedAt.Add(3 * time.Second).UnixMilli(), "type": "message", "role": "user", "content": "cron",
			"providerData": map[string]any{"isMeta": true, "startsNewUserRequest": true},
		},
	}

	got := (CodeBuddyCode{After: startedAt.Add(-time.Minute)}).parseTranscript(
		t.Context(),
		codeBuddyCodeTranscript{path: filepath.Join("project", "session.jsonl")},
		values,
	)
	require.Len(t, got, 1)
	assert.Equal(t, len("cron"), got[0].AIPromptLength)
}

func TestCodeBuddyCodeLineChanges(t *testing.T) {
	t.Run("replace all", func(t *testing.T) {
		changes, ok := codeBuddyCodeEditLineChanges(map[string]any{
			"old_string":  "old",
			"new_string":  "new\nline",
			"replace_all": true,
		}, map[string]any{
			"providerData": map[string]any{
				"toolResult": map[string]any{"title": "Made 3 replacements"},
			},
		})
		require.True(t, ok)
		assert.Equal(t, 3, changes)
	})

	t.Run("multi edit", func(t *testing.T) {
		changes, ok := codeBuddyCodeMultiEditLineChanges(map[string]any{
			"edits": []any{
				map[string]any{"old_string": "one", "new_string": "one\ntwo"},
				map[string]any{"old_string": "three\nfour", "new_string": "three"},
			},
		})
		require.True(t, ok)
		assert.Zero(t, changes)
	})

	t.Run("write overwrite is unknown", func(t *testing.T) {
		_, isWrite, changes := codeBuddyCodeToolHeartbeatInfo(codeBuddyCodeToolCall{
			name:      "Write",
			arguments: map[string]any{"file_path": "main.go", "content": "one\ntwo"},
		}, map[string]any{
			"output": "Successfully overwrote file: main.go",
		})
		assert.True(t, isWrite)
		assert.Nil(t, changes)
	})

	t.Run("write new file", func(t *testing.T) {
		_, isWrite, changes := codeBuddyCodeToolHeartbeatInfo(codeBuddyCodeToolCall{
			name:      "Write",
			arguments: map[string]any{"file_path": "main.go", "content": "one\ntwo"},
		}, map[string]any{
			"output": "Successfully created and wrote to new file: main.go",
		})
		assert.True(t, isWrite)
		assert.Equal(t, heartbeat.PointerTo(2), changes)
	})
}

func TestCodeBuddyCodeParseNoDataDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("CODEBUDDY_CONFIG_DIR", "")

	got, err := (CodeBuddyCode{After: time.Now()}).Parse(t.Context())
	require.NoError(t, err)
	assert.Empty(t, got)
}

func codeBuddyCodeTestDirs(t *testing.T) (string, string, string) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	configDir := filepath.Join(home, "codebuddy-data")
	t.Setenv("CODEBUDDY_CONFIG_DIR", configDir)

	projectDir := filepath.Join(home, "workspace", "project")
	sessionsDir := filepath.Join(configDir, "projects", "project-hash")
	require.NoError(t, os.MkdirAll(sessionsDir, 0o755))

	return configDir, projectDir, sessionsDir
}

func codeBuddyCodeToolCallRecord(
	t *testing.T,
	callID string,
	timestamp time.Time,
	cwd string,
	name string,
	arguments map[string]any,
) map[string]any {
	t.Helper()

	encoded, err := json.Marshal(arguments)
	require.NoError(t, err)

	return map[string]any{
		"id":        "item-" + callID,
		"callId":    callID,
		"timestamp": timestamp.UnixMilli(),
		"type":      "function_call",
		"name":      name,
		"arguments": string(encoded),
		"cwd":       cwd,
	}
}

func codeBuddyCodeToolResult(
	callID string,
	timestamp time.Time,
	status string,
	toolResult map[string]any,
) map[string]any {
	return map[string]any{
		"id":        "result-" + callID,
		"callId":    callID,
		"timestamp": timestamp.UnixMilli(),
		"type":      "function_call_result",
		"status":    status,
		"output":    map[string]any{"type": "text", "text": toolResult["content"]},
		"providerData": map[string]any{
			"toolResult": toolResult,
		},
	}
}

func codeBuddyCodeWriteStaleTranscript(t *testing.T, sessionsDir string, startedAt time.Time) {
	t.Helper()

	stalePath := filepath.Join(sessionsDir, "stale.jsonl")
	stale := codeBuddyCodeJSONLines(t, map[string]any{
		"timestamp": startedAt.UnixMilli(),
		"type":      "message",
		"role":      "user",
		"content":   "ignore stale transcript",
	})
	codeBuddyCodeWriteTranscript(t, stalePath, stale, startedAt.Add(-2*time.Minute))
}

func codeBuddyCodeWriteTranscript(t *testing.T, path string, contents string, modified time.Time) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
	require.NoError(t, os.Chtimes(path, modified, modified))
}

func codeBuddyCodeJSONLines(t *testing.T, records ...map[string]any) string {
	t.Helper()

	lines := make([]string, 0, len(records))
	for _, record := range records {
		encoded, err := json.Marshal(record)
		require.NoError(t, err)

		lines = append(lines, string(encoded))
	}

	return strings.Join(lines, "\n") + "\n"
}
