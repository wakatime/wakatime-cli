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

func TestGooseParse_ActualSQLiteSchema(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, ".local", "share", "goose", "sessions")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "sessions.db")
	createGooseDBWithActualSchema(t, dbPath)

	parser := ai.Goose{
		After:             time.Date(2026, 4, 20, 11, 59, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, "Goose 20260420_1", got[0].Entity)
	assert.Equal(t, "20260420_1", got[0].AISession)
	assert.Equal(t, "/workspace/goose", got[0].ProjectPathOverride)
	assert.Equal(t, len([]rune("Fix the Goose parser using the real schema")), got[0].AIPromptLength)
	assert.EqualValues(t, 13, got[0].AIInputTokens)
	assert.EqualValues(t, 8, got[0].AIOutputTokens)
	assert.Equal(t, float64(time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC).Unix()), got[0].Time)
	assert.Contains(t, got[0].UserAgent, "Goose")
}

func TestGooseParse_MissingOptionalColumnsKeepsParsing(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, ".local", "share", "goose", "sessions")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "sessions.db")
	createGooseDBWithMinimalSchema(t, dbPath)

	got, err := ai.Goose{
		After: time.Date(2026, 4, 20, 11, 59, 0, 0, time.UTC),
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, "Goose 20260420_1", got[0].Entity)
	assert.Empty(t, got[0].ProjectPathOverride)
	assert.Equal(t, len([]rune("Minimal Goose session")), got[0].AIPromptLength)
}

func TestGooseParse_MissingMinimumColumnsSkipsWithoutError(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, ".local", "share", "goose", "sessions")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "sessions.db")
	createGooseDBWithUnsupportedSchema(t, dbPath)

	got, err := ai.Goose{}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
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

func TestGooseParse_SkipsStaleSessionDB(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(home, ".local", "share", "goose", "sessions")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "sessions.db")
	require.NoError(t, os.WriteFile(dbPath, []byte("not sqlite"), 0o600))

	staleTime := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	require.NoError(t, os.Chtimes(dbPath, staleTime, staleTime))

	got, err := ai.Goose{
		After: staleTime.Add(time.Second),
	}.Parse(ctx)
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

func createGooseDBWithMinimalSchema(t *testing.T, dbPath string) {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)

	defer db.Close() // nolint:errcheck

	_, err = db.Exec(`
CREATE TABLE sessions (
	id TEXT PRIMARY KEY,
	name TEXT,
	updated_at TEXT
);
INSERT INTO sessions (
	id,
	name,
	updated_at
) VALUES (
	'20260420_1',
	'Minimal Goose session',
	'2026-04-20T12:00:00Z'
);
`)
	require.NoError(t, err)
}

func createGooseDBWithUnsupportedSchema(t *testing.T, dbPath string) {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)

	defer db.Close() // nolint:errcheck

	_, err = db.Exec(`
CREATE TABLE sessions (
	name TEXT,
	working_dir TEXT
);
INSERT INTO sessions (
	name,
	working_dir
) VALUES (
	'Unsupported Goose session',
	'/workspace/goose'
);
`)
	require.NoError(t, err)
}

