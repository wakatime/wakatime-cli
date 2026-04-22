//go:build !freebsd && !openbsd && !netbsd && !dragonfly

package ai_test

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
	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	_ "modernc.org/sqlite"
)

func TestGooseParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, ".local", "share", "goose", "sessions")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "sessions.db")
	createGooseDB(t, dbPath)

	parser := ai.Goose{
		After:             time.Date(2026, 4, 20, 11, 59, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, "Goose 20260420_1", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "20260420_1", got[0].AISession)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[0].Category)
	assert.Equal(t, "/workspace/goose", got[0].ProjectPathOverride)
	assert.Equal(t, len([]rune("Investigate the failing tests")), got[0].AIPromptLength)
	assert.Zero(t, got[0].AIInputTokens)
	assert.Zero(t, got[0].AIOutputTokens)
	assert.Equal(t, float64(time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC).Unix()), got[0].Time)
	assert.Contains(t, got[0].UserAgent, "Goose/anthropic")
	assert.True(t, strings.Index(got[0].UserAgent, "Goose/anthropic") < strings.Index(got[0].UserAgent, "plugin/0.0.1"))
}

func TestGooseParse_PlainTokenColumns(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, ".local", "share", "goose", "sessions")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "sessions.db")
	createGooseDBWithPlainTokens(t, dbPath)

	parser := ai.Goose{
		After:             time.Date(2026, 4, 20, 11, 59, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.EqualValues(t, 13, got[0].AIInputTokens)
	assert.EqualValues(t, 8, got[0].AIOutputTokens)
}

func TestGooseParse_NoSessionDB(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := ai.Goose{After: time.Now()}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func createGooseDB(t *testing.T, dbPath string) {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)

	defer db.Close() // nolint:errcheck

	_, err = db.Exec(`
CREATE TABLE sessions (
	id TEXT PRIMARY KEY,
	name TEXT,
	working_dir TEXT,
	session_type TEXT,
	provider_name TEXT,
	updated_at TEXT,
	accumulated_input_tokens INTEGER,
	accumulated_output_tokens INTEGER
);
INSERT INTO sessions (
	id,
	name,
	working_dir,
	session_type,
	provider_name,
	updated_at,
	accumulated_input_tokens,
	accumulated_output_tokens
) VALUES
	('20260419_1', 'Old session', '/workspace/old', 'Normal', 'openai', '2026-04-19T12:00:00Z', 10, 5),
	(
		'20260420_1',
		'Investigate the failing tests',
		'/workspace/goose',
		'Normal',
		'anthropic',
		'2026-04-20T12:00:00Z',
		55,
		21
	),
	('20260420_hidden', 'Hidden session', '/workspace/hidden', 'Hidden', 'anthropic', '2026-04-20T12:00:00Z', 99, 99);
`)
	require.NoError(t, err)
}

func createGooseDBWithPlainTokens(t *testing.T, dbPath string) {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)

	defer db.Close() // nolint:errcheck

	_, err = db.Exec(`
CREATE TABLE sessions (
	id TEXT PRIMARY KEY,
	name TEXT,
	working_dir TEXT,
	session_type TEXT,
	provider_name TEXT,
	updated_at TEXT,
	input_tokens INTEGER,
	output_tokens INTEGER
);
INSERT INTO sessions (
	id,
	name,
	working_dir,
	session_type,
	provider_name,
	updated_at,
	input_tokens,
	output_tokens
) VALUES (
	'20260420_1',
	'Investigate the failing tests',
	'/workspace/goose',
	'Normal',
	'anthropic',
	'2026-04-20T12:00:00Z',
	13,
	8
);
`)
	require.NoError(t, err)
}
