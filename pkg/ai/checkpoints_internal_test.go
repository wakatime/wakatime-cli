package ai

import (
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
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestSyncCheckpointsSessionsAndParsers(t *testing.T) {
	after := time.Date(2026, 9, 1, 11, 0, 0, 0, time.UTC)
	v := viper.New()
	v.Set("internal-config", filepath.Join(t.TempDir(), "internal.cfg"))
	state, err := loadSyncCheckpoints(t.Context(), v, after)
	require.NoError(t, err)

	slow := state.parser("Claude", after)
	other := state.parser("Codex", after)
	config := ParserConfig{checkpoint: slow}
	assert.Equal(t, checkpointReadAfter(after), config.sessionAfter("pending"))
	h := heartbeat.Heartbeat{AISession: "same-id", Entity: "app",
		Time: heartbeatTimestamp(after.Add(time.Hour)), AIOutputTokens: 5}
	require.Len(t, slow.filter(Heartbeats{h}), 1)
	h.Time = heartbeatTimestamp(after.Add(2 * time.Hour))
	require.Len(t, other.filter(Heartbeats{h}), 1)
	require.NoError(t, state.save())

	state, err = loadSyncCheckpoints(t.Context(), v, after.Add(2*time.Hour))
	require.NoError(t, err)

	slow = state.parser("Claude", after)
	assert.Equal(t, after, slow.Sessions["pending"].Cutoff)
	assert.Equal(t, after.Add(time.Hour), slow.Sessions["same-id"].Cutoff)
	assert.Equal(t, after.Add(2*time.Hour), state.Parsers["Codex"].Sessions["same-id"].Cutoff)
	assert.Equal(t, checkpointReadAfter(after), slow.discoveryAfter())
	require.NoError(t, state.save())

	// An explicit rewind resets both parser and session cutoffs for backfill.
	state, err = loadSyncCheckpoints(t.Context(), v, after)
	require.NoError(t, err)
	assert.Empty(t, state.Parsers)

	// Different internal configs cannot consume one another's progress.
	v.Set("internal-config", filepath.Join(t.TempDir(), "other.cfg"))
	state, err = loadSyncCheckpoints(t.Context(), v, after)
	require.NoError(t, err)
	assert.Empty(t, state.Parsers)
}

func TestSyncCheckpointBoundary(t *testing.T) {
	stamp := time.Date(2026, 9, 1, 12, 0, 0, 123456789, time.UTC)
	p := &parserCheckpoint{Cutoff: stamp.Add(-time.Hour), Sessions: make(map[string]*sessionCheckpoint)}
	h := heartbeat.Heartbeat{AISession: "session", Entity: "app", Time: heartbeatTimestamp(stamp), AIOutputTokens: 5}
	require.Len(t, p.filter(Heartbeats{h}), 1)
	assert.True(t, timestampAtOrAfterCutoff(stamp, (ParserConfig{checkpoint: p}).sessionAfter("session")))
	assert.Empty(t, p.filter(Heartbeats{h}))
	// A second identical record appended at the same timestamp is new activity.
	got := p.filter(Heartbeats{h, h})
	require.Len(t, got, 1)
	assert.EqualValues(t, 5, got[0].AIOutputTokens)
	assert.Empty(t, p.filter(Heartbeats{h, h}))
	// An in-place update emits only new tokens, without repeating the prompt.
	h.AIOutputTokens = 8
	got = p.filter(Heartbeats{h})
	require.Len(t, got, 1)
	assert.EqualValues(t, 3, got[0].AIOutputTokens)
	assert.Greater(t, got[0].Time, h.Time)
	assert.Empty(t, p.filter(Heartbeats{h}))
}

func TestSyncCheckpointInvalidState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "internal.cfg")
	v := viper.New()
	v.Set("internal-config", path)
	statePath := filepath.Join(filepath.Dir(path), "internal-ai-parsing.json")
	require.NoError(t, os.WriteFile(statePath, []byte("broken"), 0o600))

	// An unreadable file is discarded and rebuilt rather than disabling sync forever.
	state, err := loadSyncCheckpoints(t.Context(), v, time.Now())
	require.NoError(t, err)
	assert.Empty(t, state.Parsers)
	require.NoError(t, state.save())

	raw, err := os.ReadFile(statePath)
	require.NoError(t, err)
	assert.True(t, json.Valid(raw), "rebuilt checkpoint must be valid json")

	_, err = loadSyncCheckpoints(t.Context(), v, time.Now())
	require.NoError(t, err)

	// A newer format is never overwritten: that would silently discard its state.
	require.NoError(t, os.WriteFile(statePath, []byte(`{"version":2}`), 0o600))
	_, err = loadSyncCheckpoints(t.Context(), v, time.Now())
	require.ErrorContains(t, err, "unsupported ai checkpoint version")
}

func TestSyncCheckpointDropsCorruptEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "internal.cfg")
	v := viper.New()
	v.Set("internal-config", path)
	statePath := filepath.Join(filepath.Dir(path), "internal-ai-parsing.json")
	raw := `{"version":1,"parsers":{"Claude":null,"Codex":{"cutoff":"2026-09-01T12:00:00Z",` +
		`"sessions":{"good":{"cutoff":"2026-09-01T12:00:00Z"},"bad":null}}}}`
	require.NoError(t, os.WriteFile(statePath, []byte(raw), 0o600))

	state, err := loadSyncCheckpoints(t.Context(), v, time.Now())
	require.NoError(t, err, "a single corrupt entry must not disable ai sync for every parser")
	assert.NotContains(t, state.Parsers, "Claude")
	require.Contains(t, state.Parsers, "Codex")
	assert.Contains(t, state.Parsers["Codex"].Sessions, "good")
	assert.NotContains(t, state.Parsers["Codex"].Sessions, "bad")
}

func TestSyncCheckpointNewSessionStartsAtGlobalCutoff(t *testing.T) {
	base := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	p := &parserCheckpoint{Cutoff: base, global: base, Sessions: make(map[string]*sessionCheckpoint)}
	config := ParserConfig{checkpoint: p}

	// The first time a session is seen it starts at the global cutoff and is
	// recorded there, even though it emits nothing yet.
	assert.Equal(t, checkpointReadAfter(base), config.sessionAfter("b"))
	require.Contains(t, p.Sessions, "b")
	assert.Equal(t, base, p.Sessions["b"].Cutoff)

	// Another session advances the global cutoff. Session "b" keeps the cutoff
	// it was first seen with, so its later, older activity is still emitted.
	advanced := heartbeat.Heartbeat{AISession: "a", Entity: "app",
		Time: heartbeatTimestamp(base.Add(5 * time.Hour)), AIOutputTokens: 5}
	require.Len(t, p.filter(Heartbeats{advanced}), 1)

	p.global = base.Add(5 * time.Hour)
	earlier := heartbeat.Heartbeat{AISession: "b", Entity: "app",
		Time: heartbeatTimestamp(base.Add(time.Hour)), AIOutputTokens: 3}
	require.Len(t, p.filter(Heartbeats{earlier}), 1)

	// A session first seen after the global cutoff advanced starts from it.
	assert.Equal(t, checkpointReadAfter(base.Add(5*time.Hour)), config.sessionAfter("c"))
}

func TestSyncCheckpointQuietSessionDoesNotLowerDiscovery(t *testing.T) {
	base := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	p := &parserCheckpoint{Cutoff: base, global: base, Sessions: make(map[string]*sessionCheckpoint)}
	config := ParserConfig{checkpoint: p}

	active := heartbeat.Heartbeat{AISession: "active", Entity: "app",
		Time: heartbeatTimestamp(base.Add(3 * time.Hour)), AIOutputTokens: 5}
	require.Len(t, p.filter(Heartbeats{active}), 1)
	p.global = base.Add(3 * time.Hour)

	before := p.discoveryAfter()

	// A session first seen later starts at the (newer) global cutoff, so it
	// cannot lower discovery. That would make sqlite cursors look rewound.
	config.sessionAfter("quiet")
	assert.Equal(t, before, p.discoveryAfter())

	// With no recorded session at all, discovery is the global cutoff itself.
	empty := &parserCheckpoint{Cutoff: base, global: base.Add(time.Hour), Sessions: make(map[string]*sessionCheckpoint)}
	assert.Equal(t, checkpointReadAfter(base.Add(time.Hour)), empty.discoveryAfter())
}

func TestAISyncSQLiteParsersDoNotUseCheckpoints(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	v := viper.New()
	v.Set("internal-config", filepath.Join(home, "internal.cfg"))
	v.Set("internal.ai_logs_last_parsed_at", time.Now().Add(-time.Hour).Format(time.RFC3339Nano))

	sqliteParsers := []string{
		(Gemini{}).Name(), (Cody{}).Name(), (Copilot{}).Name(), (Cursor{}).Name(), (Goose{}).Name(),
		(OpenCode{}).Name(), (Qoder{}).Name(), (Windsurf{}).Name(), (Crush{}).Name(), (CursorAgent{}).Name(),
		(Forge{}).Name(), (Hermes{}).Name(), (KiloCode{}).Name(), (QuickDesk{}).Name(), (Warp{}).Name(),
		(ZCode{}).Name(), (Zed{}).Name(),
	}

	// State left behind by an earlier build that checkpointed a sqlite parser,
	// including its row cursors.
	statePath := filepath.Join(home, "internal-ai-parsing.json")
	stale := `{"version":1,"parsers":{"` + (Zed{}).Name() + `":{"cutoff":"2026-09-01T12:00:00Z",` +
		`"sessions":{},"cursors":{"/tmp/cursor.json":{"Version":2}}}}}`
	require.NoError(t, os.WriteFile(statePath, []byte(stale), 0o600))

	handler := WithAISync(Config{V: v})(func(_ context.Context, _ []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return nil, nil
	})
	_, err := handler(t.Context(), nil)
	require.NoError(t, err)

	raw, err := os.ReadFile(statePath)
	require.NoError(t, err)

	var state syncCheckpoints
	require.NoError(t, json.Unmarshal(raw, &state))
	assert.Contains(t, state.Parsers, (Claude{}).Name())

	for _, name := range sqliteParsers {
		assert.NotContains(t, state.Parsers, name, "sqlite parser %s must not be checkpointed", name)
	}

	assert.NotContains(t, string(raw), "cursors")
}

