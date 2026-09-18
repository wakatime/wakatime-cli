//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

// TestZCodeParseBeyondSQLiteRowLimit keeps the upstream regression intent of
// PR #1564, rebuilt on the real ZCode schema: with more than 5,000 in-window
// messages and parts, nothing may be silently truncated — including the oldest
// records — and telemetry tables must never duplicate token heartbeats.
func TestZCodeParseBeyondSQLiteRowLimit(t *testing.T) {
	const baseTimestamp = int64(1770000000000)

	ctx := context.Background()
	db := zCodeTestDB(t)

	// Real-named telemetry tables carry token-shaped columns; the bespoke
	// parser must ignore them entirely instead of double counting.
	_, err := db.Exec(`
CREATE TABLE model_usage (
	id TEXT PRIMARY KEY, session_id TEXT, model_id TEXT,
	started_at INTEGER, completed_at INTEGER,
	input_tokens INTEGER, output_tokens INTEGER,
	cache_creation_input_tokens INTEGER, cache_read_input_tokens INTEGER
);
CREATE TABLE turn_usage (
	id TEXT PRIMARY KEY, session_id TEXT, status TEXT,
	input_tokens INTEGER, output_tokens INTEGER,
	cache_read_input_tokens INTEGER
);
CREATE TABLE tool_usage (
	id TEXT PRIMARY KEY, session_id TEXT, tool_name TEXT, status TEXT
);
INSERT INTO model_usage VALUES ('mu-1', 'session-1', 'GLM-5.3', 1, 2, 100, 40, 10, 40);
INSERT INTO turn_usage VALUES ('tu-1', 'session-1', 'completed', 100, 40, 40);
INSERT INTO tool_usage VALUES ('toolu-1', 'session-1', 'edit', 'completed');
`)
	require.NoError(t, err)

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
	zCodeInsertPart(
		t,
		db,
		"part-edit-oldest",
		"message-oldest",
		baseTimestamp+2,
		`{
			"type": "tool",
			"tool": "Edit",
			"state": {
				"status": "completed",
				"input": {"file_path": "/workspace/target.go", "old_string": "a", "new_string": "a\nb\nc"},
				"metadata": {"display": {"additions": 3, "deletions": 2}}
			}
		}`,
	)

	_, err = db.Exec(`
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

	_, err = db.Exec(`
WITH RECURSIVE seq(x) AS (
	SELECT 2
	UNION ALL
	SELECT x + 1 FROM seq WHERE x < 5001
)
INSERT INTO part (id, message_id, session_id, time_created, data)
SELECT
	printf('part-%05d', x),
	'message-oldest',
	'session-1',
	1770000000000 + x,
	'{"type":"text"}'
FROM seq;
`)
	require.NoError(t, err)

	got, err := (ai.ZCode{After: time.UnixMilli(baseTimestamp)}).Parse(ctx)
	require.NoError(t, err)

	// 5,001 in-window messages -> 5,001 app heartbeats; the single Edit part
	// sitting at the oldest end of 5,001 parts -> exactly one file heartbeat.
	require.Len(t, got, 5002)

	expectedTime := func(millis int64) float64 {
		return float64(millis/1000) + float64((millis%1000)*int64(time.Millisecond))/float64(time.Second)
	}

	for _, h := range got {
		if h.EntityType == heartbeat.AppType {
			assert.Equal(t, "ZCode session-1", h.Entity)
		}
	}

	assert.Equal(t, expectedTime(baseTimestamp+1), got[0].Time)
	assert.EqualValues(t, 6, got[0].AIInputTokens)
	assert.EqualValues(t, 4, got[0].AICachedInputTokens)
	assert.EqualValues(t, 5, got[0].AIOutputTokens)

	assert.Equal(t, heartbeat.FileType, got[1].EntityType)
	assert.Equal(t, "/workspace/target.go", got[1].Entity)
	require.NotNil(t, got[1].AILineChanges)
	assert.Equal(t, 1, *got[1].AILineChanges) // net changes: 3 additions - 2 deletions

	assert.Equal(t, expectedTime(baseTimestamp+5001), got[5001].Time)
}
