package ai

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
)

func TestParserIDStringAndPlugins(t *testing.T) {
	assert.Equal(t, "", UnknownParser.String())
	assert.Equal(t, claudeParserString, ClaudeParser.String())
	assert.Equal(t, codexParserString, CodexParser.String())
	assert.Equal(t, cursorParserString, CursorParser.String())
	assert.Equal(t, "", ParserID(99).String())

	assert.Equal(t, "ClaudeCode", claudePlugin(""))
	assert.Equal(t, "ClaudeCode/1.2.3", claudePlugin("1.2.3"))
	assert.Equal(t, "Codex", codexPlugin(""))
	assert.Equal(t, "Codex/1.2.3", codexPlugin("1.2.3"))
	assert.Equal(t, "Cursor", cursorPlugin())
}

func TestEntityUserAgentsAndAIUserAgent(t *testing.T) {
	ctx := context.Background()

	userAgents := entityUserAgents([]heartbeat.Heartbeat{
		{Entity: "/tmp/main.go", UserAgent: "editor/1.0.0"},
		{Entity: "", UserAgent: "skip-empty-entity"},
		{Entity: "/tmp/skip.go", UserAgent: ""},
	})

	assert.Equal(t, map[string]string{"/tmp/main.go": "editor/1.0.0"}, userAgents)
	assert.Equal(
		t,
		heartbeat.UserAgent(ctx, cursorPlugin()+" "+"editor/1.0.0"),
		aiUserAgent(ctx, "/tmp/main.go", userAgents, "plugin/0.1.0", cursorPlugin()),
	)
	assert.Equal(
		t,
		heartbeat.UserAgent(ctx, cursorPlugin()+" "+"plugin/0.1.0"),
		aiUserAgent(ctx, "/tmp/other.go", userAgents, "plugin/0.1.0", cursorPlugin()),
	)
	assert.Equal(
		t,
		heartbeat.UserAgent(ctx, cursorPlugin()),
		aiUserAgent(ctx, "/tmp/other.go", userAgents, "", cursorPlugin()),
	)
}

func TestGetLastParsedAt(t *testing.T) {
	ctx := context.Background()

	t.Run("errors when viper is missing", func(t *testing.T) {
		parsed, err := getLastParsedAt(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing viper instance")
		assert.WithinDuration(t, time.Now().Add(-time.Minute), parsed, 2*time.Second)
	})

	t.Run("clamps future timestamp and writes updated value", func(t *testing.T) {
		tmpInternal, err := os.CreateTemp(t.TempDir(), "wakatime-internal")
		require.NoError(t, err)

		defer tmpInternal.Close()

		v := viper.New()
		v.Set("internal-config", tmpInternal.Name())
		v.Set("internal.ai_heartbeats_last_parsed_at", time.Now().Add(time.Hour).Format(ini.DateFormat))

		before := time.Now()
		parsed, err := getLastParsedAt(ctx, v)
		after := time.Now()

		require.NoError(t, err)
		assert.False(t, parsed.Before(before))
		assert.False(t, parsed.After(after))

		writer, err := ini.NewWriter(ctx, v, ini.InternalFilePath)
		require.NoError(t, err)
		require.NoError(t, writer.File.Reload())

		written, err := writer.File.Section("internal").Key("ai_heartbeats_last_parsed_at").TimeFormat(ini.DateFormat)
		require.NoError(t, err)
		assert.WithinDuration(t, time.Now(), written, 2*time.Second)
	})
}

func TestCursorHelpers(t *testing.T) {
	t.Run("file path prefers explicit path then code block", func(t *testing.T) {
		assert.Equal(t, "/tmp/main.go", cursorFilePath("/tmp/main.go", nil))
		assert.Equal(t, "/tmp/from-block.go", cursorFilePath("", []cursorCodeBlock{{
			URI: &cursorURI{FSPath: "/tmp/from-block.go"},
		}}))
		assert.Equal(t, "", cursorFilePath("", nil))
	})

	t.Run("line changes detect text and unified diff", func(t *testing.T) {
		assert.Equal(t, 0, cursorLineChanges(" \n\t "))
		assert.Equal(t, 2, cursorLineChanges("first\n\nsecond"))
		assert.True(t, cursorLooksLikeUnifiedDiff("--- a/main.go\n+++ b/main.go"))
		assert.True(t, cursorLooksLikeUnifiedDiff("prefix\n@@ -1 +1 @@"))
		assert.False(t, cursorLooksLikeUnifiedDiff("plain text"))
		assert.Equal(t, 1, cursorLineChangesFromDiff("--- a\n+++ b\n-old\n+new\n+extra"))
	})
}

func TestCursorHeartbeatFallbacks(t *testing.T) {
	ctx := context.Background()
	parser := Cursor{FallbackUserAgent: "plugin/0.1.0"}
	createdAt := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)

	editHeartbeats := parser.cursorHeartbeats(ctx, cursorLogLine{
		CreatedAt: createdAt,
		Type:      2,
		ToolFormerData: &cursorToolFormerData{
			Name:    "edit_file",
			Params:  `{}`,
			RawArgs: `{"target_file":"/tmp/raw-edit.go","code_edit":"one\ntwo"}`,
			Status:  "completed",
		},
	})
	require.Len(t, editHeartbeats, 1)
	edit := &editHeartbeats[0]
	require.NotNil(t, edit)
	assert.Equal(t, "/tmp/raw-edit.go", edit.Entity)
	require.NotNil(t, edit.AILineChanges)
	assert.Equal(t, 2, *edit.AILineChanges)
	require.NotNil(t, edit.IsWrite)
	assert.True(t, *edit.IsWrite)

	readHeartbeats := parser.cursorHeartbeats(ctx, cursorLogLine{
		CreatedAt: createdAt,
		Type:      2,
		ToolFormerData: &cursorToolFormerData{
			Name:    "read_file",
			Params:  `{}`,
			RawArgs: `{}`,
			Status:  "completed",
		},
		CodeBlocks: []cursorCodeBlock{{
			URI: &cursorURI{FSPath: "/tmp/from-read-block.go"},
		}},
	})
	require.Len(t, readHeartbeats, 1)
	read := &readHeartbeats[0]
	require.NotNil(t, read)
	assert.Equal(t, "/tmp/from-read-block.go", read.Entity)
	require.NotNil(t, read.IsWrite)
	assert.False(t, *read.IsWrite)

	appHeartbeats := parser.cursorHeartbeats(ctx, cursorLogLine{
		BubbleID:  "composer-1",
		CreatedAt: createdAt,
		Type:      1,
		Text:      "Please edit the file",
	})
	require.Len(t, appHeartbeats, 1)
	assert.Equal(t, "composer-1", appHeartbeats[0].Entity)
	assert.Equal(t, heartbeat.AppType, appHeartbeats[0].EntityType)
	assert.Nil(t, appHeartbeats[0].AILineChanges)
	require.NotNil(t, appHeartbeats[0].IsWrite)
	assert.False(t, *appHeartbeats[0].IsWrite)

	assert.Nil(t, parser.cursorFileHeartbeat(ctx, cursorLogLine{
		CreatedAt: createdAt,
		Type:      2,
		ToolFormerData: &cursorToolFormerData{
			Name:   "unknown_tool",
			Status: "completed",
		},
	}))
}

