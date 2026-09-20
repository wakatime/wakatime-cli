//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sqliteTestDB(t *testing.T) (*sql.DB, string) {
	t.Helper()
	t.Setenv("WAKATIME_HOME", t.TempDir())
	path := filepath.Join(t.TempDir(), "agent.sqlite")
	db, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	return db, path
}

func TestGenericAISQLitePaths(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	require.NoError(t, os.Mkdir(nested, 0o700))

	var want []string

	for _, name := range []string{"agent.db", "session.SQLITE", "history.sqlite3", "ignored.json", "agent.db-wal"} {
		path := filepath.Join(nested, name)
		require.NoError(t, os.WriteFile(path, nil, 0o600))

		if name != "ignored.json" && name != "agent.db-wal" {
			want = append(want, path)
		}
	}

	got, err := genericAISQLitePaths(ZCode{}, ParserConfig{}, []string{"", filepath.Join(root, "missing"), root, nested})
	require.NoError(t, err)
	assert.ElementsMatch(t, want, got)

	got, err = genericAISQLitePaths(ZCode{}, ParserConfig{}, []string{
		filepath.Join(nested, "agent.db"), filepath.Join(nested, "ignored.json"), filepath.Join(nested, "agent.db-wal"),
	})
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(nested, "agent.db")}, got)
}

func TestGenericAISQLiteContinuesAfterCorruptDatabase(t *testing.T) {
	db, path := sqliteTestDB(t)
	_, err := db.Exec(`CREATE TABLE events (payload TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO events VALUES (?)`,
		`{"timestamp":"2026-09-06T12:00:00Z","session_id":"s","input_tokens":10}`)
	require.NoError(t, err)

	corrupt := filepath.Join(t.TempDir(), "corrupt.db")
	require.NoError(t, os.WriteFile(corrupt, []byte("not a sqlite database"), 0o600))

	provider := genericAIProvider{parser: ZCode{}, sqliteRoots: []string{corrupt, path}}
	got, err := parseGenericAISQLite(t.Context(), provider)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, int64(10), got[0].AIInputTokens)
	assert.Equal(t, "ZCode s", got[0].Entity)
}

func TestGenericAISQLitePropagatesCancellation(t *testing.T) {
	_, path := sqliteTestDB(t)
	require.NoError(t, os.WriteFile(path, nil, 0o600))

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	got, err := parseGenericAISQLite(ctx, genericAIProvider{parser: ZCode{}, sqliteRoots: []string{path}})
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, got)
}

func TestGenericAISQLiteCursorResumesAfterBudget(t *testing.T) {
	db, path := sqliteTestDB(t)
	_, err := db.Exec(`CREATE TABLE events (payload TEXT)`)
	require.NoError(t, err)

	payload := `{"timestamp":"2026-09-06T12:00:00Z","session_id":"s","input_tokens":10,"cwd":"/project"}`
	_, err = db.Exec(`WITH RECURSIVE ids(id) AS (
		SELECT 1 UNION ALL SELECT id + 1 FROM ids WHERE id < ?
	) INSERT INTO events SELECT ? FROM ids`, aiSQLiteBatchSize+1, payload)
	require.NoError(t, err)

	provider := genericAIProvider{
		parser: ZCode{}, config: ParserConfig{After: time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)},
	}
	ctx := context.WithValue(t.Context(), aiSQLiteBudgetKey{}, &aiSQLiteBudget{
		bytes: aiSQLiteByteLimit - int64(aiSQLiteBatchSize*len(payload)),
	})
	got, err := parseGenericAISQLiteIncremental(ctx, provider, db, path, "events", "rowid")
	require.ErrorContains(t, err, "byte budget")
	require.Len(t, got, aiSQLiteBatchSize)

	// A later global timestamp must not skip the rest of an interrupted scan.
	provider.config.After = time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	got, err = parseGenericAISQLiteIncremental(t.Context(), provider, db, path, "events", "rowid")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, int64(10), got[0].AIInputTokens)

	got, err = parseGenericAISQLiteIncremental(t.Context(), provider, db, path, "events", "rowid")
	require.NoError(t, err)
	assert.Empty(t, got)

	// New rows inherit session metadata, but already parsed rows are not read.
	_, err = db.Exec(`INSERT INTO events VALUES (?)`,
		`{"timestamp":"2026-09-07T12:00:00Z","session_id":"s","file_path":"main.go","tool_name":"edit_file"}`)
	require.NoError(t, err)
	got, err = parseGenericAISQLiteIncremental(t.Context(), provider, db, path, "events", "rowid")
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, filepath.ToSlash(filepath.Join("/project", "main.go")), got[1].Entity)
}

