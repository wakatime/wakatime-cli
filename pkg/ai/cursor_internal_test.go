//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestCursorHelpers(t *testing.T) {
	t.Run("cutoff includes a four second overlap", func(t *testing.T) {
		cutoff := time.Date(2026, 7, 15, 19, 6, 59, 531000000, time.UTC)

		assert.Equal(t, cutoff.Add(-4*time.Second), Cursor{After: cutoff}.bufferedCutoff())
		assert.True(t, Cursor{}.bufferedCutoff().IsZero())
	})

	t.Run("file path prefers explicit path then code block", func(t *testing.T) {
		assert.Equal(t, "/tmp/main.go", Cursor{}.filePath("/tmp/main.go", nil))
		assert.Equal(t, "/tmp/from-block.go", Cursor{}.filePath("", []cursorCodeBlock{{
			URI: &cursorURI{FSPath: "/tmp/from-block.go"},
		}}))
		assert.Equal(t, "", Cursor{}.filePath("", nil))
	})

	t.Run("line changes detect text and unified diff", func(t *testing.T) {
		assert.Equal(t, 0, Cursor{}.lineChanges(" \n\t "))
		assert.Equal(t, 2, Cursor{}.lineChanges("first\n\nsecond"))
		assert.True(t, Cursor{}.looksLikeUnifiedDiff("--- a/main.go\n+++ b/main.go"))
		assert.True(t, Cursor{}.looksLikeUnifiedDiff("prefix\n@@ -1 +1 @@"))
		assert.False(t, Cursor{}.looksLikeUnifiedDiff("plain text"))
		assert.Equal(t, 1, Cursor{}.lineChangesFromDiff("--- a\n+++ b\n-old\n+new\n+extra"))
	})

	t.Run("token counts derive output from total minus input", func(t *testing.T) {
		input := 7
		total := 11

		assert.Equal(t,
			heartbeat.AITokens{CurrentInput: 7, CurrentOutput: 4},
			Cursor{}.cursorTokenCounts(cursorLogLine{
				TokenCount: &cursorTokenCount{
					InputTokensSnake: &input,
					TotalTokensSnake: &total,
				},
			}, heartbeat.AITokens{}),
		)
	})

	t.Run("token counts treat snake case token count as cumulative", func(t *testing.T) {
		input := 7
		output := 11

		assert.Equal(t,
			heartbeat.AITokens{LastInput: 3, LastOutput: 5, CurrentInput: 7, CurrentOutput: 11},
			Cursor{}.cursorTokenCounts(cursorLogLine{
				TokenCount: &cursorTokenCount{
					InputTokensSnake:  &input,
					OutputTokensSnake: &output,
				},
			}, heartbeat.AITokens{LastInput: 3, LastOutput: 5}),
		)
	})

	t.Run("zero token counts preserve pending tokens", func(t *testing.T) {
		zero := 0
		output := 4
		previous := heartbeat.AITokens{
			LastInput:     3,
			LastOutput:    5,
			CurrentInput:  10,
			CurrentOutput: 16,
		}

		assert.Equal(t,
			previous,
			Cursor{}.cursorTokenCounts(cursorLogLine{
				TokenCount: &cursorTokenCount{
					InputTokens:  &zero,
					OutputTokens: &zero,
				},
			}, previous),
		)

		assert.Equal(t,
			heartbeat.AITokens{
				LastInput:     3,
				LastOutput:    5,
				CurrentInput:  10,
				CurrentOutput: 16,
			},
			Cursor{}.cursorTokenCounts(cursorLogLine{
				TokenCount: &cursorTokenCount{
					InputTokens:  &zero,
					OutputTokens: &output,
				},
			}, previous),
		)

		output = 20
		assert.Equal(t,
			heartbeat.AITokens{
				LastInput:     3,
				LastOutput:    5,
				CurrentInput:  10,
				CurrentOutput: 20,
			},
			Cursor{}.cursorTokenCounts(cursorLogLine{
				TokenCount: &cursorTokenCount{
					InputTokens:  &zero,
					OutputTokens: &output,
				},
			}, previous),
		)
	})

	t.Run("output tokens take precedence over totals", func(t *testing.T) {
		input := 7
		output := 4
		total := 11

		assert.Equal(t,
			heartbeat.AITokens{CurrentInput: 7, CurrentOutput: 4},
			Cursor{}.cursorTokenCounts(cursorLogLine{
				TokenCount: &cursorTokenCount{
					InputTokensSnake:  &input,
					OutputTokensSnake: &output,
					TotalTokensSnake:  &total,
				},
			}, heartbeat.AITokens{}),
		)
	})

	t.Run("edit embedded content per tool", func(t *testing.T) {
		content, isEdit := Cursor{}.editEmbeddedContent(cursorLogLine{
			ToolFormerData: &cursorToolFormerData{
				Name:   "edit_file_v2",
				Params: `{"streamingContent":"one\ntwo"}`,
			},
		})
		assert.True(t, isEdit)
		assert.Equal(t, "one\ntwo", content)

		content, isEdit = Cursor{}.editEmbeddedContent(cursorLogLine{
			ToolFormerData: &cursorToolFormerData{
				Name:   "edit_file_v2",
				Params: "{",
			},
		})
		assert.True(t, isEdit)
		assert.Empty(t, content)

		content, isEdit = Cursor{}.editEmbeddedContent(cursorLogLine{
			ToolFormerData: &cursorToolFormerData{
				Name:    "edit_file",
				RawArgs: `{}`,
			},
			CodeBlocks: []cursorCodeBlock{{Content: "from code block"}},
		})
		assert.True(t, isEdit)
		assert.Equal(t, "from code block", content)

		_, isEdit = Cursor{}.editEmbeddedContent(cursorLogLine{
			ToolFormerData: &cursorToolFormerData{Name: "read_file"},
		})
		assert.False(t, isEdit)
	})

	t.Run("edit snapshot content ids require before and after", func(t *testing.T) {
		before, after, ok := Cursor{}.editSnapshotContentIDs(cursorLogLine{
			ToolFormerData: &cursorToolFormerData{
				Result: `{"beforeContentId":"composer.content.b","afterContentId":"composer.content.a"}`,
			},
		})
		assert.True(t, ok)
		assert.Equal(t, "composer.content.b", before)
		assert.Equal(t, "composer.content.a", after)

		_, _, ok = Cursor{}.editSnapshotContentIDs(cursorLogLine{
			ToolFormerData: &cursorToolFormerData{},
		})
		assert.False(t, ok)

		_, _, ok = Cursor{}.editSnapshotContentIDs(cursorLogLine{
			ToolFormerData: &cursorToolFormerData{Result: "{"},
		})
		assert.False(t, ok)

		_, _, ok = Cursor{}.editSnapshotContentIDs(cursorLogLine{
			ToolFormerData: &cursorToolFormerData{Result: `{"beforeContentId":"composer.content.b"}`},
		})
		assert.False(t, ok)
	})
}

