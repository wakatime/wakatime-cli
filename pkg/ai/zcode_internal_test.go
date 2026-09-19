//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZCodeParseBeyondSQLiteRowLimit(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("WAKATIME_HOME", t.TempDir())

	dbPath := filepath.Join(home, ".zcode", "cli", "db", "db.sqlite")
	require.NoError(t, os.MkdirAll(filepath.Dir(dbPath), 0o755))

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	_, err = db.Exec(`CREATE TABLE tool_usage (
		id INTEGER PRIMARY KEY, payload TEXT
	) WITHOUT ROWID`)
	require.NoError(t, err)

	// Timestamps can be nested in JSON, and tables need not have a rowid.
	// Recent activity on either side of a large historical block must survive.
	_, err = db.Exec(`WITH RECURSIVE ids(id) AS (
		SELECT 1 UNION ALL SELECT id + 1 FROM ids WHERE id < 10002
	) INSERT INTO tool_usage SELECT id, json_object(
		'timestamp', CASE WHEN id IN (1, 10002) THEN '2026-09-06T12:00:00Z' ELSE '2026-08-14T12:00:00Z' END,
		'session_id', 'session-1', 'tool_name', 'edit_file', 'file_path', ?,
		'lines_added', 3, 'lines_removed', 2,
		'input_tokens', 10, 'output_tokens', 4
	) FROM ids`, filepath.Join(home, "project", "main.go"))
	require.NoError(t, err)

	provider := genericAIProvider{parser: ZCode{}, config: ParserConfig{After: time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)}}
	got := collectZCodeSQLite(t, provider, dbPath)
	require.Len(t, got, 4)

	for _, index := range []int{0, 2} {
		assert.Equal(t, "ZCode session-1", got[index].Entity)
		assert.Equal(t, int64(10), got[index].AIInputTokens)
		assert.Equal(t, int64(4), got[index].AIOutputTokens)
		assert.Equal(t, filepath.ToSlash(filepath.Join(home, "project", "main.go")), got[index+1].Entity)
		require.NotNil(t, got[index+1].AILineChanges)
		assert.Equal(t, 5, *got[index+1].AILineChanges)
	}

	// A busy day with more than 5,000 recent records must not be truncated either.
	_, err = db.Exec(`UPDATE tool_usage SET payload = json_set(payload, '$.timestamp', '2026-09-06T12:00:00Z')`)
	require.NoError(t, err)

	// The two unchanged recent rows were already reported during the first scan.
	got = append(got, collectZCodeSQLite(t, provider, dbPath)...)
	assert.Equal(t, 20004, len(got))
}

// Exercise the production time budget without assuming that the entire history
// can be scanned in one invocation on slower or race-instrumented CI runners.
func collectZCodeSQLite(t *testing.T, provider genericAIProvider, path string) Heartbeats {
	t.Helper()

	var heartbeats Heartbeats

	cursorPath, err := genericAISQLiteCursorPath(t.Context(), provider, path, "tool_usage")
	require.NoError(t, err)

	for range 100 {
		ctx, cancel := aiSQLiteContext(t.Context())
		got, parseErr := parseGenericAISQLiteDB(ctx, provider, path)
		deadlineErr := ctx.Err()

		cancel()

		if parseErr != nil {
			require.ErrorIs(t, parseErr, deadlineErr)
		}

		heartbeats = append(heartbeats, got...)
		cursor, err := readGenericAISQLiteCursor(cursorPath)
		require.NoError(t, err)

		if !cursor.Pending {
			return heartbeats
		}
		// Reproduce the global timestamp advancing while a scan is incomplete.
		provider.config.After = time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	}

	t.Fatal("SQLite scan did not finish")

	return nil
}

func TestZCodeParseActualSQLiteSchema(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("WAKATIME_HOME", t.TempDir())

	path := filepath.Join(home, ".zcode", "cli", "db", "db.sqlite")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	db, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	schema, err := os.ReadFile("testdata/zcode_schema.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(schema))
	require.NoError(t, err)
	// Keep commits in the WAL, as in a running ZCode session.
	_, err = db.Exec(`PRAGMA journal_mode=WAL; PRAGMA wal_autocheckpoint=0; PRAGMA foreign_keys=ON;`)
	require.NoError(t, err)

	started := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	project := filepath.Join(home, "project")
	file := filepath.Join(project, "main.go")
	_, err = db.Exec(`INSERT INTO session
 (id, project_id, slug, directory, title, version, time_created, time_updated)
 VALUES ('s', 'project', 'demo', ?, 'Synthetic edit', 'test', ?, ?)`, project, started.UnixMilli(), started.UnixMilli())
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data)
 VALUES ('m', 's', ?, ?, '{"role":"assistant","modelId":"test-model"}')`, started.UnixMilli(), started.UnixMilli())
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO model_usage
 (id, logical_request_id, session_id, query_source, provider_id, model_id, status, started_at)
 VALUES ('usage', 'request', 's', 'interactive', 'test-provider', 'test-model', 'running', ?)`, started.UnixMilli())
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO tool_usage (id, session_id, tool_call_id, tool_name, status, started_at)
 VALUES ('tool', 's', 'call', 'Edit', 'running', ?)`, started.UnixMilli())
	require.NoError(t, err)

	parser := ZCode{After: started.Add(-time.Minute)}
	_, err = parser.Parse(t.Context())
	require.NoError(t, err)

	// These updates preserve rowids. model_usage and tool_usage have no updated_at
	// column, so neither can safely be treated as an append-only table.
	completed := started.Add(time.Minute)
	_, err = db.Exec(`UPDATE model_usage SET status='completed', completed_at=?,
 input_tokens=100, output_tokens=25, cache_read_input_tokens=40 WHERE id='usage'`, completed.UnixMilli())
	require.NoError(t, err)
	_, err = db.Exec(`UPDATE tool_usage SET status='completed', completed_at=? WHERE id='tool'`, completed.UnixMilli())
	require.NoError(t, err)

	payload := map[string]any{
		"type": "tool", "tool": "Edit", "callID": "call",
		"state": map[string]any{
			"status": "completed",
			"input":  map[string]any{"file_path": file, "old_string": "old", "new_string": "new"},
			"metadata": map[string]any{"display": map[string]any{
				"kind": "diff", "filePath": file, "additions": 1, "deletions": 1,
			}},
			"time": map[string]any{"start": started.UnixMilli(), "end": completed.UnixMilli()},
		},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO part (id, message_id, session_id, time_created, time_updated, data)
 VALUES ('p', 'm', 's', ?, ?, ?)`, started.UnixMilli(), completed.UnixMilli(), string(data))
	require.NoError(t, err)

	parser.After = started.Add(30 * time.Second)
	got, err := parser.Parse(t.Context())
	require.NoError(t, err)

	var (
		input, output, cached int64
		fileEdits             int
	)

	for _, hb := range got {
		input += hb.AIInputTokens
		output += hb.AIOutputTokens

		cached += hb.AICachedInputTokens
		if hb.Entity == filepath.ToSlash(file) {
			fileEdits++

			require.NotNil(t, hb.AILineChanges)
			assert.Equal(t, 2, *hb.AILineChanges)
			assert.Equal(t, float64(completed.Unix()), hb.Time)
			assert.Equal(t, "s", hb.AISession)
		}
	}

	assert.Equal(t, int64(100), input)
	assert.Equal(t, int64(25), output)
	assert.Equal(t, int64(40), cached)
	assert.Equal(t, 1, fileEdits)
	got, err = parser.Parse(t.Context())
	require.NoError(t, err)
	assert.Empty(t, got)
}