func createGooseDBWithActualSchema(t *testing.T, dbPath string) {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)

	defer db.Close() // nolint:errcheck

	_, err = db.Exec(`
CREATE TABLE schema_version (
	version INTEGER PRIMARY KEY,
	applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE sessions (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL DEFAULT '',
	description TEXT NOT NULL DEFAULT '',
	user_set_name BOOLEAN DEFAULT FALSE,
	session_type TEXT NOT NULL DEFAULT 'user',
	working_dir TEXT NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	extension_data TEXT DEFAULT '{}',
	total_tokens INTEGER,
	input_tokens INTEGER,
	output_tokens INTEGER,
	accumulated_total_tokens INTEGER,
	accumulated_input_tokens INTEGER,
	accumulated_output_tokens INTEGER,
	schedule_id TEXT,
	recipe_json TEXT,
	user_recipe_values_json TEXT,
	provider_name TEXT,
	model_config_json TEXT,
	goose_mode TEXT NOT NULL DEFAULT 'auto',
	thread_id TEXT
);
CREATE TABLE messages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	message_id TEXT,
	session_id TEXT NOT NULL REFERENCES sessions(id),
	role TEXT NOT NULL,
	content_json TEXT NOT NULL,
	created_timestamp INTEGER NOT NULL,
	timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	tokens INTEGER,
	metadata_json TEXT
);
CREATE INDEX idx_messages_message_id ON messages(message_id);
CREATE INDEX idx_messages_session ON messages(session_id);
CREATE INDEX idx_messages_timestamp ON messages(timestamp);
CREATE INDEX idx_sessions_updated ON sessions(updated_at DESC);
CREATE INDEX idx_sessions_type ON sessions(session_type);
CREATE INDEX idx_sessions_thread ON sessions(thread_id);
CREATE TABLE threads (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL DEFAULT 'New Chat',
	user_set_name BOOLEAN DEFAULT FALSE,
	working_dir TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	archived_at TIMESTAMP,
	metadata_json TEXT DEFAULT '{}'
);
CREATE TABLE thread_messages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	thread_id TEXT NOT NULL REFERENCES threads(id),
	session_id TEXT,
	message_id TEXT,
	role TEXT NOT NULL,
	content_json TEXT NOT NULL,
	created_timestamp INTEGER NOT NULL,
	metadata_json TEXT DEFAULT '{}'
);
CREATE INDEX idx_thread_messages_thread ON thread_messages(thread_id);
CREATE INDEX idx_thread_messages_message_id ON thread_messages(message_id);
CREATE TABLE provider_inventory_entries (
	inventory_key TEXT PRIMARY KEY,
	provider_id TEXT NOT NULL,
	provider_family TEXT NOT NULL,
	last_updated_at TEXT,
	last_refresh_attempt_at TEXT,
	last_refresh_error TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE provider_inventory_models (
	inventory_key TEXT NOT NULL REFERENCES provider_inventory_entries(inventory_key) ON DELETE CASCADE,
	ordinal INTEGER NOT NULL,
	model_id TEXT NOT NULL,
	name TEXT NOT NULL,
	family TEXT,
	context_limit INTEGER,
	reasoning BOOLEAN,
	recommended BOOLEAN,
	PRIMARY KEY (inventory_key, ordinal)
);
CREATE INDEX idx_provider_inventory_provider_id ON provider_inventory_entries(provider_id);
INSERT INTO sessions (
	id,
	name,
	description,
	session_type,
	working_dir,
	updated_at,
	extension_data,
	total_tokens,
	input_tokens,
	output_tokens,
	accumulated_total_tokens,
	accumulated_input_tokens,
	accumulated_output_tokens,
	provider_name,
	model_config_json,
	goose_mode
) VALUES
	(
		'20260419_1',
		'Old session name',
		'Old session description',
		'user',
		'/workspace/old',
		'2026-04-19 12:00:00',
		'{}',
		15,
		10,
		5,
		150,
		100,
		50,
		'openai',
		'{"model_name":"gpt-4o"}',
		'auto'
	),
	(
		'20260420_1',
		'Generated chat title',
		'',
		'user',
		'/workspace/goose',
		'2026-04-20 12:00:00',
		'{"enabled_extensions.v0":{"extensions":[]}}',
		21,
		13,
		8,
		210,
		130,
		80,
		'openai',
		'{"model_name":"gpt-4o"}',
		'auto'
	);
INSERT INTO messages (
	message_id,
	session_id,
	role,
	content_json,
	created_timestamp,
	timestamp,
	metadata_json
) VALUES
	(
		'msg_old',
		'20260420_1',
		'user',
		'[{"type":"text","text":"Earlier prompt"}]',
		1777540000,
		'2026-04-20 11:59:00',
		'{"userVisible":true,"agentVisible":true}'
	),
	(
		'msg_tool',
		'20260420_1',
		'user',
		'[{"type":"toolResponse","toolResult":{"value":{"content":[{"type":"text","text":"ignore tool output"}]}}}]',
		1777540003,
		'2026-04-20 12:00:01',
		'{"userVisible":true,"agentVisible":true}'
	),
	(
		'msg_prompt',
		'20260420_1',
		'user',
		'[{"type":"text","text":"Fix the Goose parser using the real schema"}]',
		1777540002,
		'2026-04-20 12:00:00',
		'{"userVisible":true,"agentVisible":true}'
	);
`)
	require.NoError(t, err)
}
