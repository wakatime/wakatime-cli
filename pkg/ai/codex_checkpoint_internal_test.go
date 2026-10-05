package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

const (
	codexTestSessionA = "019a1b2c-3d4e-7f80-9123-456789abcdea"
	codexTestSessionB = "019a1b2c-3d4e-7f80-9123-456789abcdeb"
)

type codexTestSync struct {
	t      *testing.T
	v      *viper.Viper
	global time.Time
	logs   bytes.Buffer
}

func newCodexTestSync(t *testing.T) *codexTestSync {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("CODEX_HOME", filepath.Join(home, "codex"))

	v := viper.New()
	v.Set("internal-config", filepath.Join(t.TempDir(), "wakatime-internal.cfg"))

	return &codexTestSync{t: t, v: v, global: time.Date(2025, 2, 24, 0, 0, 0, 0, time.UTC)}
}

func (*codexTestSync) transcriptPath(id string) string {
	return filepath.Join(os.Getenv("CODEX_HOME"), "sessions", "2026", "06", "20",
		"rollout-2026-06-20T09-00-00-"+id+".jsonl")
}

// run mirrors one parseAIHeartbeats pass for the Codex parser.
func (s *codexTestSync) run() Heartbeats {
	s.t.Helper()

	s.logs.Reset()
	logger := log.New(&s.logs)
	logger.SetVerbose(true)
	ctx := log.ToContext(context.Background(), logger)

	checkpoints, err := loadSyncCheckpoints(ctx, s.v, s.global)
	require.NoError(s.t, err)

	state := checkpoints.parser("Codex", s.global)
	parser := Codex{After: state.discoveryAfter(), checkpoint: state}

	heartbeats, err := parser.Parse(ctx)
	require.NoError(s.t, err)

	heartbeats = state.filter(heartbeats)
	checkpoints.PreviousGlobal = s.global

	for _, h := range heartbeats {
		if stamp := heartbeatTime(h.Time); stamp.After(checkpoints.Global) {
			checkpoints.Global = stamp
		}
	}

	require.NoError(s.t, checkpoints.save())

	s.global = checkpoints.Global

	return heartbeats
}

func codexTestLine(t *testing.T, timestamp string, kind string, payload map[string]any) string {
	t.Helper()

	raw, err := json.Marshal(map[string]any{"timestamp": timestamp, "type": kind, "payload": payload})
	require.NoError(t, err)

	return string(raw) + "\n"
}

func codexTestMeta(t *testing.T, timestamp string, id string) string {
	t.Helper()

	return codexTestLine(t, timestamp, "session_meta",
		map[string]any{"id": id, "cwd": "/workspace/project", "source": "cli", "cli_version": "0.120.0"})
}

func codexTestUserMessage(t *testing.T, timestamp string, message string) string {
	t.Helper()

	return codexTestLine(t, timestamp, "event_msg", map[string]any{"type": "user_message", "message": message})
}

func writeCodexTestFile(t *testing.T, path string, contents ...string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(strings.Join(contents, "")), 0o600))
}

func appendCodexTestFile(t *testing.T, path string, contents ...string) {
	t.Helper()

	fh, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600) // nolint:gosec
	require.NoError(t, err)

	_, err = fh.WriteString(strings.Join(contents, ""))
	require.NoError(t, err)
	require.NoError(t, fh.Close())
}

// Issue #1611: a quiet session holds back discovery, but its unchanged file
// must not be parsed again, while activity appended to it later is still sent.
func TestCodexParse_SkipsUnchangedTranscripts(t *testing.T) {
	sync := newCodexTestSync(t)
	pathA := sync.transcriptPath(codexTestSessionA)
	pathB := sync.transcriptPath(codexTestSessionB)

	writeCodexTestFile(t, pathA, codexTestMeta(t, "2026-06-20T09:00:00Z", codexTestSessionA))
	writeCodexTestFile(t, pathB,
		codexTestMeta(t, "2026-06-20T10:30:00Z", codexTestSessionB),
		codexTestUserMessage(t, "2026-06-20T11:00:00Z", "Synthetic current user activity."),
	)

	first := sync.run()
	require.Len(t, first, 1)
	assert.Equal(t, codexTestSessionB, first[0].AISession)

	second := sync.run()
	assert.Empty(t, second)
	assert.Contains(t, sync.logs.String(), "Found 2 transcript logs")
	assert.Equal(t, 2, strings.Count(sync.logs.String(), "skipping unchanged codex transcript"))

	appendCodexTestFile(t, pathA,
		codexTestUserMessage(t, "2026-06-20T10:00:00Z", "Synthetic late older activity in quiet session A."))

	third := sync.run()
	require.Len(t, third, 1)
	assert.Equal(t, codexTestSessionA, third[0].AISession)
	assert.Equal(t, 1, strings.Count(sync.logs.String(), "skipping unchanged codex transcript"))
}