func TestGenericAISQLiteCursorPreservesCumulativeCounters(t *testing.T) {
	db, path := sqliteTestDB(t)
	_, err := db.Exec(`CREATE TABLE events (payload TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO events VALUES (?)`,
		`{"timestamp":"2026-09-06T12:00:00Z","session_id":"s","input_tokens":10}`)
	require.NoError(t, err)

	provider := genericAIProvider{parser: ZCode{}, tokenCounterMode: genericAICumulativeCounters}
	got, err := parseGenericAISQLiteTable(t.Context(), provider, db, path, "events")
	require.NoError(t, err)
	require.Len(t, got, 1)

	_, err = db.Exec(`INSERT INTO events VALUES (?)`,
		`{"timestamp":"2026-09-06T12:01:00Z","session_id":"s","input_tokens":15}`)
	require.NoError(t, err)
	got, err = parseGenericAISQLiteTable(t.Context(), provider, db, path, "events")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, int64(5), got[0].AIInputTokens)

	// A truncated/rebuilt table must not remain stuck behind the old cursor.
	_, err = db.Exec(`DELETE FROM events`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO events VALUES (?)`,
		`{"timestamp":"2026-09-06T12:02:00Z","session_id":"new","input_tokens":7}`)
	require.NoError(t, err)
	got, err = parseGenericAISQLiteTable(t.Context(), provider, db, path, "events")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, int64(7), got[0].AIInputTokens)
}

func TestGenericAISQLiteMutableRows(t *testing.T) {
	db, path := sqliteTestDB(t)
	_, err := db.Exec(`CREATE TABLE conversations (updated_at TEXT, input_tokens INTEGER)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO conversations VALUES ('2026-09-06T12:00:00Z', 10)`)
	require.NoError(t, err)

	provider := genericAIProvider{parser: ZCode{}}
	got, err := parseGenericAISQLiteTable(t.Context(), provider, db, path, "conversations")
	require.NoError(t, err)
	require.Len(t, got, 1)

	_, err = db.Exec(`UPDATE conversations SET updated_at = '2026-09-06T12:01:00Z', input_tokens = 15`)
	require.NoError(t, err)

	provider.config.After = time.Date(2026, 9, 6, 12, 0, 30, 0, time.UTC)
	got, err = parseGenericAISQLiteTable(t.Context(), provider, db, path, "conversations")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, int64(15), got[0].AIInputTokens)
}

func TestGenericAISQLiteRowIDAliases(t *testing.T) {
	for _, test := range []struct{ schema, want string }{
		{`CREATE TABLE events (payload TEXT)`, "rowid"},
		{`CREATE TABLE events (rowid TEXT, payload TEXT)`, "_rowid_"},
		{`CREATE TABLE events (rowid TEXT, _rowid_ TEXT, oid TEXT)`, ""},
		{`CREATE TABLE events (id INTEGER PRIMARY KEY, payload TEXT) WITHOUT ROWID`, ""},
		{`CREATE TABLE events (updatedAt TEXT, payload TEXT)`, "rowid"},
	} {
		t.Run(test.schema, func(t *testing.T) {
			db, _ := sqliteTestDB(t)
			_, err := db.Exec(test.schema)
			require.NoError(t, err)
			rowID, err := genericAISQLiteRowID(t.Context(), db, "events")
			require.NoError(t, err)
			assert.Equal(t, test.want, rowID)
		})
	}
}

func TestAISQLiteLimitsBeforeJSON(t *testing.T) {
	db, path := sqliteTestDB(t)
	_, err := db.Exec(`CREATE TABLE events (payload TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO events VALUES (?)`, `{"text":"`+strings.Repeat("x", maxTranscriptLineSize)+`"}`)
	require.NoError(t, err)
	limited, err := openAISQLiteDB(t.Context(), path)
	require.NoError(t, err)

	defer limited.Close()

	var value string

	err = limited.QueryRowContext(t.Context(), `SELECT json_extract(payload, '$.text') FROM events`).Scan(&value)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too big")
}

func TestAISQLiteSharedBudgetAndCancellation(t *testing.T) {
	ctx, cancel := aiSQLiteContext(t.Context())
	defer cancel()

	child, childCancel := aiSQLiteContext(ctx)
	defer childCancel()

	assert.Same(t, ctx.Value(aiSQLiteBudgetKey{}), child.Value(aiSQLiteBudgetKey{}))

	payload := strings.Repeat("x", maxTranscriptLineSize)
	for range 6 {
		require.NoError(t, aiSQLiteRead(child, payload))
	}

	require.ErrorContains(t, aiSQLiteRead(ctx, payload), "byte budget")
	cancel()
	require.ErrorIs(t, aiSQLiteRead(child, ""), context.Canceled)
}

func TestAISQLiteWALModifiedAfter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.vscdb")
	cutoff := time.Now().Add(-time.Hour)

	require.NoError(t, os.WriteFile(path, nil, 0o600))
	require.NoError(t, os.Chtimes(path, cutoff.Add(-time.Hour), cutoff.Add(-time.Hour)))
	assert.False(t, aiSQLiteModifiedAfter(path, cutoff))
	require.NoError(t, os.WriteFile(path+"-wal", nil, 0o600))
	assert.True(t, aiSQLiteModifiedAfter(path, cutoff))
}

func TestGenericAISQLiteJSONText(t *testing.T) {
	db, _ := sqliteTestDB(t)
	rows, err := db.Query(`SELECT ? AS payload, ? AS prompt`,
		`{"timestamp":"2026-09-06T12:00:00Z","input_tokens":10}`, `{"example":"prompt"}`)
	require.NoError(t, err)

	defer rows.Close()

	require.True(t, rows.Next())
	row, err := genericAISQLiteRow(t.Context(), rows, []string{"payload", "prompt"})
	require.NoError(t, err)

	event := genericAIEventFromValue(row)
	assert.Equal(t, int64(10), event.input)
	assert.Equal(t, `{"example":"prompt"}`, event.prompt)
	assert.IsType(t, genericAIJSONText{}, row["payload"])
}

func TestAISQLiteMutableKeyValueRows(t *testing.T) {
	db, path := sqliteTestDB(t)
	_, err := db.Exec(`CREATE TABLE cursorDiskKV (key TEXT PRIMARY KEY, value TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE TABLE ItemTable (key TEXT PRIMARY KEY, value TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO cursorDiskKV VALUES ('bubbleId:s:prompt', ?)`,
		`{"createdAt":"2026-09-06T12:00:00Z","type":1,"text":"first"}`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO ItemTable VALUES (?, 'first')`, codyStorageKey)
	require.NoError(t, err)

	for _, value := range []string{"first", "updated"} {
		_, err = db.Exec(`UPDATE cursorDiskKV SET value = json_set(value, '$.text', ?)`, value)
		require.NoError(t, err)
		_, err = db.Exec(`UPDATE ItemTable SET value = ?`, value)
		require.NoError(t, err)

		cursorRows, err := (Cursor{}).queryRows(t.Context(), db, path)
		require.NoError(t, err)
		require.Len(t, cursorRows, 1)
		assert.Contains(t, cursorRows[0].Value, value)
		windsurfRows, err := (Windsurf{}).queryRows(t.Context(), path)
		require.NoError(t, err)
		require.Len(t, windsurfRows, 1)
		assert.Contains(t, windsurfRows[0].Value, value)
		storage, err := (Cody{}).queryStorage(t.Context(), path)
		require.NoError(t, err)
		assert.Equal(t, value, storage)
	}
}

func TestAISQLiteTimeBudgetDoesNotCancelOtherParsers(t *testing.T) {
	ctx := aiSQLiteBudgetContext(t.Context())
	budget := ctx.Value(aiSQLiteBudgetKey{}).(*aiSQLiteBudget)
	budget.spent = aiSQLiteTimeout

	parseCtx, cancel := aiSQLiteContext(ctx)
	defer cancel()

	require.ErrorIs(t, parseCtx.Err(), context.DeadlineExceeded)
	require.NoError(t, ctx.Err())
}

func BenchmarkGenericAISQLiteUnchanged(b *testing.B) {
	b.Setenv("WAKATIME_HOME", b.TempDir())
	path := filepath.Join(b.TempDir(), "agent.sqlite")
	db, err := sql.Open("sqlite", path)
	require.NoError(b, err)

	defer db.Close()

	_, err = db.Exec(`CREATE TABLE events (payload TEXT)`)
	require.NoError(b, err)
	_, err = db.Exec(`WITH RECURSIVE ids(id) AS (
		SELECT 1 UNION ALL SELECT id + 1 FROM ids WHERE id < 10000
	) INSERT INTO events SELECT '{"timestamp":"2026-09-06T12:00:00Z","input_tokens":10}' FROM ids`)
	require.NoError(b, err)

	provider := genericAIProvider{parser: ZCode{}}
	_, err = parseGenericAISQLiteTable(b.Context(), provider, db, path, "events")
	require.NoError(b, err)
	b.ResetTimer()

	for b.Loop() {
		got, err := parseGenericAISQLiteTable(b.Context(), provider, db, path, "events")
		require.NoError(b, err)
		require.Empty(b, got)
	}
}

func TestGenericAISQLiteLargeBatchResumes(t *testing.T) {
	db, path := sqliteTestDB(t)
	_, err := db.Exec(`CREATE TABLE events (payload TEXT)`)
	require.NoError(t, err)
	// Each row fits the 10 MiB limit, but eight rows exceed the 64 MiB run budget.
	payload := `{"timestamp":"2026-09-06T12:00:00Z","input_tokens":10,"padding":"` + strings.Repeat("x", 9*1024*1024) + `"}`
	_, err = db.Exec(`WITH RECURSIVE ids(id) AS (
 SELECT 1 UNION ALL SELECT id + 1 FROM ids WHERE id < 8
 ) INSERT INTO events SELECT ? FROM ids`, payload)
	require.NoError(t, err)

	provider := genericAIProvider{parser: ZCode{}}
	got, err := parseGenericAISQLiteTable(aiSQLiteBudgetContext(t.Context()), provider, db, path, "events")
	require.ErrorContains(t, err, "byte budget")
	require.Len(t, got, 7)

	provider.config.After = time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	got, err = parseGenericAISQLiteTable(aiSQLiteBudgetContext(t.Context()), provider, db, path, "events")
	require.NoError(t, err)
	require.Len(t, got, 1)
}

func TestGenericAISQLiteAllTableKindsResume(t *testing.T) {
	for _, test := range []struct{ name, schema, insert, payload, container string }{
		{"without_rowid", `CREATE TABLE events (id INTEGER PRIMARY KEY, payload TEXT) WITHOUT ROWID`,
			`INSERT INTO events VALUES (?, ?)`, `{"timestamp":"2026-09-06T12:00:00Z","input_tokens":10}`, ""},
		{"mutable", `CREATE TABLE events (updated_at TEXT, payload TEXT)`,
			`INSERT INTO events VALUES (?, ?)`, `{"timestamp":"2026-09-06T12:00:00Z","input_tokens":10}`, ""},
		{"container", `CREATE TABLE events (id INTEGER PRIMARY KEY, messages TEXT)`,
			`INSERT INTO events VALUES (?, ?)`, `[{"timestamp":"2026-09-06T12:00:00Z","input_tokens":10}]`, "messages"},
		{"composite_key", `CREATE TABLE events (payload TEXT, id INTEGER, tag TEXT DEFAULT 'a',
 PRIMARY KEY(tag,id)) WITHOUT ROWID`,
			`INSERT INTO events(id,payload) VALUES (?, ?)`, `{"timestamp":"2026-09-06T12:00:00Z","input_tokens":10}`, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, path := sqliteTestDB(t)
			_, err := db.Exec(test.schema)
			require.NoError(t, err)

			for i := range 3 {
				_, err = db.Exec(test.insert, i, test.payload)
				require.NoError(t, err)
			}

			provider := genericAIProvider{parser: ZCode{}, containerKey: test.container}
			ctx := context.WithValue(t.Context(), aiSQLiteBudgetKey{}, &aiSQLiteBudget{
				bytes: aiSQLiteByteLimit - int64(len(test.payload)+4),
			})
			got, err := parseGenericAISQLiteTable(ctx, provider, db, path, "events")
			require.ErrorContains(t, err, "byte budget")
			require.Len(t, got, 1)

			provider.config.After = time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
			got, err = parseGenericAISQLiteTable(aiSQLiteBudgetContext(t.Context()), provider, db, path, "events")
			require.NoError(t, err)
			require.Len(t, got, 2)
			got, err = parseGenericAISQLiteTable(aiSQLiteBudgetContext(t.Context()), provider, db, path, "events")
			require.NoError(t, err)
			assert.Empty(t, got)
		})
	}
}

func TestGenericAISQLiteJSONUpdateDuringScan(t *testing.T) {
	db, path := sqliteTestDB(t)
	_, err := db.Exec(`CREATE TABLE conversations (id INTEGER PRIMARY KEY, payload TEXT)`)
	require.NoError(t, err)

	payload := `{"updated_at":"2026-09-06T12:00:00Z","input_tokens":10}`
	for i := range 3 {
		_, err = db.Exec(`INSERT INTO conversations VALUES (?, ?)`, i, payload)
		require.NoError(t, err)
	}

	provider := genericAIProvider{parser: ZCode{}}
	ctx := context.WithValue(t.Context(), aiSQLiteBudgetKey{}, &aiSQLiteBudget{
		bytes: aiSQLiteByteLimit - int64(len(payload)),
	})
	got, err := parseGenericAISQLiteTable(ctx, provider, db, path, "conversations")
	require.ErrorContains(t, err, "byte budget")
	require.Len(t, got, 1)
	// Update an already checkpointed row while a scan is pending. The following
	// scan must revisit it without losing it to the newer global timestamp.
	_, err = db.Exec(`UPDATE conversations SET payload =
 '{"updated_at":"2026-09-06T12:01:00Z","input_tokens":15}' WHERE id=0`)
	require.NoError(t, err)

	provider.config.After = time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	got, err = parseGenericAISQLiteTable(t.Context(), provider, db, path, "conversations")
	require.NoError(t, err)
	require.Len(t, got, 2)
	got, err = parseGenericAISQLiteTable(t.Context(), provider, db, path, "conversations")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, int64(15), got[0].AIInputTokens)
}

