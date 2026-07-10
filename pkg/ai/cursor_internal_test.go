//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestCursorHelpers(t *testing.T) {
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

	t.Run("token counts accept camel case usage", func(t *testing.T) {
		input := 7
		total := 11
		line := cursorLogLine{
			Usage: &cursorUsage{
				InputTokensCamel: &input,
				TotalTokensCamel: &total,
			},
		}

		assert.Equal(t,
			heartbeat.AITokens{CurrentInput: 7, CurrentOutput: 4},
			Cursor{}.cursorTokenCounts(line, heartbeat.AITokens{}),
		)
		assert.Equal(t, cursorTokenSourceUsagePerRequest, Cursor{}.cursorTokenObservation(line).Source)
	})

	t.Run("token counts treat snake case token count as per request", func(t *testing.T) {
		input := 7
		output := 11
		line := cursorLogLine{
			TokenCount: &cursorTokenCount{
				InputTokensSnake:  &input,
				OutputTokensSnake: &output,
			},
		}

		assert.Equal(t,
			heartbeat.AITokens{LastInput: 3, LastOutput: 5, CurrentInput: 10, CurrentOutput: 16},
			Cursor{}.cursorTokenCounts(line, heartbeat.AITokens{LastInput: 3, LastOutput: 5}),
		)
		assert.Equal(t, cursorTokenSourceTokenCountPerRequest, Cursor{}.cursorTokenObservation(line).Source)
	})

	t.Run("token counts up until here are cumulative", func(t *testing.T) {
		input := 7
		output := 11
		line := cursorLogLine{
			TokenCountUpUntilHere: &cursorTokenCount{
				InputTokens:  &input,
				OutputTokens: &output,
			},
		}

		assert.Equal(t,
			heartbeat.AITokens{LastInput: 3, LastOutput: 5, CurrentInput: 7, CurrentOutput: 11},
			Cursor{}.cursorTokenCounts(line, heartbeat.AITokens{LastInput: 3, LastOutput: 5}),
		)
		assert.Equal(t, cursorTokenSourceCumulative, Cursor{}.cursorTokenObservation(line).Source)
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
				CurrentOutput: 20,
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
				CurrentOutput: 36,
			},
			Cursor{}.cursorTokenCounts(cursorLogLine{
				TokenCount: &cursorTokenCount{
					InputTokens:  &zero,
					OutputTokens: &output,
				},
			}, previous),
		)

		assert.Equal(t,
			previous,
			Cursor{}.cursorTokenCounts(cursorLogLine{
				Usage: &cursorUsage{
					InputTokensCamel:  &zero,
					OutputTokensCamel: &zero,
				},
			}, previous),
		)
	})

	t.Run("context window status is retained as an unpublished estimate", func(t *testing.T) {
		zero := 0
		tokensUsed := 61702
		line := cursorLogLine{
			TokenCount:          &cursorTokenCount{InputTokens: &zero, OutputTokens: &zero},
			ContextWindowStatus: &cursorContextWindowStatus{TokensUsed: &tokensUsed},
		}

		observation := Cursor{}.cursorTokenObservation(line)
		assert.Equal(t, cursorTokenSourceContextSnapshotEstimate, observation.Source)
		assert.True(t, observation.Estimated)
		assert.Equal(t, cursorInt64Pointer(61702), observation.Input)
		assert.Equal(t, heartbeat.AITokens{}, Cursor{}.cursorTokenCounts(line, heartbeat.AITokens{}))

		shrunk := 100
		previous := heartbeat.AITokens{LastInput: 61702, CurrentInput: 61702}
		assert.Equal(t,
			previous,
			Cursor{}.cursorTokenCounts(cursorLogLine{
				ContextWindowStatus: &cursorContextWindowStatus{TokensUsed: &shrunk},
			}, previous),
		)
	})

	t.Run("cumulative counter reset does not create a negative or duplicate delta", func(t *testing.T) {
		reset := 20
		previous := heartbeat.AITokens{LastInput: 100, CurrentInput: 100}

		assert.Equal(t,
			previous,
			Cursor{}.cursorTokenCounts(cursorLogLine{
				TokenCountUpUntilHere: &cursorTokenCount{InputTokens: &reset},
			}, previous),
		)

		regrown := 130
		assert.Equal(t,
			heartbeat.AITokens{LastInput: 100, CurrentInput: 130},
			Cursor{}.cursorTokenCounts(cursorLogLine{
				TokenCountUpUntilHere: &cursorTokenCount{InputTokens: &regrown},
			}, previous),
		)
	})

	t.Run("line changes from content snapshots", func(t *testing.T) {
		assert.Equal(t, 1, Cursor{}.lineChangesFromSnapshots(
			"one\ntwo\nthree",
			"one\ntwo\nthree\nfour",
		))
		assert.Equal(t, -1, Cursor{}.lineChangesFromSnapshots(
			"one\ntwo\nthree",
			"one\nthree",
		))
		assert.Equal(t, 0, Cursor{}.lineChangesFromSnapshots(
			"one\ntwo",
			"one changed\ntwo",
		))
		assert.Equal(t, 2, Cursor{}.lineChangesFromSnapshots("", "new\nfile"))
		assert.Equal(t, -1, Cursor{}.lineChangesFromSnapshots(
			strings.Repeat("line\n", 4_000)+"last",
			strings.Repeat("line\n", 3_999)+"last",
		))
	})

	t.Run("line changes from edit result content ids", func(t *testing.T) {
		lineCounts := map[string]int{
			"composer.content.before": 2,
			"composer.content.after":  3,
		}

		assert.Equal(t, heartbeat.PointerTo(1), Cursor{}.lineChangesFromEditResult(
			`{"beforeContentId":"composer.content.before","afterContentId":"composer.content.after"}`,
			lineCounts,
		))
		assert.Nil(t, Cursor{}.lineChangesFromEditResult(
			`{"beforeContentId":"composer.content.before","afterContentId":"composer.content.missing"}`,
			lineCounts,
		))
		assert.Nil(t, Cursor{}.lineChangesFromEditResult("not-json", lineCounts))
		assert.Nil(t, Cursor{}.lineChangesFromEditResult("", nil))
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

func BenchmarkCursorLineChangesFromSnapshots(b *testing.B) {
	before := strings.TrimSuffix(strings.Repeat("unchanged line\n", 10_000), "\n")
	after := before + "\nnew line"

	b.ReportAllocs()

	var got int
	for b.Loop() {
		got = (Cursor{}).lineChangesFromSnapshots(before, after)
	}

	if got != 1 {
		b.Fatalf("expected one added line, got %d", got)
	}
}
