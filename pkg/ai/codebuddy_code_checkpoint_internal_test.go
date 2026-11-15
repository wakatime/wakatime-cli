package ai

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func codeBuddyCheckpointRun(t *testing.T, state *parserCheckpoint) (Heartbeats, Heartbeats) {
	t.Helper()

	parser := CodeBuddyCode(ParserConfig{After: state.discoveryAfter(), checkpoint: state})
	raw, err := parser.Parse(t.Context())
	require.NoError(t, err)

	return raw, state.filter(raw)
}

func TestCodeBuddyCheckpointSkipsUnchangedTranscripts(t *testing.T) {
	_, projectDir, sessionsDir := codeBuddyCodeTestDirs(t)
	stamp := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	modified := stamp.Add(24 * time.Hour)
	pathA := filepath.Join(sessionsDir, "a.jsonl")
	pathB := filepath.Join(sessionsDir, "b.jsonl")
	prompt := map[string]any{
		"timestamp": stamp.UnixMilli(), "type": "message", "role": "user",
		"sessionId": "session-a", "cwd": projectDir, "content": "create a file",
	}
	other := map[string]any{
		"timestamp": stamp.UnixMilli(), "type": "message", "role": "user",
		"sessionId": "session-b", "cwd": projectDir, "content": "another session",
	}

	codeBuddyCodeWriteTranscript(t, pathA, codeBuddyCodeJSONLines(t, prompt), modified)
	codeBuddyCodeWriteTranscript(t, pathB, codeBuddyCodeJSONLines(t, other), modified)

	state := &parserCheckpoint{global: stamp.Add(-time.Hour), Sessions: make(map[string]*sessionCheckpoint)}
	_, first := codeBuddyCheckpointRun(t, state)
	require.Len(t, first, 2)
	require.Len(t, state.Sources, 2)

	raw, again := codeBuddyCheckpointRun(t, state)
	assert.Empty(t, raw, "unchanged transcripts should not be parsed again")
	assert.Empty(t, again)

	// A tool call and its later result are paired when the changed file is reread.
	call := codeBuddyCodeToolCallRecord(t, "write-1", stamp.Add(time.Minute), projectDir, "Write", map[string]any{
		"file_path": "new.go", "content": "one\ntwo",
	})
	call["sessionId"] = "session-a"
	result := codeBuddyCodeToolResult("write-1", stamp.Add(2*time.Minute), "completed", map[string]any{
		"content": "Successfully created and wrote to new file",
	})
	result["sessionId"] = "session-a"
	codeBuddyCodeWriteTranscript(t, pathA, codeBuddyCodeJSONLines(t, prompt, call, result), modified.Add(time.Second))

	raw, changed := codeBuddyCheckpointRun(t, state)
	for _, h := range raw {
		assert.Equal(t, "session-a", h.AISession, "the unchanged transcript should remain skipped")
	}

	require.Len(t, changed, 2)
	assert.Equal(t, heartbeat.FileType, changed[1].EntityType)
	assert.Equal(t, filepath.ToSlash(filepath.Join(projectDir, "new.go")), changed[1].Entity)
	assert.Equal(t, heartbeat.PointerTo(2), changed[1].AILineChanges)
}

func TestCodeBuddyCheckpointRetriesUnreadableTranscript(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("file permissions cannot make this transcript unreadable")
	}

	_, projectDir, sessionsDir := codeBuddyCodeTestDirs(t)
	stamp := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	path := filepath.Join(sessionsDir, "session.jsonl")
	prompt := map[string]any{
		"timestamp": stamp.UnixMilli(), "type": "message", "role": "user",
		"sessionId": "session", "cwd": projectDir, "content": "hello",
	}
	codeBuddyCodeWriteTranscript(t, path, codeBuddyCodeJSONLines(t, prompt), stamp.Add(24*time.Hour))
	require.NoError(t, os.Chmod(path, 0))
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })

	state := &parserCheckpoint{global: stamp.Add(-time.Hour), Sessions: make(map[string]*sessionCheckpoint)}
	_, failed := codeBuddyCheckpointRun(t, state)
	assert.Empty(t, failed)
	assert.NotContains(t, state.Sources, path, "a failed read must be retried")

	require.NoError(t, os.Chmod(path, 0o600))
	_, retried := codeBuddyCheckpointRun(t, state)
	require.Len(t, retried, 1)
}

func TestCodeBuddyCheckpointPrunesUndiscoveredSources(t *testing.T) {
	state := &parserCheckpoint{Sources: map[string]checkpointSource{"kept": {Size: 1}, "gone": {Size: 2}}}

	CodeBuddyCode{checkpoint: state}.pruneSources([]codeBuddyCodeTranscript{{path: "kept"}, {path: "new"}})

	assert.Equal(t, map[string]checkpointSource{"kept": {Size: 1}}, state.Sources)
}