func TestClaudeTranscriptPrefixUsesPrefixedSessionCutoff(t *testing.T) {
	base := time.Date(2026, 9, 1, 11, 0, 0, 0, time.UTC)
	p := &parserCheckpoint{Cutoff: base, global: base, Sessions: map[string]*sessionCheckpoint{
		"openclaude:s": {Cutoff: base.Add(2 * time.Hour)},
	}}
	path := filepath.Join(t.TempDir(), "s.jsonl")
	line := `{"timestamp":"2026-09-01T12:00:00Z","sessionId":"s","type":"user",` +
		`"message":{"role":"user","content":"continue"}}`
	require.NoError(t, os.WriteFile(path, []byte(line+"\n"), 0o600))

	g := Claude(ParserConfig{checkpoint: p})

	got, err := g.parseTranscriptWithPrefix(t.Context(), path, "openclaude:")
	require.NoError(t, err)
	assert.Empty(t, got, "cutoff lookup must use the prefixed session id")

	got, err = g.parseTranscriptWithPrefix(t.Context(), path, "")
	require.NoError(t, err)
	assert.NotEmpty(t, got, "unprefixed session has no cutoff beyond the floor")
}

func TestAISyncFutureTimestampDoesNotWipeCheckpoints(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(home, ".claude"))

	now := time.Now().UTC()
	v := viper.New()
	v.Set("internal-config", filepath.Join(home, "internal.cfg"))
	v.Set("internal.ai_logs_last_parsed_at", now.Add(-time.Hour).Format(time.RFC3339Nano))

	prompt := func(stamp time.Time, id string) string {
		return `{"timestamp":"` + stamp.Format(time.RFC3339Nano) + `","sessionId":"` + id + `","type":"user",` +
			`"message":{"role":"user","content":"continue"}}` + "\n"
	}
	write := func(name, contents string, flag int) {
		path := filepath.Join(home, ".claude", "projects", "project", name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))

		f, err := os.OpenFile(path, flag|os.O_CREATE|os.O_WRONLY, 0o600)
		require.NoError(t, err)
		_, err = f.WriteString(contents)
		require.NoError(t, err)
		require.NoError(t, f.Close())
	}

	// One record is stamped in the future, for example from clock skew.
	write("future.jsonl", prompt(now.Add(100*365*24*time.Hour), "future"), os.O_TRUNC)
	write("normal.jsonl", prompt(now.Add(-10*time.Minute), "normal"), os.O_TRUNC)

	sync := func() Heartbeats {
		var got Heartbeats

		handler := WithAISync(Config{V: v})(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			got = hh
			return nil, nil
		})
		_, err := handler(t.Context(), nil)
		require.NoError(t, err)

		return got
	}

	require.NotEmpty(t, sync())

	// Later activity in an already tracked session must still be picked up. The
	// future record must not make the next run look like a backfill request,
	// which would reset every cutoff to now and drop this line.
	write("normal.jsonl", prompt(now.Add(-30*time.Second), "normal"), os.O_APPEND)

	var sessions []string
	for _, h := range sync() {
		sessions = append(sessions, h.AISession)
	}

	assert.Contains(t, sessions, "normal")
}

