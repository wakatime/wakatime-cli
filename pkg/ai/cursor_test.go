//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	_ "modernc.org/sqlite"
)

func TestCursorParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "state.vscdb")
	testDir := filepath.Join(home, "cursor-test")
	skipPath := filepath.Join(testDir, "skip.js")
	editedPath := filepath.Join(testDir, "edited.js")
	readPath := filepath.Join(testDir, "read.go")
	diffPath := filepath.Join(testDir, "diff.go")
	pendingPath := filepath.Join(testDir, "pending.js")
	legacyPath := filepath.Join(testDir, "legacy.py")
	legacyRelativePath := filepath.Join("legacy", "relative.py")
	diffContent := strings.Join([]string{
		"@@ -1,2 +1,2 @@",
		"-old",
		"+new",
		" context",
	}, "\n")
	createCursorDB(t, dbPath, []cursorTestRow{
		{
			Key: "bubbleId:composer-1:user",
			Value: map[string]any{
				"_v":        3,
				"type":      1,
				"text":      "Please update the file",
				"createdAt": "2026-03-15T23:34:10Z",
				"modelName": "composer-2.5",
			},
		},
		{
			Key: "bubbleId:composer-1:old-edit",
			Value: map[string]any{
				"_v":        3,
				"type":      2,
				"createdAt": "2026-03-15T23:33:59Z",
				"tokenCount": map[string]any{
					"inputTokens":  3,
					"outputTokens": 1,
				},
				"toolFormerData": map[string]any{
					"status": "completed",
					"name":   "edit_file_v2",
					"params": fmt.Sprintf(`{"relativeWorkspacePath":%q,"streamingContent":"skip"}`, skipPath),
				},
			},
		},
		{
			Key: "bubbleId:composer-1:edit",
			Value: map[string]any{
				"_v":        3,
				"type":      2,
				"createdAt": "2026-03-15T23:34:39Z",
				"tokenCount": map[string]any{
					"inputTokens":  8,
					"outputTokens": 5,
				},
				"toolFormerData": map[string]any{
					"status": "completed",
					"name":   "edit_file_v2",
					"params": fmt.Sprintf(
						`{"relativeWorkspacePath":%q,"streamingContent":"first\nsecond\n\nthird"}`,
						editedPath,
					),
				},
			},
		},
		{
			Key: "bubbleId:composer-1:read",
			Value: map[string]any{
				"_v":        3,
				"type":      2,
				"createdAt": "2026-03-15T23:35:15Z",
				"toolFormerData": map[string]any{
					"status": "completed",
					"name":   "read_file_v2",
					"params": fmt.Sprintf(`{"targetFile":%q,"effectiveUri":%q}`, readPath, readPath),
				},
			},
		},
		{
			Key: "bubbleId:composer-1:token-count",
			Value: map[string]any{
				"_v":        3,
				"createdAt": "2026-03-15T23:35:20Z",
				"tokenCount": map[string]any{
					"inputTokens":  12,
					"outputTokens": 8,
				},
			},
		},
		{
			Key: "bubbleId:composer-1:assistant",
			Value: map[string]any{
				"_v":        3,
				"type":      2,
				"text":      "I updated the implementation",
				"createdAt": "2026-03-15T23:35:30Z",
				"tokenCount": map[string]any{
					"inputTokens":  0,
					"outputTokens": 0,
				},
			},
		},
		{
			Key: "bubbleId:composer-1:diff",
			Value: map[string]any{
				"_v":        3,
				"type":      2,
				"createdAt": "2026-03-15T23:36:00Z",
				"toolFormerData": map[string]any{
					"status": "completed",
					"name":   "edit_file_v2",
					"params": fmt.Sprintf(
						`{"relativeWorkspacePath":%q,"streamingContent":%q}`,
						diffPath,
						diffContent,
					),
				},
			},
		},
		{
			Key: "bubbleId:composer-1:legacy",
			Value: map[string]any{
				"_v":        3,
				"type":      2,
				"createdAt": "2026-03-15T23:36:10Z",
				"toolFormerData": map[string]any{
					"status":  "completed",
					"name":    "edit_file",
					"params":  fmt.Sprintf(`{"relativeWorkspacePath":%q}`, legacyRelativePath),
					"rawArgs": fmt.Sprintf(`{"target_file":%q,"code_edit":"alpha\nbeta"}`, legacyRelativePath),
				},
				"codeBlocks": []map[string]any{
					{
						"uri": map[string]any{
							"_fsPath": legacyPath,
						},
					},
				},
			},
		},
		{
			Key: "bubbleId:composer-1:pending",
			Value: map[string]any{
				"_v":        3,
				"type":      2,
				"createdAt": "2026-03-15T23:36:20Z",
				"toolFormerData": map[string]any{
					"status": "pending",
					"name":   "edit_file_v2",
					"params": fmt.Sprintf(`{"relativeWorkspacePath":%q,"streamingContent":"pending"}`, pendingPath),
				},
			},
		},
	})

	parser := ai.Cursor{
		After:             time.Date(2026, 3, 15, 23, 34, 0, 0, time.UTC),
		FallbackUserAgent: "Cursor/1.105.1",
		UserAgents: map[string]string{
			editedPath: heartbeat.UserAgent(ctx, "Cursor/1.105.1"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 6)

	for _, h := range got {
		assert.Equal(t, "pro", h.AISubscriptionPlan)
	}

	assert.Equal(t, "Cursor composer-1", got[0].Entity)
	assert.Equal(t, "composer-1", got[0].AISession)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[0].Category)
	assert.Nil(t, got[0].AILineChanges)
	assert.Equal(t, len([]rune("Please update the file")), got[0].AIPromptLength)
	assert.Equal(t, testDir, got[0].ProjectPathOverride)
	require.NotNil(t, got[0].IsWrite)
	assert.False(t, *got[0].IsWrite)
	assert.Equal(t, float64(time.Date(2026, 3, 15, 23, 34, 10, 0, time.UTC).Unix()), got[0].Time)
	assert.Equal(t, "composer/2.5 Cursor/1.105.1", got[0].UserAgent)

	assert.Equal(t, editedPath, got[1].Entity)
	assert.Equal(t, "composer-1", got[1].AISession)
	assert.Equal(t, heartbeat.FileType, got[1].EntityType)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[1].Category)
	require.NotNil(t, got[1].AILineChanges)
	assert.Equal(t, 3, *got[1].AILineChanges)
	assert.Zero(t, got[1].AIPromptLength)
	assert.Equal(t, int64(8), got[1].AIInputTokens)
	assert.Equal(t, int64(5), got[1].AIOutputTokens)
	require.NotNil(t, got[1].IsWrite)
	assert.True(t, *got[1].IsWrite)
	assert.Equal(t, float64(time.Date(2026, 3, 15, 23, 34, 39, 0, time.UTC).Unix()), got[1].Time)
	assert.Contains(t, got[1].UserAgent, "Cursor/1.105.1")
	assert.Contains(t, got[1].UserAgent, "composer/2.5")
	assert.True(t, strings.Index(got[1].UserAgent, "composer/2.5") <
		strings.Index(got[1].UserAgent, "Cursor/1.105.1"))

	assert.Equal(t, readPath, got[2].Entity)
	assert.Equal(t, "composer-1", got[2].AISession)
	require.NotNil(t, got[2].AILineChanges)
	assert.Equal(t, 0, *got[2].AILineChanges)
	require.NotNil(t, got[2].IsWrite)
	assert.False(t, *got[2].IsWrite)
	assert.Equal(t, "composer/2.5 Cursor/1.105.1", got[2].UserAgent)

	assert.Equal(t, "Cursor composer-1", got[3].Entity)
	assert.Equal(t, "composer-1", got[3].AISession)
	assert.Equal(t, heartbeat.AppType, got[3].EntityType)
	assert.Nil(t, got[3].AILineChanges)
	assert.Zero(t, got[3].AIPromptLength)
	assert.Equal(t, filepath.Dir(diffPath), got[3].ProjectPathOverride)
	assert.Equal(t, int64(12), got[3].AIInputTokens)
	assert.Equal(t, int64(8), got[3].AIOutputTokens)
	require.NotNil(t, got[3].IsWrite)
	assert.False(t, *got[3].IsWrite)
	assert.Equal(t, float64(time.Date(2026, 3, 15, 23, 35, 30, 0, time.UTC).Unix()), got[3].Time)

	assert.Equal(t, diffPath, got[4].Entity)
	assert.Equal(t, "composer-1", got[4].AISession)
	require.NotNil(t, got[4].AILineChanges)
	assert.Equal(t, 0, *got[4].AILineChanges)
	require.NotNil(t, got[4].IsWrite)
	assert.True(t, *got[4].IsWrite)

	assert.Equal(t, legacyRelativePath, got[5].Entity)
	assert.Equal(t, "composer-1", got[5].AISession)
	require.NotNil(t, got[5].AILineChanges)
	assert.Equal(t, 2, *got[5].AILineChanges)
}

