package ai

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	wakalog "github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClinePathHelpers(t *testing.T) {
	assert.True(t, clineUsesPosixPaths("/workspace/project"))
	assert.False(t, clineUsesPosixPaths(`C:\workspace\project`))
	assert.False(t, clineUsesPosixPaths("relative/path"))

	assert.True(t, clineIsAbsPath("/workspace/project/main.go"))

	if filepath.Separator == '\\' {
		assert.Equal(t, `/workspace/project/main.go`, clineJoinPath("/workspace/project", "main.go"))
		assert.Equal(t, filepath.Join(`C:\workspace\project`, "main.go"), clineJoinPath(`C:\workspace\project`, "main.go"))
	} else {
		assert.Equal(t, "/workspace/project/main.go", clineJoinPath("/workspace/project", "main.go"))
	}
}

func TestCodexDecodeHexEscape(t *testing.T) {
	value, consumed, ok := codexDecodeHexEscape("41 rest", 2)
	assert.True(t, ok)
	assert.Equal(t, 'A', value)
	assert.Equal(t, 2, consumed)

	value, consumed, ok = codexDecodeHexEscape("03bb", 4)
	assert.True(t, ok)
	assert.Equal(t, rune(0x03bb), value)
	assert.Equal(t, 4, consumed)

	_, _, ok = codexDecodeHexEscape("4", 2)
	assert.False(t, ok)

	_, _, ok = codexDecodeHexEscape("zz", 2)
	assert.False(t, ok)
}

func TestCopilotURIHelpers(t *testing.T) {
	c := Copilot{}

	assert.Empty(t, c.pathToFileURI(""))
	assert.Equal(t, "file:///tmp/main.go", c.pathToFileURI("/tmp/main.go"))

	path, err := c.fileURIToPath("file:///tmp/main.go")
	assert.NoError(t, err)
	assert.Equal(t, "/tmp/main.go", path)

	_, err = c.fileURIToPath("%")
	assert.Error(t, err)
}

func TestCopilotResponseHasActivity(t *testing.T) {
	c := Copilot{}

	assert.True(t, c.responseHasActivity([]copilotResponseItem{{Kind: "thinking"}}))
	assert.True(t, c.responseHasActivity([]copilotResponseItem{{Kind: "inlineReference"}}))
	assert.False(t, c.responseHasActivity([]copilotResponseItem{{Kind: "ignored"}}))
	assert.False(t, c.responseHasActivity(nil))
}

func TestWindsurfLineChangesFromDiff(t *testing.T) {
	diff := "--- a/main.go\n+++ b/main.go\n-old\n+new\n+extra\n context"

	assert.Equal(t, 1, Windsurf{}.lineChangesFromDiff(diff))
}

func TestGeminiPromptHeartbeatsFromLogs(t *testing.T) {
	root := t.TempDir()
	transcript := filepath.Join(root, "session", "transcript.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(transcript), 0700))

	logs := []geminiPromptLog{
		{
			SessionID: "session-1",
			Type:      "user",
			Message:   "write a test",
			Timestamp: "2026-01-02T03:04:05Z",
		},
		{
			SessionID: "session-1",
			Type:      "model",
			Message:   "ignored",
			Timestamp: "2026-01-02T03:05:05Z",
		},
		{
			SessionID: "other-session",
			Type:      "user",
			Message:   "ignored",
			Timestamp: "2026-01-02T03:06:05Z",
		},
	}

	contents, err := json.Marshal(logs)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root, "logs.json"), contents, 0600))

	heartbeats, err := Gemini{
		After:             time.Date(2026, 1, 2, 3, 0, 0, 0, time.UTC),
		FallbackUserAgent: "fallback/1.0.0",
	}.promptHeartbeatsFromLogs(transcript, "Gemini session-1", "session-1", "/workspace/project")
	require.NoError(t, err)
	require.Len(t, heartbeats, 1)

	assert.Equal(t, "Gemini session-1", heartbeats[0].Entity)
	assert.Equal(t, heartbeat.AppType, heartbeats[0].EntityType)
	assert.Equal(t, "session-1", heartbeats[0].AISession)
	assert.Equal(t, len([]rune("write a test")), heartbeats[0].AIPromptLength)
	assert.Equal(t, "/workspace/project", heartbeats[0].ProjectPathOverride)
	assert.Equal(t, "fallback/1.0.0", heartbeats[0].UserAgent)
}

func TestGeminiPromptHeartbeatsFromLogsMissingFile(t *testing.T) {
	heartbeats, err := Gemini{}.promptHeartbeatsFromLogs(
		filepath.Join(t.TempDir(), "session", "transcript.json"),
		"Gemini session-1",
		"session-1",
		"",
	)

	require.NoError(t, err)
	assert.Nil(t, heartbeats)
}

func TestQoderPromptHeartbeat(t *testing.T) {
	timestamp := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	h := Qoder{FallbackUserAgent: "fallback/1.0.0"}.qoderPromptHeartbeat(qoderPrompt{
		SessionID:  "session-1",
		ProjectURI: "/workspace/project",
		Timestamp:  timestamp,
		Length:     42,
	})

	assert.Equal(t, "Qoder session-1", h.Entity)
	assert.Equal(t, heartbeat.AppType, h.EntityType)
	assert.Equal(t, "session-1", h.AISession)
	assert.Equal(t, 42, h.AIPromptLength)
	assert.Equal(t, "/workspace/project", h.ProjectPathOverride)
	assert.Equal(t, float64(timestamp.UnixMilli())/1000, h.Time)
	assert.Equal(t, "fallback/1.0.0", h.UserAgent)
}

func TestQoderToolResultHelpers(t *testing.T) {
	assert.Equal(t, "/tmp/main.go", qoderToolResult{
		Parameters: qoderToolParameters{FilePath: "/tmp/main.go", Path: "/tmp/fallback.go"},
	}.parameterFilePath())
	assert.Equal(t, "/tmp/fallback.go", qoderToolResult{
		Parameters: qoderToolParameters{Path: "/tmp/fallback.go"},
	}.parameterFilePath())

	assert.Equal(t, 2, qoderToolFile{
		DiffInfo: qoderDiffInfo{Add: 3, Delete: 1},
	}.lineChanges())
	assert.Equal(t, -1, qoderToolFile{
		LastDiffInfo: qoderDiffInfo{Add: 1, Delete: 2},
	}.lineChanges())
}

func TestQoderPromptAndPathBranches(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	parser := Qoder{FallbackUserAgent: "fallback/1.0"}
	dbPath, err := parser.localDBPath(context.Background())
	require.NoError(t, err)
	assert.Empty(t, dbPath)
	assert.False(t, parser.localDBModifiedAfter(filepath.Join(home, "missing.db"), time.Now()))

	localDBPath := filepath.Join(
		home,
		"Library",
		"Application Support",
		"Qoder",
		"SharedClientCache",
		"cache",
		"db",
		"local.db",
	)
	require.NoError(t, os.MkdirAll(filepath.Dir(localDBPath), 0700))
	require.NoError(t, os.WriteFile(localDBPath, []byte("db"), 0600))
	dbPath, err = parser.localDBPath(context.Background())
	require.NoError(t, err)
	assert.Equal(t, localDBPath, dbPath)

	old := time.Now().Add(-time.Hour)
	require.NoError(t, os.Chtimes(localDBPath, old, old))
	assert.False(t, parser.localDBModifiedAfter(localDBPath, time.Now()))
	assert.True(t, parser.localDBModifiedAfter(localDBPath, time.Time{}))

	historyPath := filepath.Join(
		home,
		".qoder",
		"cache",
		"projects",
		"workspace",
		"conversation-history",
		"session",
		"abc.jsonl",
	)
	require.NoError(t, os.MkdirAll(filepath.Dir(historyPath), 0700))
	require.NoError(t, os.WriteFile(historyPath, []byte(strings.Join([]string{
		`not json`,
		`{"role":"assistant","message":{"content":[{"type":"text","text":"ignored"}]}}`,
		`{"role":"user","message":{"content":[{"type":"image","text":"ignored"}]}}`,
		`{"role":"user","message":{"content":[{"type":"text","text":"<user_query>hello</user_query>"}]}}`,
		`{"role":"user","message":{"content":[{"type":"text","text":"plain prompt"}]}}`,
	}, "\n")), 0600))

	lengths, err := qoderPromptLengths(filepath.Join(home, ".qoder", "cache", "projects"), historyPath)
	require.NoError(t, err)
	assert.Equal(t, []int{5, len([]rune("plain prompt"))}, lengths)

	_, err = qoderPromptLengths(filepath.Dir(historyPath), filepath.Join(home, "outside.jsonl"))
	require.Error(t, err)

	rows := []qoderPromptRow{
		{SessionID: "abc-session", ProjectURI: "/workspace", CreatedAtMS: time.Now().UnixMilli()},
		{SessionID: "abc-session", ProjectURI: "/workspace", CreatedAtMS: time.Now().Add(time.Second).UnixMilli()},
		{SessionID: "abc-session", ProjectURI: "/workspace", CreatedAtMS: time.Now().Add(2 * time.Second).UnixMilli()},
		{SessionID: "unknown", ProjectURI: "/workspace", CreatedAtMS: time.Now().UnixMilli()},
		{SessionID: "abc-session", ProjectURI: "/workspace"},
	}
	prompts, err := parser.qoderPrompts(context.Background(), rows)
	require.NoError(t, err)
	require.Len(t, prompts, 2)
	assert.Equal(t, 5, prompts[0].Length)
	assert.Equal(t, len([]rune("plain prompt")), prompts[1].Length)

	heartbeatTime := float64(prompts[0].Timestamp.Add(time.Second).UnixMilli()) / 1000
	heartbeats := Heartbeats{{AISession: "abc-session", Time: heartbeatTime}}
	assert.Equal(t, 0, nearestPromptHeartbeat(heartbeats, prompts[0]))
	heartbeats[0].AIPromptLength = 1
	assert.Equal(t, -1, nearestPromptHeartbeat(heartbeats, prompts[0]))

	withPrompt := parser.withPromptLengths(Heartbeats{{
		AISession: "abc-session",
		Time:      heartbeatTime,
	}}, prompts[:1])
	require.Len(t, withPrompt, 1)
	assert.Equal(t, 5, withPrompt[0].AIPromptLength)

	appPrompt := parser.withPromptLengths(nil, prompts[:1])
	require.Len(t, appPrompt, 1)
	assert.Equal(t, heartbeat.AppType, appPrompt[0].EntityType)
	assert.Equal(t, 5, appPrompt[0].AIPromptLength)

	parser.After = time.Now()
	stalePrompt := prompts[0]
	stalePrompt.Timestamp = parser.After.Add(-time.Second)
	assert.Empty(t, parser.withPromptLengths(nil, []qoderPrompt{{Length: 0}, stalePrompt}))
	assert.Equal(t, []int{2}, qoderPromptLengthsForSession("abcdef", map[string][]int{
		"ab":  {1},
		"abc": {2},
	}))
	assert.Equal(t, "last", qoderUserQuery("first <user_query>ignored</user_query> <user_query>last"))
}

func TestQwenCodeResolveDir(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")

	resolved, err := qwenCodeResolveDir("~", home)
	require.NoError(t, err)
	assert.Equal(t, filepath.Clean(home), resolved)

	resolved, err = qwenCodeResolveDir("~/runtime", home)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "runtime"), resolved)

	resolved, err = qwenCodeResolveDir(`~\runtime`, home)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "runtime"), resolved)
}

func TestSyncLockClockNow(t *testing.T) {
	assert.False(t, syncLockClock{}.Now().IsZero())
}

func TestSyncLockClockAfterHonorsDelay(t *testing.T) {
	delay := 20 * time.Millisecond
	startedAt := time.Now()

	<-syncLockClock{}.After(delay)

	assert.GreaterOrEqual(t, time.Since(startedAt), delay)
}

func TestCopilotCLITelemetryHelpers(t *testing.T) {
	telemetry := &copilotCLIToolTelemetry{
		Metrics: copilotCLIToolMetrics{
			LinesAdded:   heartbeat.PointerTo(5),
			LinesRemoved: heartbeat.PointerTo(2),
		},
		RestrictedProperties: &copilotCLIToolRestrictedProperties{
			FilePaths: json.RawMessage(`"[\"/tmp/main.go\",\"\",\"/tmp/main.go\",\"/tmp/other.go\"]"`),
		},
		Properties: copilotCLIToolProperties{
			CodeBlocks: json.RawMessage(`[{"linesAdded":3,"linesRemoved":1}]`),
		},
	}

	data := copilotCLIToolExecutionCompleteData{ResultTelemetry: telemetry}
	assert.Equal(t, telemetry, data.telemetry())
	assert.True(t, telemetry.hasWriteSignals("read_file"))
	assert.True(t, copilotCLIToolTelemetry{}.hasWriteSignals("write"))
	assert.False(t, copilotCLIToolTelemetry{}.hasWriteSignals("read_file"))
	assert.Equal(t, []string{"/tmp/main.go", "/tmp/other.go"}, telemetry.filePaths())

	lineChanges := telemetry.lineChanges(1)
	require.NotNil(t, lineChanges.value(0))
	assert.Equal(t, 2, *lineChanges.value(0))
	assert.Nil(t, lineChanges.value(-1))
	assert.Nil(t, lineChanges.value(1))

	metricLineChanges := (&copilotCLIToolTelemetry{
		Metrics: copilotCLIToolMetrics{LinesAdded: heartbeat.PointerTo(4)},
	}).lineChanges(1)
	require.NotNil(t, metricLineChanges.value(0))
	assert.Equal(t, 4, *metricLineChanges.value(0))
}

func TestCopilotCLIDecodeHelpers(t *testing.T) {
	assert.Nil(t, decodeCopilotCLIStringArray(nil))
	assert.Nil(t, decodeCopilotCLIStringArray(json.RawMessage(`null`)))
	assert.Equal(t, []string{"a", "b"}, decodeCopilotCLIStringArray(json.RawMessage(`["a","b"]`)))
	assert.Equal(t, []string{"a"}, decodeCopilotCLIStringArray(json.RawMessage(`"[\"a\"]"`)))
	assert.Nil(t, decodeCopilotCLIStringArray(json.RawMessage(`123`)))

	assert.Nil(t, decodeCopilotCLICodeBlocks(nil))
	assert.Nil(t, decodeCopilotCLICodeBlocks(json.RawMessage(`null`)))
	assert.Len(t, decodeCopilotCLICodeBlocks(json.RawMessage(`[{"linesAdded":1}]`)), 1)
	assert.Len(t, decodeCopilotCLICodeBlocks(json.RawMessage(`"[{\"linesAdded\":1}]"`)), 1)
	assert.Nil(t, decodeCopilotCLICodeBlocks(json.RawMessage(`123`)))

	assert.Equal(t, []string{"a", "b"}, uniqueNonEmptyStrings([]string{"a", "", "b", "a"}))
	assert.Equal(t, "C:/tmp/main.go", copilotCLIPathID(` C:\tmp\main.go `))
	assert.Empty(t, copilotCLIPathID("   "))
	assert.False(t, (Copilot{}).shouldTrackCLIPath(""))
	assert.True(t, (Copilot{}).shouldTrackCLIPath("/tmp/main.go"))
}

