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

func TestCursorParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "state.vscdb")
	diffContent := strings.Join([]string{
		"@@ -1,2 +1,2 @@",
		"-old",
		"+new",
		" context",
	}, "\n")
	createCursorDB(t, dbPath, []cursorTestRow{
		{
			Key: "bubbleId:composer-1:old-edit",
			Value: map[string]any{
				"_v":        3,
				"type":      2,
				"createdAt": "2026-03-15T23:33:59Z",
				"toolFormerData": map[string]any{
					"status": "completed",
					"name":   "edit_file_v2",
					"params": `{"relativeWorkspacePath":"/tmp/skip.js","streamingContent":"skip"}`,
				},
			},
		},
		{
			Key: "bubbleId:composer-1:edit",
			Value: map[string]any{
				"_v":        3,
				"type":      2,
				"createdAt": "2026-03-15T23:34:39Z",
				"toolFormerData": map[string]any{
					"status": "completed",
					"name":   "edit_file_v2",
					"params": `{"relativeWorkspacePath":"/tmp/edited.js","streamingContent":"first\nsecond\n\nthird"}`,
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
					"params": `{"targetFile":"/tmp/read.go","effectiveUri":"/tmp/read.go"}`,
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
						`{"relativeWorkspacePath":"/tmp/diff.go","streamingContent":%q}`,
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
					"params":  `{"relativeWorkspacePath":"legacy/relative.py"}`,
					"rawArgs": `{"target_file":"legacy/relative.py","code_edit":"alpha\nbeta"}`,
				},
				"codeBlocks": []map[string]any{
					{
						"uri": map[string]any{
							"_fsPath": "/tmp/legacy.py",
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
					"params": `{"relativeWorkspacePath":"/tmp/pending.js","streamingContent":"pending"}`,
				},
			},
		},
	})

	parser := ai.Cursor{
		After:             time.Date(2026, 3, 15, 23, 34, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			"/tmp/edited.js": heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 4)

	assert.Equal(t, "/tmp/edited.js", got[0].Entity)
	assert.Equal(t, heartbeat.FileType, got[0].EntityType)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[0].Category)
	require.NotNil(t, got[0].AILineChanges)
	assert.Equal(t, 3, *got[0].AILineChanges)
	require.NotNil(t, got[0].IsWrite)
	assert.True(t, *got[0].IsWrite)
	assert.Equal(t, float64(time.Date(2026, 3, 15, 23, 34, 39, 0, time.UTC).Unix()), got[0].Time)
	assert.Contains(t, got[0].UserAgent, heartbeat.UserAgent(ctx, "editor/1.2.3"))
	assert.Contains(t, got[0].UserAgent, "Cursor")

	assert.Equal(t, "/tmp/read.go", got[1].Entity)
	require.NotNil(t, got[1].AILineChanges)
	assert.Equal(t, 0, *got[1].AILineChanges)
	require.NotNil(t, got[1].IsWrite)
	assert.False(t, *got[1].IsWrite)
	assert.Contains(t, got[1].UserAgent, "plugin/0.0.1")
	assert.Contains(t, got[1].UserAgent, "Cursor")

	assert.Equal(t, "/tmp/diff.go", got[2].Entity)
	require.NotNil(t, got[2].AILineChanges)
	assert.Equal(t, 0, *got[2].AILineChanges)
	require.NotNil(t, got[2].IsWrite)
	assert.True(t, *got[2].IsWrite)

	assert.Equal(t, "legacy/relative.py", got[3].Entity)
	require.NotNil(t, got[3].AILineChanges)
	assert.Equal(t, 2, *got[3].AILineChanges)
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

type cursorTestRow struct {
	Key   string
	Value map[string]any
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

	for _, row := range rows {
		value, err := json.Marshal(row.Value)
		require.NoError(t, err)

		_, err = db.Exec(`INSERT INTO cursorDiskKV(key, value) VALUES(?, ?)`, row.Key, string(value))
		require.NoError(t, err)
	}
}
