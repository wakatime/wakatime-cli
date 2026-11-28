package ai

import (
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

func TestSprocketSparsePartsAndToolResults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("SPROCKET_DATA_DIR", "")

	thread := filepath.Join(home, ".sprocket", "transcripts", "user", "thread")
	workspace := filepath.Join(home, "project")

	for _, name := range []string{"00000000.jsonl", "00000200.jsonl"} {
		raw, err := os.ReadFile(filepath.Join("testdata", "sprocket", name))
		require.NoError(t, err)
		sprocketWrite(t, filepath.Join(thread, "parts", name),
			strings.ReplaceAll(string(raw), "/workspace", filepath.ToSlash(workspace)))
	}

	parser := Sprocket{FallbackUserAgent: "Sprocket/0.5.3 sprocket-wakatime/0.5.3"}
	got, err := parser.Parse(t.Context())
	require.NoError(t, err)
	require.Empty(t, got, "downloaded transcripts are not local activity")

	sprocketWriteRun(t, thread, "local-run", workspace, "0.5.3")
	got, err = parser.Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 6)
	assert.Equal(t, 10, got[0].AIPromptLength)
	assert.Equal(t, 1767355200.123, got[0].Time)
	assert.Equal(t, 1767355201.456, got[1].Time, "use item time, not completion persistence time")
	assert.Equal(t, 1767355206.789, got[3].Time)
	assert.Equal(t, "Sprocket thread", got[0].Entity)

	for _, h := range got {
		assert.Equal(t, "thread", h.AISession)
		assert.Equal(t, heartbeat.AICodingCategory.String(), h.Category)
		assert.Equal(t, workspace, h.ProjectPathOverride)
		assert.Equal(t, int64(0), h.AIInputTokens)
		assert.Equal(t, int64(0), h.AICachedInputTokens)
		assert.Equal(t, int64(0), h.AIOutputTokens)
	}

	assert.Equal(t, "gpt/5.4 Sprocket/0.5.3 sprocket-wakatime/0.5.3", got[4].UserAgent)
	assert.Equal(t, filepath.ToSlash(filepath.Join(workspace, "main.go")), got[4].Entity)
	require.NotNil(t, got[4].AILineChanges)
	assert.Equal(t, -1, *got[4].AILineChanges)
	assert.Equal(t, heartbeat.FileType, got[4].EntityType)
	require.NotNil(t, got[4].IsWrite)
	assert.True(t, *got[4].IsWrite)
	assert.Equal(t, filepath.ToSlash(filepath.Join(workspace, "new.go")), got[5].Entity)
	require.NotNil(t, got[5].AILineChanges)
	assert.Equal(t, 2, *got[5].AILineChanges)
}

func TestSprocketRunMetadataAndCwdBeforeCutoff(t *testing.T) {
	root := t.TempDir()
	t.Setenv("SPROCKET_DATA_DIR", root)
	thread := filepath.Join(root, "transcripts", "user", "thread")
	first := filepath.Join(root, "first")
	second := filepath.Join(root, "second")
	changed := filepath.Join(root, "changed")

	sprocketWriteRun(t, thread, "first-run", first, "0.5.3")
	sprocketWriteRun(t, thread, "second-run", second, "0.6.0")
	sprocketWriteRun(t, thread, "invalid-run", "relative", "0.6.0")

	sprocketWriteParts(t, filepath.Join(thread, "parts", "00000000.jsonl"), []sprocketPart{
		{Kind: "tool", RunID: "first-run", CreatedAt: 1000, Tool: &sprocketTool{
			ToolInvocationID: "cwd", Name: "set_cwd", Status: "completed",
			Output: sprocketJSON(t, map[string]any{"cwd": changed}),
		}},
		{Kind: "tool", RunID: "first-run", CreatedAt: 2000, Tool: &sprocketTool{
			ToolInvocationID: "read", Name: "parse_file", Status: "started",
			Input: sprocketJSON(t, map[string]any{"path": "main.go"}),
		}},
		{Kind: "tool", RunID: "first-run", CreatedAt: 2250, Tool: &sprocketTool{
			ToolInvocationID: "cwd-again", Name: "set_cwd", Status: "completed",
			Output: sprocketJSON(t, map[string]any{"cwd": first}),
		}},
		{Kind: "tool", RunID: "first-run", CreatedAt: 3000, Tool: &sprocketTool{
			ToolInvocationID: "read", Status: "completed",
		}},
		{Kind: "tool", RunID: "second-run", CreatedAt: 4000, Tool: &sprocketTool{
			ToolInvocationID: "read", Name: "parse_file", Status: "completed",
			Input: sprocketJSON(t, map[string]any{"path": "other.go"}),
		}},
		{Kind: "tool", RunID: "remote-run", CreatedAt: 5000, Tool: &sprocketTool{Status: "completed"}},
		{Kind: "tool", RunID: "invalid-run", CreatedAt: 5000, Tool: &sprocketTool{Status: "completed"}},
		{Kind: "tool", RunID: "first-run", Tool: &sprocketTool{
			ToolInvocationID: "unstamped", Status: "completed",
		}},
	})

	got, err := (Sprocket{After: time.UnixMilli(2500)}).Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 4)
	assert.Equal(t, filepath.ToSlash(filepath.Join(changed, "main.go")), got[1].Entity)
	assert.Equal(t, changed, got[1].ProjectPathOverride)
	assert.Equal(t, "Sprocket/0.5.3", got[1].UserAgent)
	assert.Equal(t, filepath.ToSlash(filepath.Join(second, "other.go")), got[3].Entity)
	assert.Equal(t, second, got[3].ProjectPathOverride)
	assert.Equal(t, "Sprocket/0.6.0", got[3].UserAgent)
}

