//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	_ "modernc.org/sqlite"
)

func TestCodyParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, "Library", "Application Support", "Code", "User", "globalStorage")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "state.vscdb")
	readPath := filepath.Join(home, "read.go")
	editPath := filepath.Join(home, "edit.go")

	createCodyDB(t, dbPath, map[string]any{
		"cody-local-chatHistory-v2": map[string]any{
			"account-key": map[string]any{
				"chat": map[string]any{
					"chat-1": map[string]any{
						"id":                       "chat-1",
						"lastInteractionTimestamp": "2026-05-01T20:00:00Z",
						"interactions": []map[string]any{
							{
								"humanMessage": map[string]any{
									"text": "old prompt",
								},
							},
							{
								"humanMessage": map[string]any{
									"text": "Please inspect and edit the file",
									"contextFiles": []map[string]any{
										{
											"type": "file",
											"uri": map[string]any{
												"scheme": "file",
												"path":   readPath,
											},
										},
									},
								},
								"assistantMessage": map[string]any{
									"text":  "Done",
									"model": "claude-3.5",
									"tokenUsage": map[string]any{
										"promptTokens":     10,
										"completionTokens": 4,
									},
									"processes": []map[string]any{
										{
											"items": []map[string]any{
												{
													"type":     "tool-state",
													"toolName": "text_editor",
													"uri": map[string]any{
														"scheme": "file",
														"path":   editPath,
													},
													"metadata": []string{"old\n", "old\nnew\n"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	parser := ai.Cody{
		After:             time.Date(2026, 5, 1, 19, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			readPath: heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 3)

	assert.Equal(t, "Cody chat-1", got[0].Entity)
	assert.Equal(t, "chat-1", got[0].AISession)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[0].Category)
	assert.Nil(t, got[0].AILineChanges)
	assert.Equal(t, len([]rune("Please inspect and edit the file")), got[0].AIPromptLength)
	assert.Equal(t, int64(10), got[0].AIInputTokens)
	assert.Equal(t, int64(4), got[0].AIOutputTokens)
	require.NotNil(t, got[0].IsWrite)
	assert.False(t, *got[0].IsWrite)
	assert.Equal(t, float64(time.Date(2026, 5, 1, 20, 0, 0, 0, time.UTC).Unix()), got[0].Time)
	assert.Contains(t, got[0].UserAgent, "claude/3.5")
	assert.True(t, strings.Index(got[0].UserAgent, "claude/3.5") <
		strings.Index(got[0].UserAgent, "plugin/0.0.1"))
	assert.Contains(t, got[0].UserAgent, "plugin/0.0.1")

	assert.Equal(t, readPath, got[1].Entity)
	assert.Equal(t, "chat-1", got[1].AISession)
	assert.Equal(t, heartbeat.FileType, got[1].EntityType)
	require.NotNil(t, got[1].AILineChanges)
	assert.Equal(t, 0, *got[1].AILineChanges)
	require.NotNil(t, got[1].IsWrite)
	assert.False(t, *got[1].IsWrite)
	assert.Contains(t, got[1].UserAgent, "claude/3.5")
	assert.True(t, strings.Index(got[1].UserAgent, "claude/3.5") <
		strings.Index(got[1].UserAgent, "editor/1.2.3"))
	assert.Contains(t, got[1].UserAgent, "editor/1.2.3")

	assert.Equal(t, editPath, got[2].Entity)
	assert.Equal(t, "chat-1", got[2].AISession)
	require.NotNil(t, got[2].AILineChanges)
	assert.Equal(t, 1, *got[2].AILineChanges)
	require.NotNil(t, got[2].IsWrite)
	assert.True(t, *got[2].IsWrite)
	assert.Contains(t, got[2].UserAgent, "claude/3.5")
	assert.Contains(t, got[2].UserAgent, "plugin/0.0.1")
}

func TestCodyParseCountsSameLengthEdits(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, "Library", "Application Support", "Code", "User", "globalStorage")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	editPath := filepath.Join(home, "rewrite.go")
	createCodyDB(t, filepath.Join(dbDir, "state.vscdb"), map[string]any{
		"cody-local-chatHistory-v2": map[string]any{
			"account-key": map[string]any{
				"chat": map[string]any{
					"chat-1": map[string]any{
						"id":                       "chat-1",
						"lastInteractionTimestamp": "2026-05-01T20:00:00Z",
						"interactions": []map[string]any{
							{
								"humanMessage": map[string]any{"text": "rewrite"},
								"assistantMessage": map[string]any{
									"processes": []map[string]any{
										{
											"items": []map[string]any{
												{
													"toolName": "text_editor",
													"uri": map[string]any{
														"scheme": "file",
														"path":   editPath,
													},
													"metadata": []string{"old\nsame\n", "new\nsame\n"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	got, err := (ai.Cody{}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.NotNil(t, got[1].AILineChanges)
	assert.Equal(t, 2, *got[1].AILineChanges)
}

func TestCodyParseSkipsMissingHistory(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, "Library", "Application Support", "Code", "User", "globalStorage")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	createCodyDB(t, filepath.Join(dbDir, "state.vscdb"), map[string]any{
		"other": true,
	})

	got, err := (ai.Cody{}).Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func createCodyDB(t *testing.T, dbPath string, storage map[string]any) {
	t.Helper()

	require.NoError(t, os.RemoveAll(dbPath))

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)

	defer db.Close() // nolint:errcheck

	_, err = db.Exec(`CREATE TABLE ItemTable (key TEXT UNIQUE, value BLOB)`)
	require.NoError(t, err)

	value, err := json.Marshal(storage)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO ItemTable (key, value) VALUES (?, ?)`, "sourcegraph.cody-ai", string(value))
	require.NoError(t, err)
}