func TestCopilotCLIUserAgentAndHeartbeats(t *testing.T) {
	parser := Copilot{
		FallbackUserAgent: "plugin/1.0",
		UserAgents: map[string]string{
			"/tmp/main.go": "editor/2.0",
		},
	}
	timestamp := time.Date(2026, 6, 24, 18, 0, 0, 123*int(time.Millisecond), time.UTC)

	assert.Equal(t, "github-copilot-cli/1.2.3 copilot/4.5.6", copilotCLIHarnessUserAgent("1.2.3", "4.5.6"))
	assert.Equal(t, "github-copilot-cli/unknown copilot/4.5.6", copilotCLIHarnessUserAgent("", "4.5.6"))
	assert.Empty(t, copilotCLIHarnessUserAgent("", ""))

	app := parser.cliAppHeartbeat(
		"Copilot session",
		"session-1",
		&heartbeat.AITokens{LastInput: 1, CurrentInput: 4, LastOutput: 2, CurrentOutput: 8},
		42,
		timestamp,
		"/project",
		"gpt-5",
		"1.2.3",
		"4.5.6",
	)
	assert.Equal(t, "session-1", app.AISession)
	assert.Equal(t, int64(3), app.AIInputTokens)
	assert.Equal(t, int64(6), app.AIOutputTokens)
	assert.Equal(t, 42, app.AIPromptLength)
	assert.Contains(t, app.UserAgent, "gpt/5")
	assert.Contains(t, app.UserAgent, "github-copilot-cli/1.2.3")

	lineChanges := 7
	file := parser.cliFileHeartbeat(
		"/tmp/main.go",
		"session-1",
		nil,
		timestamp,
		"",
		"",
		"",
		true,
		&lineChanges,
	)
	assert.Equal(t, "session-1", file.AISession)
	assert.Equal(t, &lineChanges, file.AILineChanges)
	assert.Contains(t, file.UserAgent, "editor/2.0")
}

func TestGeminiHelperBranches(t *testing.T) {
	tmpDir := t.TempDir()
	transcript := filepath.Join(tmpDir, "logs", "session", "transcript.jsonl")
	require.NoError(t, os.MkdirAll(filepath.Dir(transcript), 0700))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "logs", ".project_root"),
		[]byte("/workspace/project\n"),
		0600,
	))

	parser := Gemini{}
	assert.Equal(t, "/fallback", parser.projectPathFromLogs(transcript, "/fallback"))
	assert.Equal(t, "/workspace/project", parser.projectPathFromLogs(transcript, ""))
	assert.Equal(t, "/display.go", parser.toolFilePath(
		"/project",
		json.RawMessage(`{"filePath":"/display.go"}`),
		"relative.go",
	))
	assert.Equal(t, filepath.Join("/project", "relative.go"), parser.toolFilePath("/project", nil, "relative.go"))
	assert.Equal(t, 4, parser.replaceLineChanges(
		json.RawMessage(`{"diffStat":{"model_added_lines":5,"model_removed_lines":1}}`),
		"old",
		"new",
	))
	assert.Equal(t, 1, parser.replaceLineChanges(nil, "one", "one\ntwo"))
	assert.True(t, parser.hasTokenDelta(heartbeat.AITokens{LastInput: 1, CurrentInput: 2}))
	assert.False(t, parser.hasTokenDelta(heartbeat.AITokens{LastInput: 5, CurrentInput: 1}))
	assert.Nil(t, parser.tokensForFirstHeartbeat(false, heartbeat.AITokens{CurrentInput: 1}))
	require.NotNil(t, parser.tokensForFirstHeartbeat(true, heartbeat.AITokens{CurrentInput: 1}))
	assert.Equal(t, "hello", geminiText(json.RawMessage(`"hello"`)))
	assert.Equal(t, "hello world", geminiText(json.RawMessage(`[{"text":"hello "},{"text":"world"}]`)))
	assert.Empty(t, geminiText(json.RawMessage(`123`)))
	assert.False(t, parseGeminiTime("2026-06-24T18:00:00Z").IsZero())
	assert.True(t, parseGeminiTime("").IsZero())
	assert.True(t, parseGeminiTime("bad").IsZero())
}

func TestWindsurfHelperBranches(t *testing.T) {
	base := t.TempDir()
	editPath := filepath.Join(base, "edit.go")
	readPath := filepath.Join(base, "read.go")
	rawPath := filepath.Join(base, "raw.go")
	codeBlockPath := filepath.Join(base, "block.go")
	timestamp := time.Date(2026, 6, 24, 18, 0, 0, 0, time.UTC)

	parser := Windsurf{
		FallbackUserAgent: "fallback/1.0",
		UserAgents: map[string]string{
			editPath: "editor/1.0",
		},
	}

	input := 5
	total := 9
	usageInput := 12
	usageTotal := 15

	assert.Equal(t,
		heartbeat.AITokens{LastInput: 2, LastOutput: 3, CurrentInput: 5, CurrentOutput: 9},
		parser.windsurfTokenCounts(windsurfLogLine{
			TokenCount: &windsurfTokenCount{InputTokens: &input, TotalTokens: &total},
		}, heartbeat.AITokens{LastInput: 2, LastOutput: 3}),
	)

	zero := 0
	assert.Equal(t,
		heartbeat.AITokens{LastInput: 2, LastOutput: 3, CurrentInput: 5, CurrentOutput: 9},
		parser.windsurfTokenCounts(windsurfLogLine{
			TokenCount: &windsurfTokenCount{InputTokens: &zero, OutputTokens: &zero},
		}, heartbeat.AITokens{LastInput: 2, LastOutput: 3, CurrentInput: 5, CurrentOutput: 9}),
	)
	assert.Equal(t,
		heartbeat.AITokens{CurrentInput: 12, CurrentOutput: 15},
		parser.windsurfTokenCounts(windsurfLogLine{
			Usage: &windsurfUsage{InputTokens: &usageInput, TotalTokens: &usageTotal},
		}, heartbeat.AITokens{}),
	)
	assert.Equal(t, heartbeat.AITokens{CurrentInput: 1}, parser.windsurfTokenCounts(
		windsurfLogLine{}, heartbeat.AITokens{CurrentInput: 1}))

	require.NoError(t, os.WriteFile(editPath, []byte("package main\n"), 0600))
	assert.True(t, parser.stateDBModifiedAfter(editPath, time.Time{}))
	assert.False(t, parser.stateDBModifiedAfter(filepath.Join(base, "missing.db"), time.Now()))

	app := parser.windsurfAppHeartbeat(windsurfLogLine{
		BubbleID:  "bubble-1",
		CreatedAt: timestamp,
		Type:      1,
		Text:      "write tests",
	}, base, "session-1", "swe-1.5", nil)
	require.NotNil(t, app)
	assert.Equal(t, "Windsurf bubble-1", app.Entity)
	assert.Equal(t, len([]rune("write tests")), app.AIPromptLength)
	assert.Equal(t, base, app.ProjectPathOverride)
	assert.Contains(t, app.UserAgent, "swe/1.5")
	assert.Nil(t, parser.windsurfAppHeartbeat(windsurfLogLine{Type: 3, Text: "ignored"}, "", "", "", nil))

	editV2 := parser.windsurfFileHeartbeat(windsurfLogLine{
		CreatedAt: timestamp,
		ToolFormerData: &windsurfToolFormerData{
			Name:   "edit_file_v2",
			Params: `{"streamingContent":"one\ntwo"}`,
		},
		CodeBlocks: []windsurfCodeBlock{{URI: &windsurfURI{FSPath: editPath}}},
	}, "session-1", "swe-1.5", nil)
	require.NotNil(t, editV2)
	assert.Equal(t, editPath, editV2.Entity)
	require.NotNil(t, editV2.AILineChanges)
	assert.Equal(t, 2, *editV2.AILineChanges)
	assert.Contains(t, editV2.UserAgent, "editor/1.0")

	rawEdit := parser.windsurfFileHeartbeat(windsurfLogLine{
		CreatedAt: timestamp,
		ToolFormerData: &windsurfToolFormerData{
			Name:    "edit_file",
			Params:  `{`,
			RawArgs: `{"target_file":` + mustJSON(t, rawPath) + `,"code_edit":"--- a\n+++ b\n-old\n+new\n+extra"}`,
		},
	}, "session-1", "", nil)
	require.NotNil(t, rawEdit)
	assert.Equal(t, rawPath, rawEdit.Entity)
	require.NotNil(t, rawEdit.AILineChanges)
	assert.Equal(t, 1, *rawEdit.AILineChanges)

	blockEdit := parser.windsurfFileHeartbeat(windsurfLogLine{
		CreatedAt: timestamp,
		ToolFormerData: &windsurfToolFormerData{
			Name:   "edit_file",
			Params: `{"relativeWorkspacePath":` + mustJSON(t, codeBlockPath) + `}`,
		},
		CodeBlocks: []windsurfCodeBlock{{Content: "alpha\nbeta\n"}},
	}, "session-1", "", nil)
	require.NotNil(t, blockEdit)
	require.NotNil(t, blockEdit.AILineChanges)
	assert.Equal(t, 2, *blockEdit.AILineChanges)

	for name, line := range map[string]windsurfLogLine{
		"targetFile": {
			ToolFormerData: &windsurfToolFormerData{
				Name:   "read_file",
				Params: `{"targetFile":` + mustJSON(t, readPath) + `}`,
			},
		},
		"effectiveURI": {
			ToolFormerData: &windsurfToolFormerData{
				Name:   "read_file",
				Params: `{"effectiveUri":` + mustJSON(t, readPath) + `}`,
			},
		},
		"path": {
			ToolFormerData: &windsurfToolFormerData{Name: "read_file", Params: `{"path":` + mustJSON(t, readPath) + `}`},
		},
		"relativeWorkspacePath": {
			ToolFormerData: &windsurfToolFormerData{
				Name:   "read_file",
				Params: `{"relativeWorkspacePath":` + mustJSON(t, readPath) + `}`,
			},
		},
		"rawArgs": {
			ToolFormerData: &windsurfToolFormerData{
				Name:    "read_file",
				RawArgs: `{"target_file":` + mustJSON(t, readPath) + `}`,
			},
		},
		"codeBlock": {
			ToolFormerData: &windsurfToolFormerData{Name: "read_file", Params: `{}`},
			CodeBlocks:     []windsurfCodeBlock{{URI: &windsurfURI{FSPath: readPath}}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			line.CreatedAt = timestamp
			got := parser.windsurfFileHeartbeat(line, "session-1", "", nil)
			require.NotNil(t, got)
			assert.Equal(t, readPath, got.Entity)
			require.NotNil(t, got.IsWrite)
			assert.False(t, *got.IsWrite)
		})
	}

	assert.Nil(t, parser.windsurfFileHeartbeat(windsurfLogLine{
		ToolFormerData: &windsurfToolFormerData{Name: "read_file", Params: `{}`},
	}, "", "", nil))
	assert.Nil(t, parser.windsurfFileHeartbeat(windsurfLogLine{
		ToolFormerData: &windsurfToolFormerData{Name: "unknown"},
	}, "", "", nil))
	assert.Equal(t, filepath.Dir(readPath), parser.projectPath(windsurfLogLine{
		ToolFormerData: &windsurfToolFormerData{Name: "read_file", Params: `{"targetFile":` + mustJSON(t, readPath) + `}`},
	}))
	assert.Empty(t, parser.projectPath(windsurfLogLine{}))
	assert.True(t, parser.looksLikeUnifiedDiff("context\n@@ -1 +1 @@"))
	assert.False(t, parser.looksLikeUnifiedDiff("plain text"))
}

func TestOpenCodeHelperBranches(t *testing.T) {
	base := t.TempDir()
	timestamp := time.Date(2026, 6, 24, 18, 0, 0, 0, time.UTC)
	parser := OpenCode{FallbackUserAgent: "fallback/1.0"}

	editPath := filepath.Join(base, "edited.go")
	edit := parser.editHeartbeats("1.2.3", "claude-3.5", "session-1", base, timestamp, openCodeToolState{
		Input:    json.RawMessage(`{"filePath":"ignored.go","oldString":"old","newString":"new\nextra"}`),
		Metadata: json.RawMessage(`{"filediff":{"file":"edited.go","additions":4,"deletions":1}}`),
	})
	require.Len(t, edit, 1)
	assert.Equal(t, editPath, edit[0].Entity)
	require.NotNil(t, edit[0].AILineChanges)
	assert.Equal(t, 3, *edit[0].AILineChanges)
	assert.Nil(t, parser.editHeartbeats("", "", "", "", timestamp, openCodeToolState{Input: json.RawMessage(`{`)}))

	writeNew := parser.writeHeartbeats("1.2.3", "", "session-1", base, timestamp, openCodeToolState{
		Input:    json.RawMessage(`{"filePath":"new.go","content":"one\ntwo"}`),
		Metadata: json.RawMessage(`{"exists":false}`),
	})
	require.Len(t, writeNew, 1)
	require.NotNil(t, writeNew[0].AILineChanges)
	assert.Equal(t, 2, *writeNew[0].AILineChanges)
	assert.Nil(t, parser.writeHeartbeats("", "", "", "", timestamp, openCodeToolState{Input: json.RawMessage(`{`)}))

	patchFromMetadata := parser.applyPatchHeartbeats("1.2.3", "", "session-1", base, timestamp, openCodeToolState{
		Metadata: json.RawMessage(`{"files":[` +
			`{"filePath":"a.go","additions":3,"deletions":1},` +
			`{"filePath":"b.go","additions":1,"deletions":4}` +
			`]}`),
	})
	require.Len(t, patchFromMetadata, 2)
	assert.Equal(t, filepath.Join(base, "a.go"), patchFromMetadata[0].Entity)
	require.NotNil(t, patchFromMetadata[1].AILineChanges)
	assert.Equal(t, -3, *patchFromMetadata[1].AILineChanges)

	patchText := strings.Join([]string{
		"*** Begin Patch",
		"*** Add File: added.go",
		"+one",
		"+two",
		"*** Update File: old.go",
		"*** Move to: moved.go",
		"-old",
		"+new",
		"*** End Patch",
	}, "\n")
	patchFromInput := parser.applyPatchHeartbeats("1.2.3", "", "session-1", base, timestamp, openCodeToolState{
		Input: json.RawMessage(`{"patchText":` + mustJSON(t, patchText) + `}`),
	})
	require.Len(t, patchFromInput, 2)
	assert.Equal(t, filepath.Join(base, "added.go"), patchFromInput[0].Entity)
	assert.Equal(t, filepath.Join(base, "moved.go"), patchFromInput[1].Entity)
	assert.Nil(t, parser.applyPatchHeartbeats("", "", "", "", timestamp, openCodeToolState{Input: json.RawMessage(`{`)}))

	assert.Empty(t, openCodePatchFilePath(base, "*** Unknown File: nope.go"))
	assert.Equal(t, "relative.go", openCodeResolvePath("", "relative.go"))
	assert.Empty(t, openCodeResolvePath(base, ""))
}

