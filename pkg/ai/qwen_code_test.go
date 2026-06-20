package ai_test

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestQwenCodeParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("QWEN_HOME", "")
	t.Setenv("QWEN_RUNTIME_DIR", "")
	t.Setenv("USERPROFILE", home)

	projectDir := filepath.Join(home, "project")
	require.NoError(t, os.MkdirAll(projectDir, 0o755))

	mainFile := filepath.Join(projectDir, "main.go")
	notesFile := filepath.Join(projectDir, "notes.txt")
	readmeFile := filepath.Join(projectDir, "README.md")

	chatsDir := filepath.Join(home, ".qwen", "projects", "-project", "chats")
	require.NoError(t, os.MkdirAll(chatsDir, 0o755))

	sessionID := "426ee865-f7b9-4a3f-b9bc-5e0efd602bde"
	base := map[string]interface{}{
		"cwd":       projectDir,
		"sessionId": sessionID,
		"version":   "0.17.0",
	}
	record := func(fields map[string]interface{}) string {
		t.Helper()

		values := make(map[string]interface{}, len(base)+len(fields))
		for key, value := range base {
			values[key] = value
		}

		for key, value := range fields {
			values[key] = value
		}

		return mustJSONLine(t, values)
	}

	transcript := strings.Join([]string{
		record(map[string]interface{}{
			"type":      "assistant",
			"timestamp": "2026-05-30T11:59:59Z",
			"message": map[string]interface{}{
				"role": "model",
				"parts": []map[string]interface{}{{
					"functionCall": map[string]interface{}{
						"id":   "boundary-write",
						"name": "write_file",
						"args": map[string]interface{}{
							"file_path": "boundary.txt",
							"content":   "boundary",
						},
					},
				}},
			},
			"usageMetadata": map[string]interface{}{
				"promptTokenCount":     3,
				"candidatesTokenCount": 1,
			},
		}),
		record(map[string]interface{}{
			"type":      "user",
			"timestamp": "2026-05-30T11:59:00Z",
			"message": map[string]interface{}{
				"role":  "user",
				"parts": []map[string]interface{}{{"text": "stale prompt"}},
			},
		}),
		record(map[string]interface{}{
			"type":      "tool_result",
			"timestamp": "2026-05-30T12:00:00Z",
			"toolCallResult": map[string]interface{}{
				"callId": "boundary-write",
				"status": "success",
			},
		}),
		record(map[string]interface{}{
			"type":      "user",
			"timestamp": "2026-05-30T12:00:01Z",
			"message": map[string]interface{}{
				"role":  "user",
				"parts": []map[string]interface{}{{"text": "inspect the project and update the files"}},
			},
		}),
		record(map[string]interface{}{
			"type":      "assistant",
			"timestamp": "2026-05-30T12:00:02Z",
			"model":     "qwen3-coder-plus",
			"message": map[string]interface{}{
				"role":  "model",
				"parts": []map[string]interface{}{{"text": "I will update the files."}},
			},
			"usageMetadata": map[string]interface{}{
				"promptTokenCount":     100,
				"candidatesTokenCount": 20,
			},
		}),
		record(map[string]interface{}{
			"type":      "assistant",
			"timestamp": "2026-05-30T12:00:03Z",
			"message": map[string]interface{}{
				"role": "model",
				"parts": []map[string]interface{}{{
					"functionCall": map[string]interface{}{
						"id":   "write-1",
						"name": "write_file",
						"args": map[string]interface{}{
							"file_path": "notes.txt",
							"content":   "alpha\nbeta",
						},
					},
				}},
			},
			"usageMetadata": map[string]interface{}{
				"promptTokenCount":     50,
				"candidatesTokenCount": 10,
			},
		}),
		record(map[string]interface{}{
			"type":      "tool_result",
			"timestamp": "2026-05-30T12:00:04Z",
			"toolCallResult": map[string]interface{}{
				"callId": "write-1",
				"status": "success",
			},
		}),
		record(map[string]interface{}{
			"type":      "assistant",
			"timestamp": "2026-05-30T12:00:05Z",
			"message": map[string]interface{}{
				"role": "model",
				"parts": []map[string]interface{}{{
					"functionCall": map[string]interface{}{
						"id":   "edit-1",
						"name": "edit",
						"args": map[string]interface{}{
							"old_string": "old line",
							"new_string": "new line\nextra line",
						},
					},
				}},
			},
			"usageMetadata": map[string]interface{}{
				"promptTokenCount":     25,
				"candidatesTokenCount": 5,
			},
		}),
		record(map[string]interface{}{
			"type":      "tool_result",
			"timestamp": "2026-05-30T12:00:06Z",
			"message": map[string]interface{}{
				"role": "user",
				"parts": []map[string]interface{}{{
					"functionResponse": map[string]interface{}{
						"id":   "edit-1",
						"name": "edit",
					},
				}},
			},
			"toolCallResult": map[string]interface{}{
				"status": "success",
				"resultDisplay": map[string]interface{}{
					"fileName": "main.go",
					"diffStat": map[string]interface{}{
						"model_added_lines":   3,
						"model_removed_lines": 1,
					},
				},
			},
		}),
		record(map[string]interface{}{
			"type":      "assistant",
			"timestamp": "2026-05-30T12:00:07Z",
			"message": map[string]interface{}{
				"role": "model",
				"parts": []map[string]interface{}{{
					"functionCall": map[string]interface{}{
						"name": "read_file",
						"args": `{"file_path":"README.md"}`,
					},
				}},
			},
			"usageMetadata": map[string]interface{}{
				"promptTokenCount":     7,
				"candidatesTokenCount": 2,
			},
		}),
		record(map[string]interface{}{
			"type":      "tool_result",
			"timestamp": "2026-05-30T12:00:08Z",
			"message": map[string]interface{}{
				"role": "user",
				"parts": []map[string]interface{}{{
					"functionResponse": map[string]interface{}{
						"id":   "read_file-generated",
						"name": "read_file",
					},
				}},
			},
			"toolCallResult": map[string]interface{}{
				"callId": "read_file-generated",
				"status": "success",
			},
		}),
		record(map[string]interface{}{
			"type":      "assistant",
			"timestamp": "2026-05-30T12:00:09Z",
			"message": map[string]interface{}{
				"role": "model",
				"parts": []map[string]interface{}{{
					"functionCall": map[string]interface{}{
						"id":   "cancelled-edit",
						"name": "edit",
						"args": map[string]interface{}{"file_path": "ignored.go"},
					},
				}},
			},
		}),
		record(map[string]interface{}{
			"type":      "tool_result",
			"timestamp": "2026-05-30T12:00:10Z",
			"toolCallResult": map[string]interface{}{
				"callId": "cancelled-edit",
				"status": "cancelled",
			},
		}),
		record(map[string]interface{}{
			"subtype":   "notification",
			"type":      "user",
			"timestamp": "2026-05-30T12:00:10Z",
			"message": map[string]interface{}{
				"role":  "user",
				"parts": []map[string]interface{}{{"text": "background task completed"}},
			},
		}),
		record(map[string]interface{}{
			"forkedFrom": map[string]interface{}{"sessionId": "old-session"},
			"type":       "user",
			"timestamp":  "2026-05-30T12:00:10Z",
			"message": map[string]interface{}{
				"role":  "user",
				"parts": []map[string]interface{}{{"text": "copied prompt"}},
			},
		}),
	}, "\n") + "\n"
	transcript = strings.Replace(transcript, "}\n{", "}{", 1)
	require.NoError(t, os.WriteFile(filepath.Join(chatsDir, sessionID+".jsonl"), []byte(transcript), 0o644))

	parser := ai.QwenCode{
		After:             time.Date(2026, 5, 30, 12, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			mainFile: heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 6)

	assert.Equal(t, filepath.Join(projectDir, "boundary.txt"), got[0].Entity)
	assert.Equal(t, heartbeat.FileType, got[0].EntityType)
	assert.Equal(t, sessionID, got[0].AISession)
	require.NotNil(t, got[0].AILineChanges)
	assert.Equal(t, 1, *got[0].AILineChanges)
	assert.Zero(t, got[0].AIInputTokens)
	assert.Zero(t, got[0].AIOutputTokens)
	assert.Contains(t, got[0].UserAgent, "qwen-code-cli/0.17.0")
	assert.NotContains(t, got[0].UserAgent, "Qwen Code/")

	assert.Equal(t, "Qwen Code "+sessionID, got[1].Entity)
	assert.Equal(t, heartbeat.AppType, got[1].EntityType)
	assert.Equal(t, sessionID, got[1].AISession)
	assert.Equal(t, len([]rune("inspect the project and update the files")), got[1].AIPromptLength)
	assert.Equal(t, projectDir, got[1].ProjectPathOverride)

	assert.Equal(t, "Qwen Code "+sessionID, got[2].Entity)
	assert.EqualValues(t, 100, got[2].AIInputTokens)
	assert.EqualValues(t, 20, got[2].AIOutputTokens)
	assert.Contains(t, got[2].UserAgent, "qwen/3-coder-plus qwen-code-cli/0.17.0")
	assert.True(t, strings.Index(got[2].UserAgent, "qwen/3-coder-plus") <
		strings.Index(got[2].UserAgent, "qwen-code-cli/0.17.0"))

	assert.Equal(t, notesFile, got[3].Entity)
	require.NotNil(t, got[3].AILineChanges)
	assert.Equal(t, 2, *got[3].AILineChanges)
	assert.EqualValues(t, 50, got[3].AIInputTokens)
	assert.EqualValues(t, 10, got[3].AIOutputTokens)
	require.NotNil(t, got[3].IsWrite)
	assert.True(t, *got[3].IsWrite)

	assert.Equal(t, mainFile, got[4].Entity)
	require.NotNil(t, got[4].AILineChanges)
	assert.Equal(t, 2, *got[4].AILineChanges)
	assert.EqualValues(t, 25, got[4].AIInputTokens)
	assert.EqualValues(t, 5, got[4].AIOutputTokens)
	require.NotNil(t, got[4].IsWrite)
	assert.True(t, *got[4].IsWrite)
	assert.Contains(t, got[4].UserAgent, "qwen/3-coder-plus qwen-code-cli/0.17.0")
	assert.True(t, strings.Index(got[4].UserAgent, "qwen/3-coder-plus") <
		strings.Index(got[4].UserAgent, "qwen-code-cli/0.17.0"))
	assert.Contains(t, got[4].UserAgent, "editor/1.2.3")

	assert.Equal(t, readmeFile, got[5].Entity)
	require.NotNil(t, got[5].AILineChanges)
	assert.Zero(t, *got[5].AILineChanges)
	assert.EqualValues(t, 7, got[5].AIInputTokens)
	assert.EqualValues(t, 2, got[5].AIOutputTokens)
	require.NotNil(t, got[5].IsWrite)
	assert.False(t, *got[5].IsWrite)
}

