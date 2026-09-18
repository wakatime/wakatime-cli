//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	_ "modernc.org/sqlite"
)

func TestZCodeParse(t *testing.T) {
	ctx := context.Background()
	const baseTimestamp = int64(1770000000000)

	db := zCodeTestDB(t)

	zCodeInsertMessage(
		t,
		db,
		"message-old",
		baseTimestamp+1000,
		`{
			"role": "assistant",
			"modelID": "GLM-5.3",
			"tokens": {
				"input": 50,
				"output": 5,
				"reasoning": 0,
				"cache": {"read": 20, "write": 0}
			}
		}`,
	)
	zCodeInsertPart(
		t,
		db,
		"part-old",
		"message-old",
		baseTimestamp+1001,
		`{
			"type": "tool",
			"tool": "Edit",
			"state": {
				"status": "completed",
				"input": {
					"file_path": "/workspace/old.go",
					"old_string": "a",
					"new_string": "a\nb"
				},
				"metadata": {}
			}
		}`,
	)

	zCodeInsertMessage(t, db, "message-user", baseTimestamp+2000, `{"role":"user"}`)
	zCodeInsertPart(
		t,
		db,
		"part-user",
		"message-user",
		baseTimestamp+2001,
		`{"type":"text","text":"Refactor this file"}`,
	)

	zCodeInsertMessage(
		t,
		db,
		"message-usage",
		baseTimestamp+2500,
		`{
			"role": "assistant",
			"modelID": "GLM-5.3",
			"path": {"cwd": "/workspace", "root": "/workspace"},
			"tokens": {
				"input": 100,
				"output": 4,
				"reasoning": 2,
				"cache": {"read": 40, "write": 10}
			}
		}`,
	)
	zCodeInsertPart(
		t,
		db,
		"part-usage",
		"message-usage",
		baseTimestamp+2501,
		`{"type":"text","text":"Done"}`,
	)

	rows := []struct {
		messageID string
		partID    string
		timestamp int64
		data      string
	}{
		{
			"message-edit-display",
			"part-edit-display",
			baseTimestamp + 3000,
			`{
				"type": "tool",
				"tool": "Edit",
				"state": {
					"status": "completed",
					"input": {
						"file_path": "/workspace/edit.go",
						"old_string": "a",
						"new_string": "b"
					},
					"metadata": {
						"display": {
							"additions": 5,
							"deletions": 2,
							"filePath": "/workspace/edit.go"
						},
						"readFileState": {"path": "/workspace/edit.go"}
					}
				}
			}`,
		},
		{
			"message-edit-zero",
			"part-edit-zero",
			baseTimestamp + 4000,
			`{
				"type": "tool",
				"tool": "Edit",
				"state": {
					"status": "completed",
					"input": {
						"file_path": "/workspace/zero.go",
						"old_string": "a",
						"new_string": "b"
					},
					"metadata": {
						"display": {"additions": 1, "deletions": 1},
						"readFileState": {"path": "/workspace/zero.go"}
					}
				}
			}`,
		},
		{
			"message-edit-fallback",
			"part-edit-fallback",
			baseTimestamp + 5000,
			`{
				"type": "tool",
				"tool": "Edit",
				"state": {
					"status": "completed",
					"input": {
						"file_path": "/workspace/fallback.go",
						"old_string": "a\nb",
						"new_string": "a\nb\nc"
					},
					"metadata": {}
				}
			}`,
		},
		{
			"message-edit-error",
			"part-edit-error",
			baseTimestamp + 6000,
			`{
				"type": "tool",
				"tool": "Edit",
				"state": {
					"status": "error",
					"input": {
						"file_path": "/workspace/error.go",
						"old_string": "a",
						"new_string": "a\nb"
					},
					"metadata": {"display": {"additions": 1, "deletions": 0}}
				}
			}`,
		},
		{
			"message-write-new",
			"part-write-new",
			baseTimestamp + 7000,
			`{
				"type": "tool",
				"tool": "Write",
				"state": {
					"status": "completed",
					"input": {
						"file_path": "/workspace/new.go",
						"content": "a\nb\nc"
					},
					"output": "File created successfully at: /workspace/new.go",
					"metadata": {"readFileState": {"path": "/workspace/new.go"}}
				}
			}`,
		},
		{
			"message-write-existing",
			"part-write-existing",
			baseTimestamp + 8000,
			`{
				"type": "tool",
				"tool": "Write",
				"state": {
					"status": "completed",
					"input": {
						"file_path": "/workspace/existing.go",
						"content": "a\nb\nc\n"
					},
					"output": "The file /workspace/existing.go has been updated successfully.",
					"metadata": {
						"display": {"additions": 2, "deletions": 1},
						"readFileState": {"path": "/workspace/existing.go"}
					}
				}
			}`,
		},
		{
			"message-write-unknown",
			"part-write-unknown",
			baseTimestamp + 9000,
			`{
				"type": "tool",
				"tool": "Write",
				"state": {
					"status": "completed",
					"input": {
						"file_path": "/workspace/unknown.go",
						"content": "a\nb\nc\n"
					},
					"output": "The file /workspace/unknown.go has been updated successfully.",
					"metadata": {"readFileState": {"path": "/workspace/unknown.go"}}
				}
			}`,
		},
	}

	for _, row := range rows {
		zCodeInsertMessage(
			t,
			db,
			row.messageID,
			row.timestamp,
			`{"role":"assistant","modelID":"GLM-5.3"}`,
		)
		zCodeInsertPart(t, db, row.partID, row.messageID, row.timestamp+1, row.data)
	}

	parser := ai.ZCode{
		After: time.UnixMilli(baseTimestamp + 2000),
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 14)

	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, len([]rune("Refactor this file")), got[0].AIPromptLength)
	assert.Zero(t, got[0].AIInputTokens)
	assert.Zero(t, got[0].AICachedInputTokens)
	assert.Zero(t, got[0].AIOutputTokens)
	assert.Equal(t, "/workspace", got[0].ProjectPathOverride)
	assert.Contains(t, got[0].UserAgent, "zcode-cli/0.16.5")

	assert.Equal(t, heartbeat.AppType, got[1].EntityType)
	assert.EqualValues(t, 60, got[1].AIInputTokens)
	assert.EqualValues(t, 40, got[1].AICachedInputTokens)
	assert.EqualValues(t, 6, got[1].AIOutputTokens)
	assert.Equal(t, "/workspace", got[1].ProjectPathOverride)
	assert.Contains(t, got[1].UserAgent, "GLM/5.3")
	assert.Contains(t, got[1].UserAgent, "zcode-cli/0.16.5")

	var fileHeartbeats ai.Heartbeats
	for _, h := range got {
		if h.EntityType == heartbeat.FileType {
			fileHeartbeats = append(fileHeartbeats, h)
		}
	}
	require.Len(t, fileHeartbeats, 5)

	expected := []struct {
		entity      string
		lineChanges int
	}{
		{"/workspace/edit.go", 3},
		{"/workspace/zero.go", 0},
		{"/workspace/fallback.go", 1},
		{"/workspace/new.go", 3},
		{"/workspace/existing.go", 1},
	}

	for i, want := range expected {
		assert.Equal(t, want.entity, fileHeartbeats[i].Entity)
		assert.Equal(t, "session-1", fileHeartbeats[i].AISession)
		assert.Empty(t, fileHeartbeats[i].ProjectPathOverride)
		assert.Contains(t, fileHeartbeats[i].UserAgent, "GLM/5.3")
		assert.Contains(t, fileHeartbeats[i].UserAgent, "zcode-cli/0.16.5")
		require.NotNil(t, fileHeartbeats[i].AILineChanges)
		assert.Equal(t, want.lineChanges, *fileHeartbeats[i].AILineChanges)
		require.NotNil(t, fileHeartbeats[i].IsWrite)
		assert.True(t, *fileHeartbeats[i].IsWrite)
	}

	cutoffParser := ai.ZCode{
		After: time.UnixMilli(baseTimestamp + 8000),
	}

	atCutoff, err := cutoffParser.Parse(ctx)
	require.NoError(t, err)

	var cutoffFiles ai.Heartbeats
	for _, h := range atCutoff {
		if h.EntityType == heartbeat.FileType {
			cutoffFiles = append(cutoffFiles, h)
		}
	}
	require.Len(t, cutoffFiles, 1)
	assert.Equal(t, "/workspace/existing.go", cutoffFiles[0].Entity)
}