func TestCursorSnapshotLineChanges(t *testing.T) {
	ctx := context.Background()

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	defer db.Close() // nolint:errcheck

	_, err = db.Exec(`CREATE TABLE cursorDiskKV (key TEXT, value BLOB)`)
	require.NoError(t, err)

	for key, content := range map[string]string{
		"composer.content.before": "one\ntwo\nthree",
		"composer.content.after":  "one\nthree",
		"composer.content.empty":  "",
	} {
		_, err = db.Exec(`INSERT INTO cursorDiskKV(key, value) VALUES(?, ?)`, key, content)
		require.NoError(t, err)
	}

	editLine := func(result string) cursorLogLine {
		return cursorLogLine{
			Type: 2,
			ToolFormerData: &cursorToolFormerData{
				Name:   "edit_file_v2",
				Params: `{"relativeWorkspacePath":"/tmp/main.go","noCodeblock":true}`,
				Result: result,
				Status: "completed",
			},
		}
	}

	t.Run("signed net delta from snapshots", func(t *testing.T) {
		got := Cursor{}.snapshotLineChanges(ctx, db, editLine(
			`{"beforeContentId":"composer.content.before","afterContentId":"composer.content.after"}`,
		))
		require.NotNil(t, got)
		assert.Equal(t, -1, *got)

		got = Cursor{}.snapshotLineChanges(ctx, db, editLine(
			`{"beforeContentId":"composer.content.empty","afterContentId":"composer.content.before"}`,
		))
		require.NotNil(t, got)
		assert.Equal(t, 3, *got)
	})

	t.Run("missing snapshot is unknown not zero", func(t *testing.T) {
		assert.Nil(t, Cursor{}.snapshotLineChanges(ctx, db, editLine(
			`{"beforeContentId":"composer.content.before","afterContentId":"composer.content.missing"}`,
		)))
	})

	t.Run("skips non edit rows and embedded content", func(t *testing.T) {
		assert.Nil(t, Cursor{}.snapshotLineChanges(ctx, db, cursorLogLine{Type: 1}))
		assert.Nil(t, Cursor{}.snapshotLineChanges(ctx, db, cursorLogLine{
			Type:           2,
			ToolFormerData: &cursorToolFormerData{Name: "edit_file_v2", Status: "pending"},
		}))
		assert.Nil(t, Cursor{}.snapshotLineChanges(ctx, db, cursorLogLine{
			Type: 2,
			ToolFormerData: &cursorToolFormerData{
				Name:   "edit_file_v2",
				Params: `{"streamingContent":"one\ntwo"}`,
				Status: "completed",
			},
		}))
		assert.Nil(t, Cursor{}.snapshotLineChanges(ctx, db, editLine("")))
	})

	t.Run("legacy edit_file uses snapshot line changes when content is empty", func(t *testing.T) {
		got := Cursor{}.cursorFileHeartbeat(cursorLogLine{
			Type: 2,
			ToolFormerData: &cursorToolFormerData{
				Name:    "edit_file",
				Params:  `{"relativeWorkspacePath":"/tmp/main.go"}`,
				RawArgs: `{}`,
				Status:  "completed",
			},
		}, "", "", nil, heartbeat.PointerTo(-2))
		require.NotNil(t, got)
		require.NotNil(t, got.AILineChanges)
		assert.Equal(t, -2, *got.AILineChanges)
	})

	t.Run("null snapshot values resolve to unknown", func(t *testing.T) {
		_, err := db.Exec(
			`INSERT INTO cursorDiskKV(key, value) VALUES(?, NULL)`,
			"composer.content.null",
		)
		require.NoError(t, err)

		assert.Nil(t, Cursor{}.snapshotLineChanges(ctx, db, editLine(
			`{"beforeContentId":"composer.content.null","afterContentId":"composer.content.after"}`,
		)))
	})

	t.Run("query errors resolve to unknown", func(t *testing.T) {
		closedDB, err := sql.Open("sqlite", ":memory:")
		require.NoError(t, err)
		require.NoError(t, closedDB.Close())

		assert.Nil(t, Cursor{}.snapshotLineChanges(ctx, closedDB, editLine(
			`{"beforeContentId":"composer.content.before","afterContentId":"composer.content.after"}`,
		)))

		_, err = Cursor{}.queryContentLineCounts(ctx, closedDB, "composer.content.before")
		require.Error(t, err)
	})
}