func TestClaudeHelpers(t *testing.T) {
	t.Run("content unmarshal and line counting", func(t *testing.T) {
		var str contentValue
		require.NoError(t, json.Unmarshal([]byte(`"first\nsecond"`), &str))
		assert.Equal(t, 2, str.lineChanges())

		var arr contentValue
		require.NoError(t, json.Unmarshal([]byte(`[{"text":"first\nsecond"},"third"]`), &arr))
		assert.Equal(t, 3, arr.lineChanges())

		var unsupported contentValue
		require.Error(t, json.Unmarshal([]byte(`{"bad":true}`), &unsupported))
		assert.Equal(t, 0, ((*contentValue)(nil)).lineChanges())
		assert.Equal(t, 0, (&contentValue{}).lineChanges())
	})

	t.Run("tool use result supports object and string", func(t *testing.T) {
		var object toolUseResultValue
		require.NoError(t, json.Unmarshal([]byte(`{"filePath":"/tmp/main.go"}`), &object))
		require.NotNil(t, object.Object)
		assert.Equal(t, "/tmp/main.go", *object.Object.FilePath)

		var str toolUseResultValue
		require.NoError(t, json.Unmarshal([]byte(`"plain string result"`), &str))
		require.NotNil(t, str.String)
		assert.Equal(t, "plain string result", *str.String)

		var unsupported toolUseResultValue
		require.Error(t, json.Unmarshal([]byte(`123`), &unsupported))
	})

	t.Run("file path and line changes resolve expected sources", func(t *testing.T) {
		assert.Equal(t, "/tmp/direct.go", getClaudeFilePath(toolUseResult{
			FilePath: heartbeat.PointerTo("/tmp/direct.go"),
		}))
		assert.Equal(t, "/tmp/nested.go", getClaudeFilePath(toolUseResult{
			File: &toolUseResultFile{FilePath: heartbeat.PointerTo("/tmp/nested.go")},
		}))
		assert.Equal(t, "", getClaudeFilePath(toolUseResult{}))

		assert.Equal(t, 1, claudeLineChanges(toolUseResult{
			StructuredPatch: &[]structuredPatch{{OldLines: 1, NewLines: 2}},
		}))
		assert.Equal(t, 2, claudeLineChanges(toolUseResult{
			Content: &contentValue{String: heartbeat.PointerTo("one\ntwo")},
		}))
		assert.Equal(t, 3, claudeLineChanges(toolUseResult{
			File: &toolUseResultFile{Content: &contentValue{String: heartbeat.PointerTo("one\ntwo\nthree")}},
		}))
		assert.Equal(t, 0, claudeLineChanges(toolUseResult{
			Content:      &contentValue{String: heartbeat.PointerTo("one")},
			OriginalFile: heartbeat.PointerTo("before"),
		}))
		assert.Equal(t, 1, claudeAppLineChanges(&toolUseResultValue{
			String: heartbeat.PointerTo("summary"),
		}))
		assert.Equal(t, 2, claudeAppLineChanges(&toolUseResultValue{
			Object: &toolUseResult{
				Content: &contentValue{String: heartbeat.PointerTo("one\ntwo")},
			},
		}))
		assert.Equal(t, 0, claudeAppLineChanges(&toolUseResultValue{
			Object: &toolUseResult{
				FilePath: heartbeat.PointerTo("/tmp/skip.go"),
				Content:  &contentValue{String: heartbeat.PointerTo("one\ntwo")},
			},
		}))
		assert.Equal(t, 3, countStringLines("one\ntwo\nthree"))
	})
}

