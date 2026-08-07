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

	_ "modernc.org/sqlite"
)

func TestForgeParseSQLite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, ".forge")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))
	dbPath := filepath.Join(dbDir, ".forge.db")

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE usage (
		session_id TEXT,
		created_at TEXT,
		model TEXT,
		working_dir TEXT,
		prompt TEXT,
		input_tokens INTEGER,
		output_tokens INTEGER,
		cache_read_tokens INTEGER
	)`)
	require.NoError(t, err)
	_, err = db.Exec(
		`INSERT INTO usage VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"forge-1",
		"2026-08-06T12:00:00Z",
		"gpt-5.2",
		"/workspace/forge",
		"Ship the feature",
		100,
		25,
		10,
	)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	got, err := (Forge{After: time.Date(2026, 8, 6, 11, 0, 0, 0, time.UTC)}).Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, "Forge forge-1", got[0].Entity)
	assert.Equal(t, int64(100), got[0].AIInputTokens)
	assert.Equal(t, int64(10), got[0].AICachedInputTokens)
	assert.Equal(t, int64(25), got[0].AIOutputTokens)
	assert.Equal(t, "/workspace/forge", got[0].ProjectPathOverride)
}
