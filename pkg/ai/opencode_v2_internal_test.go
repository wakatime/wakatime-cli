//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestOpenCodeV2Tools(t *testing.T) {
	// Output shapes from packages/core/src/tool/{write,edit,apply-patch}.ts
	// at anomalyco/opencode commit 0f549842ee746e400b1f72516b0b2e292e267e2c.
	for _, tt := range []struct {
		name, tool, input, structured, file string
		changes                             int
		deleted                             bool
	}{
		{"new write", "write", `{"path":"new.go","content":"one\ntwo"}`,
			`{"operation":"write","target":"/project/new.go","resource":"new.go","existed":false}`,
			"new.go", 2, false},
		{"existing write", "write", `{"path":"existing.go","content":"one\ntwo"}`,
			`{"operation":"write","target":"/project/existing.go","resource":"existing.go","existed":true}`,
			"existing.go", 0, false},
		{"edit all occurrences", "edit", `{"path":"edited.go","oldString":"one","newString":"one\ntwo"}`,
			`{"files":[{"file":"edited.go","status":"modified","additions":6,"deletions":3}],"replacements":3}`,
			"edited.go", 3, false},
		{"edit input fallback", "edit", `{"path":"fallback.go","oldString":"one","newString":"one\ntwo"}`,
			`{}`, "fallback.go", 1, false},
		{"write input fallback", "write", `{"path":"fallback.go","content":"one"}`,
			`{}`, "fallback.go", 1, false},
		{"patch deletion", "apply_patch", `{"patchText":"*** Begin Patch\n*** Delete File: gone.go\n*** End Patch"}`,
			`{"applied":[{"type":"delete","resource":"gone.go","target":"/project/gone.go"}],` +
				`"files":[{"file":"gone.go","status":"deleted","additions":0,"deletions":8}]}`,
			"gone.go", -8, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			raw := `{"agent":"build","model":{"id":"test-model","providerID":"test"},"content":[` +
				`{"id":"call-1","type":"tool","name":"` + tt.tool + `","time":{"created":1000},` +
				`"state":{"status":"completed","input":` + tt.input + `,"structured":` + tt.structured +
				`,"content":[]}}]}`
			message, ok := openCodeV2Message("msg-1", "session", "assistant", 1000, raw)
			require.True(t, ok)

			hh := (OpenCode{}).sessionHeartbeats(openCodeSessionInfo{ID: "session", Directory: "/project"},
				[]openCodeMessageWithParts{message})
			require.Len(t, hh, 2)
			assert.Equal(t, heartbeat.FileType, hh[1].EntityType)
			assert.Equal(t, "/project/"+tt.file, filepath.ToSlash(hh[1].Entity))
			assert.Equal(t, &tt.changes, hh[1].AILineChanges)
			assert.Equal(t, tt.deleted, hh[1].IsUnsavedEntity)
		})
	}
}

func TestOpenCodeV2SharedSessionTable(t *testing.T) {
	home := updateTestHome(t)
	dbPath := filepath.Join(home, ".local", "share", "opencode", "opencode.db")
	db := updateDB(t, dbPath)
	updateSQL(t, db, updateFixture(t, "opencodemixedgenerations-1.txt"))

	// A migrated session's frozen legacy message must not reappear when its
	// v2 messages are all older than the cutoff. Legacy-only sessions still work.
	updateSQL(t, db, `UPDATE message SET time_created = 1800000003000`)

	hh, err := (OpenCode{After: time.UnixMilli(1800000003000)}).parseSQLiteDB(context.Background(), dbPath)
	require.NoError(t, err)
	require.Len(t, hh, 1)
	assert.Equal(t, "OpenCode legacy", hh[0].Entity)
	assert.EqualValues(t, 7, hh[0].AIInputTokens)

	// V2 parsing must also work with no legacy message/part tables at all.
	updateSQL(t, db, `DROP TABLE message; DROP TABLE part;`)
	updateSQL(t, db, `INSERT INTO session_message VALUES
 ('prompt','child','user',3,1800000003000,?),
 ('bad','child','assistant',4,1800000004000,'{broken')`,
		`{"text":"Fix it","files":[],"agents":[],"time":{"created":1800000003000}}`)

	hh, err = (OpenCode{}).Parse(context.Background())
	require.NoError(t, err)
	require.Len(t, hh, 4)
	totals := updateTotals(hh)
	assert.EqualValues(t, 125, totals.input)
	assert.EqualValues(t, 17, totals.output)
	assert.Equal(t, 6, totals.prompt)

	for _, h := range hh {
		assert.Contains(t, h.UserAgent, "opencode-cli/2")
	}
}

