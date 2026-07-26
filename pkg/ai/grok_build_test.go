package ai_test

import (
	"context"
	"net/url"
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

func TestGrokBuildParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("GROK_HOME", "")

	projectDir := filepath.Join(home, "project")
	require.NoError(t, os.MkdirAll(projectDir, 0o755))

	editFile := filepath.Join(projectDir, "main.go")
	skipPlan := filepath.Join(home, ".grok", "sessions", "ignore", "plan.md")

	sessionID := "019f90c0-6137-79c2-b26e-ef83b2cecfd1"
	encodedCwd := url.PathEscape(projectDir)
	sessionDir := filepath.Join(home, ".grok", "sessions", encodedCwd, sessionID)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(home, ".grok", "version.json"), []byte(`{
  "version": "0.2.111",
  "stable_version": "0.2.111"
}`), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "summary.json"), []byte(`{
  "info": {"id": "`+sessionID+`", "cwd": "`+filepath.ToSlash(projectDir)+`"},
  "current_model_id": "grok-4.5",
  "git_root_dir": "`+filepath.ToSlash(projectDir)+`/",
  "updated_at": "2026-07-23T22:00:00Z",
  "last_active_at": "2026-07-23T22:00:00Z"
}`), 0o644))

	updates := strings.Join([]string{
		mustJSONLine(t, map[string]interface{}{
			"timestamp": 1784841973,
			"method":    "session/update",
			"params": map[string]interface{}{
				"sessionId": sessionID,
				"update": map[string]interface{}{
					"sessionUpdate": "user_message_chunk",
					"content":       map[string]interface{}{"type": "text", "text": "fix main.go"},
					"_meta":         map[string]interface{}{"modelId": "grok-4.5", "promptIndex": 0},
				},
				"_meta": map[string]interface{}{
					"eventId":          sessionID + "-4",
					"agentTimestampMs": int64(1784841972430),
				},
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"timestamp": 1784844590,
			"method":    "_x.ai/session/update",
			"params": map[string]interface{}{
				"sessionId": sessionID,
				"update": map[string]interface{}{
					"sessionUpdate": "turn_completed",
					"usage": map[string]interface{}{
						"inputTokens":  100,
						"outputTokens": 40,
					},
				},
				"_meta": map[string]interface{}{
					"eventId":          sessionID + "-5113",
					"agentTimestampMs": int64(1784844590945),
				},
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"timestamp": 1784845683,
			"method":    "session/update",
			"params": map[string]interface{}{
				"sessionId": sessionID,
				"update": map[string]interface{}{
					"sessionUpdate": "user_message_chunk",
					"content":       map[string]interface{}{"type": "text", "text": "commit"},
					"_meta":         map[string]interface{}{"modelId": "grok-4.5", "promptIndex": 1},
				},
				"_meta": map[string]interface{}{
					"eventId":          sessionID + "-5114",
					"agentTimestampMs": int64(1784845681172),
				},
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"timestamp": 1784845690,
			"method":    "_x.ai/session/update",
			"params": map[string]interface{}{
				"sessionId": sessionID,
				"update": map[string]interface{}{
					"sessionUpdate": "turn_completed",
					"usage": map[string]interface{}{
						"inputTokens":  20,
						"outputTokens": 5,
					},
				},
				"_meta": map[string]interface{}{
					"eventId":          sessionID + "-5153",
					"agentTimestampMs": int64(1784845690523),
				},
			},
		}),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "updates.jsonl"), []byte(updates), 0o644))

	hunks := strings.Join([]string{
		mustJSONLine(t, map[string]interface{}{
			"hunkId":       "h1",
			"filePath":     filepath.ToSlash(editFile),
			"linesAdded":   3,
			"linesRemoved": 1,
			"authorType":   "agent",
			"sourceType":   "agentEdit",
			"eventType":    "added",
			"promptIndex":  0,
			"sessionId":    sessionID,
			"timestamp":    "2026-07-23T22:09:21.836002Z",
		}),
		mustJSONLine(t, map[string]interface{}{
			"hunkId":       "h2",
			"filePath":     filepath.ToSlash(skipPlan),
			"linesAdded":   10,
			"linesRemoved": 0,
			"authorType":   "agent",
			"sourceType":   "agentEdit",
			"eventType":    "added",
			"sessionId":    sessionID,
			"timestamp":    "2026-07-23T22:09:22.000000Z",
		}),
		mustJSONLine(t, map[string]interface{}{
			"hunkId":       "h3",
			"filePath":     filepath.ToSlash(editFile),
			"linesAdded":   0,
			"linesRemoved": 2,
			"authorType":   "human",
			"sourceType":   "humanEdit",
			"eventType":    "removed",
			"sessionId":    sessionID,
			"timestamp":    "2026-07-23T22:09:23.000000Z",
		}),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "hunk_records.jsonl"), []byte(hunks), 0o644))

	parser := ai.GrokBuild{
		After:             time.Date(2026, 7, 23, 20, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			editFile: heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 3)

	assert.Equal(t, "Grok Build "+sessionID, got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, sessionID, got[0].AISession)
	assert.Equal(t, len([]rune("fix main.go")), got[0].AIPromptLength)
	assert.Equal(t, projectDir, got[0].ProjectPathOverride)
	assert.Equal(t, int64(100), got[0].AIInputTokens)
	assert.Equal(t, int64(40), got[0].AIOutputTokens)
	require.NotNil(t, got[0].IsWrite)
	assert.False(t, *got[0].IsWrite)
	assert.Contains(t, got[0].UserAgent, "grok/4.5")
	assert.Contains(t, got[0].UserAgent, "Grok Build/0.2.111")

	assert.Equal(t, editFile, got[1].Entity)
	assert.Equal(t, heartbeat.FileType, got[1].EntityType)
	assert.Equal(t, sessionID, got[1].AISession)
	require.NotNil(t, got[1].AILineChanges)
	assert.Equal(t, 2, *got[1].AILineChanges)
	require.NotNil(t, got[1].IsWrite)
	assert.True(t, *got[1].IsWrite)
	assert.Contains(t, got[1].UserAgent, "grok/4.5")
	assert.Contains(t, got[1].UserAgent, "Grok Build/0.2.111")
	assert.Contains(t, got[1].UserAgent, "editor/1.2.3")
	assert.True(t, strings.Index(got[1].UserAgent, "grok/4.5") <
		strings.Index(got[1].UserAgent, "editor/1.2.3"))

	assert.Equal(t, "Grok Build "+sessionID, got[2].Entity)
	assert.Equal(t, heartbeat.AppType, got[2].EntityType)
	assert.Equal(t, len([]rune("commit")), got[2].AIPromptLength)
	assert.Equal(t, int64(20), got[2].AIInputTokens)
	assert.Equal(t, int64(5), got[2].AIOutputTokens)
	require.NotNil(t, got[2].IsWrite)
	assert.False(t, *got[2].IsWrite)
}

func TestGrokBuildParse_NoSessionsDir(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("GROK_HOME", "")

	got, err := ai.GrokBuild{After: time.Now()}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGrokBuildParse_OversizedLineContinues(t *testing.T) {
	ctx := context.Background()
	home, sessionDir, sessionID, projectDir := setupGrokBuildSession(t)

	editFile := filepath.Join(projectDir, "main.go")
	oversized := strings.Repeat("x", 10*1024*1024+16)
	updates := strings.Join([]string{
		`{"timestamp":1784841970,"method":"session/update","params":{"sessionId":"` + sessionID + `","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"` + oversized + `"}}}}`,
		mustJSONLine(t, map[string]interface{}{
			"timestamp": 1784841973,
			"method":    "session/update",
			"params": map[string]interface{}{
				"sessionId": sessionID,
				"update": map[string]interface{}{
					"sessionUpdate": "user_message_chunk",
					"content":       map[string]interface{}{"type": "text", "text": "still tracked"},
					"_meta":         map[string]interface{}{"modelId": "grok-4.5", "promptIndex": 0},
				},
				"_meta": map[string]interface{}{
					"agentTimestampMs": int64(1784841972430),
				},
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"timestamp": 1784844590,
			"method":    "_x.ai/session/update",
			"params": map[string]interface{}{
				"sessionId": sessionID,
				"update": map[string]interface{}{
					"sessionUpdate": "turn_completed",
					"usage": map[string]interface{}{
						"inputTokens":  10,
						"outputTokens": 2,
					},
				},
				"_meta": map[string]interface{}{
					"agentTimestampMs": int64(1784844590945),
				},
			},
		}),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "updates.jsonl"), []byte(updates), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "hunk_records.jsonl"), []byte(
		mustJSONLine(t, map[string]interface{}{
			"hunkId":       "h1",
			"filePath":     filepath.ToSlash(editFile),
			"linesAdded":   1,
			"linesRemoved": 0,
			"authorType":   "agent",
			"sourceType":   "agentEdit",
			"eventType":    "added",
			"promptIndex":  0,
			"sessionId":    sessionID,
			"timestamp":    "2026-07-23T22:09:21.836002Z",
		})+"\n",
	), 0o644))

	got, err := ai.GrokBuild{
		After:             time.Date(2026, 7, 23, 20, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, len([]rune("still tracked")), got[0].AIPromptLength)
	assert.Equal(t, editFile, got[1].Entity)
	_ = home
}

func TestGrokBuildParse_LegacyRawACP(t *testing.T) {
	ctx := context.Background()
	_, sessionDir, sessionID, _ := setupGrokBuildSession(t)

	updates := `{"sessionId":"` + sessionID + `","update":{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"legacy prompt"},"_meta":{"modelId":"grok-4.5","promptIndex":0,"agentTimestampMs":1784841972430}}}` + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "updates.jsonl"), []byte(updates), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "hunk_records.jsonl"), []byte{}, 0o644))

	got, err := ai.GrokBuild{
		After:             time.Date(2026, 7, 23, 20, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, len([]rune("legacy prompt")), got[0].AIPromptLength)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
}