func TestQwenCodeParse_NoProjectsDir(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("QWEN_HOME", "")
	t.Setenv("QWEN_RUNTIME_DIR", "")
	t.Setenv("USERPROFILE", home)

	got, err := ai.QwenCode{After: time.Now()}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestQwenCodeParse_RuntimeDirFromCommentedSettings(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("QWEN_HOME", "")
	t.Setenv("QWEN_RUNTIME_DIR", "")
	t.Setenv("USERPROFILE", home)

	runtimeDir := filepath.Join(home, "runtime")
	chatsDir := filepath.Join(runtimeDir, "projects", "-project", "chats")
	require.NoError(t, os.MkdirAll(chatsDir, 0o755))

	settings := `{
		// Qwen accepts comments in settings.json.
		"advanced": {
			"runtimeOutputDir": ` + strconv.Quote(runtimeDir) + `
		}
	}`

	require.NoError(t, os.MkdirAll(filepath.Join(home, ".qwen"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".qwen", "settings.json"), []byte(settings), 0o644))

	timestamp := time.Date(2026, 5, 30, 12, 0, 0, 0, time.UTC)
	transcript := mustJSONLine(t, map[string]interface{}{
		"cwd":       filepath.Join(home, "project"),
		"sessionId": "session-from-settings",
		"timestamp": timestamp,
		"type":      "user",
		"message": map[string]interface{}{
			"role":  "user",
			"parts": []map[string]interface{}{{"text": "use configured runtime directory"}},
		},
	}) + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(chatsDir, "session-from-settings.jsonl"), []byte(transcript), 0o644))

	got, err := ai.QwenCode{After: timestamp.Add(-time.Second)}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "Qwen Code session-from-settings", got[0].Entity)
}