func TestCursorParse_ModernSchemaContextWindowTokensAndContentSnapshots(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "state.vscdb")
	editedPath := filepath.Join(home, "cursor-test", "edited.ts")

	zeroTokenCount := map[string]any{
		"inputTokens":  0,
		"outputTokens": 0,
	}

	createCursorDB(t, dbPath, []cursorTestRow{
		{
			Key: "bubbleId:composer-modern:user",
			Value: map[string]any{
				"_v":         3,
				"type":       1,
				"text":       "Please update the file",
				"createdAt":  "2026-03-15T23:34:10Z",
				"modelInfo":  map[string]any{"modelName": "claude-sonnet-5"},
				"tokenCount": zeroTokenCount,
				"contextWindowStatusAtCreation": map[string]any{
					"percentageRemaining": 79,
					"tokensUsed":          61702,
					"tokenLimit":          300000,
				},
			},
		},
		{
			Key: "bubbleId:composer-modern:edit",
			Value: map[string]any{
				"_v":         3,
				"type":       2,
				"createdAt":  "2026-03-15T23:34:40Z",
				"tokenCount": zeroTokenCount,
				"toolFormerData": map[string]any{
					"status": "completed",
					"name":   "edit_file_v2",
					"params": fmt.Sprintf(
						`{"relativeWorkspacePath":%q,"noCodeblock":true,"cloudAgentEdit":false}`,
						editedPath,
					),
					"result": `{"beforeContentId":"composer.content.before-hash",` +
						`"afterContentId":"composer.content.after-hash"}`,
				},
			},
		},
		{
			Key: "bubbleId:composer-modern:assistant",
			Value: map[string]any{
				"_v":         3,
				"type":       2,
				"text":       "I updated the implementation",
				"createdAt":  "2026-03-15T23:35:30Z",
				"tokenCount": zeroTokenCount,
			},
		},
		{
			Key: "bubbleId:composer-modern:user-2",
			Value: map[string]any{
				"_v":         3,
				"type":       1,
				"text":       "Now add a test",
				"createdAt":  "2026-03-15T23:36:00Z",
				"tokenCount": zeroTokenCount,
				"contextWindowStatusAtCreation": map[string]any{
					"percentageRemaining": 77,
					"tokensUsed":          68073,
					"tokenLimit":          300000,
				},
			},
		},
		{
			Key: "composer.content.before-hash",
			Raw: "one\ntwo\nthree",
		},
		{
			Key: "composer.content.after-hash",
			Raw: "one\ntwo changed\nthree\nfour",
		},
	})

	parser := ai.Cursor{
		After:             time.Date(2026, 3, 15, 23, 34, 0, 0, time.UTC),
		FallbackUserAgent: "Cursor/1.105.1",
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 4)

	// Context-window snapshots are estimates and are not published as exact
	// token usage.
	assert.Equal(t, "Cursor composer-modern", got[0].Entity)
	assert.Equal(t, int64(0), got[0].AIInputTokens)
	assert.Equal(t, int64(0), got[0].AIOutputTokens)
	assert.Contains(t, got[0].UserAgent, "sonnet/5")

	// edit heartbeat computes line changes from before/after content snapshots
	assert.Equal(t, editedPath, got[1].Entity)
	require.NotNil(t, got[1].AILineChanges)
	assert.Equal(t, 1, *got[1].AILineChanges)
	require.NotNil(t, got[1].IsWrite)
	assert.True(t, *got[1].IsWrite)
	assert.Equal(t, int64(0), got[1].AIInputTokens)

	// zero tokenCount placeholder on the assistant bubble adds no tokens
	assert.Equal(t, "Cursor composer-modern", got[2].Entity)
	assert.Equal(t, int64(0), got[2].AIInputTokens)

	// Later context snapshots remain unpublished estimates.
	assert.Equal(t, "Cursor composer-modern", got[3].Entity)
	assert.Equal(t, int64(0), got[3].AIInputTokens)
}

