package heartbeat_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gandarez/go-realpath"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cmdheartbeat "github.com/wakatime/wakatime-cli/cmd/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
)

func TestRunGrokBuildAsEditor(t *testing.T) {
	for _, tc := range []struct {
		name       string
		setting    string
		syncOnly   bool
		aiProducts string
	}{
		{
			name:       "default preserves host editor",
			aiProducts: "grok/4.5 Cursor/1.2.3 grok-build/0.2.111",
		},
		{
			name:       "flag prefers Grok Build",
			setting:    "grok-build-as-editor",
			aiProducts: "grok/4.5 grok-build/0.2.111 Cursor/1.2.3",
		},
		{
			name:       "config prefers Grok Build",
			setting:    "settings.grok_build_as_editor",
			aiProducts: "grok/4.5 grok-build/0.2.111 Cursor/1.2.3",
		},
		{
			name:       "sync activity honors config",
			setting:    "settings.grok_build_as_editor",
			syncOnly:   true,
			aiProducts: "grok/4.5 grok-build/0.2.111 Cursor/1.2.3",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resetSingleton(t)

			home, err := realpath.Realpath(t.TempDir())
			require.NoError(t, err)

			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			t.Setenv("GROK_HOME", filepath.Join(home, ".grok"))
			t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

			base := time.Date(2026, 7, 23, 22, 0, 0, 0, time.UTC)
			entity, sessionID := writeGrokBuildCommandSession(t, home, base)

			testServerURL, router, tearDown := setupTestServer()
			defer tearDown()

			var received []struct {
				Entity         string `json:"entity"`
				AISession      string `json:"ai_session"`
				Category       string `json:"category"`
				UserAgent      string `json:"user_agent"`
				AIPromptLength int    `json:"ai_prompt_length"`
				AILineChanges  *int   `json:"ai_line_changes"`
			}

			var numCalls int

			router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
				numCalls++

				require.NoError(t, json.NewDecoder(req.Body).Decode(&received))

				responses := make([]any, len(received))
				for i := range responses {
					responses[i] = []any{map[string]any{"data": map[string]string{"id": "heartbeat"}}, http.StatusCreated}
				}

				w.WriteHeader(http.StatusCreated)
				require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"responses": responses}))
			})

			v := viper.New()
			v.Set("api-url", testServerURL)
			v.Set("category", "debugging")
			v.Set("entity", entity)
			v.Set("entity-type", "file")
			v.Set("heartbeat-rate-limit-seconds", 0)
			v.Set("internal-config", filepath.Join(home, "wakatime-internal.cfg"))
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("offline-queue-file", filepath.Join(home, "offline.db"))
			v.Set("plugin", "Cursor/1.2.3")
			v.Set("project", "myproject")
			v.Set("time", float64(base.Add(time.Hour).Unix()))
			v.Set("timeout", 5)
			v.Set("internal.ai_logs_last_parsed_at", base.Add(-time.Minute).Format(ini.DateFormat))

			if tc.setting != "" {
				v.Set(tc.setting, true)
			}

			var code int

			if tc.syncOnly {
				code, err = cmdheartbeat.RunAISyncActivity(t.Context(), v)
			} else {
				code, err = cmdheartbeat.Run(t.Context(), v)
			}

			require.NoError(t, err)
			assert.Zero(t, code)
			require.Equal(t, 1, numCalls)

			var aiCount, humanCount, promptLength, lineChanges int

			for _, h := range received {
				assert.Equal(t, entity, h.Entity)

				if h.AISession == "" {
					humanCount++

					assert.Equal(t, "debugging", h.Category)
					assert.Equal(t, heartbeat.UserAgent(t.Context(), "Cursor/1.2.3"), h.UserAgent)

					continue
				}

				aiCount++

				assert.Equal(t, sessionID, h.AISession)
				assert.Equal(t, "ai coding", h.Category)
				assert.True(t, strings.HasSuffix(h.UserAgent, " "+tc.aiProducts), h.UserAgent)
				promptLength += h.AIPromptLength

				if h.AILineChanges != nil {
					lineChanges += *h.AILineChanges
				}
			}

			assert.Equal(t, 2, aiCount)
			assert.Equal(t, len([]rune("fix main.go")), promptLength)
			assert.Equal(t, 2, lineChanges)

			if tc.syncOnly {
				assert.Zero(t, humanCount)
			} else {
				assert.Equal(t, 1, humanCount)
			}
		})
	}
}

func writeGrokBuildCommandSession(t *testing.T, home string, base time.Time) (string, string) {
	t.Helper()

	projectDir := filepath.Join(home, "project")
	require.NoError(t, os.MkdirAll(projectDir, 0o755))

	entity := filepath.Join(projectDir, "main.go")
	copyFile(t, "testdata/main.go", entity)

	sessionID := "019f90c0-6137-79c2-b26e-ef83b2cecfd1"
	grokHome := filepath.Join(home, ".grok")
	sessionDir := filepath.Join(grokHome, "sessions", "project", sessionID)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	writeGrokBuildCommandJSON(t, filepath.Join(grokHome, "version.json"), map[string]any{"version": "0.2.111"})
	writeGrokBuildCommandJSON(t, filepath.Join(sessionDir, "summary.json"), map[string]any{
		"info":             map[string]any{"id": sessionID, "cwd": filepath.ToSlash(projectDir)},
		"current_model_id": "grok-4.5",
		"git_root_dir":     filepath.ToSlash(projectDir),
	})
	writeGrokBuildCommandJSON(t, filepath.Join(sessionDir, "updates.jsonl"), map[string]any{
		"timestamp": base.Add(time.Second).Unix(),
		"method":    "session/update",
		"params": map[string]any{
			"sessionId": sessionID,
			"_meta":     map[string]any{"agentTimestampMs": base.Add(time.Second).UnixMilli()},
			"update": map[string]any{
				"sessionUpdate": "user_message_chunk",
				"content":       map[string]any{"type": "text", "text": "fix main.go"},
				"_meta":         map[string]any{"modelId": "grok-4.5", "promptIndex": 0},
			},
		},
	})
	writeGrokBuildCommandJSON(t, filepath.Join(sessionDir, "hunk_records.jsonl"), map[string]any{
		"hunkId":       "agent-hunk",
		"filePath":     filepath.ToSlash(entity),
		"linesAdded":   3,
		"linesRemoved": 1,
		"authorType":   "agent",
		"eventType":    "added",
		"promptIndex":  0,
		"timestamp":    base.Add(2 * time.Second),
	})

	return filepath.ToSlash(entity), sessionID
}

func writeGrokBuildCommandJSON(t *testing.T, path string, value any) {
	t.Helper()

	data, err := json.Marshal(value)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, append(data, '\n'), 0o644))
}
