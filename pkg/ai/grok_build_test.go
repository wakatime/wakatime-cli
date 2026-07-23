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
			"filePath":     editFile,
			"linesAdded":   3,
			"linesRemoved": 1,
			"authorType":   "agent",
			"sourceType":   "agentEdit",
			"eventType":    "added",
			"sessionId":    sessionID,
			"timestamp":    "2026-07-23T22:09:21.836002Z",
		}),
		mustJSONLine(t, map[string]interface{}{
			"hunkId":       "h2",
			"filePath":     skipPlan,
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
			"filePath":     editFile,
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

	got, err := ai.GrokBuild{After: time.Now()}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}