// An unchanged file whose size and mtime match is skipped even if its bytes
// differ, which proves the skip comes from the checkpoint rather than parsing.
func TestCodexParse_UnchangedTranscriptIsNotRead(t *testing.T) {
	sync := newCodexTestSync(t)
	path := sync.transcriptPath(codexTestSessionA)

	meta := codexTestMeta(t, "2026-06-20T09:00:00Z", codexTestSessionA)
	writeCodexTestFile(t, path, meta, codexTestUserMessage(t, "2026-06-20T09:01:00Z", "first"))
	require.Len(t, sync.run(), 1)

	info, err := os.Stat(path)
	require.NoError(t, err)

	writeCodexTestFile(t, path, meta, codexTestUserMessage(t, "2026-06-20T09:02:00Z", "other"))
	require.NoError(t, os.Chtimes(path, info.ModTime(), info.ModTime()))

	assert.Empty(t, sync.run())
}

// Issue #1609: activity older than the last 10 MiB of a transcript is still
// read the first time a transcript is synced.
func TestCodexParse_ReadsWholeTranscriptOnFirstSync(t *testing.T) {
	sync := newCodexTestSync(t)
	path := sync.transcriptPath(codexTestSessionA)

	contents := []string{
		codexTestMeta(t, "2026-06-20T07:00:00Z", codexTestSessionA),
		codexTestLine(t, "2026-06-20T07:01:00Z", "response_item", map[string]any{
			"type": "custom_tool_call", "name": "apply_patch", "call_id": "patch-1",
			"input": "*** Begin Patch\n*** Add File: demo.txt\n+hello\n*** End Patch",
		}),
		codexTestLine(t, "2026-06-20T07:01:01Z", "response_item", map[string]any{
			"type": "custom_tool_call_output", "call_id": "patch-1", "output": "Done",
		}),
	}

	reasoning := strings.Repeat("r", 1024*1024)
	for range 12 {
		contents = append(contents, codexTestLine(t, "2026-06-20T07:02:00Z", "response_item",
			map[string]any{"type": "reasoning", "summary": []any{}, "encrypted_content": reasoning}))
	}

	contents = append(contents, codexTestLine(t, "2026-06-20T07:03:00Z", "response_item", map[string]any{
		"type": "message", "role": "assistant",
		"content": []map[string]any{{"type": "output_text", "text": "Finished."}},
	}))

	writeCodexTestFile(t, path, contents...)

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Greater(t, info.Size(), int64(maxTranscriptLineSize))

	got := sync.run()

	writes := 0

	for _, h := range got {
		if h.IsWrite != nil && *h.IsWrite {
			writes++

			assert.Equal(t, filepath.Join("/workspace/project", "demo.txt"), h.Entity)
		}
	}

	assert.Equal(t, 1, writes)
	assert.Len(t, got, 2)
}

// Issue #1609: when a transcript grows by more than 10 MiB between syncs, for
// example after downtime, every appended record is still read.
func TestCodexParse_ReadsAllGrowthSincePreviousSync(t *testing.T) {
	sync := newCodexTestSync(t)
	path := sync.transcriptPath(codexTestSessionA)

	writeCodexTestFile(t, path,
		codexTestMeta(t, "2026-06-20T07:00:00Z", codexTestSessionA),
		codexTestLine(t, "2026-06-20T07:00:01Z", "turn_context",
			map[string]any{"cwd": "/workspace/project", "model": "gpt-5.5"}),
		codexTestUserMessage(t, "2026-06-20T07:00:02Z", "first"),
	)
	require.Len(t, sync.run(), 1)

	reasoning := strings.Repeat("r", 1024*1024)
	appended := []string{codexTestUserMessage(t, "2026-06-20T08:00:00Z", "early in the gap")}

	for range 12 {
		appended = append(appended, codexTestLine(t, "2026-06-20T08:01:00Z", "response_item",
			map[string]any{"type": "reasoning", "summary": []any{}, "encrypted_content": reasoning}))
	}

	appended = append(appended, codexTestUserMessage(t, "2026-06-20T08:02:00Z", "late in the gap"))
	appendCodexTestFile(t, path, appended...)

	got := sync.run()
	require.Len(t, got, 2)
	assert.Equal(t, len([]rune("early in the gap")), got[0].AIPromptLength)
	assert.Equal(t, len([]rune("late in the gap")), got[1].AIPromptLength)
	assert.Contains(t, got[1].UserAgent, "gpt/5.5")
}