func TestCursorHeartbeatFallbacks(t *testing.T) {
	parser := Cursor{FallbackUserAgent: "plugin/0.1.0"}
	createdAt := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)

	editHeartbeats := parser.cursorHeartbeats(cursorLogLine{
		CreatedAt: createdAt,
		Type:      2,
		ToolFormerData: &cursorToolFormerData{
			Name:    "edit_file",
			Params:  `{}`,
			RawArgs: `{"target_file":"/tmp/raw-edit.go","code_edit":"one\ntwo"}`,
			Status:  "completed",
		},
	}, "", "composer-2.5", heartbeat.AITokens{}, nil)
	require.Len(t, editHeartbeats, 1)
	edit := &editHeartbeats[0]
	require.NotNil(t, edit)
	assert.Equal(t, "/tmp/raw-edit.go", edit.Entity)
	assert.Contains(t, edit.UserAgent, "composer/2.5")
	assert.Contains(t, edit.UserAgent, "Cursor")
	require.NotNil(t, edit.AILineChanges)
	assert.Equal(t, 2, *edit.AILineChanges)
	require.NotNil(t, edit.IsWrite)
	assert.True(t, *edit.IsWrite)

	readHeartbeats := parser.cursorHeartbeats(cursorLogLine{
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
	}, "", "", heartbeat.AITokens{}, nil)
	require.Len(t, readHeartbeats, 1)
	read := &readHeartbeats[0]
	require.NotNil(t, read)
	assert.Equal(t, "/tmp/from-read-block.go", read.Entity)
	require.NotNil(t, read.IsWrite)
	assert.False(t, *read.IsWrite)

	appHeartbeats := parser.cursorHeartbeats(cursorLogLine{
		BubbleID:  "composer-1",
		CreatedAt: createdAt,
		Type:      1,
		Text:      "Please edit the file",
	}, "/tmp", "", heartbeat.AITokens{}, nil)
	require.Len(t, appHeartbeats, 1)
	assert.Equal(t, "Cursor composer-1", appHeartbeats[0].Entity)
	assert.Equal(t, heartbeat.AppType, appHeartbeats[0].EntityType)
	assert.Equal(t, "/tmp", appHeartbeats[0].ProjectPathOverride)
	assert.Nil(t, appHeartbeats[0].AILineChanges)
	require.NotNil(t, appHeartbeats[0].IsWrite)
	assert.False(t, *appHeartbeats[0].IsWrite)

	assert.Nil(t, parser.cursorFileHeartbeat(cursorLogLine{
		CreatedAt: createdAt,
		Type:      2,
		ToolFormerData: &cursorToolFormerData{
			Name:   "unknown_tool",
			Status: "completed",
		},
	}, "", "", nil, nil))
}