func TestOpenCodeV2Compaction(t *testing.T) {
	// Compaction records are summaries, not assistant responses with usage.
	raw := `{"reason":"auto","summary":"Earlier work","recent":"Latest work","time":{"created":1000}}`
	require.True(t, json.Valid([]byte(raw)))
	_, ok := openCodeV2Message("compact", "session", "compaction", 1000, raw)
	assert.False(t, ok)
}

func TestOpenCodeV2SkipsLegacyFiles(t *testing.T) {
	home := updateTestHome(t)
	root := filepath.Join(home, ".local", "share", "opencode")
	db := updateDB(t, filepath.Join(root, "opencode.db"))
	updateSQL(t, db, updateFixture(t, "opencodemixedgenerations-1.txt"))
	updateWrite(t, filepath.Join(root, "storage", "session", "project", "migrated.json"),
		`{"id":"migrated","directory":"/project"}`)
	updateWrite(t, filepath.Join(root, "storage", "message", "migrated", "old.json"),
		`{"id":"old","sessionID":"migrated","role":"assistant","time":{"created":1800000000000},`+
			`"tokens":{"input":999,"output":999}}`)

	hh, err := (OpenCode{}).Parse(context.Background())
	require.NoError(t, err)
	assert.EqualValues(t, 132, updateTotals(hh).input)
	assert.EqualValues(t, 20, updateTotals(hh).output)
}

func TestOpenCodeV2MultiplePatchFilesNativePaths(t *testing.T) {
	cwd := t.TempDir()
	external := filepath.Join(t.TempDir(), "external.go")
	structured := mustJSON(t, map[string]any{
		"files": []map[string]any{
			{"file": "nested/new.go", "status": "added", "additions": 3, "deletions": 0},
			{"file": filepath.ToSlash(external), "status": "modified", "additions": 4, "deletions": 2},
			{"file": "gone.go", "status": "deleted", "additions": 0, "deletions": 5},
		},
	})
	hh := (OpenCode{}).toolHeartbeats("2", "model", "session", cwd, time.UnixMilli(1000), openCodePart{
		Type: "tool", Tool: "apply_patch", State: &openCodeToolState{
			Status: "completed", Structured: json.RawMessage(structured),
		},
	})
	require.Len(t, hh, 3)
	assert.Equal(t, filepath.ToSlash(filepath.Join(cwd, "nested", "new.go")), filepath.ToSlash(hh[0].Entity))
	assert.Equal(t, filepath.ToSlash(external), filepath.ToSlash(hh[1].Entity))
	assert.Equal(t, filepath.ToSlash(filepath.Join(cwd, "gone.go")), filepath.ToSlash(hh[2].Entity))
	assert.Equal(t, heartbeat.PointerTo(3), hh[0].AILineChanges)
	assert.Equal(t, heartbeat.PointerTo(2), hh[1].AILineChanges)
	assert.Equal(t, heartbeat.PointerTo(-5), hh[2].AILineChanges)
	assert.True(t, hh[2].IsUnsavedEntity)
}

func TestOpenCodeV2UnsuccessfulTools(t *testing.T) {
	for _, status := range []string{"pending", "running", "error"} {
		t.Run(status, func(t *testing.T) {
			input := `{"path":"main.go","content":"hello"}`
			if status == "pending" {
				input = `""`
			}

			raw := `{"content":[{"type":"tool","name":"write","state":{"status":"` + status +
				`","input":` + input + `,"structured":{},"content":[],` +
				`"error":{"type":"unknown","message":"failed"}}}]}`
			message, ok := openCodeV2Message("msg", "session", "assistant", 1000, raw)
			require.True(t, ok)

			hh := (OpenCode{}).sessionHeartbeats(openCodeSessionInfo{ID: "session", Directory: t.TempDir()},
				[]openCodeMessageWithParts{message})
			require.Len(t, hh, 1)
			assert.Equal(t, heartbeat.AppType, hh[0].EntityType)
		})
	}
}
