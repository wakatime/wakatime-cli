//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"database/sql"
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

	got, err := (ZCode{After: time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)}).Parse(t.Context())
	require.NoError(t, err)
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

	got, err = (ZCode{After: time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)}).Parse(t.Context())
	require.NoError(t, err)
	assert.Len(t, got, 20004)
}
