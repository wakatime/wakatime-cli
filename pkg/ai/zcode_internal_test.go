//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
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
		SELECT 1 UNION ALL SELECT id + 1 FROM ids WHERE id < 5002
	) INSERT INTO tool_usage SELECT id, json_object(
		'timestamp', CASE WHEN id IN (1, 5002) THEN '2026-09-06T12:00:00Z' ELSE '2026-08-14T12:00:00Z' END,
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
	assert.Equal(t, 10004, len(got))
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
	_, err = db.Exec(`UPDATE message SET time_updated=?, data=? WHERE id='m'`, completed.UnixMilli(),
		`{"role":"assistant","modelID":"test-model","tokens":{
		"input":100,"output":25,"reasoning":5,"cache":{"read":40,"write":10}}}`)
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

	assert.Equal(t, int64(60), input)
	assert.Equal(t, int64(25), output)
	assert.Equal(t, int64(40), cached)
	assert.Equal(t, 1, fileEdits)
	got, err = parser.Parse(t.Context())
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestZCodeParseRequestUsageAndPrompts(t *testing.T) {
	db, _, started := zCodeFixture(t)
	insertMessage := func(id, data string) {
		t.Helper()

		_, err := db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data)
 VALUES (?, 's', ?, ?, ?)`, id, started.UnixMilli(), started.UnixMilli(), data)
		require.NoError(t, err)
	}
	insertMessage("user", `{"role":"user"}`)
	insertMessage("first", `{"role":"assistant","modelID":"GLM-5.3","tokens":{
 "input":100,"output":25,"reasoning":5,"cache":{"read":40,"write":10}}}`)
	// A smaller second request is not a reset of a session-wide counter.
	insertMessage("second", `{"role":"assistant","tokens":{
 "input":10,"output":4,"reasoning":2,"cache":{"read":2,"write":1}}}`)
	// Guard against negative regular input and negative provider counters.
	insertMessage("invalid-tokens", `{"role":"assistant","tokens":{
 "input":2,"output":-1,"reasoning":-1,"cache":{"read":8}}}`)
	insertMessage("malformed", `{broken`)

	for _, table := range []string{"model_usage", "turn_usage"} {
		columns, values := "", ""
		if table == "model_usage" {
			columns = "id, logical_request_id, query_source, provider_id, model_id,"
			values = "'usage', 'request', 'interactive', 'provider', 'model',"
		}

		_, err := db.Exec("INSERT INTO " + table + " (" + columns + `
 session_id, turn_id, status, started_at, completed_at, input_tokens, output_tokens,
 reasoning_tokens, cache_read_input_tokens, cache_creation_input_tokens)
 VALUES (` + values + ` 's', 'turn', 'completed', 1, 1, 100, 25, 5, 40, 10)`)
		require.NoError(t, err)
	}

	parts := []struct {
		message, data string
	}{
		{"user", `{"type":"text","text":"Hi 世界"}`},
		{"user", `{"type":"text","text":" 👋 "}`},
		{"user", `{"type":"text","text":"ignored","ignored":true}`},
		{"user", `{"type":"text","text":"synthetic","synthetic":true}`},
		{"user", `{"type":"text","text":"   "}`},
		{"first", `{"type":"text","text":"assistant response"}`},
		{"first", `{"type":"step-finish","tokens":{"input":100,"output":25,"cache":{"read":40}}}`},
		{"missing", `{"type":"text","text":"orphan"}`},
		{"user", `{broken`},
	}
	for i, part := range parts {
		_, err := db.Exec(`INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
 VALUES (?, 's', ?, ?, ?, ?)`, fmt.Sprint(i), part.message, started.UnixMilli(), started.UnixMilli(), part.data)
		require.NoError(t, err)
	}

	parser := ZCode{After: started.Add(-time.Minute)}
	got, err := parser.Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 5)
	assertZCodeTotals(t, got, 68, 50, 29, 6)

	for _, hb := range got {
		assert.Equal(t, "s", hb.AISession)
		assert.Equal(t, "ZCode s", hb.Entity)
	}

	assert.Contains(t, got[0].UserAgent, "GLM/5.3")

	// Repeated scans and metadata-only updates must not replay usage or prompts.
	got, err = parser.Parse(t.Context())
	require.NoError(t, err)
	assert.Empty(t, got)

	updated := started.Add(time.Minute)
	_, err = db.Exec(`UPDATE message SET time_updated=?; UPDATE part SET time_updated=?`,
		updated.UnixMilli(), updated.UnixMilli())
	require.NoError(t, err)

	parser.After = started.Add(30 * time.Second)
	got, err = parser.Parse(t.Context())
	require.NoError(t, err)
	assert.Empty(t, got)

	// Only increases on the same request are new usage. Changing one request
	// must not reset another request's baseline or recount its prompt.
	_, err = db.Exec(`UPDATE message SET time_updated=?, data=? WHERE id='second'`,
		updated.Add(time.Second).UnixMilli(), `{"role":"assistant","tokens":{
 "input":30,"output":9,"reasoning":3,"cache":{"read":8,"write":1}}}`)
	require.NoError(t, err)
	got, err = parser.Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assertZCodeTotals(t, got, 14, 6, 5, 0)
}

func TestZCodeParseResumesRequestUsageAndPrompts(t *testing.T) {
	db, path, started := zCodeFixture(t)
	// Cross a batch boundary; the separate row-limit regression covers large histories.
	const rows = aiSQLiteBatchSize + 1

	_, err := db.Exec(`WITH RECURSIVE ids(id) AS (
 SELECT 1 UNION ALL SELECT id+1 FROM ids WHERE id<?
 ) INSERT INTO message (id, session_id, time_created, time_updated, data)
 SELECT printf('message-%05d', id), 's', ?, ?,
 '{"role":"assistant","tokens":{"input":10,"output":4,"reasoning":2,"cache":{"read":2}}}' FROM ids`,
		rows, started.UnixMilli(), started.UnixMilli())
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data)
 VALUES ('user', 's', ?, ?, '{"role":"user"}')`, started.UnixMilli(), started.UnixMilli())
	require.NoError(t, err)
	_, err = db.Exec(`WITH RECURSIVE ids(id) AS (
 SELECT 1 UNION ALL SELECT id+1 FROM ids WHERE id<?
 ) INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
 SELECT printf('part-%05d', id), 's', 'user', ?, ?, '{"type":"text","text":"Hi 世界"}' FROM ids`,
		rows, started.UnixMilli(), started.UnixMilli())
	require.NoError(t, err)

	provider := (ZCode{After: started.Add(-time.Minute)}).sqliteProvider(path)
	// Force interruption inside a batch, after some complete rows. The real
	// byte-budget path must checkpoint exactly the usage already returned.
	ctx := aiSQLiteBudgetContext(t.Context())
	ctx.Value(aiSQLiteBudgetKey{}).(*aiSQLiteBudget).bytes = aiSQLiteByteLimit - 1500
	got, err := parseGenericAISQLiteDB(ctx, provider, path)
	// The DB wrapper logs a per-table byte-limit error and returns the completed
	// rows. Both tables must retain their original cutoff for the next run.
	require.NoError(t, err)
	require.NotEmpty(t, got)
	require.Less(t, len(got), aiSQLiteBatchSize)

	for _, table := range provider.sqliteTables {
		cursorPath, err := genericAISQLiteCursorPath(t.Context(), provider, path, table)
		require.NoError(t, err)
		cursor, err := readGenericAISQLiteCursor(cursorPath)
		require.NoError(t, err)
		require.True(t, cursor.Pending)
		require.Equal(t, provider.config.After, cursor.Cutoff)
	}

	provider.config.After = started.Add(time.Hour)

	for range 100 {
		ctx, cancel := aiSQLiteContext(t.Context())
		more, parseErr := parseGenericAISQLiteDB(ctx, provider, path)
		deadlineErr := ctx.Err()

		cancel()

		if parseErr != nil {
			require.ErrorIs(t, parseErr, deadlineErr)
		}

		got = append(got, more...)
		pending := false

		for _, table := range provider.sqliteTables {
			cursorPath, err := genericAISQLiteCursorPath(t.Context(), provider, path, table)
			require.NoError(t, err)
			cursor, err := readGenericAISQLiteCursor(cursorPath)
			require.NoError(t, err)

			pending = pending || cursor.Pending || cursor.Rescan
		}

		if !pending {
			require.Len(t, got, 2*rows)
			assertZCodeTotals(t, got, rows*8, rows*2, rows*4, rows*5)

			return
		}
	}

	t.Fatal("ZCode scan did not finish")
}

func TestZCodeParseUpgradesGenericCursors(t *testing.T) {
	for _, scan := range []string{"complete", "pending", "rescan"} {
		t.Run(scan, func(t *testing.T) {
			db, path, started := zCodeFixture(t)
			_, err := db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data)
 VALUES ('m', 's', ?, ?, '{"role":"assistant","tokens":{"input":10,"output":4,"cache":{"read":2}}}')`,
				started.UnixMilli(), started.UnixMilli())
			require.NoError(t, err)

			oldProvider := genericAIProvider{parser: ZCode{}, config: ParserConfig{After: started.Add(-time.Minute)}}
			_, err = parseGenericAISQLiteDB(t.Context(), oldProvider, path)
			require.NoError(t, err)
			cursorPath, err := genericAISQLiteCursorPath(t.Context(), oldProvider, path, "message")
			require.NoError(t, err)
			cursor, err := readGenericAISQLiteCursor(cursorPath)
			require.NoError(t, err)

			cursor.Pending = scan == "pending"
			cursor.Rescan = scan == "rescan"
			require.NoError(t, writeGenericAISQLiteCursor(cursorPath, cursor))

			parser := ZCode{After: oldProvider.config.After}
			if scan != "complete" {
				parser.After = started.Add(time.Hour)
			}
			// No database write or fingerprint change: the parser version must
			// invalidate old hashes while preserving unfinished scans' cutoffs.
			got, err := parser.Parse(t.Context())
			require.NoError(t, err)
			require.Len(t, got, 1)
			assertZCodeTotals(t, got, 8, 2, 4, 0)
		})
	}
}

func TestZCodeParseSkipsHistoricalPayloads(t *testing.T) {
	db, path, started := zCodeFixture(t)

	var project string
	require.NoError(t, db.QueryRow(`SELECT directory FROM session WHERE id='s'`).Scan(&project))

	_, err := db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data)
 VALUES ('m', 's', ?, ?, '{"role":"assistant","tokens":{"input":10,"output":4}}')`,
		started.UnixMilli(), started.UnixMilli())
	require.NoError(t, err)

	// This row exceeds even SQLite's per-value limit. Filtering it by scalar
	// timestamps must happen before reading data or invoking any JSON function.
	_, err = db.Exec(`INSERT INTO part (id, message_id, session_id, time_created, time_updated, data)
 VALUES ('old', 'm', 's', ?, ?, ?)`, started.UnixMilli(), started.UnixMilli(),
		`{"type":"text","text":"`+strings.Repeat("x", maxTranscriptLineSize+1)+`"}`)
	require.NoError(t, err)

	provider := (ZCode{After: started.Add(time.Minute)}).sqliteProvider(path)
	ctx := aiSQLiteBudgetContext(t.Context())
	got, err := parseGenericAISQLiteDB(ctx, provider, path)
	require.NoError(t, err)
	require.Empty(t, got)
	assertZCodeScanCompleted(t, provider, path)
	assert.Less(t, ctx.Value(aiSQLiteBudgetKey{}).(*aiSQLiteBudget).bytes, int64(1024))

	// The same rowid can become a completed edit after its parent message has
	// aged out. The timestamp predicate must use the part, not the message.
	updated := started.Add(2 * time.Minute)
	_, err = db.Exec(`UPDATE part SET time_updated=?, data=? WHERE id='old'`, updated.UnixMilli(),
		`{"type":"tool","tool":"Edit","state":{"status":"completed",`+
			`"input":{"file_path":"main.go"},"metadata":{"display":{"additions":2,"deletions":1}}}}`)
	require.NoError(t, err)
	got, err = parseGenericAISQLiteDB(aiSQLiteBudgetContext(t.Context()), provider, path)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.NotNil(t, got[1].AILineChanges)
	assert.Equal(t, 3, *got[1].AILineChanges)
	assert.Equal(t, filepath.ToSlash(filepath.Join(project, "main.go")), got[1].Entity)
	assert.Equal(t, float64(updated.Unix()), got[1].Time)
	assertZCodeScanCompleted(t, provider, path)

	// Historical request counters still seed later token deltas.
	_, err = db.Exec(`UPDATE message SET time_updated=?, data=? WHERE id='m'`, updated.UnixMilli(),
		`{"role":"assistant","tokens":{"input":25,"output":9}}`)
	require.NoError(t, err)
	got, err = parseGenericAISQLiteDB(aiSQLiteBudgetContext(t.Context()), provider, path)
	require.NoError(t, err)
	require.Len(t, got, 1) // A parent's token update must not replay a file edit.
	assertZCodeTotals(t, got, 15, 0, 5, 0)
}

func TestZCodeParseProjectsLargeRecentParts(t *testing.T) {
	db, path, started := zCodeFixture(t)
	_, err := db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data)
 VALUES ('m', 's', ?, ?, '{"role":"assistant","tokens":{"input":10,"output":4}}')`,
		started.UnixMilli(), started.UnixMilli())
	require.NoError(t, err)

	output := strings.Repeat("x", 4*1024)
	_, err = db.Exec(`WITH RECURSIVE ids(id) AS (
 SELECT 1 UNION ALL SELECT id+1 FROM ids WHERE id<70
 ) INSERT INTO part (id, message_id, session_id, time_created, time_updated, data)
 SELECT printf('p-%03d', id), 'm', 's', ?, ?,
 CASE WHEN id%2=0 THEN json_object('type', 'text', 'text', ?)
 ELSE json_object('type', 'tool', 'tool', 'Read', 'state',
   json_object('status', 'completed', 'input', json_object('file_path', 'main.go'), 'output', ?)) END
 FROM ids`, started.UnixMilli(), started.UnixMilli(), output, output)
	require.NoError(t, err)

	provider := (ZCode{After: started.Add(-time.Minute)}).sqliteProvider(path)
	// Leave less budget than the raw payloads need, but enough for projections.
	const usedBytes = aiSQLiteByteLimit - 64*1024

	ctx := aiSQLiteBudgetContext(t.Context())
	ctx.Value(aiSQLiteBudgetKey{}).(*aiSQLiteBudget).bytes = usedBytes
	got, err := parseGenericAISQLiteDB(ctx, provider, path)
	require.NoError(t, err)
	assertZCodeTotals(t, got, 10, 0, 4, 0)
	require.Len(t, got, 71) // One usage heartbeat and 35 app/file read pairs.
	assertZCodeScanCompleted(t, provider, path)
	assert.Less(t, ctx.Value(aiSQLiteBudgetKey{}).(*aiSQLiteBudget).bytes-usedBytes, int64(64*1024))

	// Unrelated writes must not make these unchanged payloads exhaust the
	// budget on every subsequent heartbeat, or replay existing activity.
	_, err = db.Exec(`UPDATE session SET title='changed'`)
	require.NoError(t, err)
	ctx = aiSQLiteBudgetContext(t.Context())
	ctx.Value(aiSQLiteBudgetKey{}).(*aiSQLiteBudget).bytes = usedBytes
	got, err = parseGenericAISQLiteDB(ctx, provider, path)
	require.NoError(t, err)
	assert.Empty(t, got)
	assertZCodeScanCompleted(t, provider, path)
	assert.Less(t, ctx.Value(aiSQLiteBudgetKey{}).(*aiSQLiteBudget).bytes-usedBytes, int64(64*1024))
}

func assertZCodeScanCompleted(t testing.TB, provider genericAIProvider, path string) {
	t.Helper()

	for _, table := range provider.sqliteTables {
		cursorPath, err := genericAISQLiteCursorPath(t.Context(), provider, path, table)
		require.NoError(t, err)
		cursor, err := readGenericAISQLiteCursor(cursorPath)
		require.NoError(t, err)
		assert.False(t, cursor.Pending, "%s scan did not finish", table)
		assert.False(t, cursor.Rescan, "%s scan needs another pass", table)
	}
}

func BenchmarkZCodeSQLiteLargeParts(b *testing.B) {
	db, path, started := zCodeFixture(b)
	_, err := db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data)
 VALUES ('m', 's', ?, ?, '{"role":"assistant","tokens":{"input":10,"output":4}}')`,
		started.UnixMilli(), started.UnixMilli())
	require.NoError(b, err)

	output := strings.Repeat("x", 1024*1024)
	_, err = db.Exec(`WITH RECURSIVE ids(id) AS (
 SELECT 1 UNION ALL SELECT id+1 FROM ids WHERE id<834
 ) INSERT INTO part (id, message_id, session_id, time_created, time_updated, data)
 SELECT printf('p-%04d', id), 'm', 's', ?, ?,
 CASE WHEN id%2=0 THEN json_object('type', 'text', 'text', ?)
 ELSE json_object('type', 'tool', 'tool', 'Read', 'state',
   json_object('status', 'completed', 'input', json_object('file_path', 'main.go'), 'output', ?)) END
 FROM ids`, started.UnixMilli(), started.UnixMilli(), output, output)
	require.NoError(b, err)

	provider := (ZCode{After: started.Add(-time.Minute)}).sqliteProvider(path)

	for b.Loop() {
		// First iteration has cold cursors; later iterations mimic a busy agent
		// changing unrelated database content between heartbeats.
		_, err := db.Exec(`UPDATE session SET time_updated=time_updated+1`)
		require.NoError(b, err)
		ctx, cancel := aiSQLiteContext(b.Context())
		_, err = parseGenericAISQLiteDB(ctx, provider, path)
		require.NoError(b, err)
		require.NoError(b, ctx.Err())
		cancel()
		assertZCodeScanCompleted(b, provider, path)
		b.ReportMetric(float64(ctx.Value(aiSQLiteBudgetKey{}).(*aiSQLiteBudget).bytes)/1024, "KiB/read")
	}
}