func TestCodexPatchDecodeBranches(t *testing.T) {
	decoded := codexDecodePatch(`line\nnext\rcr\tab\bback\f form\vvert\\slash\"quote\'apos\/path\x41\u03bb\q`)
	assert.Contains(t, decoded, "line\nnext\rcr\tab")
	assert.Contains(t, decoded, "A")
	assert.Contains(t, decoded, string(rune(0x03bb)))
	assert.Contains(t, decoded, `\q`)

	input := `prefix tools.apply_patch("*** Begin Patch\n*** Add File: a.go\n+one\n*** End Patch") suffix ` +
		`tools.apply_patch("*** Begin Patch\u000a*** Add File: b.go\u000a+two\u000a*** End Patch")`
	patches := codexExecPatchInputs(input)
	require.Len(t, patches, 2)
	assert.Contains(t, patches[0], "*** Add File: a.go")
	assert.Contains(t, patches[1], "*** Add File: b.go")
	assert.Nil(t, codexExecPatchInputs("echo no patch"))
	assert.Nil(t, codexExecPatchInputs("tools.apply_patch(\"*** Begin Patch"))

	applyPatch := "patch"
	nameApply := "apply_patch"
	nameExec := "exec"

	assert.Equal(t, []string{"patch"}, codexPatchInputs(codexPayload{Name: &nameApply, Input: &applyPatch}))

	execInput := `tools.apply_patch("*** Begin Patch\n*** Add File: c.go\n+three\n*** End Patch")`
	assert.Len(t, codexPatchInputs(codexPayload{Name: &nameExec, Input: &execInput}), 1)
	assert.Nil(t, codexPatchInputs(codexPayload{}))

	assert.True(t, codexToolCallSucceeded(nil))
	assert.True(t, codexToolCallSucceeded(json.RawMessage(`"all good"`)))
	assert.False(t, codexToolCallSucceeded(json.RawMessage(`"invalid patch"`)))
	assert.False(t, codexToolCallSucceeded(json.RawMessage(`[{"text":"tool failed"}]`)))
	assert.True(t, codexToolCallSucceeded(json.RawMessage(`123`)))
}

func TestCodexAdditionalHelperBranches(t *testing.T) {
	base := t.TempDir()
	parser := Codex{FallbackUserAgent: "fallback/1.0"}

	assert.Empty(t, codexStripHarnessPrefix("   "))
	assert.Empty(t, codexStripHarnessPrefix("<bad"))
	assert.Empty(t, codexStripHarnessPrefix("</bad>"))
	assert.Empty(t, codexStripHarnessPrefix("<tag>missing close"))
	assert.Equal(t, "request", codexUserMessageText("<system>ignore</system> request"))
	assert.Equal(t, "actual request", codexUserMessageText(strings.Join([]string{
		"# Context from my IDE setup:",
		"ignored",
		"## My request for Codex:",
		"actual request",
	}, "\n")))

	assert.Empty(t, codexFilePath(base, "*** Unknown File: main.go"))
	assert.Empty(t, codexFilePath(base, "*** Add File: "))
	assert.Equal(t, filepath.Join(base, "main.go"), codexFilePath(base, "*** Add File: main.go"))
	assert.Equal(t, "/tmp/main.go", codexFilePath(base, "*** Delete File: /tmp/main.go"))
	assert.Empty(t, codexMoveFilePath(base, "*** Move from: old.go"))
	assert.Empty(t, codexMoveFilePath(base, "*** Move to: "))
	assert.Equal(t, filepath.Join(base, "new.go"), codexMoveFilePath(base, "*** Move to: new.go"))
	assert.Equal(t, "/tmp/new.go", codexMoveFilePath(base, "*** Move to: /tmp/new.go"))

	assert.Empty(t, codexSourceEditor("", ""))
	assert.Equal(t, "codex-cli/unknown", codexSourceEditor("cli", ""))
	assert.Equal(t, "codex-cli/1.2.3", codexSourceEditor("cli", "1.2.3"))
	assert.Equal(t, "codex-vs-code/unknown", codexSourceEditor(" VS Code ", ""))
	assert.Empty(t, codexSourceProduct(" / \\ "))

	callID := "call-1"
	payloadType := "custom_tool_call"
	name := "apply_patch"
	input := "*** Begin Patch\n*** Add File: main.go\n+one\n*** End Patch"
	state := &codexParseState{pendingPatches: map[string]codexPendingPatch{}}
	session := codexSessionState{cwd: base, id: "session-1", version: "0.1.0", source: "cli"}

	heartbeats, handled := parser.handlePendingPatch(codexLogLine{}, session, state)
	assert.False(t, handled)
	assert.Nil(t, heartbeats)

	heartbeats, handled = parser.handlePendingPatch(codexLogLine{Payload: &codexPayload{
		Type: &payloadType,
	}}, session, state)
	assert.False(t, handled)
	assert.Nil(t, heartbeats)

	emptyCallID := " "
	heartbeats, handled = parser.handlePendingPatch(codexLogLine{Payload: &codexPayload{
		Type:   &payloadType,
		CallID: &emptyCallID,
	}}, session, state)
	assert.False(t, handled)
	assert.Nil(t, heartbeats)

	heartbeats, handled = parser.handlePendingPatch(codexLogLine{Payload: &codexPayload{
		Type:   &payloadType,
		CallID: &callID,
		Name:   &name,
		Input:  &input,
	}}, session, state)
	assert.True(t, handled)
	assert.Nil(t, heartbeats)
	assert.Contains(t, state.pendingPatches, callID)

	unknownType := "unknown"
	heartbeats, handled = parser.handlePendingPatch(codexLogLine{Payload: &codexPayload{
		Type:   &unknownType,
		CallID: &callID,
	}}, session, state)
	assert.False(t, handled)
	assert.Nil(t, heartbeats)

	endType := "patch_apply_end"
	success := false
	heartbeats, handled = parser.handlePendingPatch(codexLogLine{Payload: &codexPayload{
		Type:    &endType,
		CallID:  &callID,
		Success: &success,
	}}, session, state)
	assert.True(t, handled)
	assert.Nil(t, heartbeats)
	assert.NotContains(t, state.pendingPatches, callID)

	state.pendingPatches[callID] = codexPendingPatch{
		inputs:            []string{input},
		cwd:               base,
		sessionID:         "session-1",
		fallbackUserAgent: "fallback/1.0",
	}
	outputType := "custom_tool_call_output"
	heartbeats, handled = parser.handlePendingPatch(codexLogLine{Payload: &codexPayload{
		Type:   &outputType,
		CallID: &callID,
		Output: json.RawMessage(`"invalid patch"`),
	}}, session, state)
	assert.True(t, handled)
	assert.Nil(t, heartbeats)
	assert.NotContains(t, state.pendingPatches, callID)

	state.pendingPatches[callID] = codexPendingPatch{
		inputs:            []string{input},
		cwd:               base,
		sessionID:         "session-1",
		fallbackUserAgent: "fallback/1.0",
	}
	heartbeats, handled = parser.handlePendingPatch(codexLogLine{
		Timestamp: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		Payload: &codexPayload{
			Type:   &outputType,
			CallID: &callID,
			Output: json.RawMessage(`"done"`),
		},
	}, session, state)
	assert.True(t, handled)
	require.NotEmpty(t, heartbeats)
	assert.Equal(t, filepath.Join(base, "main.go"), heartbeats[0].Entity)
}

func TestQwenCodeHelperBranches(t *testing.T) {
	home := t.TempDir()
	resolvedHome, ok := qwenCodeConfiguredRuntimeDir("~", home)
	require.True(t, ok)
	assert.Equal(t, filepath.Clean(home), resolvedHome)
	resolvedSubdir, ok := qwenCodeConfiguredRuntimeDir("~/runtime", home)
	require.True(t, ok)
	assert.Equal(t, filepath.Join(home, "runtime"), resolvedSubdir)
	absDir := filepath.Join(home, "absolute")
	resolvedAbs, ok := qwenCodeConfiguredRuntimeDir(absDir, home)
	require.True(t, ok)
	assert.Equal(t, filepath.Clean(absDir), resolvedAbs)

	_, ok = qwenCodeConfiguredRuntimeDir("relative", home)
	assert.False(t, ok)
	_, ok = qwenCodeConfiguredRuntimeDir(" ", home)
	assert.False(t, ok)

	stripped := string(qwenCodeStripJSONComments([]byte(`{
  // line comment
  "url": "https://example.test//kept",
  "escaped": "quote \" // kept",
  /* block
     comment */
  "value": 1
}`)))
	assert.Contains(t, stripped, `"url": "https://example.test//kept"`)
	assert.NotContains(t, stripped, "line comment")
	assert.NotContains(t, stripped, "block")

	assert.Equal(t, map[string]interface{}{"path": "a.go"}, qwenCodeFunctionCallArgs(json.RawMessage(`{"path":"a.go"}`)))
	assert.Equal(
		t,
		map[string]interface{}{"path": "b.go"},
		qwenCodeFunctionCallArgs(json.RawMessage(`"{\"path\":\"b.go\"}"`)),
	)
	assert.Nil(t, qwenCodeFunctionCallArgs(json.RawMessage(`123`)))

	tokens := QwenCode{}.tokenCounts(qwenCodeRecord{
		Type:          "assistant",
		UsageMetadata: &qwenCodeUsageMetadata{PromptTokenCount: 5, CandidatesTokenCount: 2},
	}, heartbeat.AITokens{LastInput: 10, LastOutput: 20})
	assert.Equal(t, heartbeat.AITokens{LastInput: 10, LastOutput: 20, CurrentInput: 15, CurrentOutput: 22}, tokens)
	assert.Equal(t, heartbeat.AITokens{CurrentInput: 1}, QwenCode{}.tokenCounts(
		qwenCodeRecord{Type: "user"}, heartbeat.AITokens{CurrentInput: 1}))

	base := t.TempDir()
	assert.Equal(t, filepath.Join(base, "a.go"), qwenCodeToolPath(map[string]interface{}{"file_path": "a.go"}, nil, base))
	assert.Equal(t, filepath.Join(base, "b.go"), qwenCodeToolPath(map[string]interface{}{"path": "b.go"}, nil, base))
	assert.Equal(
		t,
		filepath.Join(base, "display.go"),
		qwenCodeToolPath(nil, json.RawMessage(`{"fileName":"display.go"}`), base),
	)
	assert.Empty(t, qwenCodeToolPath(nil, json.RawMessage(`{`), base))

	changes, write, ok := qwenCodeToolLineChanges(qwenCodeToolCall{Name: "read_file"}, nil)
	assert.Equal(t, 0, changes)
	assert.False(t, write)
	assert.True(t, ok)
	changes, write, ok = qwenCodeToolLineChanges(qwenCodeToolCall{
		Name: "edit",
		Args: map[string]interface{}{"old_string": "one", "new_string": "one\ntwo"},
	}, nil)
	assert.Equal(t, 1, changes)
	assert.True(t, write)
	assert.True(t, ok)
	changes, write, ok = qwenCodeToolLineChanges(qwenCodeToolCall{Name: "write_file"}, json.RawMessage(
		`{"diffStat":{"model_added_lines":4,"model_removed_lines":1}}`))
	assert.Equal(t, 3, changes)
	assert.True(t, write)
	assert.True(t, ok)
	_, _, ok = qwenCodeToolLineChanges(qwenCodeToolCall{Name: "unknown"}, nil)
	assert.False(t, ok)
	assert.Empty(t, qwenCodeStringValue(map[string]interface{}{"x": 1}, "x"))
	assert.Empty(t, qwenCodeAbsPath("  ", base))
}

func TestPiHelperBranches(t *testing.T) {
	base := t.TempDir()
	pathFromArg := filepath.Join(base, "arg.go")
	assert.Equal(t, pathFromArg, piToolPath(piMessage{}, map[string]interface{}{"target_file": "arg.go"}, base))

	successPath := filepath.Join(base, "created.go")
	assert.Equal(t, successPath, piToolPath(piMessage{
		Content: []piContentBlock{{Type: "text", Text: "Successfully wrote file to " + successPath}},
	}, nil, ""))
	assert.Empty(t, piToolPath(piMessage{Content: []piContentBlock{{Type: "image", Text: successPath}}}, nil, ""))

	changes, ok := Pi{}.toolLineChanges("edit", piMessage{
		Details: json.RawMessage(`{"diff":"--- a\n+++ b\n-old\n+new\n+extra"}`),
	}, nil)
	assert.True(t, ok)
	assert.Equal(t, 1, changes)
	changes, ok = Pi{}.toolLineChanges("write", piMessage{}, map[string]interface{}{"content": "one\ntwo"})
	assert.True(t, ok)
	assert.Equal(t, 2, changes)
	changes, ok = Pi{}.toolLineChanges("read", piMessage{}, nil)
	assert.False(t, ok)
	assert.Equal(t, 0, changes)
	changes, ok = Pi{}.toolLineChanges("patch", piMessage{}, nil)
	assert.True(t, ok)
	assert.Equal(t, 0, changes)

	assert.Empty(t, Pi{}.diffFromDetails(nil))
	assert.Empty(t, Pi{}.diffFromDetails(json.RawMessage(`{`)))
	assert.Empty(t, piStringValue(map[string]interface{}{"x": 1}, "x"))
	assert.Empty(t, piAbsPath("  ", base))
	assert.Equal(t, filepath.Join(base, "relative.go"), piAbsPath("relative.go", base))
	assert.Equal(t, "session", Pi{}.sessionIDFromPath(filepath.Join(base, "prefix_session.jsonl")))
}

