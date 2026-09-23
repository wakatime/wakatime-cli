//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"context"
	"database/sql"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func updateTestHome(t *testing.T) string {
	t.Helper()

	home := t.TempDir()
	for _, key := range []string{"HOME", "USERPROFILE", "WAKATIME_HOME"} {
		t.Setenv(key, home)
	}

	for _, key := range []string{"XDG_DATA_HOME", "OPENCODE_DATA_DIR", "OPENCODE_DB_PREFIX", "CODEX_HOME",
		"DSH_HOME", "HERMES_HOME", "KIRO_HOME"} {
		t.Setenv(key, "")
	}

	return home
}

func updateWrite(t *testing.T, path, contents string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
}

func updateDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	db, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	return db
}

func updateSQL(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()

	_, err := db.Exec(query, args...)
	require.NoError(t, err)
}

type updateUsageTotals struct {
	input, cached, output int64
	prompt                int
}

func updateTotals(hh Heartbeats) updateUsageTotals {
	var totals updateUsageTotals
	for _, h := range hh {
		totals.input += h.AIInputTokens
		totals.cached += h.AICachedInputTokens
		totals.output += h.AIOutputTokens
		totals.prompt += h.AIPromptLength
	}

	return totals
}

func TestProviderUpdatesOpenCodeMixedGenerations(t *testing.T) {
	home := updateTestHome(t)
	root := filepath.Join(home, ".local", "share", "opencode")
	db := updateDB(t, filepath.Join(root, "opencode.db"))
	updateSQL(t, db, updateFixture(t, "opencodemixedgenerations-1.txt"))
	updateWrite(t, filepath.Join(root, "storage", "session", "project", "legacy.json"),
		`{"id":"legacy","directory":"/project"}`)
	updateWrite(t, filepath.Join(root, "storage", "message", "legacy", "legacy-msg.json"), updateFixture(t,
		"opencodemixedgenerations-2.txt"))

	hh, err := (OpenCode{}).Parse(context.Background())
	require.NoError(t, err)

	totals := updateTotals(hh)
	input, cached, output, _ := totals.input, totals.cached, totals.output, totals.prompt
	assert.EqualValues(t, 132, input)
	assert.EqualValues(t, 4, cached)
	assert.EqualValues(t, 20, output)

	var files int

	for _, h := range hh {
		if h.EntityType == heartbeat.FileType {
			files++

			assert.Equal(t, "/project/main.go", filepath.ToSlash(h.Entity))
		}
	}

	assert.Equal(t, 1, files)
}

func TestProviderUpdatesCodexUsageAndStructuredSource(t *testing.T) {
	home := updateTestHome(t)
	path := filepath.Join(home, ".codex", "archived_sessions", "child.jsonl")
	updateWrite(t, path, updateFixture(t, "codexusageandstructuredsource-1.txt"))

	hh, err := (Codex{}).Parse(context.Background())
	require.NoError(t, err)

	totals := updateTotals(hh)
	input, cached, output, prompt := totals.input, totals.cached, totals.output, totals.prompt
	assert.EqualValues(t, 23, input)
	assert.EqualValues(t, 7, cached)
	assert.EqualValues(t, 7, output)
	assert.Equal(t, 3, prompt)
	require.NotEmpty(t, hh)
	assert.Equal(t, "/project", hh[0].ProjectPathOverride)
	// An archived copy of the same session must not add another set of tokens.
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	updateWrite(t, filepath.Join(home, ".codex", "sessions", "copy.jsonl"), string(raw))

	again, err := (Codex{}).Parse(context.Background())
	require.NoError(t, err)
	require.Len(t, again, len(hh))
	totals = updateTotals(again)
	input, cached, output, prompt = totals.input, totals.cached, totals.output, totals.prompt
	assert.EqualValues(t, 23, input)
	assert.EqualValues(t, 7, cached)
	assert.EqualValues(t, 7, output)
	assert.Equal(t, 3, prompt)
}

