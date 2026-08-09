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

	_, err = db.Exec(`CREATE TABLE conversations (
		conversation_id TEXT PRIMARY KEY NOT NULL,
		title TEXT,
		workspace_id BIGINT NOT NULL,
		context TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP,
		metrics TEXT
	)`)
	require.NoError(t, err)
	_, err = db.Exec(
		`INSERT INTO conversations VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"forge-1",
		"Ship the feature",
		12345,
		`{
			"conversation_id":"forge-1",
			"messages":[
				{"message":{"text":{"role":"User","content":"Ship the feature","model":"gpt-5.2"}}},
				{
					"message":{"text":{"role":"Assistant","content":"First response","model":"gpt-5.2"}},
					"usage":{
						"prompt_tokens":{"actual":100},
						"completion_tokens":{"actual":25},
						"total_tokens":{"actual":125},
						"cached_tokens":{"actual":40}
					}
				},
				{"message":{"text":{"role":"User","content":"Make it smaller","model":"gpt-5.2"}}},
				{
					"message":{"text":{"role":"Assistant","content":"Second response","model":"gpt-5.2"}},
					"usage":{
						"prompt_tokens":{"Actual":30},
						"completion_tokens":{"Actual":10},
						"total_tokens":{"Actual":40},
						"cached_tokens":{"Actual":5}
					}
				}
			]
		}`,
		"2026-08-06 11:00:00",
		"2026-08-06 12:00:00",
		`{"started_at":"2026-08-06T11:00:00Z"}`,
	)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	got, err := (Forge{After: time.Date(2026, 8, 6, 11, 0, 0, 0, time.UTC)}).Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 4)

	assert.Equal(t, "Forge forge-1", got[1].Entity)
	assert.Equal(t, int64(60), got[1].AIInputTokens)
	assert.Equal(t, int64(40), got[1].AICachedInputTokens)
	assert.Equal(t, int64(25), got[1].AIOutputTokens)
	assert.Equal(t, len([]rune("Ship the feature")), got[0].AIPromptLength)
	assert.Equal(t, float64(time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC).Unix()), got[1].Time)

	// Forge stores independent Usage values on assistant messages. A later,
	// smaller response must not be treated as a cumulative-counter reset.
	assert.Equal(t, int64(25), got[3].AIInputTokens)
	assert.Equal(t, int64(5), got[3].AICachedInputTokens)
	assert.Equal(t, int64(10), got[3].AIOutputTokens)
	assert.Equal(t, len([]rune("Make it smaller")), got[2].AIPromptLength)
}
