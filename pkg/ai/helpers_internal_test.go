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
	"github.com/wakatime/wakatime-cli/pkg/params"
)

func TestParserIDStringAndPlugins(t *testing.T) {
	assert.Equal(t, "Claude", aiPlugin(Claude{}, ""))
	assert.Equal(t, "Claude/1.2.3", aiPlugin(Claude{}, "1.2.3"))
	assert.Equal(t, "Codex", aiPlugin(Codex{}, ""))
	assert.Equal(t, "Codex/1.2.3", aiPlugin(Codex{}, "1.2.3"))
	assert.Equal(t, "Continue", aiPlugin(Continue{}, ""))
	assert.Equal(t, "Continue/gpt-5.2", aiPlugin(Continue{}, "gpt-5.2"))
	assert.Equal(t, "Cody", aiPlugin(Cody{}, ""))
	assert.Equal(t, "Cody/claude-3.5", aiPlugin(Cody{}, "claude-3.5"))
	assert.Equal(t, "Roo Code", aiPlugin(RooCode{}, ""))
	assert.Equal(t, "OpenCode", aiPlugin(OpenCode{}, ""))
	assert.Equal(t, "Copilot", aiPlugin(Copilot{}, Copilot{}.version(nil)))
	assert.Equal(t, "Copilot/0.42.3", aiPlugin(Copilot{}, Copilot{}.version(&copilotAgent{ExtensionVersion: "0.42.3"})))
	assert.Equal(t, "Cursor", aiPlugin(Cursor{}, ""))
	assert.Equal(t, "Windsurf", aiPlugin(Windsurf{}, ""))
	assert.Equal(t, "Qoder", aiPlugin(Qoder{}, ""))
	assert.Equal(t, "Kiro", aiPlugin(Kiro{}, ""))
	assert.Equal(t, "Cline", aiPlugin(Cline{}, ""))
	assert.Equal(t, "Gemini", aiPlugin(Gemini{}, ""))
	assert.Equal(t, "Pi", aiPlugin(Pi{}, ""))
	assert.Equal(t, "Goose", aiPlugin(Goose{}, ""))
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
		aiPlugin(Cursor{}, "")+" "+"editor/1.0.0",
		aiUserAgent("/tmp/main.go", userAgents, "plugin/0.1.0", aiPlugin(Cursor{}, "")),
	)
	assert.Equal(
		t,
		aiPlugin(Cursor{}, "")+" "+"plugin/0.1.0",
		aiUserAgent("/tmp/other.go", userAgents, "plugin/0.1.0", aiPlugin(Cursor{}, "")),
	)
	assert.Equal(
		t,
		aiPlugin(Cursor{}, "")+" "+heartbeat.UserAgent(ctx, "editor/1.0.0"),
		aiUserAgent(
			"/tmp/rendered.go",
			map[string]string{
				"/tmp/rendered.go": heartbeat.UserAgent(ctx, "editor/1.0.0"),
			},
			"plugin/0.1.0",
			aiPlugin(Cursor{}, ""),
		),
	)
	assert.Equal(
		t,
		aiPlugin(Cursor{}, ""),
		aiUserAgent("/tmp/other.go", userAgents, "", aiPlugin(Cursor{}, "")),
	)
}

func TestPreserveHumanAttributesAppliesConfigOverridesWhenEmpty(t *testing.T) {
	aiHeartbeats := Heartbeats{
		{Entity: "/tmp/main.go", EntityType: heartbeat.FileType},
		{Entity: "Codex session", EntityType: heartbeat.AppType},
	}

	got, _ := preserveHumanAttributes(aiHeartbeats, nil, Config{
		Plugin: "",
		Project: params.ProjectParams{
			BranchAlternate: "mybranch",
			Alternate:       "fallback-project",
			Override:        "myproject",
		},
		Sanitize: params.SanitizeParams{
			ProjectPathOverride: "/path/to/project",
		},
	}, 0)

	require.Len(t, got, 2)

	for _, h := range got {
		assert.Equal(t, "fallback-project", h.ProjectAlternate)
		assert.Equal(t, "myproject", h.ProjectOverride)
		assert.Equal(t, "/path/to/project", h.ProjectPathOverride)
		assert.Equal(t, "mybranch", h.BranchAlternate)
	}
}