func TestProviderUpdatesDeepSeekGenerationsAndInterleaving(t *testing.T) {
	home := updateTestHome(t)
	dir := filepath.Join(home, ".dsh", "sessions", "project", "session-one")
	updateWrite(t, filepath.Join(dir, "session.jsonl"), updateFixture(t, "deepseekgenerationsandinterleaving-1.txt"))
	updateWrite(t, filepath.Join(dir, "session.v3.jsonl"), updateFixture(t, "deepseekgenerationsandinterleaving-2.txt"))

	hh, err := (DeepSeek{}).Parse(context.Background())
	require.NoError(t, err)

	totals := updateTotals(hh)
	input, _, output, _ := totals.input, totals.cached, totals.output, totals.prompt
	assert.EqualValues(t, 32, input)
	assert.EqualValues(t, 9, output)

	for _, h := range hh {
		assert.Equal(t, "current", h.AISession)
		assert.Contains(t, h.UserAgent, "deepseek/4")
	}

	updateWrite(t, filepath.Join(dir, "session.v4.jsonl"), `{"type":"session","version":4}`)

	hh, err = (DeepSeek{}).Parse(context.Background())
	require.NoError(t, err)
	assert.Empty(t, hh)
}

func TestProviderUpdatesHermesGrowth(t *testing.T) {
	home := updateTestHome(t)
	db := updateDB(t, filepath.Join(home, ".hermes", "state.db"))
	updateSQL(t, db, updateFixture(t, "hermesgrowth-1.txt"))

	hh, err := (Hermes{}).Parse(context.Background())
	require.NoError(t, err)

	totals := updateTotals(hh)
	input, cached, output, _ := totals.input, totals.cached, totals.output, totals.prompt
	assert.EqualValues(t, 105, input)
	assert.EqualValues(t, 20, cached)
	assert.EqualValues(t, 13, output)

	after := time.Now().Add(-time.Minute)

	updateSQL(t, db, `UPDATE sessions SET input_tokens=110, output_tokens=12 WHERE id='one'`)

	hh, err = (Hermes{After: after}).Parse(context.Background())
	require.NoError(t, err)

	totals = updateTotals(hh)
	input, _, output, _ = totals.input, totals.cached, totals.output, totals.prompt
	assert.EqualValues(t, 10, input)
	assert.EqualValues(t, 2, output)
	require.NotEmpty(t, hh)
	assert.GreaterOrEqual(t, hh[0].Time, heartbeatTimestamp(after))
	hh, err = (Hermes{After: after}).Parse(context.Background())
	require.NoError(t, err)
	assert.Empty(t, hh)
}