func TestZCodeParse_MessageHeartbeatAfter5000Rows(t *testing.T) {
	const baseTimestamp = int64(1770000000000)

	ctx := context.Background()
	db := zCodeTestDB(t)

	// The oldest in-window message carries distinctive tokens; more than 5,000
	// newer messages in the same window must not push it out of the parse.
	zCodeInsertMessage(
		t,
		db,
		"message-oldest",
		baseTimestamp+1,
		`{
			"role": "assistant",
			"modelID": "GLM-5.3",
			"tokens": {
				"input": 10, "output": 4, "reasoning": 1,
				"cache": {"read": 4, "write": 0}
			}
		}`,
	)

	_, err := db.Exec(`
WITH RECURSIVE seq(x) AS (
	SELECT 2
	UNION ALL
	SELECT x + 1 FROM seq WHERE x < 5001
)
INSERT INTO message (id, session_id, time_created, data)
SELECT
	printf('message-%05d', x),
	'session-1',
	1770000000000 + x,
	'{"role":"assistant","modelID":"GLM-5.3","tokens":{
		"input":1,"output":1,"reasoning":0,"cache":{"read":0,"write":0}
	}}'
FROM seq;
`)
	require.NoError(t, err)

	parser := ai.ZCode{
		After: time.UnixMilli(baseTimestamp),
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 5001)

	expectedTime := func(millis int64) float64 {
		return float64(millis/1000) + float64((millis%1000)*int64(time.Millisecond))/float64(time.Second)
	}

	for _, h := range got {
		assert.Equal(t, "ZCode session-1", h.Entity)
	}

	assert.Equal(t, expectedTime(baseTimestamp+1), got[0].Time)
	assert.EqualValues(t, 6, got[0].AIInputTokens)
	assert.EqualValues(t, 4, got[0].AICachedInputTokens)
	assert.EqualValues(t, 5, got[0].AIOutputTokens)

	assert.Equal(t, expectedTime(baseTimestamp+5001), got[5000].Time)
}