func TestSprocketPatchOperations(t *testing.T) {
	root := t.TempDir()
	t.Setenv("SPROCKET_DATA_DIR", root)
	thread := filepath.Join(root, "transcripts", "user", "thread")
	sprocketWriteRun(t, thread, "run", root, "0.5.3")

	patch := "*** Begin Patch\n*** Delete File: deleted.go\n*** Update File: old.go\n" +
		"*** Move to: moved.go\n@@\n-old\n+new\n+extra\n*** Copy File: source.go\n" +
		"*** Copy to: copied.go\n*** Update File: reduced.go\n@@\n-a\n-b\n-c\n+one\n*** End Patch"
	changes := []map[string]any{
		{"path": filepath.Join(root, "deleted.go"), "operation": "deleted"},
		{"path": filepath.Join(root, "moved.go"), "operation": "renamed"},
		{"path": filepath.Join(root, "copied.go"), "operation": "copied"},
		{"path": filepath.Join(root, "reduced.go"), "operation": "updated"},
	}
	sprocketWriteParts(t, filepath.Join(thread, "parts", "00000000.jsonl"), []sprocketPart{
		{Kind: "tool", RunID: "run", CreatedAt: 1000, Tool: &sprocketTool{
			ToolInvocationID: "patch", Name: "apply_patch", Status: "completed",
			Input:  sprocketJSON(t, map[string]any{"patch": patch}),
			Output: sprocketJSON(t, map[string]any{"changes": changes}),
		}},
		{Kind: "tool", RunID: "run", CreatedAt: 2000, Tool: &sprocketTool{
			ToolInvocationID: "failed-patch", Name: "apply_patch", Status: "completed",
			Input:  sprocketJSON(t, map[string]any{"patch": patch}),
			Output: sprocketJSON(t, map[string]any{"success": false, "changes": changes}),
		}},
	})
	got, err := (Sprocket{}).Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 6)
	assert.True(t, got[1].IsUnsavedEntity)
	assert.Nil(t, got[1].AILineChanges, "the deleted file's original contents are unavailable")
	require.NotNil(t, got[2].AILineChanges)
	assert.Equal(t, 1, *got[2].AILineChanges)
	assert.Nil(t, got[3].AILineChanges, "the copied file's full contents are unavailable")
	require.NotNil(t, got[4].AILineChanges)
	assert.Equal(t, -2, *got[4].AILineChanges)
}

func TestSprocketSyncRollingChunk(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("SPROCKET_DATA_DIR", home)
	thread := filepath.Join(home, "transcripts", "user", "thread")
	sprocketWriteRun(t, thread, "run", home, "0.5.3")
	path := filepath.Join(thread, "parts", "00000400.jsonl")
	line := `{"number":400,"kind":"prompt","runId":"run","createdAt":1767355200123,"prompt":{"text":"First"}}` + "\n"
	sprocketWrite(t, path, line)

	v := viper.New()
	v.Set("internal-config", filepath.Join(home, "internal.cfg"))

	after := time.UnixMilli(1767355200000)
	config := Config{V: v}
	got, err := parseAIHeartbeats(t.Context(), after, nil, config)
	require.NoError(t, err)
	require.Len(t, got, 1)

	got, err = parseAIHeartbeats(t.Context(), after, nil, config)
	require.NoError(t, err)
	require.Empty(t, got)

	line += `{"number":401,"kind":"prompt","runId":"run","createdAt":1767355200123,"prompt":{"text":"Second"}}` + "\n"
	sprocketWrite(t, path, line)
	got, err = parseAIHeartbeats(t.Context(), after, nil, config)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 6, got[0].AIPromptLength)

	sprocketWrite(t, filepath.Join(thread, "parts", "00000000.jsonl"),
		`{"number":0,"kind":"prompt","runId":"remote","createdAt":1767355300000,"prompt":{"text":"Downloaded"}}`+"\n")
	got, err = parseAIHeartbeats(t.Context(), after, nil, config)
	require.NoError(t, err)
	require.Empty(t, got)
}

func sprocketWriteRun(t *testing.T, thread, run, workspace, version string) {
	t.Helper()
	sprocketWrite(t, filepath.Join(thread, "wakatime", run+".json"),
		string(sprocketJSON(t, map[string]any{"workspacePath": workspace, "version": version})))
}

func sprocketWriteParts(t *testing.T, path string, parts []sprocketPart) {
	t.Helper()

	var text strings.Builder
	for _, part := range parts {
		text.Write(sprocketJSON(t, part))
		text.WriteByte('\n')
	}

	sprocketWrite(t, path, text.String())
}

func sprocketJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()

	raw, err := json.Marshal(value)
	require.NoError(t, err)

	return raw
}

func sprocketWrite(t *testing.T, path, text string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte(text), 0o600))
}