func TestCursorParse_RealShapeFixtures(t *testing.T) {
	for _, name := range []string{
		"legacy-token-count.json",
		"zero-token-placeholder.json",
		"modern-context-snapshots.json",
	} {
		t.Run(name, func(t *testing.T) {
			fixture := loadCursorFixture(t, name)

			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)

			dbDir := filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage")
			require.NoError(t, os.MkdirAll(dbDir, 0o755))
			createCursorDB(t, filepath.Join(dbDir, "state.vscdb"), fixture.Rows)

			after, err := time.Parse(time.RFC3339Nano, fixture.After)
			require.NoError(t, err)

			got, err := (ai.Cursor{
				After:             after,
				FallbackUserAgent: "cursor/fixture",
			}).Parse(t.Context())
			require.NoError(t, err)
			require.Len(t, got, len(fixture.Expected.Entities))

			for i := range got {
				assert.Equal(t, fixture.Expected.Entities[i], got[i].Entity)
				assert.Equal(t, fixture.Expected.InputTokens[i], got[i].AIInputTokens)
				assert.Equal(t, fixture.Expected.OutputTokens[i], got[i].AIOutputTokens)
				assert.Equal(t, fixture.Expected.LineChanges[i], got[i].AILineChanges)
			}
		})
	}
}