func TestZCodeParse_SQLiteWALModifiedAfter(t *testing.T) {
	ctx := context.Background()
	db := zCodeTestDB(t)

	var journalMode string
	require.NoError(t, db.QueryRow(`PRAGMA journal_mode = WAL;`).Scan(&journalMode))
	require.Equal(t, "wal", journalMode)
	_, err := db.Exec(`PRAGMA wal_checkpoint(TRUNCATE);`)
	require.NoError(t, err)

	cutoff := time.Now().Add(-time.Minute).Truncate(time.Millisecond)
	dbPath := filepath.Join(os.Getenv("HOME"), ".zcode", "cli", "db", "db.sqlite")
	oldTime := cutoff.Add(-time.Hour)
	require.NoError(t, os.Chtimes(dbPath, oldTime, oldTime))

	zCodeInsertMessage(
		t,
		db,
		"message-wal",
		cutoff.Add(time.Second).UnixMilli(),
		`{
			"role":"assistant",
			"modelID":"GLM-5.3-Flash",
			"tokens":{
				"input":10,
				"output":4,
				"reasoning":0,
				"cache":{"read":2,"write":0}
			}
		}`,
	)

	dbInfo, err := os.Stat(dbPath)
	require.NoError(t, err)
	assert.True(t, dbInfo.ModTime().Before(cutoff))

	walInfo, err := os.Stat(dbPath + "-wal")
	require.NoError(t, err)
	assert.False(t, walInfo.ModTime().Before(cutoff))

	parser := ai.ZCode{After: cutoff}
	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.EqualValues(t, 8, got[0].AIInputTokens)
	assert.EqualValues(t, 2, got[0].AICachedInputTokens)
	assert.EqualValues(t, 4, got[0].AIOutputTokens)
	assert.Contains(t, got[0].UserAgent, "GLM/5.3-Flash")
}