func TestCodexHelpers(t *testing.T) {
	ctx := context.Background()
	timestamp := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)

	assert.Nil(t, getCodexEntities(ctx, timestamp, "session.jsonl", "1.2.3", "/workspace", nil, "", codexPayload{}))
	assert.Nil(t, getCodexEntities(ctx, timestamp, "session.jsonl", "1.2.3", "/workspace", nil, "", codexPayload{
		Name:  heartbeat.PointerTo("not_apply_patch"),
		Input: heartbeat.PointerTo("*** Update File: pkg/main.go\n+one"),
	}))

	userHeartbeats := getCodexEntities(
		ctx,
		timestamp,
		"session.jsonl",
		"1.2.3",
		"/workspace",
		nil,
		"plugin/0.1.0",
		codexPayload{
			Type: "message",
			Role: heartbeat.PointerTo("user"),
			Content: []codexContentItem{
				{Type: "input_text", Text: "Please implement this"},
			},
		},
	)
	require.Len(t, userHeartbeats, 1)
	assert.Equal(t, "session.jsonl", userHeartbeats[0].Entity)
	assert.Equal(t, heartbeat.AppType, userHeartbeats[0].EntityType)
	assert.Nil(t, userHeartbeats[0].AILineChanges)
	require.NotNil(t, userHeartbeats[0].IsWrite)
	assert.False(t, *userHeartbeats[0].IsWrite)
	assert.Contains(t, userHeartbeats[0].UserAgent, "Codex/1.2.3")

	assistantHeartbeats := getCodexEntities(
		ctx,
		timestamp,
		"session.jsonl",
		"1.2.3",
		"/workspace",
		nil,
		"plugin/0.1.0",
		codexPayload{
			Type: "message",
			Role: heartbeat.PointerTo("assistant"),
			Content: []codexContentItem{
				{Type: "output_text", Text: "I am on it"},
			},
		},
	)
	require.Len(t, assistantHeartbeats, 1)
	assert.Equal(t, "session.jsonl", assistantHeartbeats[0].Entity)
	assert.Equal(t, heartbeat.AppType, assistantHeartbeats[0].EntityType)
	assert.Nil(t, assistantHeartbeats[0].AILineChanges)
	require.NotNil(t, assistantHeartbeats[0].IsWrite)
	assert.False(t, *assistantHeartbeats[0].IsWrite)

	heartbeats := getCodexEntities(
		ctx,
		timestamp,
		"session.jsonl",
		"1.2.3",
		"/workspace",
		nil,
		"plugin/0.1.0",
		codexPayload{
			Name: heartbeat.PointerTo("apply_patch"),
			Input: heartbeat.PointerTo(
				"*** Update File: pkg/main.go\n+one\n-two\n*** Add File: /tmp/extra.go\n+alpha\n+beta",
			),
		},
	)
	require.Len(t, heartbeats, 2)
	assert.Equal(t, filepath.Join("/workspace", "pkg/main.go"), heartbeats[0].Entity)
	require.NotNil(t, heartbeats[0].AILineChanges)
	assert.Equal(t, 0, *heartbeats[0].AILineChanges)
	assert.Contains(t, heartbeats[0].UserAgent, "Codex/1.2.3")

	assert.Equal(t, "/tmp/extra.go", heartbeats[1].Entity)
	require.NotNil(t, heartbeats[1].AILineChanges)
	assert.Equal(t, 2, *heartbeats[1].AILineChanges)

	assert.Equal(t, filepath.Join("/workspace", "pkg/main.go"), codexFilePath("/workspace", "*** Update File: pkg/main.go"))
	assert.Equal(t, "/tmp/main.go", codexFilePath("/workspace", "*** Add File: /tmp/main.go"))
	assert.Equal(t, "", codexFilePath("/workspace", "*** Move to: pkg/main.go"))
}