func TestProviderUpdatesCopilotSQLiteReconciliation(t *testing.T) {
	home := updateTestHome(t)
	db := updateDB(t, filepath.Join(home, ".copilot", "session-store.db"))
	updateSQL(t, db, updateFixture(t, "copilotsqlitereconciliation-1.txt"))

	end := time.Date(2026, 9, 20, 10, 0, 3, 0, time.UTC)
	timed := []copilotTimedHeartbeat{{timestamp: end, shutdown: true,
		heartbeat: heartbeat.Heartbeat{AISession: "one", AIInputTokens: 135, AICachedInputTokens: 25, AIOutputTokens: 7}}}
	got, err := (Copilot{}).reconcileSQLiteUsage(context.Background(), timed)
	require.NoError(t, err)

	var hh Heartbeats
	for _, row := range got {
		hh = append(hh, row.heartbeat)
	}

	totals := updateTotals(hh)
	input, cached, output, _ := totals.input, totals.cached, totals.output, totals.prompt
	assert.EqualValues(t, 135, input)
	assert.EqualValues(t, 25, cached)
	assert.EqualValues(t, 7, output)
	// A crash has no shutdown record: usage still comes from the database.
	got, err = (Copilot{}).reconcileSQLiteUsage(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.EqualValues(t, 0, got[0].heartbeat.AIOutputTokens)
	assert.EqualValues(t, 7, got[1].heartbeat.AIOutputTokens)
}

func TestProviderUpdatesKiroLayouts(t *testing.T) {
	home := updateTestHome(t)
	root := filepath.Join(home, ".kiro", "sessions")
	updateWrite(t, filepath.Join(root, "hash", "sess_one", "session.json"),
		`{"id":"ide","workspacePaths":["/project"],"modelId":"claude-4"}`)
	updateWrite(t, filepath.Join(root, "hash", "sess_one", "messages.jsonl"),
		`{"timestamp":"2026-09-20T10:00:00Z","payload":{"type":"user","content":"hello"}}
{"timestamp":"2026-09-20T10:00:01Z","payload":{"type":"assistant","content":"done"}}
`)
	updateWrite(t, filepath.Join(root, "cli", "one.json"), updateFixture(t, "kirolayouts-1.txt"))
	updateWrite(t, filepath.Join(root, "cli", "one.jsonl"),
		`{"kind":"Prompt","data":{"content":[{"kind":"text","data":"test"}]}}
{"kind":"AssistantMessage","data":{"content":[{"kind":"text","data":"done"}]}}
`)

	hh, err := (Kiro{}).Parse(context.Background())
	require.NoError(t, err)
	require.Len(t, hh, 3)
	totals := updateTotals(hh)
	input, _, output, prompt := totals.input, totals.cached, totals.output, totals.prompt
	assert.Zero(t, input)
	assert.Zero(t, output)
	assert.Equal(t, 9, prompt)

	for _, h := range hh {
		assert.Equal(t, "/project", h.ProjectPathOverride)
		assert.Contains(t, h.UserAgent, "claude/4")
	}
}

func TestProviderUpdatesFallbackTimestampIsStable(t *testing.T) {
	home := updateTestHome(t)
	path := filepath.Join(home, "events.jsonl")
	updateWrite(t, path, `{"id":"one"}`)
	clock, err := newTranscriptClock(context.Background(), path)
	require.NoError(t, err)

	first := clock.resolve("one", time.Time{})
	require.NoError(t, clock.save())

	next := first.Add(time.Hour)
	require.NoError(t, os.Chtimes(path, next, next))
	clock, err = newTranscriptClock(context.Background(), path)
	require.NoError(t, err)
	// JSON can change the location from Local to UTC without changing the instant.
	assert.WithinDuration(t, first, clock.resolve("one", time.Time{}), 0)

	var normalized map[string]any
	require.NoError(t, json.Unmarshal(clock.normalize([]byte(`{"timestamp":"broken","id":"two"}`)), &normalized))
	assert.False(t, genericAITime(normalized["timestamp"]).IsZero())
}

func TestProviderUpdatesGrokBotAndOpenClaude(t *testing.T) {
	home := updateTestHome(t)
	sliceKey := []byte("sand.client.slice.account.user.transcript.replicas.bot")
	key := strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(sliceKey))
	root := filepath.Join(home, "mirror")
	updateWrite(t, filepath.Join(root, key+".blob"), updateFixture(t, "grokbotandopenclaude-1.txt"))
	hh, err := (GrokBot{}).parseMirror(context.Background(), root)
	require.NoError(t, err)
	require.Len(t, hh, 3)
	totals := updateTotals(hh)
	input, _, output, prompt := totals.input, totals.cached, totals.output, totals.prompt
	assert.Zero(t, input)
	assert.Zero(t, output)
	assert.Equal(t, 5, prompt)
	updateWrite(t, filepath.Join(home, ".openclaude", "projects", "project", "one.jsonl"),
		updateFixture(t, "grokbotandopenclaude-2.txt"))

	hh, err = (OpenClaude{}).Parse(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, hh)
	assert.Contains(t, hh[0].UserAgent, "openclaude/")
	assert.Equal(t, "openclaude:one", hh[0].AISession)
}

func updateProtoBytes(number uint64, data []byte) []byte {
	result := binary.AppendUvarint(nil, number<<3|2)
	result = binary.AppendUvarint(result, uint64(len(data)))

	return append(result, data...)
}