func TestZCodeParse_FileHeartbeatAfter5000Rows(t *testing.T) {
	ctx := context.Background()
	db := zCodeTestDB(t)

	zCodeInsertMessage(t, db, "message-1", 6000, `{"role":"assistant","modelID":"GLM-5.3"}`)

	// The only Edit part sits at the oldest end of the part table; more than
	// 5,000 newer filler parts must not push it out of the parsed window.
	zCodeInsertPart(
		t,
		db,
		"part-edit-oldest",
		"message-1",
		5001,
		`{
			"type": "tool",
			"tool": "Edit",
			"state": {
				"status": "completed",
				"input": {"file_path": "/workspace/target.go", "old_string": "a", "new_string": "a\nb"},
				"metadata": {"display": {"additions": 1, "deletions": 0}}
			}
		}`,
	)

	_, err := db.Exec(`
WITH RECURSIVE seq(x) AS (
	SELECT 1
	UNION ALL
	SELECT x + 1 FROM seq WHERE x < 5000
)
INSERT INTO part (id, message_id, session_id, time_created, data)
SELECT printf('part-%05d', x), 'message-1', 'session-1', 10000 + x, '{"type":"text"}' FROM seq;
`)
	require.NoError(t, err)

	parser := ai.ZCode{}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)

	var fileHeartbeats ai.Heartbeats
	for _, h := range got {
		if h.EntityType == heartbeat.FileType {
			fileHeartbeats = append(fileHeartbeats, h)
		}
	}
	require.Len(t, fileHeartbeats, 1)
	assert.Equal(t, "/workspace/target.go", fileHeartbeats[0].Entity)
	assert.Empty(t, fileHeartbeats[0].ProjectPathOverride)
	assert.Contains(t, fileHeartbeats[0].UserAgent, "GLM/5.3")
	assert.Contains(t, fileHeartbeats[0].UserAgent, "zcode-cli/0.16.5")
	require.NotNil(t, fileHeartbeats[0].AILineChanges)
	assert.Equal(t, 1, *fileHeartbeats[0].AILineChanges)
}

func TestZCodeParse_MissingSessionMetadata(t *testing.T) {
	ctx := context.Background()
	db := zCodeTestDB(t)

	// A message referencing a session absent from the session table still
	// produces a heartbeat, falling back to the message's session ID.
	_, err := db.Exec(
		`INSERT INTO message (id, session_id, time_created, data) VALUES (?, 'session-unknown', ?, ?)`,
		"message-orphan",
		7000,
		`{"role":"user"}`,
	)
	require.NoError(t, err)
	_, err = db.Exec(
		`INSERT INTO part (id, message_id, session_id, time_created, data) VALUES (?, ?, 'session-unknown', ?, ?)`,
		"part-orphan",
		"message-orphan",
		7001,
		`{"type":"text","text":"Hello there"}`,
	)
	require.NoError(t, err)

	parser := ai.ZCode{}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "ZCode session-unknown", got[0].Entity)
	assert.EqualValues(t, len([]rune("Hello there")), got[0].AIPromptLength)
	assert.Empty(t, got[0].ProjectPathOverride)
	assert.Contains(t, got[0].UserAgent, "zcode-cli/unknown")
}

func TestZCodeParse_JSONTimeCutoffMismatch(t *testing.T) {
	const baseTimestamp = int64(1770000000000)

	ctx := context.Background()
	db := zCodeTestDB(t)

	// SQL column time is inside the window but the JSON time.created is before
	// the cutoff: the message must be skipped without emitting a heartbeat,
	// while the next message still parses normally.
	zCodeInsertMessage(
		t,
		db,
		"message-json-old",
		baseTimestamp+3000,
		`{"role":"user","time":{"created":`+fmt.Sprint(baseTimestamp+1000)+`}}`,
	)
	zCodeInsertPart(
		t,
		db,
		"part-json-old",
		"message-json-old",
		baseTimestamp+3001,
		`{"type":"text","text":"should be skipped"}`,
	)

	zCodeInsertMessage(t, db, "message-current", baseTimestamp+3000+1, `{"role":"user"}`)
	zCodeInsertPart(
		t,
		db,
		"part-current",
		"message-current",
		baseTimestamp+3002,
		`{"type":"text","text":"kept"}`,
	)

	parser := ai.ZCode{After: time.UnixMilli(baseTimestamp + 2000)}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, "ZCode session-1", got[0].Entity)
	assert.EqualValues(t, len([]rune("kept")), got[0].AIPromptLength)
}

