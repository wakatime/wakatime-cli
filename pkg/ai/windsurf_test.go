//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai_test

import (
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

	_ "modernc.org/sqlite"
)

func TestWindsurfParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, "Library", "Application Support", "Windsurf", "User", "globalStorage")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "state.vscdb")
	testDir := filepath.Join(home, "windsurf-test")
	editedPath := filepath.Join(testDir, "edited.js")

	createWindsurfDB(t, dbPath, []windsurfTestRow{
		{
			Key: "bubbleId:cascade-1:old-edit",
			Value: map[string]any{
				"_v":        3,
				"type":      2,
				"createdAt": "2026-04-20T11:58:00Z",
				"toolFormerData": map[string]any{
					"status": "completed",
					"name":   "edit_file_v2",
					"params": fmt.Sprintf(
						`{"relativeWorkspacePath":%q,"streamingContent":"skip"}`,
						editedPath,
					),
				},
				"tokenCount": map[string]any{
					"inputTokens":  3,
					"outputTokens": 1,
				},
			},
		},
		{
			Key: "bubbleId:cascade-1:user",
			Value: map[string]any{
				"_v":        3,
				"type":      1,
				"text":      "Refactor this file",
				"model":     "swe-1.5",
				"createdAt": "2026-04-20T12:00:00Z",
			},
		},
		{
			Key: "bubbleId:cascade-1:edit",
			Value: map[string]any{
				"_v":        3,
				"type":      2,
				"createdAt": "2026-04-20T12:00:05Z",
				"toolFormerData": map[string]any{
					"status": "completed",
					"name":   "edit_file_v2",
					"params": fmt.Sprintf(
						`{"relativeWorkspacePath":%q,"streamingContent":"one\ntwo"}`,
						editedPath,
					),
				},
				"tokenCount": map[string]any{
					"inputTokens":  8,
					"outputTokens": 3,
				},
			},
		},
	})

	parser := ai.Windsurf{
		After:             time.Date(2026, 4, 20, 11, 59, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			editedPath: heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "Windsurf cascade-1", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Contains(t, got[0].UserAgent, "swe/1.5")
	assert.NotContains(t, got[0].UserAgent, "Cursor")
	assert.True(t, strings.Index(got[0].UserAgent, "Windsurf") < strings.Index(got[0].UserAgent, "plugin/0.0.1"))

	assert.Equal(t, editedPath, got[1].Entity)
	assert.Equal(t, heartbeat.FileType, got[1].EntityType)
	require.NotNil(t, got[1].AILineChanges)
	assert.Equal(t, 2, *got[1].AILineChanges)
	assert.EqualValues(t, 5, got[1].AIInputTokens)
	assert.EqualValues(t, 2, got[1].AIOutputTokens)
	assert.Contains(t, got[1].UserAgent, "swe/1.5")
	assert.NotContains(t, got[1].UserAgent, "Cursor")
	assert.True(t, strings.Index(got[1].UserAgent, "Windsurf") < strings.Index(got[1].UserAgent, "editor/1.2.3"))
}

func TestWindsurfParse_NoStateDB(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := ai.Windsurf{After: time.Now()}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestWindsurfParse_SkipsMalformedSQLiteRows(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, "Library", "Application Support", "Windsurf", "User", "globalStorage")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "state.vscdb")
	createWindsurfDB(t, dbPath, []windsurfTestRow{
		{
			Key: "bubbleId:cascade-1:user",
			Value: map[string]any{
				"_v":        3,
				"type":      1,
				"text":      "Refactor this file",
				"createdAt": "2026-04-20T12:00:00Z",
			},
		},
	})

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO cursorDiskKV(key, value) VALUES(?, ?)`, "bubbleId:cascade-1:bad", "{")
	require.NoError(t, err)
	require.NoError(t, db.Close())

	got, err := ai.Windsurf{
		After:             time.Date(2026, 4, 20, 11, 59, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "Windsurf cascade-1", got[0].Entity)
}

type windsurfTestRow struct {
	Key   string
	Value map[string]any
}

func createWindsurfDB(t *testing.T, dbPath string, rows []windsurfTestRow) {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close() // nolint:errcheck

	_, err = db.Exec(`CREATE TABLE cursorDiskKV (key TEXT UNIQUE ON CONFLICT REPLACE, value BLOB);`)
	require.NoError(t, err)

	for _, row := range rows {
		value, err := json.Marshal(row.Value)
		require.NoError(t, err)

		_, err = db.Exec(`INSERT INTO cursorDiskKV(key, value) VALUES(?, ?)`, row.Key, string(value))
		require.NoError(t, err)
	}
}