func TestCursorModelName(t *testing.T) {
	tests := map[string]string{
		`{"modelName":"claude-3.7-sonnet"}`:                                "claude-3.7-sonnet",
		`{"selectedModel":{"name":"gpt-5 high"}}`:                          "gpt-5-high",
		`{"modelDetails":{"model_id":"anthropic/claude-sonnet-4.5"}}`:      "anthropic/claude-sonnet-4.5",
		`{"nestedModelConfig":{"value":{"slug":"gemini-3-pro-preview"}}}`:  "gemini-3-pro-preview",
		`{"modelName":"` + strings.Repeat("x", 129) + `"}`:                 "",
		`{"toolFormerData":{"name":"edit_file_v2"},"model_added_lines":3}`: "",
	}

	for raw, expected := range tests {
		assert.Equal(t, expected, Cursor{}.modelName([]byte(raw)))
	}
}

func TestCursorModelUserAgentToken(t *testing.T) {
	tests := map[string]string{
		"composer-2.5":                "composer/2.5",
		"composer/2.5":                "composer/2.5",
		"claude-3.7-sonnet":           "claude/3.7-sonnet",
		"gpt 5 high":                  "gpt/5-high",
		"anthropic/claude-sonnet-4.5": "anthropic/claude-sonnet-4.5",
		"":                            "",
		"/":                           "",
	}

	for raw, expected := range tests {
		assert.Equal(t, expected, cursorModelUserAgentToken(raw))
	}
}

func TestCursorUserAgentWithModel(t *testing.T) {
	assert.Equal(
		t,
		"composer/2.5 Cursor/1.105.1",
		cursorUserAgentWithModel("Cursor/1.105.1", "composer/2.5"),
	)
	assert.Equal(
		t,
		"composer/2.5 wakatime/13.0.7 (macOS-arm64-arm64) go1.26 Cursor/1.105.1",
		cursorUserAgentWithModel(
			"wakatime/13.0.7 (macOS-arm64-arm64) go1.26 Cursor/1.105.1",
			"composer/2.5",
		),
	)
	assert.Equal(
		t,
		"composer/2.5 Cursor plugin/0.1.0",
		cursorUserAgentWithModel("Cursor plugin/0.1.0", "composer/2.5"),
	)
	assert.Equal(
		t,
		"composer/2.5 Cursor/1.105.1",
		cursorUserAgentWithModel("composer/2.5 Cursor/1.105.1", "composer/2.5"),
	)
}

func TestCursorUserAgentIncludesModelAndCursorAttribution(t *testing.T) {
	parser := Cursor{
		FallbackUserAgent: "codex-cli/0.144.4 codex-cli-wakatime/1.0.0",
		UserAgents: map[string]string{
			"/tmp/dependencies.ts": "neovim/0.12 wakatime.nvim/12.0.0",
		},
	}

	assert.Equal(
		t,
		"grok/4.5 Cursor codex-cli/0.144.4 codex-cli-wakatime/1.0.0",
		parser.userAgent("Cursor session-id", "grok-4.5"),
	)
	assert.Equal(
		t,
		"grok/4.5 Cursor neovim/0.12 wakatime.nvim/12.0.0",
		parser.userAgent("/tmp/dependencies.ts", "grok-4.5"),
	)
	assert.Equal(t, "grok/4.5 Cursor", Cursor{}.userAgent("Cursor session-id", "grok-4.5"))
	assert.Equal(
		t,
		"grok/4.5 Cursor/3.11.25",
		Cursor{FallbackUserAgent: "Cursor/3.11.25"}.userAgent("Cursor session-id", "grok-4.5"),
	)
}
