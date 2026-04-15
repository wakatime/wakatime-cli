package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestRemoveNoopHeartbeatsReplacesAppHeartbeatsAndMergesDuplicates(t *testing.T) {
	got := replaceAppHeartbeats([]heartbeat.Heartbeat{
		{
			Entity:         "Codex session",
			EntityType:     heartbeat.AppType,
			AISession:      "session-1",
			AILineChanges:  heartbeat.PointerTo(2),
			AIInputTokens:  10,
			AIOutputTokens: 4,
			AIPromptLength: 20,
			Category:       "ai coding",
			Time:           100,
			UserAgent:      "Codex/1.0.0",
		},
		{
			Entity:         "/tmp/main.go",
			EntityType:     heartbeat.FileType,
			AISession:      "session-1",
			AILineChanges:  heartbeat.PointerTo(3),
			AIInputTokens:  7,
			AIOutputTokens: 5,
			Category:       "ai coding",
			Time:           130,
			UserAgent:      "Codex/1.0.0",
		},
		{
			Entity:         "Codex session",
			EntityType:     heartbeat.AppType,
			AISession:      "session-1",
			AILineChanges:  heartbeat.PointerTo(6),
			AIInputTokens:  1,
			AIOutputTokens: 2,
			Category:       "ai coding",
			Time:           150,
			UserAgent:      "Codex/1.0.0",
		},
		{
			Entity:         "/tmp/main.go",
			EntityType:     heartbeat.FileType,
			AISession:      "session-1",
			AIInputTokens:  8,
			AIOutputTokens: 9,
			Category:       "ai coding",
			Time:           191,
			UserAgent:      "Codex/1.0.0",
		},
	})

	require.Len(t, got, 3)

	assert.Equal(t, "/tmp/main.go", got[0].Entity)
	assert.Equal(t, heartbeat.FileType, got[0].EntityType)
	require.NotNil(t, got[0].AILineChanges)
	assert.Equal(t, 2, *got[0].AILineChanges)
	assert.Nil(t, got[0].HumanLineChanges)
	assert.Equal(t, 20, got[0].AIPromptLength)
	assert.Equal(t, int64(10), got[0].AIInputTokens)
	assert.Equal(t, int64(4), got[0].AIOutputTokens)
	assert.Equal(t, float64(100), got[0].Time)

	assert.Equal(t, "/tmp/main.go", got[1].Entity)
	assert.Equal(t, heartbeat.FileType, got[1].EntityType)
	require.NotNil(t, got[1].AILineChanges)
	assert.Equal(t, 3, *got[1].AILineChanges)
	assert.Equal(t, int64(7), got[1].AIInputTokens)
	assert.Equal(t, int64(5), got[1].AIOutputTokens)
	assert.Equal(t, float64(130), got[1].Time)

	assert.Equal(t, "/tmp/main.go", got[2].Entity)
	assert.Equal(t, heartbeat.FileType, got[2].EntityType)
	require.NotNil(t, got[2].AILineChanges)
	assert.Equal(t, 6, *got[2].AILineChanges)
	assert.Equal(t, int64(9), got[2].AIInputTokens)
	assert.Equal(t, int64(11), got[2].AIOutputTokens)
	assert.Equal(t, float64(191), got[2].Time)
}

func TestRemoveNoopHeartbeatsKeepsDifferentIdentityHeartbeats(t *testing.T) {
	got := replaceAppHeartbeats([]heartbeat.Heartbeat{
		{
			Entity:     "/tmp/main.go",
			EntityType: heartbeat.FileType,
			AISession:  "session-1",
			Category:   "ai coding",
			Time:       100,
			UserAgent:  "Codex/1.0.0",
		},
		{
			Entity:     "/tmp/main.go",
			EntityType: heartbeat.FileType,
			AISession:  "session-1",
			Category:   "debugging",
			Time:       130,
			UserAgent:  "Codex/1.0.0",
		},
	})

	require.Len(t, got, 2)
	assert.Equal(t, "ai coding", got[0].Category)
	assert.Equal(t, "debugging", got[1].Category)
}