func TestZCodeParse_PerRequestTokenUsage(t *testing.T) {
	const baseTimestamp = int64(1770000000000)

	ctx := context.Background()
	db := zCodeTestDB(t)

	// Two assistant messages whose per-request usage drops sharply. Values are
	// per request, not session counters, so the second message must be
	// reported as-is instead of being treated as a counter reset or differenced
	// away against the previous message.
	zCodeInsertMessage(
		t,
		db,
		"message-first",
		baseTimestamp+1000,
		`{
			"role": "assistant",
			"modelID": "GLM-5.3",
			"tokens": {
				"input": 100, "output": 50, "reasoning": 5,
				"cache": {"read": 80, "write": 10}
			}
		}`,
	)
	zCodeInsertMessage(
		t,
		db,
		"message-second",
		baseTimestamp+2000,
		`{
			"role": "assistant",
			"modelID": "GLM-5.3",
			"tokens": {
				"input": 10, "output": 4, "reasoning": 1,
				"cache": {"read": 2, "write": 1}
			}
		}`,
	)

	full, err := (ai.ZCode{}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, full, 2)

	assert.EqualValues(t, 20, full[0].AIInputTokens)
	assert.EqualValues(t, 80, full[0].AICachedInputTokens)
	assert.EqualValues(t, 55, full[0].AIOutputTokens)

	assert.EqualValues(t, 8, full[1].AIInputTokens)
	assert.EqualValues(t, 2, full[1].AICachedInputTokens)
	assert.EqualValues(t, 5, full[1].AIOutputTokens)

	// A window that starts after the first message needs no baseline: the
	// second message still reports its own request usage.
	windowed, err := (ai.ZCode{After: time.UnixMilli(baseTimestamp + 1500)}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, windowed, 1)

	assert.EqualValues(t, 8, windowed[0].AIInputTokens)
	assert.EqualValues(t, 2, windowed[0].AICachedInputTokens)
	assert.EqualValues(t, 5, windowed[0].AIOutputTokens)
}

func zCodeTestDB(t *testing.T) *sql.DB {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbPath := filepath.Join(home, ".zcode", "cli", "db", "db.sqlite")
	require.NoError(t, os.MkdirAll(filepath.Dir(dbPath), 0o700))

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`
CREATE TABLE session (
	id TEXT PRIMARY KEY,
	directory TEXT NOT NULL,
	version TEXT
);
CREATE TABLE message (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL,
	time_created INTEGER NOT NULL,
	data TEXT NOT NULL
);
CREATE TABLE part (
	id TEXT PRIMARY KEY,
	message_id TEXT NOT NULL,
	session_id TEXT NOT NULL,
	time_created INTEGER NOT NULL,
	data TEXT NOT NULL
);
INSERT INTO session (id, directory, version) VALUES ('session-1', '/workspace', '0.16.5');
`)
	require.NoError(t, err)

	return db
}

func zCodeInsertMessage(t *testing.T, db *sql.DB, id string, timestamp int64, data string) {
	t.Helper()

	_, err := db.Exec(
		`INSERT INTO message (id, session_id, time_created, data) VALUES (?, 'session-1', ?, ?)`,
		id,
		timestamp,
		data,
	)
	require.NoError(t, err)
}

func zCodeInsertPart(t *testing.T, db *sql.DB, id string, messageID string, timestamp int64, data string) {
	t.Helper()

	_, err := db.Exec(
		`INSERT INTO part (id, message_id, session_id, time_created, data) VALUES (?, ?, 'session-1', ?, ?)`,
		id,
		messageID,
		timestamp,
		data,
	)
	require.NoError(t, err)
}