func TestClineAndRooToolBranches(t *testing.T) {
	base := t.TempDir()

	edited := Cline{}.toolHeartbeat(clineToolMessage{
		Tool: "editedExistingFile",
		Path: "edited.go",
		Diff: "<<<<<<< SEARCH\nold\n=======\nnew\nextra\n>>>>>>> REPLACE",
	}, base)
	require.True(t, edited.ok)
	assert.Equal(t, filepath.Join(base, "edited.go"), edited.filePath)
	assert.Equal(t, 1, edited.lineChanges)
	assert.True(t, edited.isWrite)

	created := Cline{}.toolHeartbeat(clineToolMessage{Tool: "newFileCreated", Path: "new.go", Content: "one\ntwo"}, base)
	require.True(t, created.ok)
	assert.Equal(t, 2, created.lineChanges)

	read := Cline{}.toolHeartbeat(clineToolMessage{Tool: "readFile", Path: "read.go"}, base)
	require.True(t, read.ok)
	assert.False(t, read.isWrite)

	deleted := Cline{}.toolHeartbeat(clineToolMessage{Tool: "fileDeleted", Path: "gone.go", Content: "one\ntwo"}, base)
	require.True(t, deleted.ok)
	assert.Equal(t, -2, deleted.lineChanges)
	assert.False(t, Cline{}.toolHeartbeat(clineToolMessage{Tool: "fileDeleted", Path: "gone.go"}, base).ok)
	assert.False(t, Cline{}.toolHeartbeat(clineToolMessage{Tool: "unknown", Path: "x.go"}, base).ok)

	assert.Equal(t, "task text", clineTaskText("prefix <task> task text </task> suffix"))
	assert.Empty(t, clineTaskText("missing"))
	assert.Equal(
		t,
		filepath.Clean(base),
		clineCurrentWorkingDirectory("# Current Working Directory ("+filepath.Clean(base)+")"),
	)
	assert.Empty(t, clineCurrentWorkingDirectory("# Current Working Directory ("))

	value, ok := clineBetween("prefix [value] suffix", "[", "]")
	require.True(t, ok)
	assert.Equal(t, "value", value)

	_, ok = clineBetween("prefix [value", "[", "]")
	assert.False(t, ok)

	rooPath, rooChanges, ok := rooToolHeartbeat(rooToolAsk{
		Tool: "appliedDiff",
		Path: "roo.go",
		Diff: "<<<<<<< SEARCH\nold\n=======\nnew\nextra\n>>>>>>> REPLACE",
	}, base)
	require.True(t, ok)
	assert.Equal(t, filepath.Join(base, "roo.go"), rooPath)
	assert.Equal(t, 1, rooChanges)
	rooPath, rooChanges, ok = rooToolHeartbeat(rooToolAsk{Tool: "writeToFile", Path: "write.go", Content: "one\ntwo"}, base)
	require.True(t, ok)
	assert.Equal(t, filepath.Join(base, "write.go"), rooPath)
	assert.Equal(t, 2, rooChanges)

	_, _, ok = rooToolHeartbeat(rooToolAsk{Tool: "writeToFile", Path: "empty.go"}, base)
	assert.False(t, ok)
	_, _, ok = rooToolHeartbeat(rooToolAsk{Tool: "unknown", Path: "x.go"}, base)
	assert.False(t, ok)
	assert.Empty(t, rooCurrentWorkingDirectory("# Current Working Directory ("))
}

func TestCopilotHelperBranches(t *testing.T) {
	base := t.TempDir()
	filePath := filepath.Join(base, "main.go")
	require.NoError(t, os.WriteFile(filePath, []byte("package main\n"), 0600))

	parser := Copilot{FallbackUserAgent: "fallback/1.0"}
	assert.Empty(t, parser.messageText(nil))
	assert.Equal(t, "hello", parser.messageText(&copilotMessage{Text: "hello"}))
	assert.Empty(t, parser.version(nil))
	assert.Equal(t, "1.2.3", parser.version(&copilotAgent{ExtensionVersion: "1.2.3"}))
	assert.True(t, parser.millisTime(0).IsZero())
	assert.Equal(t, time.UnixMilli(1234), parser.millisTime(1234))

	request := copilotRequest{
		Timestamp: 1000,
		VariableData: &copilotVariableData{Variables: []copilotVariable{
			{Value: &copilotPathValue{URI: &copilotFileURI{FSPath: filePath}}},
			{ID: parser.pathToFileURI(filePath)},
		}},
		Response: []copilotResponseItem{
			{
				Kind: "markdown",
				InvocationMessage: &copilotMessageWithURIs{URIs: map[string]copilotReferenceURI{
					"file://ignored": {Path: filePath},
				}},
			},
			{InlineReference: &copilotInlineReference{URI: &struct {
				FSPath string `json:"fsPath"`
			}{FSPath: filePath}}},
		},
	}
	assert.Equal(t, time.UnixMilli(1000).Add(time.Second), parser.requestCompletedAt(request))
	assert.Equal(t, filepath.Dir(filePath), parser.requestProjectPath(request))
	assert.Equal(t, []string{filePath}, parser.readFiles(request))
	assert.Empty(t, parser.requestProjectPath(copilotRequest{}))
	assert.True(t, parser.requestCompletedAt(copilotRequest{}).IsZero())
	assert.Equal(t, time.UnixMilli(2000), parser.requestCompletedAt(copilotRequest{
		ModelState: &copilotModelState{CompletedAt: 2000},
	}))

	assert.Equal(t, "replacement", parser.applyTextEdit("one\ntwo", copilotTextEdit{Text: "replacement"}))
	assert.Equal(t, "one\nTWO", parser.applyTextEdit("one\ntwo", copilotTextEdit{
		Text: "TWO",
		Range: &copilotTextRange{
			StartLineNumber: 2,
			StartColumn:     1,
			EndLineNumber:   2,
			EndColumn:       4,
		},
	}))
	assert.Equal(t, "Xtwo", parser.applyTextEdit("one\ntwo", copilotTextEdit{
		Text: "X",
		Range: &copilotTextRange{
			StartLineNumber: 2,
			StartColumn:     1,
			EndLineNumber:   1,
			EndColumn:       1,
		},
	}))
	assert.Equal(t, 0, parser.offset("one", 1, 1))
	assert.Equal(t, len("one"), parser.offset("one", 9, 9))
	assert.Equal(t, 0, parser.lineCount(""))
	assert.Equal(t, 2, parser.lineCount("one\ntwo"))
}

func TestCopilotJSONLAndSessionBranches(t *testing.T) {
	base := t.TempDir()
	readPath := filepath.Join(base, "read.go")
	require.NoError(t, os.WriteFile(readPath, []byte("package main\n"), 0600))

	parser := Copilot{
		FallbackUserAgent: "fallback/1.0",
		UserAgents: map[string]string{
			readPath: "editor/1.0",
		},
	}

	sessionPath := filepath.Join(base, "session.jsonl")
	lines := []string{
		"",
		`not json`,
		mustJSON(t, map[string]interface{}{
			"kind": 1,
			"k":    []interface{}{"requests", 0, "message"},
			"v":    map[string]interface{}{"text": "ignored before root"},
		}),
		mustJSON(t, map[string]interface{}{
			"kind": 0,
			"v": map[string]interface{}{
				"sessionId": "session-1",
				"requests": []map[string]interface{}{
					{
						"requestId": "request-1",
						"timestamp": int64(1000),
						"agent":     map[string]interface{}{"extensionVersion": "1.2.3"},
						"usage":     map[string]interface{}{"promptTokens": 5, "completionTokens": 2},
					},
				},
			},
		}),
		mustJSON(t, map[string]interface{}{
			"kind": 1,
			"k":    []interface{}{"requests", 0, "modelState"},
			"v":    map[string]interface{}{"completedAt": int64(2200)},
		}),
		mustJSON(t, map[string]interface{}{
			"kind": 1,
			"k":    []interface{}{"requests", 0, "message"},
			"v":    map[string]interface{}{"text": "inspect this file"},
		}),
		mustJSON(t, map[string]interface{}{
			"kind": 1,
			"k":    []interface{}{"requests", 0, "variableData"},
			"v": map[string]interface{}{"variables": []map[string]interface{}{
				{"value": map[string]interface{}{"fsPath": readPath}},
			}},
		}),
		mustJSON(t, map[string]interface{}{
			"kind": 1,
			"k":    []interface{}{"requests", 0, "response"},
			"v": []map[string]interface{}{
				{"kind": "markdown"},
			},
		}),
		mustJSON(t, map[string]interface{}{
			"kind": 1,
			"k":    []interface{}{"requests", 0, "result"},
			"v": map[string]interface{}{
				"metadata": map[string]interface{}{
					"usage": map[string]interface{}{"inputTokens": 3, "outputTokens": 4},
				},
			},
		}),
		mustJSON(t, map[string]interface{}{
			"kind": 2,
			"k":    []interface{}{"requests"},
			"i":    1,
			"v": []map[string]interface{}{
				{
					"requestId": "request-2",
					"timestamp": int64(3000),
					"message":   map[string]interface{}{"text": "second prompt"},
				},
			},
		}),
		mustJSON(t, map[string]interface{}{
			"kind": 2,
			"k":    []interface{}{"requests", 0, "response"},
			"i":    0,
			"v": []map[string]interface{}{
				{
					"kind":            "inlineReference",
					"inlineReference": map[string]interface{}{"fsPath": readPath},
				},
			},
		}),
	}
	require.NoError(t, os.WriteFile(sessionPath, []byte(strings.Join(lines, "\n")+"\n"), 0600))

	session, err := parser.parseJSONLSession(sessionPath)
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Len(t, session.Requests, 2)
	assert.Equal(t, "inspect this file", session.Requests[0].Message.Text)
	require.NotNil(t, session.Requests[0].ModelState)
	assert.EqualValues(t, 2200, session.Requests[0].ModelState.CompletedAt)
	assert.Len(t, session.Requests[0].Response, 2)
	assert.Equal(t, "second prompt", session.Requests[1].Message.Text)

	timed, requestMeta := parser.sessionHeartbeats(copilotWorkspaceState{
		sessions: map[string]*copilotSession{session.SessionID: session},
	})
	require.Len(t, timed, 4)
	assert.Equal(t, "session-1", requestMeta["request-1"].sessionID)
	assert.Equal(t, "github-copilot/1.2.3", requestMeta["request-1"].plugin)
	assert.Equal(t, heartbeat.AppType, timed[0].heartbeat.EntityType)
	assert.Equal(t, len([]rune("inspect this file")), timed[0].heartbeat.AIPromptLength)
	assert.EqualValues(t, 5, timed[0].heartbeat.AIInputTokens)
	assert.EqualValues(t, 2, timed[0].heartbeat.AIOutputTokens)
	assert.Equal(t, readPath, timed[1].heartbeat.Entity)
	assert.Equal(t, heartbeat.FileType, timed[1].heartbeat.EntityType)
	assert.Contains(t, timed[1].heartbeat.UserAgent, "editor/1.0")
	assert.Equal(t, "Copilot session-1", timed[3].heartbeat.Entity)
	assert.Equal(t, len([]rune("second prompt")), timed[3].heartbeat.AIPromptLength)

	setSession := &copilotSession{Requests: []copilotRequest{{RequestID: "request"}}}
	parser.applyJSONLSet(setSession, copilotJSONLPatch{})
	parser.applyJSONLSet(setSession, copilotJSONLPatch{K: []json.RawMessage{json.RawMessage(`"other"`)}})
	parser.applyJSONLSet(setSession, copilotJSONLPatch{
		K: []json.RawMessage{json.RawMessage(`"requests"`), json.RawMessage(`99`)},
	})
	parser.applyJSONLSet(setSession, copilotJSONLPatch{
		K: []json.RawMessage{
			json.RawMessage(`"requests"`),
			json.RawMessage(`0`),
			json.RawMessage(`123`),
		},
	})
	assert.Equal(t, "request", setSession.Requests[0].RequestID)

	insertSession := &copilotSession{Requests: []copilotRequest{{RequestID: "request"}}}
	parser.applyJSONLInsert(insertSession, copilotJSONLPatch{})
	parser.applyJSONLInsert(insertSession, copilotJSONLPatch{K: []json.RawMessage{json.RawMessage(`"other"`)}})
	parser.applyJSONLInsert(insertSession, copilotJSONLPatch{
		K: []json.RawMessage{json.RawMessage(`"requests"`), json.RawMessage(`0`), json.RawMessage(`"message"`)},
		V: json.RawMessage(`[]`),
	})
	assert.Len(t, insertSession.Requests, 1)

	assert.Equal(t,
		[]copilotRequest{{RequestID: "a"}, {RequestID: "b"}},
		parser.insertRequests([]copilotRequest{{RequestID: "a"}}, 99, []copilotRequest{{RequestID: "b"}}),
	)
	assert.Equal(t,
		[]copilotResponseItem{{Kind: "a"}, {Kind: "b"}},
		parser.insertResponses([]copilotResponseItem{{Kind: "a"}}, -1, []copilotResponseItem{{Kind: "b"}}),
	)
	_, ok := parser.parseJSONString(json.RawMessage(`123`))
	assert.False(t, ok)
	_, ok = parser.parseJSONInt(json.RawMessage(`"not-int"`))
	assert.False(t, ok)
}