func updateProtoInt(number, value uint64) []byte {
	return binary.AppendUvarint(binary.AppendUvarint(nil, number<<3), value)
}

func TestProviderUpdatesAntigravitySQLite(t *testing.T) {
	home := updateTestHome(t)
	path := filepath.Join(home, ".gemini", "antigravity", "conversations", "one.db")
	db := updateDB(t, path)
	updateSQL(t, db, `CREATE TABLE gen_metadata(idx INTEGER,data BLOB);CREATE TABLE steps(idx INTEGER,metadata BLOB);`)

	usage := append(updateProtoInt(2, 100), updateProtoInt(3, 20)...)
	usage = append(usage, updateProtoInt(9, 15)...)
	usage = append(usage, updateProtoInt(10, 5)...)
	usage = append(usage, updateProtoBytes(11, []byte("response-one"))...)
	chat := append(updateProtoBytes(4, usage), updateProtoBytes(19, []byte("gemini-3"))...)
	chat = append(chat, updateProtoBytes(9, updateProtoBytes(4, []byte("2026-09-20T10:00:00Z")))...)
	root := append(updateProtoBytes(1, chat), updateProtoBytes(2, []byte{1})...)
	updateSQL(t, db, `INSERT INTO gen_metadata VALUES(1,?),(2,?)`, root, root)

	args := []byte(`{"TargetFile":"/project/main.go","CodeContent":"one\ntwo"}`)
	tool := append(updateProtoBytes(2, []byte("write_to_file")), updateProtoBytes(3, args)...)
	updateSQL(t, db, `INSERT INTO steps VALUES(1,?),(2,?)`, updateProtoBytes(4, tool), updateProtoBytes(4, tool))

	hh, err := (Gemini{}).antigravitySQLite(context.Background(), home)
	require.NoError(t, err)

	totals := updateTotals(hh)
	input, _, output, _ := totals.input, totals.cached, totals.output, totals.prompt
	assert.EqualValues(t, 100, input)
	assert.EqualValues(t, 20, output)

	var files int

	for _, h := range hh {
		if h.EntityType == heartbeat.FileType {
			files++

			assert.Equal(t, "/project/main.go", h.Entity)
		}
	}

	assert.Equal(t, 1, files, "unreferenced steps and duplicate response IDs must not produce edits")
	assert.Nil(t, antigravityProto([]byte{10, 127, 1}))
}

func TestProviderUpdatesWSLDiscovery(t *testing.T) {
	names := parseWSLDistros([]byte("Ubuntu\r\ndocker-desktop\r\n../bad\r\nDebian\r\n"))
	assert.Equal(t, []string{"Ubuntu", "Debian"}, names)

	var raw []byte
	for _, r := range "Ubuntu\r\n" {
		raw = binary.LittleEndian.AppendUint16(raw, uint16(r))
	}

	assert.Equal(t, []string{"Ubuntu"}, parseWSLDistros(raw))
	home := updateTestHome(t)
	wsl := t.TempDir()
	updateWrite(t, filepath.Join(wsl, ".codex", "archived_sessions", "one.jsonl"),
		`{"type":"session_meta","payload":{"id":"wsl-one"}}`)
	ctx := context.WithValue(context.Background(), wslHomesKey{}, []string{wsl})
	paths, err := (Codex{}).transcriptPaths(ctx)
	require.NoError(t, err)
	require.Len(t, paths, 1)

	dirs, err := claudeConfigDirs(ctx)
	require.NoError(t, err)
	assert.Contains(t, dirs, filepath.Join(home, ".claude"))
	assert.Contains(t, dirs, filepath.Join(wsl, ".claude"))
}

func updateFixture(t *testing.T, name string) string {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("testdata", "provider_updates", name))
	require.NoError(t, err)

	return string(raw)
}

