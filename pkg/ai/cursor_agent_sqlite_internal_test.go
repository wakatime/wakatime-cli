//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

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
)

func TestCursorAgentSkipsIrrelevantSQLiteTables(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("WAKATIME_HOME", t.TempDir())

	dbPath := filepath.Join(home, ".cursor", "ai-tracking", "ai-code-tracking.db")
	require.NoError(t, os.MkdirAll(filepath.Dir(dbPath), 0o755))

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	_, err = db.Exec(`CREATE TABLE tracked_file_content (content TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO tracked_file_content VALUES (?)`, strings.Repeat("x", 2048))
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE ai_code_hashes (
		hash TEXT PRIMARY KEY,
		fileName TEXT,
		conversationId TEXT,
		timestamp INTEGER,
		model TEXT
	)`)
	require.NoError(t, err)

	filePath := filepath.Join(home, "project", "main.go")
	_, err = db.Exec(`INSERT INTO ai_code_hashes VALUES (?, ?, ?, ?, ?)`,
		"hash", filePath, "conversation", time.Now().UnixMilli(), "grok-4.5")
	require.NoError(t, err)

	ctx := context.WithValue(t.Context(), aiSQLiteBudgetKey{}, &aiSQLiteBudget{
		bytes: aiSQLiteByteLimit - 1024,
	})
	got, err := (CursorAgent{}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "Cursor Agent conversation", got[0].Entity)
	assert.Equal(t, "grok/4.5 cursor-agent", got[0].UserAgent)
	assert.Equal(t, filepath.ToSlash(filePath), got[1].Entity)
	assert.Equal(t, "grok/4.5 cursor-agent", got[1].UserAgent)
}

func TestCursorAgentSkipsHistoricalSQLiteRows(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("WAKATIME_HOME", t.TempDir())

	dbPath := filepath.Join(home, ".cursor", "ai-tracking", "ai-code-tracking.db")
	require.NoError(t, os.MkdirAll(filepath.Dir(dbPath), 0o755))
	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	_, err = db.Exec(`CREATE TABLE ai_code_hashes (
		hash TEXT PRIMARY KEY,
		fileName TEXT,
		conversationId TEXT,
		timestamp INTEGER,
		model TEXT
	)`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE TABLE tracking_state (value TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO tracking_state VALUES ('initial')`)
	require.NoError(t, err)

	cutoff := time.Date(2026, 9, 6, 12, 0, 0, 123*int(time.Millisecond), time.UTC)
	filePath := filepath.Join(home, "project", "main.go")
	// History in the retained table also exceeds the available budget. Exclude
	// it in SQLite before transferring, hashing, or charging its text values.
	_, err = db.Exec(`WITH RECURSIVE ids(id) AS (
		SELECT 1 UNION ALL SELECT id + 1 FROM ids WHERE id < ?
	) INSERT INTO ai_code_hashes
	SELECT printf('%064d', id), ?, 'old-conversation', ?, 'grok-4.5' FROM ids`,
		aiSQLiteBatchSize+1, filePath, cutoff.Add(-time.Minute).UnixMilli())
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO ai_code_hashes VALUES (?, ?, ?, ?, ?)`,
		"recent-hash", filePath, "recent-conversation", cutoff.UnixMilli(), "grok-4.5")
	require.NoError(t, err)

	parser := CursorAgent{After: cutoff}
	budgetContext := func() context.Context {
		return context.WithValue(t.Context(), aiSQLiteBudgetKey{}, &aiSQLiteBudget{
			bytes: aiSQLiteByteLimit - 1024,
		})
	}
	ctx := budgetContext()
	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "Cursor Agent recent-conversation", got[0].Entity)
	assert.Equal(t, "grok/4.5 cursor-agent", got[0].UserAgent)
	assert.Equal(t, filepath.ToSlash(filePath), got[1].Entity)
	assert.Equal(t, "grok/4.5 cursor-agent", got[1].UserAgent)
	assert.LessOrEqual(t, ctx.Value(aiSQLiteBudgetKey{}).(*aiSQLiteBudget).bytes, int64(aiSQLiteByteLimit))

	for _, h := range got {
		encoded, err := json.Marshal(h)
		require.NoError(t, err)
		// Provenance has no billed token counters: do not invent zero-usage data.
		assert.NotContains(t, string(encoded), "ai_input_tokens")
		assert.NotContains(t, string(encoded), "ai_cached_input_tokens")
		assert.NotContains(t, string(encoded), "ai_output_tokens")
	}

	ctx = budgetContext()
	got, err = parser.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
	assert.Equal(t, int64(aiSQLiteByteLimit-1024), ctx.Value(aiSQLiteBudgetKey{}).(*aiSQLiteBudget).bytes)

	// An unrelated tracking-table write changes the database fingerprint. The
	// rescan must still exclude history and must not replay existing attribution.
	_, err = db.Exec(`UPDATE tracking_state SET value = 'changed'`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO ai_code_hashes VALUES (?, ?, ?, ?, ?)`,
		"new-hash", filePath, "new-conversation", cutoff.Add(time.Minute).UnixMilli(), "grok-4.5")
	require.NoError(t, err)

	ctx = budgetContext()
	got, err = parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "Cursor Agent new-conversation", got[0].Entity)
	assert.Equal(t, filepath.ToSlash(filePath), got[1].Entity)
	assert.LessOrEqual(t, ctx.Value(aiSQLiteBudgetKey{}).(*aiSQLiteBudget).bytes, int64(aiSQLiteByteLimit))
}
