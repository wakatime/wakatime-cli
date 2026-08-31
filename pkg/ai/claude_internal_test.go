package ai

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestClaudeTranscriptStatePath(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("WAKATIME_HOME", home)

	v := viper.New()

	path, err := claudeTranscriptStatePath(ctx, v)
	require.NoError(t, err)
	assert.Equal(t, home+"/wakatime-internal-ai-claude-transcripts.json", path)

	v.Set("internal-config", home+"/profiles/work.cfg")

	path, err = claudeTranscriptStatePath(ctx, v)
	require.NoError(t, err)
	assert.Equal(t, home+"/profiles/work-ai-claude-transcripts.json", path)
}

func TestClaudeNextCheckpoint(t *testing.T) {
	noon := time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC)

	hb := func(t time.Time) heartbeat.Heartbeat {
		return heartbeat.Heartbeat{Time: heartbeatTimestamp(t)}
	}

	// no heartbeats: previous checkpoint is kept
	previous := claudeTranscriptCheckpoint{Time: noon, Count: 2}
	next, ok := claudeNextCheckpoint(nil, previous, true)
	assert.True(t, ok)
	assert.Equal(t, previous, next)

	_, ok = claudeNextCheckpoint(nil, claudeTranscriptCheckpoint{}, false)
	assert.False(t, ok)

	// newer heartbeats restart the count
	next, ok = claudeNextCheckpoint(Heartbeats{hb(noon.Add(time.Hour)), hb(noon.Add(time.Hour))}, previous, true)
	assert.True(t, ok)
	assert.Equal(t, claudeTranscriptCheckpoint{Time: noon.Add(time.Hour), Count: 2}, next)

	// heartbeats at the checkpoint timestamp add to the count
	next, ok = claudeNextCheckpoint(Heartbeats{hb(noon)}, previous, true)
	assert.True(t, ok)
	assert.Equal(t, claudeTranscriptCheckpoint{Time: noon, Count: 3}, next)
}

func TestClaudeSaveTranscriptState_Errors(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	state := claudeTranscriptState{}

	// no state path configured: nothing to do
	Claude{}.saveTranscriptState(ctx, state)

	// parent of the state dir is a regular file: MkdirAll fails
	blocked := tmpDir + "/blocked"
	require.NoError(t, os.WriteFile(blocked, nil, 0o644))
	Claude{StateFilePath: blocked + "/dir/state.json"}.saveTranscriptState(ctx, state)

	// state dir not writable: creating the temp file fails
	readonly := tmpDir + "/readonly"
	require.NoError(t, os.Mkdir(readonly, 0o500))
	Claude{StateFilePath: readonly + "/state.json"}.saveTranscriptState(ctx, state)

	// state path is an existing directory: the final rename fails
	asDir := tmpDir + "/as-dir"
	require.NoError(t, os.MkdirAll(asDir+"/state.json", 0o755))
	require.NoError(t, os.WriteFile(asDir+"/state.json/keep", nil, 0o644))
	Claude{StateFilePath: asDir + "/state.json"}.saveTranscriptState(ctx, state)

	assert.NoFileExists(t, readonly+"/state.json")
}

func TestClaudeLoadTranscriptState_Unreadable(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	path := tmpDir + "/state.json"
	require.NoError(t, os.WriteFile(path, []byte("{}"), 0o000))

	state := Claude{StateFilePath: path}.loadTranscriptState(ctx)
	assert.Empty(t, state)
}