func TestCopilotEditStateBranches(t *testing.T) {
	base := t.TempDir()
	statePath := filepath.Join(base, "state.json")
	filePath := filepath.Join(base, "main.go")
	require.NoError(t, os.WriteFile(filePath, []byte("one\ntwo\n"), 0600))

	parser := Copilot{
		FallbackUserAgent: "fallback/1.0",
		UserAgents: map[string]string{
			filePath: "editor/1.0",
		},
	}
	timestamp := time.Date(2026, 6, 24, 18, 0, 0, 0, time.UTC)
	state := map[string]interface{}{
		"timeline": map[string]interface{}{
			"fileBaselines": []interface{}{
				[]interface{}{"bad"},
				[]interface{}{123, map[string]interface{}{"content": "ignored"}},
				[]interface{}{
					"file://" + filepath.ToSlash(filePath) + "::request-1",
					map[string]interface{}{"content": "one\ntwo\n"},
				},
				[]interface{}{
					"file://" + filePath + "::request-1",
					map[string]interface{}{"content": "one\ntwo\n"},
				},
				[]interface{}{
					parser.pathToFileURI(filePath) + "::request-1",
					map[string]interface{}{"content": "one\ntwo\n"},
				},
			},
			"operations": []map[string]interface{}{
				{"type": "other"},
				{
					"type":      "textEdit",
					"requestId": "missing",
					"uri":       map[string]interface{}{"fsPath": filePath},
					"edits":     []interface{}{map[string]interface{}{"text": "x"}},
				},
				{
					"type":      "textEdit",
					"requestId": "request-1",
					"uri":       map[string]interface{}{"fsPath": filePath},
					"edits": []interface{}{map[string]interface{}{
						"text": "three",
						"range": map[string]interface{}{
							"startLineNumber": 2,
							"startColumn":     1,
							"endLineNumber":   2,
							"endColumn":       4,
						},
					}},
				},
				{
					"type":      "textEdit",
					"requestId": "request-1",
					"uri":       map[string]interface{}{"fsPath": filePath},
					"edits": []interface{}{map[string]interface{}{
						"text": "\nfour",
						"range": map[string]interface{}{
							"startLineNumber": 2,
							"startColumn":     6,
							"endLineNumber":   2,
							"endColumn":       6,
						},
					}},
				},
			},
		},
	}
	require.NoError(t, os.WriteFile(statePath, []byte(mustJSON(t, state)), 0600))

	timed, err := parser.parseEditState(statePath, map[string]copilotRequestMeta{
		"request-1": {
			timestamp: timestamp,
			plugin:    "github-copilot/1.2.3",
			sessionID: "session-1",
		},
	})
	require.NoError(t, err)
	require.Len(t, timed, 2)
	assert.Equal(t, filePath, timed[0].heartbeat.Entity)
	require.NotNil(t, timed[0].heartbeat.AILineChanges)
	assert.Equal(t, 0, *timed[0].heartbeat.AILineChanges)
	require.NotNil(t, timed[1].heartbeat.AILineChanges)
	assert.Equal(t, 1, *timed[1].heartbeat.AILineChanges)
	assert.Contains(t, timed[0].heartbeat.UserAgent, "editor/1.0")

	missing, err := parser.parseEditState(filepath.Join(base, "missing.json"), nil)
	require.NoError(t, err)
	assert.Nil(t, missing)
	require.NoError(t, os.WriteFile(statePath, []byte(`{`), 0600))
	_, err = parser.parseEditState(statePath, nil)
	require.Error(t, err)
	require.NoError(t, os.WriteFile(statePath, []byte(`{}`), 0600))
	empty, err := parser.parseEditState(statePath, nil)
	require.NoError(t, err)
	assert.Nil(t, empty)

	assert.False(t, parser.fileModifiedAfter(filepath.Join(base, "missing.json"), time.Time{}))
}

func TestCopilotCLIStateBranches(t *testing.T) {
	parser := Copilot{
		FallbackUserAgent: "fallback/1.0",
		UserAgents: map[string]string{
			"/tmp/main.go": "editor/1.0",
		},
	}
	state := copilotCLIParseState{
		sessionID:        "session-1",
		model:            "gpt-5",
		cliVersion:       "1.2.3",
		agentVersion:     "4.5.6",
		gitRoot:          "/repo",
		cwd:              "/fallback",
		fileHeartbeatIDs: make(map[string]struct{}),
	}
	assert.Equal(t, "/repo", state.projectPath())
	state.appendFileHeartbeat(parser, "/tmp/main.go", time.Date(2026, 6, 24, 18, 0, 0, 0, time.UTC), heartbeat.PointerTo(3))
	require.Len(t, state.heartbeats, 1)
	assert.Contains(t, state.fileHeartbeatIDs, "/tmp/main.go")
	require.NotNil(t, state.heartbeats[0].heartbeat.AILineChanges)
	assert.Equal(t, 3, *state.heartbeats[0].heartbeat.AILineChanges)
	state.trackCLIFilePath(" ")
	assert.Len(t, state.fileHeartbeatIDs, 1)

	telemetry := &copilotCLIToolTelemetry{}
	result := &copilotCLIToolResult{ToolTelemetry: telemetry}
	assert.Equal(t, telemetry, copilotCLIToolExecutionCompleteData{ToolTelemetry: telemetry}.telemetry())
	assert.Equal(t, telemetry, copilotCLIToolExecutionCompleteData{Result: result}.telemetry())
	assert.Nil(t, copilotCLIToolExecutionCompleteData{}.telemetry())
	assert.False(t, telemetry.hasWriteSignals("read_file"))

	pathsTelemetry := copilotCLIToolTelemetry{
		RestrictedProperties: &copilotCLIToolRestrictedProperties{
			AddedPaths:   json.RawMessage(`["/tmp/a.go"]`),
			DeletedPaths: json.RawMessage(`["/tmp/b.go","/tmp/a.go"]`),
		},
	}
	assert.Equal(t, []string{"/tmp/a.go", "/tmp/b.go"}, pathsTelemetry.filePaths())
	assert.True(t, pathsTelemetry.hasWriteSignals("read_file"))

	value, ok := (copilotCLIToolMetrics{LinesRemoved: heartbeat.PointerTo(2)}).lineChanges()
	assert.True(t, ok)
	assert.Equal(t, -2, value)

	_, ok = (copilotCLIToolMetrics{}).lineChanges()
	assert.False(t, ok)
}

