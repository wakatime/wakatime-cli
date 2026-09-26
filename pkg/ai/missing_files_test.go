package ai_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/filter"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/regex"
)

func TestWithAISync_DeletedFiles(t *testing.T) {
	for _, provider := range []string{"Codex", "Claude"} {
		for _, operation := range []string{"delete", "create", "update"} {
			for _, missing := range []bool{false, true} {
				name := provider + "/" + operation + "/existing"
				if missing {
					name = provider + "/" + operation + "/missing"
				}

				t.Run(name, func(t *testing.T) {
					home := t.TempDir()
					t.Setenv("HOME", home)
					t.Setenv("USERPROFILE", home)
					t.Setenv("CODEX_HOME", "")

					entity := filepath.Join(home, "main.go")
					require.NoError(t, os.WriteFile(entity, []byte("package main\n"), 0o600))

					if missing {
						require.NoError(t, os.Remove(entity))
					}

					entity = filepath.ToSlash(entity)
					encodedEntity, err := json.Marshal(entity)
					require.NoError(t, err)

					dir := filepath.Join(home, ".claude", "projects", "sample")
					transcript := `{"timestamp":"2026-06-20T12:00:00Z","sessionId":"session","type":"user",` +
						`"message":{"role":"user","content":"Update the file"}}` + "\n" +
						`{"timestamp":"2026-06-20T12:02:00Z","sessionId":"session","toolUseResult":{"filePath":ENTITY,` +
						`"type":"` + operation + `","structuredPatch":[{"oldLines":3,"newLines":1}]}}`

					if provider == "Codex" {
						dir = filepath.Join(home, ".codex", "sessions", "2026", "06", "20")
						patchType := map[string]string{"delete": "Delete", "create": "Add", "update": "Update"}[operation]
						patch, err := json.Marshal(
							"*** Begin Patch\n*** " + patchType + " File: " + entity + "\n*** End Patch",
						)
						require.NoError(t, err)

						transcript = `{"timestamp":"2026-06-20T12:00:00Z","type":"session_meta",` +
							`"payload":{"id":"session"}}` + "\n" +
							`{"timestamp":"2026-06-20T12:00:00Z","type":"event_msg",` +
							`"payload":{"type":"user_message","message":"Delete the file"}}` + "\n" +
							`{"timestamp":"2026-06-20T12:02:00Z","type":"response_item",` +
							`"payload":{"type":"custom_tool_call","name":"apply_patch","input":` + string(patch) + `}}`
					}

					transcript = strings.ReplaceAll(transcript, "ENTITY", string(encodedEntity))

					require.NoError(t, os.MkdirAll(dir, 0o755))
					require.NoError(t, os.WriteFile(filepath.Join(dir, "session.jsonl"), []byte(transcript+"\n"), 0o600))

					v := viper.New()
					v.Set("internal-config", filepath.Join(home, "internal.cfg"))
					v.Set("internal.ai_logs_last_parsed_at", "2026-06-20T11:00:00Z")

					var got []heartbeat.Heartbeat

					handle := ai.WithAISync(ai.Config{V: v})(func(
						_ context.Context, hh []heartbeat.Heartbeat,
					) ([]heartbeat.Result, error) {
						got = hh
						return nil, nil
					})
					_, err = handle(t.Context(), nil)
					require.NoError(t, err)
					require.Len(t, got, 2, "file activity and the session prompt must both survive")

					for _, h := range got {
						assert.Equal(t, entity, h.Entity)
						assert.Equal(t, heartbeat.FileType, h.EntityType)
						assert.Equal(t, operation == "delete", h.IsUnsavedEntity)

						if missing && operation != "delete" {
							assert.Error(t, filter.Filter(t.Context(), h, filter.Config{}))
						} else {
							assert.NoError(t, filter.Filter(t.Context(), h, filter.Config{}))
						}

						assert.Error(t, filter.Filter(t.Context(), h, filter.Config{
							Exclude: []regex.Regex{regex.MustCompile(regexp.QuoteMeta(entity))},
						}))
					}
				})
			}
		}
	}
}