func TestCodexParse_SkipsOversizedLines(t *testing.T) {
	sync := newCodexTestSync(t)
	path := sync.transcriptPath(codexTestSessionA)

	writeCodexTestFile(t, path,
		codexTestMeta(t, "2026-06-20T07:00:00Z", codexTestSessionA),
		codexTestLine(t, "2026-06-20T07:01:00Z", "response_item",
			map[string]any{"type": "reasoning", "encrypted_content": strings.Repeat("r", maxTranscriptLineSize+1)}),
		codexTestUserMessage(t, "2026-06-20T07:02:00Z", "after the oversized line"),
	)

	got := sync.run()
	require.Len(t, got, 1)
	assert.Equal(t, len([]rune("after the oversized line")), got[0].AIPromptLength)
}

func TestCodexPlanTranscriptReads(t *testing.T) {
	modified := time.Date(2026, 6, 20, 9, 0, 0, 0, time.UTC)
	later := modified.Add(time.Minute)
	large := int64(3 * maxTranscriptLineSize)

	state := &parserCheckpoint{Sources: map[string]checkpointSource{
		"unchanged":        {Size: 100, Modified: modified},
		"grown":            {Size: 100, Modified: modified},
		"grown-large":      {Size: large, Modified: modified},
		"truncated":        {Size: 500, Modified: modified},
		"fork-unchanged":   {Size: 100, Modified: modified},
		"parent-changed":   {Size: 100, Modified: modified},
		"touched-same-len": {Size: 100, Modified: modified},
	}}

	transcripts := []codexTranscript{
		{path: "new", owner: "new", source: checkpointSource{Size: 50, Modified: later}},
		{path: "unchanged", owner: "unchanged", source: checkpointSource{Size: 100, Modified: modified}},
		{path: "grown", owner: "grown", source: checkpointSource{Size: 200, Modified: later}},
		{path: "grown-large", owner: "grown-large", source: checkpointSource{Size: large + 10, Modified: later}},
		{path: "truncated", owner: "truncated", source: checkpointSource{Size: 50, Modified: later}},
		{path: "fork-unchanged", owner: "root", source: checkpointSource{Size: 100, Modified: modified}},
		{path: "parent-changed", owner: "root", source: checkpointSource{Size: 150, Modified: later}},
		{path: "touched-same-len", owner: "touched", source: checkpointSource{Size: 100, Modified: later}},
	}

	Codex{checkpoint: state}.planTranscriptReads(transcripts)

	expected := map[string]struct {
		skip  bool
		start int64
	}{
		"new":              {false, 0},
		"unchanged":        {true, 0},
		"grown":            {false, 0},
		"grown-large":      {false, large - maxTranscriptLineSize},
		"truncated":        {false, 0},
		"fork-unchanged":   {false, 0},
		"parent-changed":   {false, 0},
		"touched-same-len": {false, 0},
	}

	for _, transcript := range transcripts {
		assert.Equal(t, expected[transcript.path].skip, transcript.skip, transcript.path)
		assert.Equal(t, expected[transcript.path].start, transcript.start, transcript.path)
	}
}

func TestCodexReaderStartsAtFullLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	require.NoError(t, os.WriteFile(path, []byte("one\ntwo\nthree\n"), 0o600))

	read := func(start int64) string {
		fh, err := os.Open(path) // nolint:gosec
		require.NoError(t, err)

		defer fh.Close() // nolint:errcheck

		reader, err := codexReader(fh, path, start)
		require.NoError(t, err)

		line, err := grokBuildReadJSONLLine(reader, maxTranscriptLineSize)
		require.NoError(t, err)

		return string(line)
	}

	assert.Equal(t, "one", read(0))
	assert.Equal(t, "two", read(4))
	assert.Equal(t, "three", read(5))
}

func TestCodexPruneSources(t *testing.T) {
	state := &parserCheckpoint{Sources: map[string]checkpointSource{"kept": {Size: 1}, "gone": {Size: 2}}}

	Codex{checkpoint: state}.pruneSources([]codexTranscript{{path: "kept"}, {path: "new"}})

	assert.Equal(t, map[string]checkpointSource{"kept": {Size: 1}}, state.Sources)
}
