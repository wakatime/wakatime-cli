package ai

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestAITokenHelperBranches(t *testing.T) {
	tokens := heartbeat.AITokens{
		LastInput:     10,
		LastOutput:    20,
		CurrentInput:  5,
		CurrentOutput: 10,
	}

	claudeInput, claudeOutput := Claude{}.tokenDelta(tokens)
	assert.Zero(t, claudeInput)
	assert.Zero(t, claudeOutput)
	assert.False(t, Claude{}.hasTokenDelta(tokens))
	assert.Nil(t, Claude{}.tokensForFirstHeartbeat(false, tokens))
	require.NotNil(t, Claude{}.tokensForFirstHeartbeat(true, tokens))

	cursorInput, cursorOutput := Cursor{}.tokenDelta(tokens)
	assert.Zero(t, cursorInput)
	assert.Zero(t, cursorOutput)
	assert.False(t, Cursor{}.hasTokenDelta(tokens))
	assert.Nil(t, Cursor{}.tokensForFirstHeartbeat(false, tokens))
	require.NotNil(t, Cursor{}.tokensForFirstHeartbeat(true, tokens))

	windsurfInput, windsurfOutput := Windsurf{}.tokenDelta(tokens)
	assert.Zero(t, windsurfInput)
	assert.Zero(t, windsurfOutput)
	assert.False(t, Windsurf{}.hasTokenDelta(tokens))
	assert.Nil(t, Windsurf{}.tokensForFirstHeartbeat(false, tokens))
	require.NotNil(t, Windsurf{}.tokensForFirstHeartbeat(true, tokens))
}

func TestAIHeartbeatHelperBranches(t *testing.T) {
	cursor := Cursor{}
	windsurf := Windsurf{}
	tokens := heartbeat.AITokens{CurrentInput: 7, CurrentOutput: 9}

	withTokens := cursor.newHeartbeat(nil, "cursor-session", &tokens, "Cursor", heartbeat.AppType, nil, "", 1, "ua")
	assert.Equal(t, "cursor-session", withTokens.AISession)
	assert.Equal(t, int64(7), withTokens.AIInputTokens)

	withoutTokens := cursor.newHeartbeat(nil, "cursor-session", nil, "Cursor", heartbeat.AppType, nil, "", 1, "ua")
	assert.Equal(t, "cursor-session", withoutTokens.AISession)
	assert.Zero(t, withoutTokens.AIInputTokens)

	withTokens = windsurf.newHeartbeat(nil, "windsurf-session", &tokens, "Windsurf", heartbeat.AppType, nil, "", 1, "ua")
	assert.Equal(t, "windsurf-session", withTokens.AISession)
	assert.Equal(t, int64(9), withTokens.AIOutputTokens)

	withoutTokens = windsurf.newHeartbeat(nil, "windsurf-session", nil, "Windsurf", heartbeat.AppType, nil, "", 1, "ua")
	assert.Equal(t, "windsurf-session", withoutTokens.AISession)
	assert.Zero(t, withoutTokens.AIOutputTokens)
}

func TestCursorAndWindsurfProjectPathFallbacks(t *testing.T) {
	cursor := Cursor{}
	windsurf := Windsurf{}

	assert.Empty(t, cursor.projectPath(cursorLogLine{}))
	assert.Empty(t, windsurf.projectPath(windsurfLogLine{}))

	cursorReadRaw, err := json.Marshal(cursorReadRawArgs{TargetFile: "/workspace/raw/main.go"})
	require.NoError(t, err)
	assert.Equal(t, filepath.FromSlash("/workspace/raw"), cursor.projectPath(cursorLogLine{
		ToolFormerData: &cursorToolFormerData{Name: "read_file", RawArgs: string(cursorReadRaw)},
	}))

	windsurfReadRaw, err := json.Marshal(windsurfReadRawArgs{TargetFile: "/workspace/raw/main.go"})
	require.NoError(t, err)
	assert.Equal(t, filepath.FromSlash("/workspace/raw"), windsurf.projectPath(windsurfLogLine{
		ToolFormerData: &windsurfToolFormerData{Name: "read_file", RawArgs: string(windsurfReadRaw)},
	}))

	assert.Equal(t, filepath.FromSlash("/workspace/block"), cursor.projectPath(cursorLogLine{
		ToolFormerData: &cursorToolFormerData{Name: "read_file"},
		CodeBlocks:     []cursorCodeBlock{{URI: &cursorURI{FSPath: "/workspace/block/main.go"}}},
	}))
	assert.Equal(t, filepath.FromSlash("/workspace/block"), windsurf.projectPath(windsurfLogLine{
		ToolFormerData: &windsurfToolFormerData{Name: "read_file"},
		CodeBlocks:     []windsurfCodeBlock{{URI: &windsurfURI{FSPath: "/workspace/block/main.go"}}},
	}))
}