func TestAISyncIndependentSessions(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("CODEX_HOME", filepath.Join(home, ".codex"))
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(home, ".claude"))
	t.Setenv("FACTORY_DIR", filepath.Join(home, ".factory"))

	after := time.Date(2026, 9, 1, 11, 0, 0, 0, time.UTC)
	v := viper.New()
	v.Set("internal-config", filepath.Join(home, "internal.cfg"))
	v.Set("internal.ai_logs_last_parsed_at", after.Format(time.RFC3339Nano))

	write := func(relative string, lines ...string) {
		path := filepath.Join(home, filepath.FromSlash(relative))
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
		require.NoError(t, os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600))
	}
	claudePrompt := func(stamp, id string) string {
		return `{"timestamp":"2026-09-01T` + stamp + `Z","sessionId":"` + id + `","type":"user",` +
			`"message":{"role":"user","content":"continue"}}`
	}
	usage := `{"timestamp":"2026-09-01T12:30:00Z","sessionId":"slow","type":"assistant",` +
		`"message":{"id":"msg-1","usage":{"input_tokens":10,"cache_read_input_tokens":20,"output_tokens":30}}}`
	slowPath := ".claude/projects/project/slow.jsonl"
	write(slowPath, claudePrompt("12:00:00", "slow"), usage)
	// A session without any heartbeat must also retain its initial cutoff.
	write(".claude/projects/project/pending.jsonl", strings.ReplaceAll(usage, "slow", "pending"))
	write(".claude/projects/project/fast.jsonl", claudePrompt("12:45:00", "fast"))

	meta := `{"timestamp":"2026-09-01T11:00:00Z","type":"session_meta","payload":{"id":"codex-session"}}`
	codexPrompt := `{"timestamp":"2026-09-01T12:00:00Z","type":"event_msg",` +
		`"payload":{"type":"user_message","message":"continue"}}`
	write(".codex/sessions/session.jsonl", meta, codexPrompt)
	write(".factory/sessions/session.jsonl", `{"timestamp":"2026-09-01T12:15:00Z","session_id":"generic",`+
		`"role":"assistant","usage":{"input_tokens":3,"output_tokens":4}}`)

	sync := func() Heartbeats {
		var got Heartbeats

		handler := WithAISync(Config{V: v})(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			got = hh
			return nil, nil
		})
		_, err := handler(t.Context(), nil)
		require.NoError(t, err)

		return got
	}
	first := sync()
	require.NotEmpty(t, first)
	require.Empty(t, sync(), "unchanged rerun must not resend boundary heartbeats")

	// Simulate another parser having advanced the global cutoff beyond these sessions.
	require.NoError(t, UpdateLastParsedAt(t.Context(), v, after.Add(4*time.Hour)))
	write(slowPath, claudePrompt("12:00:00", "slow"), usage, claudePrompt("13:00:00", "slow"))
	write(".claude/projects/project/pending.jsonl", strings.ReplaceAll(usage, "slow", "pending"),
		claudePrompt("13:00:00", "pending"))

	tokenCount := `{"timestamp":"2026-09-01T12:30:00Z","type":"event_msg",` +
		`"payload":{"type":"token_count","info":{"total_token_usage":` +
		`{"input_tokens":30,"cached_input_tokens":20,"output_tokens":30}}}}`
	write(".codex/sessions/session.jsonl", meta, codexPrompt, tokenCount)
	write(".factory/sessions/session.jsonl", `{"timestamp":"2026-09-01T12:15:00Z","session_id":"generic",`+
		`"role":"assistant","usage":{"input_tokens":3,"output_tokens":4}}`,
		`{"timestamp":"2026-09-01T12:30:00Z","session_id":"generic",`+
			`"role":"assistant","usage":{"input_tokens":10,"cache_read_input_tokens":20,"output_tokens":30}}`)

	second := sync()

	totals := make(map[string][3]int64)
	for _, h := range second {
		sum := totals[h.AISession]
		sum[0] += h.AIInputTokens
		sum[1] += h.AICachedInputTokens
		sum[2] += h.AIOutputTokens
		totals[h.AISession] = sum
	}

	for _, id := range []string{"slow", "pending", "codex-session", "generic"} {
		assert.Equal(t, [3]int64{10, 20, 30}, totals[id], "session %s", id)
	}

	assert.Empty(t, sync())
	persisted, err := getSyncLastParsedAt(t.Context(), v)
	require.NoError(t, err)
	assert.Equal(t, after.Add(4*time.Hour), persisted, "older sessions cannot move the global cutoff backward")

	raw, err := os.ReadFile(filepath.Join(home, "internal-ai-parsing.json"))
	require.NoError(t, err)

	var state syncCheckpoints
	require.NoError(t, json.Unmarshal(raw, &state))
	assert.Contains(t, state.Parsers, "Claude")
	assert.Contains(t, state.Parsers, "Codex")
	assert.Contains(t, state.Parsers, "Droid")

	// Incremental sync and one complete replay must produce the same totals.
	v.Set("internal-config", filepath.Join(home, "replay.cfg"))

	replayed := sync()
	tokenTotals := func(hh Heartbeats) [3]int64 {
		var totals [3]int64
		for _, h := range hh {
			totals[0] += h.AIInputTokens
			totals[1] += h.AICachedInputTokens
			totals[2] += h.AIOutputTokens
		}

		return totals
	}
	assert.Equal(t, tokenTotals(replayed), tokenTotals(append(first, second...)))
}