func TestCursorParse_UnicodePathsAndText(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "state.vscdb")
	editedPath := filepath.Join(home, "项目", "文件.go")
	prompt := "请更新这个文件"

	createCursorDB(t, dbPath, []cursorTestRow{
		{
			Key: "bubbleId:composer-zh:old-read",
			Value: map[string]any{
				"_v":        3,
				"type":      2,
				"createdAt": "2026-03-15T23:33:59Z",
				"toolFormerData": map[string]any{
					"status": "completed",
					"name":   "read_file_v2",
					"params": fmt.Sprintf(`{"targetFile":%q}`, editedPath),
				},
			},
		},
		{
			Key: "bubbleId:composer-zh:user",
			Value: map[string]any{
				"_v":        3,
				"type":      1,
				"text":      prompt,
				"createdAt": "2026-03-15T23:34:10Z",
			},
		},
		{
			Key: "bubbleId:composer-zh:edit",
			Value: map[string]any{
				"_v":        3,
				"type":      2,
				"createdAt": "2026-03-15T23:34:39Z",
				"toolFormerData": map[string]any{
					"status": "completed",
					"name":   "edit_file_v2",
					"params": fmt.Sprintf(
						`{"relativeWorkspacePath":%q,"streamingContent":"第一行\n第二行"}`,
						editedPath,
					),
				},
			},
		},
	})

	got, err := ai.Cursor{
		After:             time.Date(2026, 3, 15, 23, 34, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "Cursor composer-zh", got[0].Entity)
	assert.Equal(t, len([]rune(prompt)), got[0].AIPromptLength)
	assert.Equal(t, filepath.Dir(editedPath), got[0].ProjectPathOverride)

	assert.Equal(t, editedPath, got[1].Entity)
	require.NotNil(t, got[1].AILineChanges)
	assert.Equal(t, 2, *got[1].AILineChanges)
}

