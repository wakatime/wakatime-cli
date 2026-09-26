package ai

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
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
	assert.Equal(t, filepath.Join(home, "wakatime-internal.cfg-ai-claude-transcripts.json"), path)

	// distinct internal config files must not share transcript state
	for _, name := range []string{"work.cfg", "work.ini", "work"} {
		v.Set("internal-config", filepath.Join(home, "profiles", name))

		path, err = claudeTranscriptStatePath(ctx, v)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(home, "profiles", name+"-ai-claude-transcripts.json"), path)
	}
}

func TestClaudeNextCheckpoint(t *testing.T) {
	noon := time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC)

	hb := func(t time.Time) heartbeat.Heartbeat {
		return heartbeat.Heartbeat{Time: heartbeatTimestamp(t)}
	}

	// no heartbeats: previous checkpoint is kept
	previous := claudeTranscriptCheckpoint{Time: noon, Count: 2}
	assert.Equal(t, previous, claudeNextCheckpoint(nil, previous))

	// newer heartbeats restart the count
	next := claudeNextCheckpoint(Heartbeats{hb(noon.Add(time.Hour)), hb(noon.Add(time.Hour))}, previous)
	assert.Equal(t, claudeTranscriptCheckpoint{Time: noon.Add(time.Hour), Count: 2}, next)

	// heartbeats at the checkpoint timestamp add to the count
	next = claudeNextCheckpoint(Heartbeats{hb(noon)}, previous)
	assert.Equal(t, claudeTranscriptCheckpoint{Time: noon, Count: 3}, next)
}

func TestClaudeTranscriptCheckpoint_Unchanged(t *testing.T) {
	path := t.TempDir() + "/session.jsonl"
	require.NoError(t, os.WriteFile(path, []byte("{}\n"), 0o644))

	info, err := os.Stat(path)
	require.NoError(t, err)

	cp := claudeTranscriptCheckpoint{Size: info.Size(), ModTime: info.ModTime()}
	assert.True(t, cp.unchanged(info))

	// zero mtime never matches, so legacy entries without stat info are re-read
	assert.False(t, claudeTranscriptCheckpoint{Size: info.Size()}.unchanged(info))
	assert.False(t, claudeTranscriptCheckpoint{Size: info.Size() + 1, ModTime: info.ModTime()}.unchanged(info))
}

func TestClaudeTranscriptState_Prune(t *testing.T) {
	// nothing exists on disk: pruning decides from the walk alone
	root := filepath.Join(t.TempDir(), "projects")
	visited := filepath.Join(root, "p", "visited.jsonl")
	deleted := filepath.Join(root, "p", "deleted.jsonl")
	sibling := filepath.Join(root+"-other", "p", "x.jsonl")
	unwalked := filepath.Join(t.TempDir(), "stopped-distro", "projects", "p", "x.jsonl")

	state := claudeTranscriptState{visited: {}, deleted: {}, sibling: {}, unwalked: {}}
	seen := map[string]struct{}{visited: {}}

	assert.True(t, state.prune([]string{root}, seen))
	assert.Len(t, state, 3)
	assert.NotContains(t, state, deleted)
	assert.Contains(t, state, visited)
	assert.Contains(t, state, sibling)
	assert.Contains(t, state, unwalked)

	assert.False(t, state.prune([]string{root}, seen))
}

func TestClaudeSaveTranscriptState_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	state := claudeTranscriptState{}

	// no state path configured: nothing to do
	require.NoError(t, Claude{}.saveTranscriptState(state))

	// parent of the state dir is a regular file: MkdirAll fails
	blocked := filepath.Join(tmpDir, "blocked")
	require.NoError(t, os.WriteFile(blocked, nil, 0o644))
	assert.Error(t, Claude{StateFilePath: filepath.Join(blocked, "dir", "state.json")}.saveTranscriptState(state))

	// state path is an existing directory: the final rename fails
	asDir := filepath.Join(tmpDir, "as-dir", "state.json")
	require.NoError(t, os.MkdirAll(asDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(asDir, "keep"), nil, 0o644))
	assert.Error(t, Claude{StateFilePath: asDir}.saveTranscriptState(state))
	assert.DirExists(t, asDir)

	// state dir not writable: creating the temp file fails. Windows ignores
	// unix permission bits on directories, so this case is unix-only.
	if runtime.GOOS == "windows" {
		t.Skip("Skipping because OS is windows.")
	}

	readonly := filepath.Join(tmpDir, "readonly")
	require.NoError(t, os.Mkdir(readonly, 0o500))
	assert.Error(t, Claude{StateFilePath: filepath.Join(readonly, "state.json")}.saveTranscriptState(state))
	assert.NoFileExists(t, filepath.Join(readonly, "state.json"))
}

func TestClaudePersistTranscriptState_RemovesStaleState(t *testing.T) {
	ctx := context.Background()

	// an empty directory at the state path: saving fails, removing succeeds
	statePath := filepath.Join(t.TempDir(), "state.json")
	require.NoError(t, os.Mkdir(statePath, 0o755))

	require.NoError(t, Claude{StateFilePath: statePath}.persistTranscriptState(ctx, claudeTranscriptState{}))
	assert.NoDirExists(t, statePath)

	// a non-empty directory cannot be removed either: the heartbeats are withheld
	require.NoError(t, os.MkdirAll(filepath.Join(statePath, "keep"), 0o755))
	assert.Error(t, Claude{StateFilePath: statePath}.persistTranscriptState(ctx, claudeTranscriptState{}))
}

func TestClaudeLoadTranscriptState_Unreadable(t *testing.T) {
	ctx := context.Background()

	// Windows ignores unix permission bits on files, so this case is unix-only.
	if runtime.GOOS == "windows" {
		t.Skip("Skipping because OS is windows.")
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "state.json")
	require.NoError(t, os.WriteFile(path, []byte("{}"), 0o000))

	state := Claude{StateFilePath: path}.loadTranscriptState(ctx)
	assert.Empty(t, state)
}