func TestKiroHelperBranches(t *testing.T) {
	rawBlocks, err := json.Marshal([]kiroContentBlock{
		{Type: "text", Text: " first "},
		{Type: "image", Text: "ignored"},
		{Type: "text", Text: "second"},
	})
	require.NoError(t, err)
	assert.Equal(t, "first \nsecond", kiroMessageText(rawBlocks))

	rawString, err := json.Marshal(" plain text ")
	require.NoError(t, err)
	assert.Equal(t, "plain text", kiroMessageText(rawString))
	assert.Empty(t, kiroMessageText(json.RawMessage(`{"unsupported": true}`)))

	session := kiroExecutionSession(kiroExecution{
		ChatSessionID: "session-1",
	}, map[string]kiroSessionInfo{
		"session-1": {SessionID: "session-1", Workspace: "/workspace/from-session"},
	})
	assert.Equal(t, "/workspace/from-session", session.Workspace)

	assert.Empty(t, kiroFilePath("", "/workspace"))
	assert.Equal(t, "/workspace/main.go", kiroFilePath("main.go", "/workspace"))
	assert.Equal(t, filepath.FromSlash("relative/main.go"), kiroFilePath("relative/main.go", ""))
	assert.Equal(t, filepath.FromSlash("C:/workspace/main.go"), kiroFilePath("/C:/workspace/main.go", ""))
	assert.Equal(t, "/tmp/main.go", kiroFilePath("file:///tmp/main.go", ""))
	assert.Equal(t, 3, kiroLineChanges("", "a\nb\n"))
	assert.Equal(t, -3, kiroLineChanges("a\nb\n", ""))

	after := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	got := Kiro{After: after.Add(-time.Minute), FallbackUserAgent: "ua"}.actionHeartbeats(kiroExecution{
		StartTimeMS: after.UnixMilli(),
		Actions: []kiroAction{
			{
				ActionType:  "readFiles",
				ActionState: "Rejected",
				Input:       kiroActionInput{Files: []kiroActionFile{{Path: "skip.go"}}},
			},
			{
				ActionType:  "readFiles",
				ActionState: "Accepted",
				Input:       kiroActionInput{Files: []kiroActionFile{{Path: "main.go"}}},
			},
			{
				ActionType:  "replace",
				ActionState: "Accepted",
				Input:       kiroActionInput{Path: "main.go", ModifiedContent: "a\nb\n"},
			},
			{ActionType: "unknown", ActionState: "Accepted"},
		},
	}, kiroSessionInfo{SessionID: "session-1", Workspace: "/workspace"})

	require.Len(t, got, 2)
	assert.False(t, *got[0].IsWrite)
	assert.True(t, *got[1].IsWrite)
}

func TestClaudeLineWriteAndPromptBranches(t *testing.T) {
	claude := Claude{}

	newText := "one\ntwo\n"
	oldText := "one\n"
	assert.Equal(t, 1, claude.lineChanges(toolUseResult{NewString: &newText, OldString: &oldText}))
	assert.Equal(t, 3, claude.lineChanges(toolUseResult{NewString: &newText}))

	original := "existing"
	assert.Zero(t, claude.lineChanges(toolUseResult{OriginalFile: &original}))
	assert.False(t, claude.isWrite(toolUseResult{OriginalFile: &original}))

	update := "update"
	assert.True(t, claude.isWrite(toolUseResult{Type: &update}))
	assert.True(t, claude.isWrite(toolUseResult{NewString: &newText}))

	content := &contentValue{String: &newText}
	assert.Equal(t, 3, claude.lineChanges(toolUseResult{Content: content}))
	assert.True(t, claude.isWrite(toolUseResult{Content: content}))

	filePath := "/workspace/main.go"
	assert.Zero(t, claude.appLineChanges(&toolUseResultValue{Object: &toolUseResult{FilePath: &filePath, Content: content}}))
	assert.Zero(t, claude.appLineChanges(&toolUseResultValue{Object: &toolUseResult{Raw: map[string]json.RawMessage{
		"taskId": json.RawMessage(`"task-1"`),
	}}}))
	assert.Equal(t, 3, claude.appLineChanges(&toolUseResultValue{String: &newText}))
	assert.Zero(t, claude.appLineChanges(&toolUseResultValue{}))

	sideChain := true
	userType := "user"

	assert.Zero(t, claudePromptLength(claudeLogLine{IsSideChain: &sideChain}))
	assert.Zero(t, claudePromptLength(claudeLogLine{Type: &userType, Message: &claudeMessage{Role: "assistant"}}))
	assert.Equal(t, len([]rune("hello")), claudePromptLength(claudeLogLine{
		Type: &userType,
		Message: &claudeMessage{
			Role:    "user",
			Content: claudeMessageContentList{{Type: "text", Text: "<system>ignored</system> hello"}},
		},
	}))

	assert.Zero(t, claudePromptTextLength("<bad"))
	assert.Zero(t, claudePromptTextLength("</bad>"))
	assert.Zero(t, claudePromptTextLength("<tag>missing close"))
	assert.True(t, claudeHasIDEContext(claudeLogLine{Message: &claudeMessage{
		Content: claudeMessageContentList{{Type: "text", Text: "<ide_file>main.go</ide_file>"}},
	}}))
}

func TestCursorAndWindsurfLineChangeBranches(t *testing.T) {
	cursor := Cursor{}
	windsurf := Windsurf{}

	assert.Zero(t, cursor.lineChanges("   "))
	assert.Equal(t, 1, cursor.lineChanges("line one\n\n"))
	assert.Equal(t, 1, cursor.lineChanges("--- a/main.go\n+++ b/main.go\n-old\n+new\n+extra"))
	assert.True(t, cursor.looksLikeUnifiedDiff("\n@@ -1 +1 @@"))

	assert.Zero(t, windsurf.lineChanges("   "))
	assert.Equal(t, 1, windsurf.lineChanges("line one\n\n"))
	assert.Equal(t, 1, windsurf.lineChanges("--- a/main.go\n+++ b/main.go\n-old\n+new\n+extra"))
	assert.True(t, windsurf.looksLikeUnifiedDiff("\n+++ b/main.go"))
}