func TestPreserveHumanAttributesKeepsExistingProjectOverrides(t *testing.T) {
	aiHeartbeats := Heartbeats{
		{
			Entity:              "Codex session",
			EntityType:          heartbeat.AppType,
			ProjectPathOverride: "/detected/cwd",
			ProjectAlternate:    "alternate",
			ProjectOverride:     "override",
			Project:             heartbeat.PointerTo("myproj"),
		},
	}
	humanHeartbeats := []heartbeat.Heartbeat{
		{Entity: "/tmp/human.go", EntityType: heartbeat.FileType, Time: 100},
	}

	got, _ := preserveHumanAttributes(aiHeartbeats, humanHeartbeats, Config{
		Plugin: "",
		Project: params.ProjectParams{
			Alternate: "fallback-project",
			Override:  "myproject",
		},
		Sanitize: params.SanitizeParams{
			ProjectPathOverride: "/path/to/project",
		},
	}, 0)

	require.Len(t, got, 1)

	for _, h := range got {
		assert.Equal(t, "/detected/cwd", h.ProjectPathOverride)
		assert.Equal(t, "override", h.ProjectOverride)
		assert.Equal(t, "alternate", h.ProjectAlternate)
		assert.Equal(t, heartbeat.PointerTo("myproj"), h.Project)
	}
}

func TestAppHeartbeatEntity(t *testing.T) {
	assert.Equal(t, "Codex rollout-2026-04-02T11-15-29-019d4ec3-83f2-77b2-805a-3a1461effcc7",
		appHeartbeatEntity("Codex", "/tmp/rollout-2026-04-02T11-15-29-019d4ec3-83f2-77b2-805a-3a1461effcc7.jsonl"))
	assert.Equal(t, "Claude session", appHeartbeatEntity("Claude", "session.jsonl"))
	assert.Equal(t, "Cursor composer-1", appHeartbeatEntity("Cursor", "composer-1"))
	assert.Equal(t, "Cursor", appHeartbeatEntity("Cursor", ""))
}

func TestCodexSessionIDFromPath(t *testing.T) {
	assert.Equal(t, "019d4ec3-83f2-77b2-805a-3a1461effcc7",
		Codex{}.sessionIDFromPath("/tmp/rollout-2026-04-02T11-15-29-019d4ec3-83f2-77b2-805a-3a1461effcc7.jsonl"))
	assert.Equal(t, "session",
		Codex{}.sessionIDFromPath("/tmp/session.jsonl"))
}