func TestProviderUpdatesCopilotOTel(t *testing.T) {
	home := updateTestHome(t)
	path := filepath.Join(home, "agent-traces.db")
	db := updateDB(t, path)
	updateSQL(t, db, `CREATE TABLE spans(span_id TEXT,trace_id TEXT,operation_name TEXT,
 start_time_ms INTEGER,response_model TEXT);
 CREATE TABLE span_attributes(span_id TEXT,key TEXT,value TEXT);
 INSERT INTO spans VALUES ('s1','trace','chat',1800000000000,'claude-4'),('tool','trace','execute_tool',1800000000000,'');
 INSERT INTO span_attributes VALUES ('s1','gen_ai.conversation.id','conversation'),
 ('s1','gen_ai.usage.input_tokens','100'),('s1','gen_ai.usage.cache_read.input_tokens','20'),
 ('s1','gen_ai.usage.output_tokens','5');`)

	parser := Copilot{}
	got, err := parser.parseOTelDB(context.Background(), path, nil, make(map[string]bool))
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.EqualValues(t, 80, got[0].heartbeat.AIInputTokens)
	assert.EqualValues(t, 20, got[0].heartbeat.AICachedInputTokens)
	assert.EqualValues(t, 5, got[0].heartbeat.AIOutputTokens)
	again, err := parser.parseOTelDB(context.Background(), path, got, make(map[string]bool))
	require.NoError(t, err)
	assert.Empty(t, again)
}

func TestProviderUpdatesPiMissingTimestamp(t *testing.T) {
	home := updateTestHome(t)
	path := filepath.Join(home, "pi.jsonl")
	updateWrite(t, path, updateFixture(t, "pi-missing-timestamp.txt"))
	hh, err := (Pi{}).parseTranscript(context.Background(), path)
	require.NoError(t, err)
	require.NotEmpty(t, hh)
	totals := updateTotals(hh)
	assert.EqualValues(t, 10, totals.input)
	assert.EqualValues(t, 5, totals.output)

	stamp := heartbeatTime(hh[0].Time)
	future := stamp.Add(time.Hour)
	require.NoError(t, os.Chtimes(path, future, future))
	again, err := (Pi{After: stamp.Add(time.Second)}).parseTranscript(context.Background(), path)
	require.NoError(t, err)
	assert.Empty(t, again, "mtime growth must not replay an unstamped message")
}

func TestProviderUpdatesCodexInheritedUsageDoesNotSuppressParent(t *testing.T) {
	stamp := time.Date(2026, time.September, 1, 12, 0, 0, 0, time.UTC)

	var line codexLogLine
	require.NoError(t, json.Unmarshal([]byte(`{"payload":{"type":"token_count","info":{
"total_token_usage":{"input_tokens":100,"output_tokens":20}}}}`), &line))
	line.Timestamp = stamp
	seen := make(map[string]bool)
	child := codexParseState{seenUsage: seen, usageOwner: "parent"}
	child.trackTokenCount(line, stamp.Add(time.Minute))

	parent := codexParseState{seenUsage: seen, usageOwner: "parent", heartbeats: Heartbeats{{}}}
	parent.trackTokenCount(line, time.Time{})
	assert.EqualValues(t, 100, parent.heartbeats[0].AIInputTokens)
	assert.EqualValues(t, 20, parent.heartbeats[0].AIOutputTokens)
	assert.EqualValues(t, 100, child.tokens.LastInput)
}

func TestHermesOnlyReadsSessionStores(t *testing.T) {
	home := updateTestHome(t)
	root := filepath.Join(home, "custom hermes")
	t.Setenv("HERMES_HOME", root)

	for _, name := range []string{"state.db", "profiles/work/state.db", "profiles/personal/state.db",
		"state-snapshots/old/state.db", "profiles/work/state-snapshots/old/state.db",
		"chrome-cdp-profile/Default/browser.db", "cron/executions.db", "projects.db"} {
		db := updateDB(t, filepath.Join(root, filepath.FromSlash(name)))
		updateSQL(t, db, updateFixture(t, "hermesgrowth-1.txt"))
	}

	hh, err := (Hermes{}).Parse(t.Context())
	require.NoError(t, err)

	totals := updateTotals(hh)
	assert.EqualValues(t, 3*105, totals.input)
}