func TestGrokBuildParse_AggregatesUserChunks(t *testing.T) {
	ctx := context.Background()
	_, sessionDir, sessionID, _ := setupGrokBuildSession(t)

	updates := strings.Join([]string{
		mustJSONLine(t, map[string]interface{}{
			"timestamp": 1784841973,
			"method":    "session/update",
			"params": map[string]interface{}{
				"sessionId": sessionID,
				"update": map[string]interface{}{
					"sessionUpdate": "user_message_chunk",
					"content":       map[string]interface{}{"type": "text", "text": "fix "},
					"_meta":         map[string]interface{}{"modelId": "grok-4.5", "promptIndex": 0},
				},
				"_meta": map[string]interface{}{"agentTimestampMs": int64(1784841972430)},
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"timestamp": 1784841974,
			"method":    "session/update",
			"params": map[string]interface{}{
				"sessionId": sessionID,
				"update": map[string]interface{}{
					"sessionUpdate": "user_message_chunk",
					"content": map[string]interface{}{
						"type":  "text",
						"text":  "! echo hi",
						"_meta": map[string]interface{}{"bash_command": "echo hi"},
					},
					"_meta": map[string]interface{}{"modelId": "grok-4.5", "promptIndex": 0},
				},
				"_meta": map[string]interface{}{"agentTimestampMs": int64(1784841972500)},
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"timestamp": 1784841975,
			"method":    "session/update",
			"params": map[string]interface{}{
				"sessionId": sessionID,
				"update": map[string]interface{}{
					"sessionUpdate": "user_message_chunk",
					"content":       map[string]interface{}{"type": "text", "text": "main.go"},
					"_meta":         map[string]interface{}{"modelId": "grok-4.5", "promptIndex": 0},
				},
				"_meta": map[string]interface{}{"agentTimestampMs": int64(1784841972600)},
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"timestamp": 1784841976,
			"method":    "session/update",
			"params": map[string]interface{}{
				"sessionId": sessionID,
				"update": map[string]interface{}{
					"sessionUpdate": "user_message_chunk",
					"content":       map[string]interface{}{"type": "text", "text": "host"},
					"_meta":         map[string]interface{}{"modelId": "grok-4.5", "promptIndex": 0, "hostTurn": true},
				},
				"_meta": map[string]interface{}{"agentTimestampMs": int64(1784841972700)},
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"timestamp": 1784841978,
			"method":    "_x.ai/session/update",
			"params": map[string]interface{}{
				"sessionId": sessionID,
				"update": map[string]interface{}{
					"sessionUpdate": "turn_completed",
					"usage": map[string]interface{}{
						"inputTokens":  11,
						"outputTokens": 1,
					},
				},
				"_meta": map[string]interface{}{"agentTimestampMs": int64(1784841972900)},
			},
		}),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "updates.jsonl"), []byte(updates), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "hunk_records.jsonl"), []byte{}, 0o644))

	got, err := ai.GrokBuild{
		After:             time.Date(2026, 7, 23, 20, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, len([]rune("fix main.go")), got[0].AIPromptLength)
	assert.Equal(t, int64(11), got[0].AIInputTokens)
}

func TestGrokBuildParse_RejectedHunkNegates(t *testing.T) {
	ctx := context.Background()
	_, sessionDir, sessionID, projectDir := setupGrokBuildSession(t)

	editFile := filepath.Join(projectDir, "main.go")
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "updates.jsonl"), []byte{}, 0o644))

	hunks := strings.Join([]string{
		mustJSONLine(t, map[string]interface{}{
			"hunkId":       "h-reject",
			"filePath":     filepath.ToSlash(editFile),
			"linesAdded":   2,
			"linesRemoved": 0,
			"authorType":   "agent",
			"sourceType":   "agentEdit",
			"eventType":    "added",
			"promptIndex":  0,
			"sessionId":    sessionID,
			"timestamp":    "2026-07-23T22:09:21.000000Z",
		}),
		mustJSONLine(t, map[string]interface{}{
			"hunkId":        "h-reject",
			"filePath":      filepath.ToSlash(editFile),
			"linesAdded":    -2,
			"linesRemoved":  0,
			"eventType":     "removed",
			"removalReason": "rejected",
			"sessionId":     sessionID,
			"timestamp":     "2026-07-23T22:09:22.000000Z",
		}),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "hunk_records.jsonl"), []byte(hunks), 0o644))

	got, err := ai.GrokBuild{
		After:             time.Date(2026, 7, 23, 20, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.NotNil(t, got[0].AILineChanges)
	require.NotNil(t, got[1].AILineChanges)
	assert.Equal(t, 2, *got[0].AILineChanges)
	assert.Equal(t, -2, *got[1].AILineChanges)
	assert.Equal(t, 0, *got[0].AILineChanges+*got[1].AILineChanges)
}

func TestGrokBuildParse_HunkUsesPromptIndexModel(t *testing.T) {
	ctx := context.Background()
	_, sessionDir, sessionID, projectDir := setupGrokBuildSession(t)

	editFile := filepath.Join(projectDir, "main.go")
	updates := strings.Join([]string{
		mustJSONLine(t, map[string]interface{}{
			"timestamp": 1784841973,
			"method":    "session/update",
			"params": map[string]interface{}{
				"sessionId": sessionID,
				"update": map[string]interface{}{
					"sessionUpdate": "user_message_chunk",
					"content":       map[string]interface{}{"type": "text", "text": "first"},
					"_meta":         map[string]interface{}{"modelId": "grok-3", "promptIndex": 0},
				},
				"_meta": map[string]interface{}{"agentTimestampMs": int64(1784841972430)},
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"timestamp": 1784841974,
			"method":    "_x.ai/session/update",
			"params": map[string]interface{}{
				"sessionId": sessionID,
				"update": map[string]interface{}{
					"sessionUpdate": "turn_completed",
					"usage":         map[string]interface{}{"inputTokens": 1, "outputTokens": 1},
				},
				"_meta": map[string]interface{}{"agentTimestampMs": int64(1784841972500)},
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"timestamp": 1784841975,
			"method":    "session/update",
			"params": map[string]interface{}{
				"sessionId": sessionID,
				"update": map[string]interface{}{
					"sessionUpdate": "user_message_chunk",
					"content":       map[string]interface{}{"type": "text", "text": "second"},
					"_meta":         map[string]interface{}{"modelId": "grok-4.5", "promptIndex": 1},
				},
				"_meta": map[string]interface{}{"agentTimestampMs": int64(1784841972600)},
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"timestamp": 1784841976,
			"method":    "_x.ai/session/update",
			"params": map[string]interface{}{
				"sessionId": sessionID,
				"update": map[string]interface{}{
					"sessionUpdate": "turn_completed",
					"usage":         map[string]interface{}{"inputTokens": 1, "outputTokens": 1},
				},
				"_meta": map[string]interface{}{"agentTimestampMs": int64(1784841972700)},
			},
		}),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "updates.jsonl"), []byte(updates), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "hunk_records.jsonl"), []byte(
		mustJSONLine(t, map[string]interface{}{
			"hunkId":       "h-old",
			"filePath":     filepath.ToSlash(editFile),
			"linesAdded":   1,
			"linesRemoved": 0,
			"authorType":   "agent",
			"sourceType":   "agentEdit",
			"eventType":    "added",
			"promptIndex":  0,
			"sessionId":    sessionID,
			"timestamp":    "2026-07-23T22:09:21.000000Z",
		})+"\n",
	), 0o644))

	got, err := ai.GrokBuild{
		After:             time.Date(2026, 7, 23, 20, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}.Parse(ctx)
	require.NoError(t, err)

	var fileHB *heartbeat.Heartbeat
	for i := range got {
		if got[i].EntityType == heartbeat.FileType {
			fileHB = &got[i]
			break
		}
	}

	require.NotNil(t, fileHB)
	assert.Contains(t, fileHB.UserAgent, "grok/3")
	assert.NotContains(t, fileHB.UserAgent, "grok/4.5")
}

func TestGrokBuildParse_GrokHome(t *testing.T) {
	ctx := context.Background()

	userHome := t.TempDir()
	grokHome := t.TempDir()
	t.Setenv("HOME", userHome)
	t.Setenv("USERPROFILE", userHome)
	t.Setenv("GROK_HOME", grokHome)

	projectDir := filepath.Join(userHome, "project")
	require.NoError(t, os.MkdirAll(projectDir, 0o755))

	sessionID := "grok-home-session"
	sessionDir := filepath.Join(grokHome, "sessions", url.PathEscape(projectDir), sessionID)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(grokHome, "version.json"), []byte(`{"version":"9.9.9"}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "summary.json"), []byte(`{
  "info": {"id": "`+sessionID+`", "cwd": "`+filepath.ToSlash(projectDir)+`"},
  "current_model_id": "grok-4.5",
  "git_root_dir": "`+filepath.ToSlash(projectDir)+`/",
  "updated_at": "2026-07-23T22:00:00Z",
  "last_active_at": "2026-07-23T22:00:00Z"
}`), 0o644))

	updates := mustJSONLine(t, map[string]interface{}{
		"timestamp": 1784841973,
		"method":    "session/update",
		"params": map[string]interface{}{
			"sessionId": sessionID,
			"update": map[string]interface{}{
				"sessionUpdate": "user_message_chunk",
				"content":       map[string]interface{}{"type": "text", "text": "from grok home"},
				"_meta":         map[string]interface{}{"modelId": "grok-4.5", "promptIndex": 0},
			},
			"_meta": map[string]interface{}{"agentTimestampMs": int64(1784841972430)},
		},
	}) + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "updates.jsonl"), []byte(updates), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "hunk_records.jsonl"), []byte{}, 0o644))

	got, err := ai.GrokBuild{
		After:             time.Date(2026, 7, 23, 20, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, len([]rune("from grok home")), got[0].AIPromptLength)
	assert.Contains(t, got[0].UserAgent, "Grok Build/9.9.9")
}

func setupGrokBuildSession(t *testing.T) (home, sessionDir, sessionID, projectDir string) {
	t.Helper()

	home = t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("GROK_HOME", "")

	projectDir = filepath.Join(home, "project")
	require.NoError(t, os.MkdirAll(projectDir, 0o755))

	sessionID = "019f90c0-6137-79c2-b26e-ef83b2cecfd1"
	sessionDir = filepath.Join(home, ".grok", "sessions", url.PathEscape(projectDir), sessionID)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(home, ".grok", "version.json"), []byte(`{"version":"0.2.111"}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "summary.json"), []byte(`{
  "info": {"id": "`+sessionID+`", "cwd": "`+filepath.ToSlash(projectDir)+`"},
  "current_model_id": "grok-4.5",
  "git_root_dir": "`+filepath.ToSlash(projectDir)+`/",
  "updated_at": "2026-07-23T22:00:00Z",
  "last_active_at": "2026-07-23T22:00:00Z"
}`), 0o644))

	return home, sessionDir, sessionID, projectDir
}