func zCodeFixture(t testing.TB) (*sql.DB, string, time.Time) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("WAKATIME_HOME", t.TempDir())

	path := filepath.Join(home, ".zcode", "cli", "db", "db.sqlite")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	db, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	schema, err := os.ReadFile("testdata/zcode_schema.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(schema))
	require.NoError(t, err)

	started := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	_, err = db.Exec(`INSERT INTO session
 (id, project_id, slug, directory, title, version, time_created, time_updated)
 VALUES ('s', 'project', 'demo', ?, 'Not a user prompt', 'test', ?, ?)`,
		filepath.Join(home, "project"), started.UnixMilli(), started.UnixMilli())
	require.NoError(t, err)

	return db, path, started
}

func assertZCodeTotals(t *testing.T, got Heartbeats, input, cached, output int64, prompt int) {
	t.Helper()

	var (
		actualInput, actualCached, actualOutput int64
		actualPrompt                            int
	)

	for _, hb := range got {
		actualInput += hb.AIInputTokens
		actualCached += hb.AICachedInputTokens
		actualOutput += hb.AIOutputTokens
		actualPrompt += hb.AIPromptLength
	}

	assert.Equal(t, input, actualInput)
	assert.Equal(t, cached, actualCached)
	assert.Equal(t, output, actualOutput)
	assert.Equal(t, prompt, actualPrompt)
}