func TestGenericAISQLitePendingTablesRunFirst(t *testing.T) {
	db, path := sqliteTestDB(t)
	_, err := db.Exec(`CREATE TABLE first (payload TEXT); CREATE TABLE second (payload TEXT)`)
	require.NoError(t, err)

	payload := `{"timestamp":"2026-09-06T12:00:00Z","input_tokens":10}`
	_, err = db.Exec(`INSERT INTO first VALUES (?);`, payload)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO second VALUES (?);`, payload)
	require.NoError(t, err)

	provider := genericAIProvider{parser: ZCode{}}
	budget := func() context.Context {
		return context.WithValue(t.Context(), aiSQLiteBudgetKey{}, &aiSQLiteBudget{
			bytes: aiSQLiteByteLimit - int64(len(payload)),
		})
	}
	got, err := parseGenericAISQLiteDB(budget(), provider, path)
	require.NoError(t, err)
	require.Len(t, got, 1)

	_, err = db.Exec(`UPDATE first SET payload = replace(payload, '10', '20')`)
	require.NoError(t, err)

	provider.config.After = time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	got, err = parseGenericAISQLiteDB(budget(), provider, path)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, int64(10), got[0].AIInputTokens)
}

func TestGenericAISQLiteMutableJSONRows(t *testing.T) {
	db, path := sqliteTestDB(t)
	_, err := db.Exec(`CREATE TABLE conversations (id TEXT PRIMARY KEY, payload TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO conversations VALUES ('s', '{"updated_at":"2026-09-06T12:00:00Z","input_tokens":10}')`)
	require.NoError(t, err)

	provider := genericAIProvider{parser: ZCode{}}
	got, err := parseGenericAISQLiteTable(t.Context(), provider, db, path, "conversations")
	require.NoError(t, err)
	require.Len(t, got, 1)

	_, err = db.Exec(`UPDATE conversations SET payload = '{"updated_at":"2026-09-06T12:01:00Z","input_tokens":15}'`)
	require.NoError(t, err)

	provider.config.After = time.Date(2026, 9, 6, 12, 0, 30, 0, time.UTC)
	got, err = parseGenericAISQLiteTable(t.Context(), provider, db, path, "conversations")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, int64(15), got[0].AIInputTokens)
}

func TestGenericAISQLiteInsertBeforePendingOffset(t *testing.T) {
	db, path := sqliteTestDB(t)
	_, err := db.Exec(`CREATE TABLE events (id INTEGER PRIMARY KEY, payload TEXT) WITHOUT ROWID`)
	require.NoError(t, err)

	payload := `{"timestamp":"2026-09-06T12:00:00Z","input_tokens":10}`
	for _, id := range []int{1, 3} {
		_, err = db.Exec(`INSERT INTO events VALUES (?, ?)`, id, payload)
		require.NoError(t, err)
	}

	provider := genericAIProvider{parser: ZCode{}}
	ctx := context.WithValue(t.Context(), aiSQLiteBudgetKey{}, &aiSQLiteBudget{
		bytes: aiSQLiteByteLimit - int64(len(payload)),
	})
	got, err := parseGenericAISQLiteTable(ctx, provider, db, path, "events")
	require.ErrorContains(t, err, "byte budget")
	require.Len(t, got, 1)

	_, err = db.Exec(`INSERT INTO events VALUES (0, ?)`, payload)
	require.NoError(t, err)

	provider.config.After = time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	// OFFSET now encounters id=1 again. It must not emit a duplicate.
	got, err = parseGenericAISQLiteTable(t.Context(), provider, db, path, "events")
	require.NoError(t, err)
	require.Len(t, got, 1)
	// The follow-up scan catches the inserted id=0 using the original cutoff.
	got, err = parseGenericAISQLiteTable(t.Context(), provider, db, path, "events")
	require.NoError(t, err)
	require.Len(t, got, 1)
}