func TestAntigravityHelperBranches(t *testing.T) {
	home := t.TempDir()
	appRoot := filepath.Join(home, "antigravity")
	require.NoError(t, os.MkdirAll(filepath.Join(appRoot, "bin"), 0700))
	executable := filepath.Join(home, "Antigravity.app", "Contents", "MacOS", "agentapi")
	require.NoError(t, os.WriteFile(filepath.Join(appRoot, "bin", "agentapi"), []byte(`exec "`+executable+`"`), 0600))

	plistPath := filepath.Join(home, "Antigravity.app", "Contents", "Info.plist")
	require.NoError(t, os.MkdirAll(filepath.Dir(plistPath), 0700))
	require.NoError(t, os.WriteFile(plistPath, []byte(
		`<key>CFBundleShortVersionString</key><string>1.2.3</string>`), 0600))
	assert.Equal(t, "1.2.3", antigravityProductVersion(home, appRoot, antigravityProduct{
		userAgentProduct: "antigravity-desktop",
	}))

	ideRoot := filepath.Join(home, "ide", "resources", "app")
	require.NoError(t, os.MkdirAll(filepath.Join(ideRoot, "bin"), 0700))
	ideExecutable := filepath.Join(ideRoot, "extensions", "agentapi")
	require.NoError(t, os.WriteFile(filepath.Join(ideRoot, "bin", "agentapi"), []byte(`exec "`+ideExecutable+`"`), 0600))
	productPath := filepath.Join(ideRoot, "product.json")
	require.NoError(t, os.WriteFile(productPath, []byte(`{"version":"2.0.0-beta"}`), 0600))
	assert.Equal(t, "2.0.0-beta", antigravityProductVersion(home, ideRoot, antigravityProduct{
		userAgentProduct: "antigravity-ide",
	}))
	assert.Equal(t, "unknown", antigravityProductVersion(home, appRoot, antigravityProduct{
		userAgentProduct: "antigravity-cli",
	}))
	assert.Empty(t, antigravityAgentAPIExecutable(filepath.Join(home, "missing")))
	assert.Empty(t, antigravityJSONVersion(filepath.Join(home, "missing.json")))
	assert.Empty(t, antigravityPlistVersion(filepath.Join(home, "missing.plist")))
	assert.Equal(t, "2.0.0-beta", antigravityVersionString("Gemini v2.0.0-beta"))
	assert.Empty(t, antigravityVersionString("singleword"))

	actionPath := filepath.Join(home, "project", "main.go")
	content := "The following changes were made by the model to: `" + actionPath + "`. If relevant\n" +
		"[diff_block_start]\n--- a\n+++ b\n-old\n+new\n+extra\n[diff_block_end]"
	assert.Equal(t, actionPath, antigravityCodeActionPath(content))
	changes, ok := antigravityCodeActionLineChanges(content)
	require.True(t, ok)
	assert.Equal(t, 1, changes)

	_, ok = antigravityCodeActionLineChanges("no diff")
	assert.False(t, ok)
	assert.Empty(t, antigravityCodeActionPath("no marker"))

	transcriptPath := filepath.Join(appRoot, "a", "b", "c", "d", "session.jsonl")
	require.NoError(t, os.MkdirAll(filepath.Dir(transcriptPath), 0700))

	cachePath := filepath.Join(appRoot, "cache", "last_conversations.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0700))

	workspace := filepath.Join(home, "workspace")
	require.NoError(t, os.WriteFile(cachePath, []byte(`{`+mustJSON(t, workspace)+`:"session-1"}`), 0600))
	assert.Equal(t, workspace, antigravityProjectPath(antigravityTranscript{path: transcriptPath, sessionID: "session-1"}))

	require.NoError(t, os.WriteFile(cachePath, []byte(`{}`), 0600))

	historyPath := filepath.Join(appRoot, "history.jsonl")
	require.NoError(t, os.WriteFile(
		historyPath,
		[]byte(`{"conversationId":"session-2","workspace":`+mustJSON(t, workspace)+`}`+"\n"),
		0600,
	))
	assert.Equal(t, workspace, antigravityProjectPath(antigravityTranscript{path: transcriptPath, sessionID: "session-2"}))
	assert.Empty(t, antigravityProjectPath(antigravityTranscript{path: transcriptPath, sessionID: "missing"}))

	assert.Empty(t, antigravityFilePath(" "))
	assert.Equal(t, filepath.FromSlash("/tmp/main.go"), antigravityFilePath("file:///tmp/main.go"))
	assert.Equal(t, filepath.FromSlash("C:/tmp/main.go"), antigravityFilePath("file:///C:/tmp/main.go"))
}

func TestContinueAdditionalBranches(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	parser := Continue{
		After:             time.Date(2026, 1, 2, 3, 4, 4, 0, time.UTC),
		FallbackUserAgent: "fallback/1.0",
		UserAgents: map[string]string{
			filepath.Join(home, "workspace", "read.go"): "editor/1.0",
		},
	}

	root, err := parser.root(context.Background())
	require.NoError(t, err)
	assert.Empty(t, root)

	continueRoot := filepath.Join(home, ".continue")
	require.NoError(t, os.MkdirAll(continueRoot, 0700))

	root, err = parser.root(context.Background())
	require.NoError(t, err)
	assert.Equal(t, continueRoot, root)

	events, err := parser.eventsFromFile(filepath.Join(home, "missing.jsonl"), nil)
	require.NoError(t, err)
	assert.Nil(t, events)

	sessionPath := filepath.Join(continueRoot, "sessions", "sessions.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(sessionPath), 0700))
	require.NoError(t, os.WriteFile(sessionPath, []byte(mustJSON(t, []continueSession{
		{SessionID: "session-1", WorkspaceDirectory: "file://" + filepath.ToSlash(filepath.Join(home, "workspace"))},
		{SessionID: "", WorkspaceDirectory: "/ignored"},
	})), 0600))
	workspaces := parser.sessionWorkspaces(sessionPath)
	assert.Equal(t, filepath.Join(home, "workspace"), workspaces["session-1"])
	assert.Empty(t, parser.sessionWorkspaces(filepath.Join(home, "bad-sessions.json")))
	require.NoError(t, os.WriteFile(sessionPath, []byte(`{`), 0600))
	assert.Empty(t, parser.sessionWorkspaces(sessionPath))

	logPath := filepath.Join(home, "events.jsonl")
	lines := []string{
		"",
		`not json`,
		mustJSON(t, map[string]interface{}{
			"eventName":     "chatInteraction",
			"timestamp":     "2026-01-02T03:04:05Z",
			"userAgent":     "continue/1.0",
			"prompt":        "<system>ignore</system><user>write tests</user><meta>",
			"modelName":     "gpt-5",
			"modelProvider": "openai",
			"sessionId":     "session-1",
		}),
		mustJSON(t, map[string]interface{}{
			"eventName":    "toolUsage",
			"timestamp":    "2026-01-02T03:04:06Z",
			"userAgent":    "continue/1.0",
			"functionName": "read_file",
			"accepted":     true,
			"succeeded":    true,
			"toolCallArgs": mustJSON(t, map[string]string{"filepath": filepath.Join(home, "workspace", "read.go")}),
		}),
		mustJSON(t, map[string]interface{}{
			"eventName": "toolUsage",
			"timestamp": "2026-01-02T03:04:07Z",
		}),
		mustJSON(t, map[string]interface{}{
			"eventName":     "editOutcome",
			"timestamp":     "2026-01-02T03:04:08Z",
			"userAgent":     "continue/1.0",
			"prompt":        "edit file",
			"modelName":     "gpt-5",
			"modelProvider": "openai",
			"accepted":      true,
			"lineChange":    2,
			"filepath":      "relative.go",
		}),
		mustJSON(t, map[string]interface{}{
			"eventName": "editOutcome",
			"timestamp": "2026-01-02T03:04:09Z",
		}),
		mustJSON(t, map[string]interface{}{
			"eventName":       "tokensGenerated",
			"timestamp":       "2026-01-02T03:04:05Z",
			"model":           "gpt-5",
			"provider":        "openai",
			"promptTokens":    7,
			"generatedTokens": 3,
		}),
		mustJSON(t, map[string]interface{}{"eventName": "ignored"}),
	}
	require.NoError(t, os.WriteFile(logPath, []byte(strings.Join(lines, "\n")+"\n"), 0600))

	events, err = parser.eventsFromFile(logPath, workspaces)
	require.NoError(t, err)
	require.Len(t, events, 4)

	heartbeats := parser.heartbeats(events)
	require.Len(t, heartbeats, 3)
	assert.Equal(t, "Continue session-1", heartbeats[0].Entity)
	assert.EqualValues(t, 7, heartbeats[0].AIInputTokens)
	assert.EqualValues(t, 3, heartbeats[0].AIOutputTokens)
	assert.Equal(t, filepath.Join(home, "workspace", "read.go"), heartbeats[1].Entity)
	assert.Equal(t, filepath.Join(home, "workspace", "relative.go"), heartbeats[2].Entity)

	_, ok := parser.eventFromLine([]byte(`{"eventName":"chatInteraction","timestamp":1}`), nil)
	assert.False(t, ok)
	_, ok = parser.eventFromLine([]byte(`{"eventName":"tokensGenerated","timestamp":1}`), nil)
	assert.False(t, ok)
	assert.Nil(t, Continue{}.tokensForEvent(continueEvent{}, nil))
	assert.Nil(t, Continue{}.tokensForEvent(
		continueEvent{Timestamp: time.Unix(10, 0), Model: "a"},
		&continueEvent{Timestamp: time.Unix(10, 0), Model: "b"},
	))
	assert.Nil(t, Continue{}.tokensForEvent(
		continueEvent{Timestamp: time.Unix(10, 0), Provider: "a"},
		&continueEvent{Timestamp: time.Unix(10, 0), Provider: "b"},
	))
	assert.Nil(t, Continue{}.tokensForEvent(
		continueEvent{Timestamp: time.Unix(10, 0)},
		&continueEvent{Timestamp: time.Unix(16, 0)},
	))
	assert.NotNil(t, Continue{}.tokensForEvent(
		continueEvent{Timestamp: time.Unix(10, 0), Model: "GPT-5", Provider: "OpenAI"},
		&continueEvent{Timestamp: time.Unix(8, 0), Model: "gpt-5", Provider: "openai"},
	))
	assert.Equal(t, len([]rune("keep")), Continue{}.promptLength("<user> keep <ignore>"))

	encodedArgs, err := json.Marshal(`{"filepath":"file:///tmp/main.go"}`)
	require.NoError(t, err)
	assert.Equal(t, filepath.FromSlash("/tmp/main.go"), Continue{}.toolFilePath(encodedArgs))
	assert.Equal(t, filepath.FromSlash("C:/tmp/main.go"), Continue{}.filePath("file:///C:/tmp/main.go"))
}

func TestCodyAdditionalBranches(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	parser := Cody{
		After:             time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		FallbackUserAgent: "fallback/1.0",
	}

	paths, err := parser.stateDBPaths(context.Background())
	require.NoError(t, err)
	assert.Empty(t, paths)

	dbPath := filepath.Join(home, ".config", "Code", "User", "globalStorage", "state.vscdb")
	require.NoError(t, os.MkdirAll(filepath.Dir(dbPath), 0700))
	require.NoError(t, os.WriteFile(dbPath, []byte("db"), 0600))

	paths, err = parser.stateDBPaths(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{dbPath}, paths)
	assert.True(t, parser.stateDBModifiedAfter(dbPath, time.Time{}))
	assert.False(t, parser.stateDBModifiedAfter(filepath.Join(home, "missing.vscdb"), time.Now()))

	old := time.Now().Add(-time.Hour)
	require.NoError(t, os.Chtimes(dbPath, old, old))
	assert.False(t, parser.stateDBModifiedAfter(dbPath, time.Now()))

	emptyDBPath := filepath.Join(home, "empty.vscdb")
	db, err := sql.Open("sqlite", emptyDBPath)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE TABLE ItemTable (key TEXT UNIQUE, value BLOB)`)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	value, err := parser.queryStorage(context.Background(), emptyDBPath)
	require.NoError(t, err)
	assert.Empty(t, value)

	badDBPath := filepath.Join(home, "bad.vscdb")
	db, err = sql.Open("sqlite", badDBPath)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	_, err = parser.queryStorage(context.Background(), badDBPath)
	require.Error(t, err)

	assert.Nil(t, parser.heartbeatsFromStorage(""))
	assert.Nil(t, parser.heartbeatsFromStorage(`{`))
	assert.Nil(t, parser.chatHistory(`{"cody-local-chatHistory-v2":{`))

	encodedHistory := mustJSON(t, map[string]interface{}{
		"account": map[string]interface{}{
			"chat": map[string]interface{}{
				"fallback-id": map[string]interface{}{
					"lastInteractionTimestamp": "2026-01-02T03:04:06Z",
					"interactions": []map[string]interface{}{
						{"humanMessage": map[string]interface{}{"text": "latest prompt"}},
					},
				},
			},
		},
	})
	storage := mustJSON(t, map[string]interface{}{"cody-local-chatHistory-v2": encodedHistory})
	heartbeats := parser.heartbeatsFromStorage(storage)
	require.Len(t, heartbeats, 1)
	assert.Equal(t, "Cody fallback-id", heartbeats[0].Entity)
	assert.Equal(t, len([]rune("latest prompt")), heartbeats[0].AIPromptLength)

	staleStorage := mustJSON(t, map[string]interface{}{
		"cody-local-chatHistory-v2": map[string]interface{}{
			"account": map[string]interface{}{
				"chat": map[string]interface{}{
					"old": map[string]interface{}{
						"lastInteractionTimestamp": "2026-01-02T03:04:04Z",
						"interactions": []map[string]interface{}{
							{"humanMessage": map[string]interface{}{"text": "old prompt"}},
						},
					},
					"blank": map[string]interface{}{
						"lastInteractionTimestamp": "2026-01-02T03:04:06Z",
						"interactions": []map[string]interface{}{
							{"humanMessage": map[string]interface{}{"text": "  "}},
						},
					},
				},
			},
		},
	})
	assert.Nil(t, parser.heartbeatsFromStorage(staleStorage))
}

func TestCodyContextAndDiffBranches(t *testing.T) {
	timestamp := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	parser := Cody{FallbackUserAgent: "fallback/1.0"}

	_, ok := codyLastInteraction([]codyInteraction{{HumanMessage: codyMessage{Text: "  "}}})
	assert.False(t, ok)

	interaction, ok := codyLastInteraction([]codyInteraction{{
		AssistantMessage: &codyMessage{Model: "claude-3"},
	}})
	require.True(t, ok)
	assert.Equal(t, "claude-3", interaction.AssistantMessage.Model)

	assert.Nil(t, parser.tokens(codyTokenUsage{}))
	require.NotNil(t, parser.tokens(codyTokenUsage{PromptTokens: 2}))
	require.NotNil(t, parser.tokens(codyTokenUsage{CompletionTokens: 3}))

	readPath := filepath.Join(t.TempDir(), "read.go")
	items := codyContextItems(codyInteraction{
		HumanMessage: codyMessage{ContextFiles: []codyContextItem{{URI: codyURI{FSPath: readPath}}}},
	})
	require.Len(t, items, 1)
	assert.Equal(t, readPath, items[0].filePath())

	allItems := codyContextItems(codyInteraction{
		HumanMessage: codyMessage{ContextFiles: []codyContextItem{{URI: codyURI{FSPath: "human.go"}}}},
		AssistantMessage: &codyMessage{
			ContextFiles: []codyContextItem{{URI: codyURI{FSPath: "assistant.go"}}},
			Processes: []codyProcessStep{{
				Items: []codyContextItem{{URI: codyURI{FSPath: "process.go"}}},
			}},
			SubMessages: []codySubMessage{{
				ContextFiles: []codyContextItem{{URI: codyURI{FSPath: "sub.go"}}},
			}},
		},
	})
	assert.Len(t, allItems, 4)

	assert.Nil(t, parser.fileHeartbeat(timestamp, "session", "", codyContextItem{}))
	assert.Nil(t, parser.fileHeartbeat(timestamp, "session", "", codyContextItem{
		ToolName: "text_editor",
		URI:      codyURI{FSPath: readPath},
		Metadata: json.RawMessage(`["only one side"]`),
	}))

	h := parser.fileHeartbeat(timestamp, "session", "gpt-5", codyContextItem{
		URI: codyURI{Path: "file://%"},
	})
	require.NotNil(t, h)
	assert.Equal(t, "%", h.Entity)

	assert.Equal(t, filepath.FromSlash("C:/tmp/main.go"), codyContextItem{
		URI: codyURI{Path: "file://C:/tmp/main.go"},
	}.filePath())
	assert.Equal(t, filepath.FromSlash("C:/tmp/main.go"), codyContextItem{
		URI: codyURI{FSPath: filepath.FromSlash("/C:/tmp/main.go")},
	}.filePath())

	changes, ok := codyContextItem{Metadata: json.RawMessage(`["old\n","new\n"]`)}.lineChanges()
	require.True(t, ok)
	assert.Equal(t, 2, changes)

	_, ok = codyContextItem{Metadata: json.RawMessage(`{}`)}.lineChanges()
	assert.False(t, ok)
	assert.Equal(t, 2, codyDiffLineChanges("", "a\nb\n"))
	assert.Equal(t, 2, codyDiffLineChanges("a\nb\n", ""))
	assert.Equal(t, 1, codyDiffLineChanges(strings.Repeat("a\n", 2001), strings.Repeat("a\n", 2002)))
	assert.Equal(t, 1, codyDiffLineChanges(strings.Repeat("a\n", 2002), strings.Repeat("a\n", 2001)))
	assert.Nil(t, codyContentLines(""))
	assert.Equal(t, 1, codyLCSLineCount([]string{"a", "b"}, []string{"b", "c"}))
}

func TestContinueMatchingAndSkipBranches(t *testing.T) {
	parser := Continue{After: time.Unix(100, 0), FallbackUserAgent: "fallback/1.0"}

	heartbeats := parser.heartbeats([]continueEvent{
		{Kind: continueEventChat, Timestamp: time.Unix(90, 0), SessionID: "old", Prompt: "skip"},
		{Kind: continueEventRead, Timestamp: time.Unix(90, 0), FilePath: "old.go"},
		{Kind: continueEventEdit, Timestamp: time.Unix(90, 0), FilePath: "old.go"},
		{Kind: continueEventChat, Timestamp: time.Unix(101, 0), SessionID: "new", WorkspacePath: "/workspace"},
		{Kind: continueEventEdit, Timestamp: time.Unix(102, 0), FilePath: "edit.go"},
	})
	require.Len(t, heartbeats, 2)
	assert.Equal(t, "Continue new", heartbeats[0].Entity)
	assert.Equal(t, filepath.Join("/workspace", "edit.go"), heartbeats[1].Entity)
	require.NotNil(t, heartbeats[1].AILineChanges)
	assert.Equal(t, 0, *heartbeats[1].AILineChanges)

	events := []continueEvent{
		{Kind: continueEventChat, Timestamp: time.Unix(100, 0), Model: "gpt-5"},
		{Kind: continueEventChat, Timestamp: time.Unix(101, 0), Model: "gpt-5"},
		{Kind: continueEventTokens, Timestamp: time.Unix(100, 0), Model: "other"},
		{
			Kind:      continueEventTokens,
			Timestamp: time.Unix(101, 0),
			Model:     "gpt-5",
			Tokens:    heartbeat.AITokens{CurrentInput: 5},
		},
		{
			Kind:      continueEventTokens,
			Timestamp: time.Unix(110, 0),
			Model:     "gpt-5",
			Tokens:    heartbeat.AITokens{CurrentInput: 10},
		},
	}
	matches := parser.tokenMatches(events)
	require.Len(t, matches, 1)
	assert.EqualValues(t, 5, matches[1].CurrentInput)

	assert.Equal(t, "fallback/1.0", parser.userAgent(continueEvent{}, "Continue session"))
	assert.Equal(t, "%", parser.filePath("file://%"))
	assert.Equal(t, "relative.go", parser.resolvePath("relative.go", ""))
}

func TestSQLiteParserErrorBranches(t *testing.T) {
	tests := []struct {
		name       string
		parser     Parser
		dbPath     func(string) string
		errMessage string
	}{
		{
			name:   "Cursor",
			parser: Cursor{},
			dbPath: func(home string) string {
				return filepath.Join(home, ".config", "Cursor", "User", "globalStorage", "state.vscdb")
			},
			errMessage: "failed querying cursor sqlite db",
		},
		{
			name:   "Windsurf",
			parser: Windsurf{},
			dbPath: func(home string) string {
				return filepath.Join(home, ".config", "Windsurf", "User", "globalStorage", "state.vscdb")
			},
			errMessage: "failed querying windsurf sqlite db",
		},
		{
			name:   "Qoder",
			parser: Qoder{},
			dbPath: func(home string) string {
				return filepath.Join(home, ".config", "Qoder", "SharedClientCache", "cache", "db", "local.db")
			},
			errMessage: "failed querying qoder sqlite db",
		},
		{
			name:   "Goose",
			parser: Goose{},
			dbPath: func(home string) string {
				return filepath.Join(home, ".local", "share", "goose", "sessions", "sessions.db")
			},
			errMessage: "failed reading goose sqlite schema",
		},
		{
			name:   "OpenCode",
			parser: OpenCode{},
			dbPath: func(home string) string {
				return filepath.Join(home, ".local", "share", "opencode", "opencode.db")
			},
			errMessage: "failed querying OpenCode sqlite sessions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)

			dbPath := tt.dbPath(home)
			require.NoError(t, os.MkdirAll(filepath.Dir(dbPath), 0700))
			require.NoError(t, os.WriteFile(dbPath, []byte("not sqlite"), 0600))

			_, err := tt.parser.Parse(context.Background())
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMessage)
		})
	}
}

func TestSQLiteParserModifiedAfterBranches(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.db")
	assert.False(t, Cursor{}.stateDBModifiedAfter(missing, time.Now()))
	assert.False(t, Windsurf{}.stateDBModifiedAfter(missing, time.Now()))
	assert.False(t, Qoder{}.localDBModifiedAfter(missing, time.Now()))
	assert.False(t, Goose{}.dbModifiedAfter(missing, time.Now()))
	assert.False(t, openCodeSQLiteDBModifiedAfter(missing, time.Now()))

	dbPath := filepath.Join(t.TempDir(), "state.db")
	require.NoError(t, os.WriteFile(dbPath, []byte("db"), 0600))
	assert.True(t, Cursor{}.stateDBModifiedAfter(dbPath, time.Time{}))
	assert.True(t, Windsurf{}.stateDBModifiedAfter(dbPath, time.Time{}))
	assert.True(t, Qoder{}.localDBModifiedAfter(dbPath, time.Time{}))
	assert.True(t, Goose{}.dbModifiedAfter(dbPath, time.Time{}))
	assert.True(t, openCodeSQLiteDBModifiedAfter(dbPath, time.Time{}))

	info, err := os.Stat(dbPath)
	require.NoError(t, err)

	cutoff := info.ModTime()
	assert.True(t, Cursor{}.stateDBModifiedAfter(dbPath, cutoff))
	assert.True(t, Windsurf{}.stateDBModifiedAfter(dbPath, cutoff))
	assert.True(t, Qoder{}.localDBModifiedAfter(dbPath, cutoff))
	assert.True(t, Goose{}.dbModifiedAfter(dbPath, cutoff))
	assert.True(t, openCodeSQLiteDBModifiedAfter(dbPath, cutoff))
}

func TestTranscriptParserOpenContextAndScannerErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.jsonl")
	_, err := Codex{}.parseTranscript(context.Background(), missing)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open codex transcript")

	_, err = Pi{}.parseTranscript(context.Background(), missing)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open pi transcript")

	_, err = QwenCode{}.parseTranscript(context.Background(), missing)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open qwen code transcript")

	_, err = Amp{}.parseTranscript(context.Background(), missing, ampSessionMetadata{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open amp transcript")

	transcript := filepath.Join(t.TempDir(), "session.jsonl")
	require.NoError(t, os.WriteFile(transcript, []byte("{}\n"), 0600))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = Codex{}.parseTranscript(ctx, transcript)
	require.ErrorIs(t, err, context.Canceled)

	_, err = Pi{}.parseTranscript(ctx, transcript)
	require.ErrorIs(t, err, context.Canceled)

	_, err = QwenCode{}.parseTranscript(ctx, transcript)
	require.ErrorIs(t, err, context.Canceled)

	_, err = Amp{}.parseTranscript(ctx, transcript, ampSessionMetadata{})
	require.ErrorIs(t, err, context.Canceled)

	assertScannerOversizedTranscript(t, "codex", codexScanner)
	assertScannerOversizedTranscript(t, "pi", piScanner)
	assertScannerOversizedTranscript(t, "qwen code", qwenCodeScanner)
	assertScannerOversizedTranscript(t, "amp", ampScanner)
}

func assertScannerOversizedTranscript(
	t *testing.T,
	name string,
	scannerFn func(*os.File, string) (*bufio.Scanner, error),
) {
	t.Helper()

	transcript := filepath.Join(t.TempDir(), name+".jsonl")
	fh, err := os.Create(transcript)
	require.NoError(t, err)

	defer fh.Close() //nolint:errcheck

	require.NoError(t, fh.Truncate(maxTranscriptLineSize+2))
	require.NoError(t, fh.Sync())
	_, err = fh.Seek(0, 0)
	require.NoError(t, err)

	_, err = scannerFn(fh, transcript)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read "+name+" transcript")
}

func TestGeminiAdditionalBranches(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	tmpDir := filepath.Join(home, ".gemini", "tmp")
	slugDir := filepath.Join(tmpDir, "slug")
	require.NoError(t, os.MkdirAll(filepath.Join(slugDir, "chats"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(slugDir, ".project_root"), []byte("/workspace/project\n"), 0600))
	require.NoError(t, os.WriteFile(
		filepath.Join(home, ".gemini", "projects.json"),
		[]byte(mustJSON(t, geminiProjectsRegistry{
			Projects: map[string]string{"/registry/project": "registry-slug", "": "ignored"},
		})),
		0600,
	))

	paths, err := Gemini{}.projectPaths(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "/workspace/project", paths.bySlug["slug"])
	assert.Equal(t, "/registry/project", paths.bySlug["registry-slug"])

	require.NoError(t, os.WriteFile(filepath.Join(home, ".gemini", "projects.json"), []byte(`{`), 0600))

	_, err = Gemini{}.projectPaths(context.Background())
	require.Error(t, err)

	parser := Gemini{FallbackUserAgent: "fallback/1.0"}
	_, err = parser.parseTranscript(filepath.Join(home, "missing.json"), "")
	require.Error(t, err)

	transcriptRoot := filepath.Join(home, "gemini-fixture")
	transcriptDir := filepath.Join(transcriptRoot, "chats")
	transcriptPath := filepath.Join(transcriptDir, "session-fallback.json")
	require.NoError(t, os.MkdirAll(transcriptDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(transcriptRoot, "logs.json"), []byte(mustJSON(t, []geminiPromptLog{
		{SessionID: "session-fallback", Type: "user", Message: "from logs", Timestamp: "2026-01-02T03:04:05Z"},
		{SessionID: "session-fallback", Type: "user", Message: " ", Timestamp: "2026-01-02T03:04:06Z"},
	})), 0600))
	require.NoError(t, os.WriteFile(transcriptPath, []byte(mustJSON(t, geminiSession{
		Messages: []geminiMessage{
			{
				Type:      "gemini",
				Timestamp: "2026-01-02T03:04:07Z",
				Content:   json.RawMessage(`"assistant text"`),
				Model:     "gemini-3-pro",
				Tokens:    &geminiMessageToken{Input: 4, Output: 2},
			},
		},
	})), 0600))
	got, err := parser.parseTranscript(transcriptPath, "")
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "Gemini session-fallback", got[0].Entity)
	assert.Equal(t, len([]rune("from logs")), got[0].AIPromptLength)

	timestamp := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	_, ok := parser.toolHeartbeat(timestamp, "session-1", "/workspace/project", "gemini-3-pro", geminiToolCall{
		Status: "failed",
	}, nil)
	assert.False(t, ok)
	write, ok := parser.toolHeartbeat(timestamp, "session-1", "/workspace/project", "gemini-3-pro", geminiToolCall{
		DisplayName:   "write_file",
		Args:          json.RawMessage(`{"file_path":"relative.go","content":"one\ntwo"}`),
		ResultDisplay: json.RawMessage(`{"filePath":"/override.go"}`),
	}, nil)
	require.True(t, ok)
	assert.Equal(t, "/override.go", write.Entity)

	replace, ok := parser.toolHeartbeat(timestamp, "session-1", "/workspace/project", "gemini-3-pro", geminiToolCall{
		Name:          "replace",
		Args:          json.RawMessage(`{"file_path":"relative.go","old_string":"one","new_string":"one\ntwo"}`),
		ResultDisplay: json.RawMessage(`{"diffStat":{"model_added_lines":4,"model_removed_lines":1}}`),
	}, nil)
	require.True(t, ok)
	require.NotNil(t, replace.AILineChanges)
	assert.Equal(t, 3, *replace.AILineChanges)

	_, ok = parser.toolHeartbeat(
		timestamp,
		"session-1",
		"",
		"",
		geminiToolCall{Name: "write_file", Args: json.RawMessage(`{`)},
		nil,
	)
	assert.False(t, ok)
	_, ok = parser.toolHeartbeat(timestamp, "session-1", "", "", geminiToolCall{Name: "unknown"}, nil)
	assert.False(t, ok)

	assert.False(t, Gemini{}.hasTokenDelta(heartbeat.AITokens{
		CurrentInput:  1,
		LastInput:     2,
		CurrentOutput: 1,
		LastOutput:    2,
	}))
	assert.Nil(t, Gemini{}.tokensForFirstHeartbeat(false, heartbeat.AITokens{CurrentInput: 1}))
	assert.Equal(t, "", geminiResolvePath("/workspace", ""))
	assert.Equal(t, "relative.go", geminiResolvePath("", "relative.go"))
}

func TestOpenCodeAdditionalBranches(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	roots, err := openCodeDataRoots(context.Background())
	require.NoError(t, err)
	require.Len(t, roots, 4)
	assert.Equal(t, int64(0), OpenCode{}.afterUnixMilli())
	assert.Greater(t, OpenCode{After: time.Unix(10, 0)}.afterUnixMilli(), int64(0))
	assert.False(t, openCodeSQLiteDBModifiedAfter(filepath.Join(home, "missing.db"), time.Now()))

	parser := OpenCode{FallbackUserAgent: "fallback/1.0"}
	sessionPath := filepath.Join(home, "session.json")
	_, err = parser.readSessionInfo(sessionPath)
	require.Error(t, err)
	require.NoError(t, os.WriteFile(sessionPath, []byte(`{`), 0600))
	_, err = parser.readSessionInfo(sessionPath)
	require.Error(t, err)
	require.NoError(t, os.WriteFile(sessionPath, []byte(`{"id":"session-1"}`), 0600))
	session, err := parser.readSessionInfo(sessionPath)
	require.NoError(t, err)
	assert.Equal(t, "session-1", session.ID)

	messagePath := filepath.Join(home, "message.json")
	_, err = parser.readMessageInfo(messagePath)
	require.Error(t, err)
	require.NoError(t, os.WriteFile(messagePath, []byte(`{`), 0600))
	_, err = parser.readMessageInfo(messagePath)
	require.Error(t, err)
	require.NoError(t, os.WriteFile(messagePath, []byte(`{"id":"message-1"}`), 0600))
	message, err := parser.readMessageInfo(messagePath)
	require.NoError(t, err)
	assert.Equal(t, "message-1", message.ID)

	baseStorageDir := filepath.Join(home, "storage")
	parts, err := parser.readParts(baseStorageDir, "missing")
	require.NoError(t, err)
	assert.Nil(t, parts)

	partsDir := filepath.Join(baseStorageDir, "part", "message-1")
	require.NoError(t, os.MkdirAll(filepath.Join(partsDir, "subdir"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(partsDir, "skip.txt"), []byte(`{}`), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(partsDir, "b.json"), []byte(`{"id":"b"}`), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(partsDir, "a.json"), []byte(`{"id":"a"}`), 0600))

	parts, err = parser.readParts(baseStorageDir, "message-1")
	require.NoError(t, err)
	require.Len(t, parts, 2)
	assert.Equal(t, "a", parts[0].ID)
	require.NoError(t, os.WriteFile(filepath.Join(partsDir, "bad.json"), []byte(`{`), 0600))

	_, err = parser.readParts(baseStorageDir, "message-1")
	require.Error(t, err)

	assert.Nil(t, parser.sessionHeartbeats(openCodeSessionInfo{ID: "session-1"}, nil))
	old := parser.sessionHeartbeats(openCodeSessionInfo{ID: "session-1"}, []openCodeMessageWithParts{{
		info: openCodeMessageInfo{Role: "assistant", Time: struct {
			Created int64 `json:"created"`
		}{Created: 1000}},
	}})
	assert.Nil(t, old)
	assert.Nil(t, parser.userHeartbeat("OpenCode session-1", "", "", "session-1", "", openCodeMessageWithParts{
		info: openCodeMessageInfo{Time: struct {
			Created int64 `json:"created"`
		}{Created: 1000}},
		parts: []openCodePart{{Type: "text", Ignored: true, Text: "ignored"}},
	}, heartbeat.AITokens{}))
	assert.Nil(t, parser.assistantHeartbeat("OpenCode session-1", "", "", "session-1", "", openCodeMessageWithParts{
		info: openCodeMessageInfo{Time: struct {
			Created int64 `json:"created"`
		}{Created: 1000}},
	}, heartbeat.AITokens{}))
	assert.Nil(t, parser.toolHeartbeats("", "", "", "", time.Now(), openCodePart{}))
	assert.Nil(t, parser.toolHeartbeats("", "", "", "", time.Now(), openCodePart{Type: "tool"}))
	assert.Nil(t, parser.toolHeartbeats("", "", "", "", time.Now(), openCodePart{
		Type:  "tool",
		Tool:  "unknown",
		State: &openCodeToolState{Status: "completed"},
	}))
	assert.Equal(t, "/workspace/rel.go", openCodeResolvePath("/workspace", "rel.go"))
	assert.Equal(t, "rel.go", openCodeResolvePath("", "rel.go"))
	assert.Empty(t, openCodePatchFilePath("/workspace", "*** Unknown File: rel.go"))
}

func TestOpenCodeLegacyAndSQLiteBranches(t *testing.T) {
	parser := OpenCode{FallbackUserAgent: "fallback/1.0", After: time.UnixMilli(2000)}

	dataRoot := t.TempDir()
	sessionsDir := filepath.Join(dataRoot, "storage", "session")
	require.NoError(t, os.MkdirAll(filepath.Join(sessionsDir, "ignored-dir"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(sessionsDir, "ignored.txt"), []byte(`{}`), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(sessionsDir, "bad.json"), []byte(`{`), 0600))

	_, err := parser.parseLegacyRoot(dataRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to walk OpenCode sessions directory")

	require.NoError(t, os.Remove(filepath.Join(sessionsDir, "bad.json")))
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionsDir, "old.json"),
		[]byte(`{"id":"old","time":{"updated":1000}}`),
		0600,
	))
	require.NoError(t, os.WriteFile(filepath.Join(sessionsDir, "zero.json"), []byte(`{"id":"zero"}`), 0600))

	legacy, err := parser.parseLegacyRoot(dataRoot)
	require.NoError(t, err)
	assert.Nil(t, legacy)

	_, err = parser.parseLegacySession(filepath.Join(dataRoot, "missing.json"))
	require.Error(t, err)

	messageDataRoot := t.TempDir()
	sessionPath := filepath.Join(messageDataRoot, "storage", "session", "session.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(sessionPath), 0700))
	require.NoError(t, os.WriteFile(sessionPath, []byte(`{"id":"session-1"}`), 0600))
	require.NoError(t, os.MkdirAll(filepath.Join(messageDataRoot, "message"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(messageDataRoot, "message", "session-1"), []byte("file"), 0600))

	_, err = parser.parseLegacySession(sessionPath)
	if err != nil {
		assert.Contains(t, err.Error(), "failed to read OpenCode messages directory")
	}

	require.NoError(t, os.Remove(filepath.Join(messageDataRoot, "message", "session-1")))
	require.NoError(t, os.MkdirAll(filepath.Join(messageDataRoot, "message", "session-1", "subdir"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(messageDataRoot, "message", "session-1", "bad.json"), []byte(`{`), 0600))

	_, err = parser.parseLegacySession(sessionPath)
	require.Error(t, err)

	require.NoError(t, os.WriteFile(
		filepath.Join(messageDataRoot, "message", "session-1", "bad.json"),
		[]byte(`{"id":"message-1","role":"user","time":{"created":3000}}`),
		0600,
	))
	require.NoError(t, os.MkdirAll(filepath.Join(messageDataRoot, "part", "message-1"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(messageDataRoot, "part", "message-1", "bad.json"), []byte(`{`), 0600))

	_, err = parser.parseLegacySession(sessionPath)
	require.Error(t, err)

	dbPath := filepath.Join(t.TempDir(), "opencode.db")

	db := openOpenCodeTestDB(t, dbPath)
	defer db.Close()

	_, err = queryOpenCodeSQLiteSessions(context.Background(), db, dbPath)
	require.Error(t, err)

	require.NoError(t, execOpenCodeSQL(db, `
CREATE TABLE session (id TEXT, directory TEXT, version TEXT);
CREATE TABLE message (id TEXT, session_id TEXT, data TEXT, time_created INTEGER);
CREATE TABLE part (id TEXT, message_id TEXT, session_id TEXT, data TEXT, time_created INTEGER);
INSERT INTO message VALUES ('bad-json', 's1', '123', 3000);
INSERT INTO message VALUES ('old-by-payload', 's1', '{"id":"old","time":{"created":1000}}', 3000);
INSERT INTO message VALUES ('m1', 's1', '{"role":"user"}', 3000);
INSERT INTO message VALUES ('other-seed', 'other', '{"id":"other","sessionID":"other"}', 1900);
INSERT INTO message VALUES ('bad-seed', 's1', '123', 1800);
INSERT INTO message VALUES ('zero-seed', 's1', '{"id":"zero-seed"}', 0);
INSERT INTO message VALUES ('seed', 's1', '{"id":"seed","sessionID":"s1"}', 1600);
INSERT INTO part VALUES ('skip', 'unknown', 's1', '{"type":"text","text":"skip"}', 3000);
INSERT INTO part VALUES ('bad-part', 'm1', 's1', '123', 3001);
INSERT INTO part VALUES ('p1', 'm1', 's1', '{"type":"text","text":"hello"}', 3002);
`))

	messagesBySession, messageIDs, err := parser.querySQLiteMessages(context.Background(), db, dbPath)
	require.NoError(t, err)
	assert.Contains(t, messageIDs, "m1")
	require.NoError(t, queryOpenCodeSQLiteParts(context.Background(), db, dbPath, messagesBySession, messageIDs))

	heartbeats, err := parser.parseSQLiteDB(context.Background(), dbPath)
	require.NoError(t, err)
	require.NotEmpty(t, heartbeats)
	assert.Equal(t, "OpenCode s1", heartbeats[0].Entity)

	noMessagesPath := filepath.Join(t.TempDir(), "no-messages.db")

	noMessagesDB := openOpenCodeTestDB(t, noMessagesPath)
	defer noMessagesDB.Close()

	require.NoError(t, execOpenCodeSQL(noMessagesDB, `
CREATE TABLE session (id TEXT, directory TEXT, version TEXT);
CREATE TABLE message (id TEXT, session_id TEXT, data TEXT, time_created INTEGER);
`))

	heartbeats, err = parser.parseSQLiteDB(context.Background(), noMessagesPath)
	require.NoError(t, err)
	assert.Nil(t, heartbeats)

	missingMessageDB := openOpenCodeTestDB(t, filepath.Join(t.TempDir(), "missing-message.db"))
	defer missingMessageDB.Close()

	require.NoError(t, execOpenCodeSQL(missingMessageDB, `CREATE TABLE session (id TEXT, directory TEXT, version TEXT);`))
	_, _, err = parser.querySQLiteMessages(context.Background(), missingMessageDB, "missing-message.db")
	require.Error(t, err)

	missingPartPath := filepath.Join(t.TempDir(), "missing-part.db")

	missingPartDB := openOpenCodeTestDB(t, missingPartPath)
	defer missingPartDB.Close()

	require.NoError(t, execOpenCodeSQL(missingPartDB, `
CREATE TABLE session (id TEXT, directory TEXT, version TEXT);
CREATE TABLE message (id TEXT, session_id TEXT, data TEXT, time_created INTEGER);
INSERT INTO message VALUES ('m1', 's1', '{"role":"user"}', 3000);
`))

	_, err = parser.parseSQLiteDB(context.Background(), missingPartPath)
	require.Error(t, err)

	require.Error(t, parser.querySQLiteSeedMessages(
		context.Background(),
		missingMessageDB,
		"missing-message.db",
		map[string][]openCodeMessageWithParts{"s1": nil},
	))
}

func TestQwenCodeAdditionalBranches(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("QWEN_RUNTIME_DIR", "~/runtime")

	runtimeDir, err := qwenCodeRuntimeDir(context.Background())
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "runtime"), runtimeDir)

	t.Setenv("QWEN_RUNTIME_DIR", "")

	qwenHome := filepath.Join(home, "configured-qwen")
	t.Setenv("QWEN_HOME", qwenHome)
	require.NoError(t, os.MkdirAll(qwenHome, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(qwenHome, "settings.json"), []byte(`{
  "advanced": {"runtimeOutputDir": "~/qwen-runtime"}
}`), 0600))

	runtimeDir, err = qwenCodeRuntimeDir(context.Background())
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "qwen-runtime"), runtimeDir)

	parser := QwenCode{FallbackUserAgent: "fallback/1.0"}
	state := &qwenCodeParseState{
		cwd:       filepath.Join(home, "workspace"),
		sessionID: "session-1",
		toolCalls: map[string]qwenCodeToolCall{
			"call-1": {ID: "call-1", Name: "read_file", Args: map[string]interface{}{"path": "read.go"}},
		},
		toolQueue: []qwenCodeToolCall{
			{ID: "call-1", Name: "read_file", Args: map[string]interface{}{"path": "read.go"}},
			{Name: "write_file", Args: map[string]interface{}{"file_path": "write.go", "content": "one\ntwo"}},
		},
	}
	record := qwenCodeRecord{
		Timestamp: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		Type:      "tool_result",
		ToolCallResult: &qwenCodeToolCallResult{
			CallID: "call-1",
			Status: "success",
		},
	}
	h := parser.toolResultHeartbeat(record, state)
	require.NotNil(t, h)
	assert.Equal(t, filepath.Join(home, "workspace", "read.go"), h.Entity)
	assert.Len(t, state.toolQueue, 1)
	assert.NotContains(t, state.toolCalls, "call-1")

	record.ToolCallResult = &qwenCodeToolCallResult{Status: "failed"}
	assert.Nil(t, parser.toolResultHeartbeat(record, state))
	record.ToolCallResult = &qwenCodeToolCallResult{}
	record.Message = qwenCodeMessage{Parts: []qwenCodePart{{
		FunctionResponse: &qwenCodeFunctionResponse{Name: "write_file"},
	}}}
	h = parser.toolResultHeartbeat(record, state)
	require.NotNil(t, h)
	assert.Equal(t, filepath.Join(home, "workspace", "write.go"), h.Entity)

	assert.Nil(t, parser.messageHeartbeat(qwenCodeRecord{Type: "user", Subtype: "cron"}, qwenCodeParseState{}))
	assert.Nil(t, parser.messageHeartbeat(qwenCodeRecord{
		Type:    "assistant",
		Message: qwenCodeMessage{Parts: []qwenCodePart{{Thought: true, Text: "hidden"}}},
	}, qwenCodeParseState{}))
	assert.NotNil(t, parser.messageHeartbeat(qwenCodeRecord{
		Type:      "assistant",
		Timestamp: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		Message:   qwenCodeMessage{Parts: []qwenCodePart{{Text: "visible"}}},
	}, qwenCodeParseState{sessionID: "session-1"}))

	var raw qwenCodeFunctionCall
	require.NoError(t, json.Unmarshal([]byte(`{"name":"read_file","args":"{\"path\":\"read.go\"}"}`), &raw))
	assert.Equal(t, "read.go", raw.Args["path"])
	assert.Empty(t, qwenCodeFunctionResponseName(qwenCodeMessage{}))
	assert.Empty(t, qwenCodeFunctionResponseID(qwenCodeMessage{}))

	_, ok := qwenCodeRemoveQueuedToolCall(&qwenCodeParseState{}, func(qwenCodeToolCall) bool { return true })
	assert.False(t, ok)
	changes, ok := qwenCodeDiffLineChanges(json.RawMessage(`{"diffStat":{"model_added_lines":1,"model_removed_lines":3}}`))
	require.True(t, ok)
	assert.Equal(t, -2, changes)
	assert.Equal(t, "relative.go", qwenCodeAbsPath("relative.go", ""))
}

func TestCopilotAdditionalBranches(t *testing.T) {
	var variable copilotVariable
	require.NoError(t, json.Unmarshal([]byte(`{"kind":"file","value":null}`), &variable))
	assert.Nil(t, variable.Value)
	require.Error(t, json.Unmarshal([]byte(`{`), &variable))

	var message copilotMessageWithURIs
	require.NoError(t, json.Unmarshal([]byte(`null`), &message))
	require.NoError(t, json.Unmarshal([]byte(`"plain"`), &message))
	assert.Equal(t, "plain", message.Value)
	require.Error(t, json.Unmarshal([]byte(`123`), &message))

	base := t.TempDir()
	parser := Copilot{FallbackUserAgent: "fallback/1.0"}
	_, err := parser.parseJSONSession(filepath.Join(base, "missing.json"))
	require.Error(t, err)

	badSessionPath := filepath.Join(base, "bad.json")
	require.NoError(t, os.WriteFile(badSessionPath, []byte(`{`), 0600))
	_, err = parser.parseJSONSession(badSessionPath)
	require.Error(t, err)

	workspace := filepath.Join(base, "workspace")
	sessionDir := filepath.Join(workspace, "chatSessions")
	require.NoError(t, os.MkdirAll(sessionDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "skip.txt"), []byte(`{}`), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "bad.json"), []byte(`{`), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "fallback-id.json"), []byte(`{"requests":[]}`), 0600))

	sessions, err := parser.loadWorkspaceSessions(wakalog.New(os.Stdout), workspace)
	require.NoError(t, err)
	require.Contains(t, sessions, "fallback-id")

	editFile := filepath.Join(workspace, "chatEditingSessions")
	require.NoError(t, os.WriteFile(editFile, []byte("not a dir"), 0600))

	_, err = parser.editHeartbeats(workspace, nil)
	require.Error(t, err)

	assert.Nil(t, parser.tokensForFirstHeartbeat(false, heartbeat.AITokens{CurrentInput: 1}))
	input, output := parser.tokenDelta(heartbeat.AITokens{CurrentInput: 1, LastInput: 2, CurrentOutput: 1, LastOutput: 2})
	assert.Zero(t, input)
	assert.Zero(t, output)
	assert.False(t, parser.hasTokenDelta(heartbeat.AITokens{}))
	in, out, cumulative := parser.usageCounts(copilotRequest{
		Result: json.RawMessage(`{"metadata":{"usage":{"inputTokens":4,"outputTokens":2}}}`),
	})
	require.NotNil(t, in)
	require.NotNil(t, out)
	assert.Equal(t, 4, *in)
	assert.Equal(t, 2, *out)
	assert.False(t, cumulative)

	in, out = parser.parseTokenCounts(json.RawMessage(`{"inputTokens":1,"totalTokens":9}`))
	require.NotNil(t, in)
	require.NotNil(t, out)
	assert.Equal(t, 1, *in)
	assert.Equal(t, 9, *out)
	in, out = parser.parseResultMetadataTokenCounts(json.RawMessage(`{"promptTokens":3,"completionTokens":1}`))
	require.NotNil(t, in)
	require.NotNil(t, out)
	assert.Equal(t, 3, *in)
	assert.Equal(t, 1, *out)

	dirPath := filepath.Join(base, "dir")
	require.NoError(t, os.MkdirAll(dirPath, 0700))
	assert.False(t, parser.shouldTrackReadPath(""))
	assert.False(t, parser.shouldTrackReadPath(dirPath))
	assert.True(t, parser.shouldTrackReadPath(filepath.Join(base, "missing.go")))
	assert.Empty(t, parser.variablePath(copilotVariable{ID: "file://%zz"}))
	assert.Equal(t, []string{filepath.Join(base, "past.go")}, parser.responsePaths(copilotResponseItem{
		PastTenseMessage: &copilotMessageWithURIs{URIs: map[string]copilotReferenceURI{
			"file://ignored": {FSPath: filepath.Join(base, "past.go")},
		}},
	}))

	timestamp := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	app := parser.appHeartbeat(
		"Copilot session-1",
		"session-1",
		nil,
		5,
		timestamp,
		"/workspace",
		nil,
		"fallback/1.0",
		"plugin/1",
	)
	assert.Equal(t, "session-1", app.AISession)

	file := parser.fileHeartbeat(
		"/workspace/main.go",
		"session-1",
		nil,
		timestamp,
		nil,
		"fallback/1.0",
		"plugin/1",
		true,
		nil,
	)
	assert.Equal(t, "session-1", file.AISession)

	cliApp := parser.cliAppHeartbeat(
		"Copilot session-1",
		"session-1",
		nil,
		5,
		timestamp,
		"/workspace",
		"gpt-5",
		"1.0",
		"2.0",
	)
	assert.Equal(t, "session-1", cliApp.AISession)

	cliFile := parser.cliFileHeartbeat("/workspace/main.go", "session-1", nil, timestamp, "gpt-5", "1.0", "2.0", true, nil)
	assert.Equal(t, "session-1", cliFile.AISession)
	assert.Equal(t, "github-copilot-cli/1.0 copilot/unknown", copilotCLIHarnessUserAgent("1.0", ""))
}

func TestPiAmpAndGooseAdditionalBranches(t *testing.T) {
	assert.Equal(t, piSessionState{cwd: "cwd", id: "id"}, func() piSessionState {
		cwd, id := Pi{}.sessionInfo("cwd", "id", nil)
		return piSessionState{cwd: cwd, id: id}
	}())
	assert.Equal(t, "model", Pi{}.sessionVersion("", piLogLine{Type: "model_change", ModelID: "model"}))
	assert.Equal(t, "provider", Pi{}.sessionVersion("", piLogLine{Type: "model_change", Provider: "provider"}))
	assert.Equal(t, "message-provider", Pi{}.sessionVersion("", piLogLine{
		Message: &piMessage{Provider: "message-provider"},
	}))
	assert.Nil(t, Pi{}.getHeartbeats(time.Now(), "Pi session", "session", "", "", nil, "", piLogLine{}, &piParseState{}))
	assert.Nil(t, Pi{}.toolResultHeartbeat(
		time.Now(),
		"session",
		"",
		"",
		nil,
		"",
		piMessage{},
		piToolCall{},
		heartbeat.AITokens{},
	))
	assert.NotNil(t, Pi{}.messageHeartbeat(
		time.Now(),
		"Pi session",
		"session",
		"gpt-5",
		"/workspace",
		nil,
		"fallback/1.0",
		piMessage{
			Role:    "assistant",
			Content: []piContentBlock{{Type: "text", Text: "assistant\ntext"}},
		},
		heartbeat.AITokens{},
	))
	assert.True(t, parseGooseTime("bad").IsZero())
	assert.Equal(t, int64(0), parseGooseInt("bad"))
	assert.Equal(t, "hello", goosePromptFromContent(`[{"type":"text","text":"hello"}]`))
	assert.Empty(t, goosePromptFromContent(`{`))
	assert.Empty(t, ampWorkspacePath(" "))
	assert.Equal(t, filepath.FromSlash("/tmp/main.go"), ampWorkspacePath("file:///tmp/main.go"))
	assert.Equal(t, filepath.FromSlash("C:/tmp/main.go"), ampWorkspacePath("file:///C:/tmp/main.go"))
	assert.Equal(t, ampProcessMetadata{}, ampSessionMetadata{}.process(123, time.Now()))
	assert.Equal(t, "amp/unknown-high", ampAgentVersion("", "high"))
}

func mustJSON(t *testing.T, value interface{}) string {
	t.Helper()

	encoded, err := json.Marshal(value)
	require.NoError(t, err)

	return string(encoded)
}

func openOpenCodeTestDB(t *testing.T, dbPath string) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)

	return db
}

func execOpenCodeSQL(db *sql.DB, statement string) error {
	_, err := db.Exec(statement)

	return err
}