func TestCursorParse_SkipsMalformedSQLiteRows(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "state.vscdb")
	createCursorDB(t, dbPath, []cursorTestRow{
		{
			Key: "bubbleId:composer-1:user",
			Value: map[string]any{
				"_v":        3,
				"type":      1,
				"text":      "Please update the file",
				"createdAt": "2026-03-15T23:34:10Z",
			},
		},
	})

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO cursorDiskKV(key, value) VALUES(?, ?)`, "bubbleId:composer-1:bad", "{")
	require.NoError(t, err)
	require.NoError(t, db.Close())

	got, err := ai.Cursor{
		After:             time.Date(2026, 3, 15, 23, 34, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "Cursor composer-1", got[0].Entity)
}

func TestCursorParse_NoCursorStateDB(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := ai.Cursor{After: time.Now()}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestCursorParse_SubscriptionPlanErrorLogsAndDoesNotBlockHeartbeats(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "state.vscdb")

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE cursorDiskKV (key TEXT UNIQUE ON CONFLICT REPLACE, value BLOB);`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE TABLE ItemTable (unexpected TEXT);`)
	require.NoError(t, err)

	value, err := json.Marshal(map[string]any{
		"_v":        3,
		"type":      1,
		"text":      "Please update the file",
		"createdAt": "2026-03-15T23:34:10Z",
	})
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO cursorDiskKV(key, value) VALUES(?, ?)`, "bubbleId:composer-1:user", string(value))
	require.NoError(t, err)
	require.NoError(t, db.Close())

	var logs bytes.Buffer

	ctx := log.ToContext(context.Background(), log.New(&logs, log.WithVerbose(true)))

	got, err := ai.Cursor{
		After:             time.Date(2026, 3, 15, 23, 34, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Empty(t, got[0].AISubscriptionPlan)
	assert.Contains(t, logs.String(), `"level":"debug"`)
	assert.Contains(t, logs.String(), "failed reading cursor subscription plan")
}

func TestCursorParse_SkipsStaleCursorStateDB(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "state.vscdb")
	staleEditedPath := filepath.Join(home, "cursor-test", "edited.js")
	createCursorDB(t, dbPath, []cursorTestRow{
		{
			Key: "bubbleId:composer-1:edit",
			Value: map[string]any{
				"_v":        3,
				"type":      2,
				"createdAt": "2026-03-15T23:34:39Z",
				"toolFormerData": map[string]any{
					"status": "completed",
					"name":   "edit_file_v2",
					"params": fmt.Sprintf(
						`{"relativeWorkspacePath":%q,"streamingContent":"first"}`,
						staleEditedPath,
					),
				},
			},
		},
	})

	staleTime := time.Date(2026, 3, 15, 23, 33, 0, 0, time.UTC)
	require.NoError(t, os.Chtimes(dbPath, staleTime, staleTime))

	got, err := ai.Cursor{
		After: staleTime.Add(time.Second),
	}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestCursorParse_BoundsRecentRows(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	rows := []cursorTestRow{{
		Key: "bubbleId:outside-window:user",
		Value: map[string]any{
			"_v":        3,
			"type":      1,
			"text":      "this row must be outside the bounded scan",
			"createdAt": "2026-07-10T14:00:00Z",
		},
	}}

	for i := range 5000 {
		rows = append(rows, cursorTestRow{
			Key: fmt.Sprintf("bubbleId:noise-%04d:row", i),
			Value: map[string]any{
				"_v":        3,
				"createdAt": time.Date(2026, 7, 10, 14, 1, 0, i, time.UTC).Format(time.RFC3339Nano),
			},
		})
	}

	createCursorDB(t, filepath.Join(dbDir, "state.vscdb"), rows)

	got, err := (ai.Cursor{
		After: time.Date(2026, 7, 10, 13, 59, 0, 0, time.UTC),
	}).Parse(t.Context())
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestCursorParse_LoadsOnlyPostCheckpointSnapshotsAndLogsSafeStats(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	privatePath := filepath.Join(home, "private-project", "main.go")
	createCursorDB(t, filepath.Join(dbDir, "state.vscdb"), []cursorTestRow{
		{
			Key: "bubbleId:historical:edit",
			Value: map[string]any{
				"_v":        3,
				"type":      2,
				"createdAt": "2026-07-10T14:00:00Z",
				"toolFormerData": map[string]any{
					"status": "completed",
					"name":   "edit_file_v2",
					"params": fmt.Sprintf(
						`{"relativeWorkspacePath":%q,"noCodeblock":true}`,
						privatePath,
					),
					"result": `{"beforeContentId":"composer.content.old-before",` +
						`"afterContentId":"composer.content.old-after"}`,
				},
			},
		},
		{
			Key: "bubbleId:current:user",
			Value: map[string]any{
				"_v":        3,
				"type":      1,
				"text":      "sanitized current prompt",
				"createdAt": "2026-07-10T14:02:00Z",
			},
		},
		{Key: "composer.content.old-before", Raw: "one"},
		{Key: "composer.content.old-after", Raw: "one\ntwo"},
	})

	var logs bytes.Buffer

	ctx := log.ToContext(t.Context(), log.New(&logs, log.WithVerbose(true)))

	got, err := (ai.Cursor{
		After: time.Date(2026, 7, 10, 14, 1, 0, 0, time.UTC),
	}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Contains(t, logs.String(), "cursor parser stats")
	assert.Contains(t, logs.String(), "content_ids=0")
	assert.NotContains(t, logs.String(), privatePath)
	assert.NotContains(t, logs.String(), "sanitized current prompt")
}

func TestCursorParse_MissingSnapshotIsUnknownNotZero(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	editPath := filepath.Join(home, "project", "main.go")
	createCursorDB(t, filepath.Join(dbDir, "state.vscdb"), []cursorTestRow{{
		Key: "bubbleId:missing-snapshot:edit",
		Value: map[string]any{
			"_v":        3,
			"type":      2,
			"createdAt": "2026-07-10T15:00:00Z",
			"toolFormerData": map[string]any{
				"status": "completed",
				"name":   "edit_file_v2",
				"params": fmt.Sprintf(
					`{"relativeWorkspacePath":%q,"noCodeblock":true}`,
					editPath,
				),
				"result": `{"beforeContentId":"composer.content.before",` +
					`"afterContentId":"composer.content.missing"}`,
			},
		},
	}})

	got, err := (ai.Cursor{
		After: time.Date(2026, 7, 10, 14, 59, 0, 0, time.UTC),
	}).Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, editPath, got[0].Entity)
	assert.Nil(t, got[0].AILineChanges)
}

type cursorTestRow struct {
	Key   string
	Value map[string]any
	Raw   string
}

type cursorFixtureExpected struct {
	Entities     []string `json:"entities"`
	InputTokens  []int64  `json:"inputTokens"`
	LineChanges  []*int   `json:"lineChanges"`
	OutputTokens []int64  `json:"outputTokens"`
	TokenSource  string   `json:"tokenSource"`
}

type cursorSchemaFixture struct {
	After    string                `json:"after"`
	Expected cursorFixtureExpected `json:"expected"`
	Rows     []struct {
		Key   string          `json:"key"`
		Raw   *string         `json:"raw"`
		Value json.RawMessage `json:"value"`
	} `json:"rows"`
	Schema string `json:"schema"`
}

type loadedCursorFixture struct {
	After    string
	Expected cursorFixtureExpected
	Rows     []cursorTestRow
}

func loadCursorFixture(t *testing.T, name string) loadedCursorFixture {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", "cursor", name))
	require.NoError(t, err)

	var fixture cursorSchemaFixture
	require.NoError(t, json.Unmarshal(data, &fixture))
	require.NotEmpty(t, fixture.Schema)
	require.NotEmpty(t, fixture.Expected.TokenSource)

	rows := make([]cursorTestRow, 0, len(fixture.Rows))
	for _, row := range fixture.Rows {
		raw := string(row.Value)
		if row.Raw != nil {
			raw = *row.Raw
		}

		rows = append(rows, cursorTestRow{Key: row.Key, Raw: raw})
	}

	return loadedCursorFixture{
		After:    fixture.After,
		Expected: fixture.Expected,
		Rows:     rows,
	}
}

func createCursorDB(t *testing.T, dbPath string, rows []cursorTestRow) {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close() // nolint:errcheck

	_, err = db.Exec(`CREATE TABLE cursorDiskKV (key TEXT UNIQUE ON CONFLICT REPLACE, value BLOB);`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE TABLE ItemTable (key TEXT UNIQUE ON CONFLICT REPLACE, value BLOB);`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO ItemTable(key, value) VALUES('cursorAuth/stripeMembershipType', 'pro')`)
	require.NoError(t, err)

	for _, row := range rows {
		raw := row.Raw
		if raw == "" {
			value, err := json.Marshal(row.Value)
			require.NoError(t, err)

			raw = string(value)
		}

		_, err = db.Exec(`INSERT INTO cursorDiskKV(key, value) VALUES(?, ?)`, row.Key, raw)
		require.NoError(t, err)
	}
}