func TestContinueFilePathHandlesWindowsFileURI(t *testing.T) {
	assert.Equal(t, `C:\Users\runner\project`, Continue{}.filePath(`file://C:\Users\runner\project`))
	assert.Equal(t, filepath.FromSlash(`C:/Users/runner/project`), Continue{}.filePath(`file:///C:/Users/runner/project`))
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

func TestClaudeHelpers(t *testing.T) {
	t.Run("content unmarshal and line counting", func(t *testing.T) {
		var str contentValue
		require.NoError(t, json.Unmarshal([]byte(`"first\nsecond"`), &str))
		assert.Equal(t, 2, str.lineChanges())

		var arr contentValue
		require.NoError(t, json.Unmarshal([]byte(`[{"text":"first\nsecond"},"third"]`), &arr))
		assert.Equal(t, 3, arr.lineChanges())

		var unsupported contentValue
		require.NoError(t, json.Unmarshal([]byte(`{"bad":true}`), &unsupported))
		assert.Equal(t, 0, ((*contentValue)(nil)).lineChanges())
		assert.Equal(t, 0, (&contentValue{}).lineChanges())
	})

	t.Run("tool use result supports object string and unknown shapes", func(t *testing.T) {
		var object toolUseResultValue
		require.NoError(t, json.Unmarshal([]byte(`{"filePath":"/tmp/main.go"}`), &object))
		require.NotNil(t, object.Object)
		assert.Equal(t, "/tmp/main.go", *object.Object.FilePath)

		var str toolUseResultValue
		require.NoError(t, json.Unmarshal([]byte(`"plain string result"`), &str))
		require.NotNil(t, str.String)
		assert.Equal(t, "plain string result", *str.String)

		var unsupported toolUseResultValue
		require.NoError(t, json.Unmarshal([]byte(`123`), &unsupported))
		assert.Nil(t, unsupported.Object)
		assert.Nil(t, unsupported.String)

		var list toolUseResultValue
		require.NoError(t, json.Unmarshal([]byte(`[{"type":"text","text":"ok"}]`), &list))
		assert.Nil(t, list.Object)
		assert.Nil(t, list.String)
	})

	t.Run("file path and line changes resolve expected sources", func(t *testing.T) {
		parser := Claude{}

		assert.Equal(t, "/tmp/direct.go", parser.getFilePath(toolUseResult{
			FilePath: heartbeat.PointerTo("/tmp/direct.go"),
		}))
		assert.Equal(t, "/tmp/nested.go", parser.getFilePath(toolUseResult{
			File: &toolUseResultFile{FilePath: heartbeat.PointerTo("/tmp/nested.go")},
		}))
		assert.Equal(t, "", parser.getFilePath(toolUseResult{}))
		assert.Equal(t, filepath.Dir("/tmp/direct.go"), parser.projectPath(claudeLogLine{
			ToolUseResult: &toolUseResultValue{
				Object: &toolUseResult{FilePath: heartbeat.PointerTo("/tmp/direct.go")},
			},
		}))
		assert.Equal(t, "/workspace", parser.projectPath(claudeLogLine{
			Cwd: heartbeat.PointerTo("/workspace"),
		}))
		assert.Equal(t, "", parser.projectPath(claudeLogLine{}))

		assert.Equal(t, 1, parser.lineChanges(toolUseResult{
			StructuredPatch: &[]structuredPatch{{OldLines: 1, NewLines: 2}},
		}))
		assert.Equal(t, 1, parser.lineChanges(toolUseResult{
			OldString: heartbeat.PointerTo("one"),
			NewString: heartbeat.PointerTo("one\ntwo"),
		}))
		assert.Equal(t, 2, parser.lineChanges(toolUseResult{
			Content: &contentValue{String: heartbeat.PointerTo("one\ntwo")},
		}))
		// File subfield represents a read result; line changes should be zero.
		assert.Equal(t, 0, parser.lineChanges(toolUseResult{
			File: &toolUseResultFile{Content: &contentValue{String: heartbeat.PointerTo("one\ntwo\nthree")}},
		}))
		// Create: originalFile is empty string, content at top level should count.
		assert.Equal(t, 2, parser.lineChanges(toolUseResult{
			Content:      &contentValue{String: heartbeat.PointerTo("one\ntwo")},
			OriginalFile: heartbeat.PointerTo(""),
		}))
		assert.Equal(t, 0, parser.lineChanges(toolUseResult{
			Content:      &contentValue{String: heartbeat.PointerTo("one")},
			OriginalFile: heartbeat.PointerTo("before"),
		}))
		assert.Equal(t, 1, parser.appLineChanges(&toolUseResultValue{
			String: heartbeat.PointerTo("summary"),
		}))
		assert.Equal(t, 2, parser.appLineChanges(&toolUseResultValue{
			Object: &toolUseResult{
				Content: &contentValue{String: heartbeat.PointerTo("one\ntwo")},
			},
		}))
		assert.Equal(t, 0, parser.appLineChanges(&toolUseResultValue{
			Object: &toolUseResult{
				FilePath: heartbeat.PointerTo("/tmp/skip.go"),
				Content:  &contentValue{String: heartbeat.PointerTo("one\ntwo")},
			},
		}))
		assert.Equal(t, 3, countStringLines("one\ntwo\nthree"))
	})
}

func TestCodexHelpers(t *testing.T) {
	timestamp := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	parser := Codex{}

	assert.Nil(t, parser.getHeartbeats(
		timestamp,
		"session.jsonl",
		"session",
		"1.2.3",
		"/workspace",
		nil,
		"",
		codexPayload{},
		heartbeat.AITokens{},
	))
	assert.Nil(t, parser.getHeartbeats(timestamp, "session.jsonl", "session", "1.2.3", "/workspace", nil, "", codexPayload{
		Name:  heartbeat.PointerTo("not_apply_patch"),
		Input: heartbeat.PointerTo("*** Update File: pkg/main.go\n+one"),
	}, heartbeat.AITokens{}))

	userHeartbeats := parser.getHeartbeats(
		timestamp,
		"session.jsonl",
		"session",
		"1.2.3",
		"/workspace",
		nil,
		"plugin/0.1.0",
		codexPayload{
			Type: heartbeat.PointerTo("message"),
			Role: heartbeat.PointerTo("user"),
			Content: []codexContentItem{
				{Type: "input_text", Text: "Please implement this"},
			},
		},
		heartbeat.AITokens{},
	)
	require.Len(t, userHeartbeats, 1)
	assert.Equal(t, "session.jsonl", userHeartbeats[0].Entity)
	assert.Equal(t, heartbeat.AppType, userHeartbeats[0].EntityType)
	assert.Nil(t, userHeartbeats[0].AILineChanges)
	require.NotNil(t, userHeartbeats[0].IsWrite)
	assert.False(t, *userHeartbeats[0].IsWrite)
	assert.Contains(t, userHeartbeats[0].UserAgent, "Codex/1.2.3")

	assistantHeartbeats := parser.getHeartbeats(
		timestamp,
		"session.jsonl",
		"session",
		"1.2.3",
		"/workspace",
		nil,
		"plugin/0.1.0",
		codexPayload{
			Type: heartbeat.PointerTo("message"),
			Role: heartbeat.PointerTo("assistant"),
			Content: []codexContentItem{
				{Type: "output_text", Text: "I am on it"},
			},
		},
		heartbeat.AITokens{},
	)
	require.Len(t, assistantHeartbeats, 1)
	assert.Equal(t, "session.jsonl", assistantHeartbeats[0].Entity)
	assert.Equal(t, heartbeat.AppType, assistantHeartbeats[0].EntityType)
	assert.Nil(t, assistantHeartbeats[0].AILineChanges)
	require.NotNil(t, assistantHeartbeats[0].IsWrite)
	assert.False(t, *assistantHeartbeats[0].IsWrite)

	heartbeats := parser.getHeartbeats(
		timestamp,
		"session.jsonl",
		"session",
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
		heartbeat.AITokens{},
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

func TestPreserveAttributesMutatesAIHeartbeats(t *testing.T) {
	project := "sample-project"
	branch := "main"
	language := "Go"
	lines := 120
	projectRootCount := 2

	aiHeartbeats := []heartbeat.Heartbeat{
		{Entity: "/tmp/main.go", Time: 100, UserAgent: "Codex/0.116.0-alpha.1"},
	}
	humanHeartbeats := []heartbeat.Heartbeat{
		{
			Entity:              "/tmp/main.go",
			Project:             &project,
			ProjectAlternate:    project,
			Branch:              &branch,
			BranchAlternate:     branch,
			Language:            &language,
			LanguageAlternate:   language,
			Lines:               &lines,
			ProjectOverride:     "override-project",
			ProjectPath:         "/tmp/project",
			ProjectPathOverride: "/tmp/override-project",
			ProjectRootCount:    &projectRootCount,
			Time:                99,
		},
	}

	got, _ := preserveHumanAttributes(aiHeartbeats, humanHeartbeats, Config{}, 0)

	require.Len(t, got, 1)
	assert.Same(t, &aiHeartbeats[0], &got[0])
	assert.Equal(t, &project, got[0].Project)
	assert.Equal(t, project, got[0].ProjectAlternate)
	assert.Equal(t, &branch, got[0].Branch)
	assert.Equal(t, branch, got[0].BranchAlternate)
	assert.Equal(t, &language, got[0].Language)
	assert.Equal(t, language, got[0].LanguageAlternate)
	assert.Equal(t, &lines, got[0].Lines)
	assert.Equal(t, "override-project", got[0].ProjectOverride)
	assert.Equal(t, "/tmp/project", got[0].ProjectPath)
	assert.Equal(t, "/tmp/override-project", got[0].ProjectPathOverride)
	assert.Equal(t, &projectRootCount, got[0].ProjectRootCount)
	assert.Equal(t, "Codex/0.116.0-alpha.1", got[0].UserAgent)

	assert.Equal(t, &project, aiHeartbeats[0].Project)
	assert.Equal(t, branch, aiHeartbeats[0].BranchAlternate)
	assert.Equal(t, "/tmp/project", aiHeartbeats[0].ProjectPath)
	assert.Equal(t, "Codex/0.116.0-alpha.1", aiHeartbeats[0].UserAgent)
}

func TestPreserveAttributes_AppHeartbeatFallsBackToHumanProjectFolder(t *testing.T) {
	aiHeartbeats := []heartbeat.Heartbeat{
		{
			Entity:     "session.jsonl",
			EntityType: heartbeat.AppType,
			Time:       100,
		},
	}
	humanHeartbeats := []heartbeat.Heartbeat{
		{
			Entity:              "/tmp/main.go",
			EntityType:          heartbeat.FileType,
			ProjectPath:         "/tmp/project",
			ProjectPathOverride: "/tmp/project-override",
			Time:                99,
		},
	}

	got, _ := preserveHumanAttributes(aiHeartbeats, humanHeartbeats, Config{}, 0)

	require.Len(t, got, 1)
	assert.Equal(t, "/tmp/project-override", got[0].ProjectPathOverride)
	assert.Equal(t, "", got[0].ProjectPath)
}
