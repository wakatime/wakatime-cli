package heartbeat_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gandarez/go-realpath"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cmdheartbeat "github.com/wakatime/wakatime-cli/cmd/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"
)

func TestRunAIEditorAttribution(t *testing.T) {
	const (
		host    = "cursor/1.2.3"
		plugin  = "vscode-wakatime/30.2.1"
		agent   = "grok-build/1.0.67"
		model   = "grok/4.5"
		session = "grok-build-session"
		prompt  = "fix main.go"
	)

	for _, tc := range []struct {
		name       string
		run        func(context.Context, *viper.Viper) (int, error)
		editors    string
		aiProducts string
		wantHuman  bool
	}{
		{"default host editor", cmdheartbeat.Run, "", model + " " + host + " " + agent + " " + plugin, true},
		{"selected agent", cmdheartbeat.Run, "grok-build", model + " " + agent + " " + host + " " + plugin, true},
		{"standalone sync", cmdheartbeat.RunAISyncActivity, "codex-cli, GROK-BUILD amp",
			model + " " + agent + " " + host + " " + plugin, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resetSingleton(t)

			home, err := realpath.Realpath(t.TempDir())
			require.NoError(t, err)
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			t.Setenv("WAKATIME_HOME", home)
			t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

			base := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
			entity := writeGrokBuildAttributionSession(t, home, base, session, prompt)
			config := filepath.Join(home, ".wakatime.cfg")
			require.NoError(t, os.WriteFile(config, []byte("[settings]\nai_editors = "+tc.editors+"\n"), 0o600))

			testServerURL, router, tearDown := setupTestServer()
			defer tearDown()

			requests := make(chan []heartbeat.Heartbeat, 4)

			router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
				var received []heartbeat.Heartbeat
				if !assert.NoError(t, json.NewDecoder(req.Body).Decode(&received)) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				requests <- received

				responses := make([][]any, len(received))

				for i := range responses {
					responses[i] = []any{map[string]any{"data": map[string]string{"id": "heartbeat"}}, http.StatusCreated}
				}

				w.WriteHeader(http.StatusCreated)
				assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"responses": responses}))
			})

			v := vipertools.MustNew()
			require.NoError(t, ini.ReadInConfig(v, config))
			v.Set("api-url", testServerURL)
			v.Set("category", "debugging")
			v.Set("entity", entity)
			v.Set("entity-type", "file")
			v.Set("heartbeat-rate-limit-seconds", 0)
			v.Set("internal-config", filepath.Join(home, "wakatime-internal.cfg"))
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("offline-queue-file", filepath.Join(home, "offline.db"))
			v.Set("plugin", host+" "+plugin)
			v.Set("project", "myproject")
			v.Set("time", float64(base.Add(time.Hour).Unix()))
			v.Set("timeout", 5)
			v.Set("internal.ai_logs_last_parsed_at", base.Add(-time.Minute).Format(ini.DateFormat))

			code, err := tc.run(t.Context(), v)
			require.NoError(t, err)
			require.Zero(t, code)
			require.Len(t, requests, 1)

			type attribution struct {
				session      string
				category     string
				userAgent    string
				promptLength int
				lineChanges  *int
			}

			received := <-requests
			got := make([]attribution, 0, len(received))

			for _, h := range received {
				assert.Equal(t, entity, h.Entity)
				got = append(got, attribution{h.AISession, h.Category, h.UserAgent, h.AIPromptLength, h.AILineChanges})
			}

			aiUserAgent := heartbeat.UserAgent(t.Context(), tc.aiProducts)
			want := []attribution{
				{session, "ai coding", aiUserAgent, len(prompt), nil},
				{session, "ai coding", aiUserAgent, 0, heartbeat.PointerTo(2)},
			}

			if tc.wantHuman {
				humanUserAgent := heartbeat.UserAgent(t.Context(), host+" "+plugin)
				want = append(want, attribution{"", "debugging", humanUserAgent, 0, nil})
			}

			assert.ElementsMatch(t, want, got)
		})
	}
}

func writeGrokBuildAttributionSession(t *testing.T, home string, base time.Time, session string, prompt string) string {
	t.Helper()

	projectDir := filepath.Join(home, "project")
	require.NoError(t, os.MkdirAll(projectDir, 0o755))
	entity := filepath.Join(projectDir, "main.go")
	copyFile(t, "testdata/main.go", entity)

	grokHome := filepath.Join(home, ".grok")
	t.Setenv("GROK_HOME", grokHome)
	sessionDir := filepath.Join(grokHome, "sessions", session)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	files := map[string]any{
		filepath.Join(grokHome, "version.json"): map[string]any{"version": "1.0.67"},
		filepath.Join(sessionDir, "summary.json"): map[string]any{
			"info":             map[string]any{"id": session, "cwd": filepath.ToSlash(projectDir)},
			"current_model_id": "grok-4.5",
		},
		filepath.Join(sessionDir, "updates.jsonl"): map[string]any{
			"timestamp": base.Add(time.Second).Unix(),
			"method":    "session/update",
			"params": map[string]any{
				"sessionId": session,
				"update": map[string]any{
					"sessionUpdate": "user_message_chunk",
					"content":       map[string]any{"type": "text", "text": prompt},
				},
			},
		},
		filepath.Join(sessionDir, "hunk_records.jsonl"): map[string]any{
			"hunkId":       "agent-hunk",
			"filePath":     filepath.ToSlash(entity),
			"linesAdded":   3,
			"linesRemoved": 1,
			"authorType":   "agent",
			"eventType":    "added",
			"timestamp":    base.Add(2 * time.Second),
		},
	}

	for path, content := range files {
		data, err := json.Marshal(content)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(path, append(data, '\n'), 0o600))
	}

	return filepath.ToSlash(entity)
}
