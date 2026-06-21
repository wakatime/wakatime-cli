package ai_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestAmpParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))

	projectDir := filepath.Join(home, "project")
	require.NoError(t, os.MkdirAll(projectDir, 0o755))

	transcriptDir := filepath.Join(home, ".cache", "amp", "logs", "threads")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	const (
		threadID        = "T-019eea92-63e4-70dd-83d7-bfac30818089"
		failedCallID    = "TU-failed"
		successCallID   = "TU-033cGsbeTNZliT2kAC05xW"
		shellToolCallID = "TU-033cGsVzq3lpr0uZliulfb"
	)

	transcriptPath := filepath.Join(transcriptDir, threadID+".log")
	transcript := mustJSONLine(t, map[string]interface{}{
		"@timestamp":      "2026-06-21T14:28:00.500Z",
		"message":         "[observer] onAgentState",
		"type":            "agent_state",
		"agentMode":       "deep",
		"reasoningEffort": "medium",
		"threadId":        threadID,
	}) + "\n" + mustJSONLine(t, map[string]interface{}{
		"@timestamp": "2026-06-21T14:28:03.137Z",
		"level":      "INFO",
		"message":    "onToolLease",
		"logger":     "executor",
		"data": map[string]interface{}{
			"type":       "tool_lease",
			"toolCallId": shellToolCallID,
			"toolName":   "shell_command",
			"args": map[string]interface{}{
				"command":    "sed -n '1,120p' README.md",
				"workdir":    projectDir,
				"timeout_ms": 10000,
			},
		},
		"threadId": threadID,
	}) + "\n" + mustJSONLine(t, map[string]interface{}{
		"@timestamp":  "2026-06-21T14:28:03.518Z",
		"message":     "websocket message",
		"type":        "executor_tool_result",
		"toolCallId":  shellToolCallID,
		"runStatus":   "done",
		"hasRunError": false,
		"threadId":    threadID,
	}) + "\n" + mustJSONLine(t, map[string]interface{}{
		"@timestamp": "2026-06-21T14:28:05.000Z",
		"message":    "onToolLease",
		"data": map[string]interface{}{
			"type":       "tool_lease",
			"toolCallId": failedCallID,
			"toolName":   "apply_patch",
			"args": map[string]interface{}{
				"patchText": "*** Begin Patch\n*** Update File: failed.go\n-old\n+new\n*** End Patch",
			},
		},
		"threadId": threadID,
	}) + "\n" + mustJSONLine(t, map[string]interface{}{
		"@timestamp":  "2026-06-21T14:28:05.200Z",
		"message":     "websocket message",
		"type":        "executor_tool_result",
		"toolCallId":  failedCallID,
		"runStatus":   "done",
		"hasRunError": true,
		"threadId":    threadID,
	}) + "\n" + mustJSONLine(t, map[string]interface{}{
		"@timestamp": "2026-06-21T14:28:07.006Z",
		"level":      "INFO",
		"message":    "onToolLease",
		"logger":     "executor",
		"data": map[string]interface{}{
			"type":       "tool_lease",
			"toolCallId": successCallID,
			"toolName":   "apply_patch",
			"args": map[string]interface{}{
				"patchText": "*** Begin Patch\n*** Update File: README.md\n@@\n" +
					"-Made with :heart: by the WakaTime Team.\n" +
					"+Made with :heart: by the WakaTime Team and Amp.\n" +
					"*** End of File\n+Amp docs footer.\n" +
					"*** Update File: pkg/ai/amp_test.go\n@@\n" +
					"-old test\n+new test\n+second new test\n*** End Patch",
			},
		},
		"threadId": threadID,
	}) + "\n" + mustJSONLine(t, map[string]interface{}{
		"@timestamp":  "2026-06-21T14:28:07.330Z",
		"level":       "INFO",
		"message":     "websocket message",
		"logger":      "executor",
		"type":        "executor_tool_result",
		"toolCallId":  successCallID,
		"runStatus":   "done",
		"hasRunError": false,
		"threadId":    threadID,
	}) + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	modifiedAt := time.Date(2026, 6, 21, 14, 29, 0, 0, time.UTC)
	require.NoError(t, os.Chtimes(transcriptPath, modifiedAt, modifiedAt))

	readmePath := filepath.Join(projectDir, "README.md")
	ampTestPath := filepath.Join(projectDir, "pkg", "ai", "amp_test.go")
	parser := ai.Amp{
		After:             time.Date(2026, 6, 21, 14, 28, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			readmePath: heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, readmePath, got[0].Entity)
	assert.Equal(t, heartbeat.FileType, got[0].EntityType)
	assert.Equal(t, threadID, got[0].AISession)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[0].Category)
	require.NotNil(t, got[0].AILineChanges)
	assert.Equal(t, 1, *got[0].AILineChanges)
	require.NotNil(t, got[0].IsWrite)
	assert.True(t, *got[0].IsWrite)
	assert.Equal(t, float64(time.Date(2026, 6, 21, 14, 28, 7, 330000000, time.UTC).UnixMilli())/1000, got[0].Time)
	assert.Equal(t, "amp/unknown-medium "+heartbeat.UserAgent(ctx, "editor/1.2.3"), got[0].UserAgent)

	assert.Equal(t, ampTestPath, got[1].Entity)
	assert.Equal(t, threadID, got[1].AISession)
	require.NotNil(t, got[1].AILineChanges)
	assert.Equal(t, 1, *got[1].AILineChanges)
	require.NotNil(t, got[1].IsWrite)
	assert.True(t, *got[1].IsWrite)
	assert.Equal(t, "amp/unknown-medium plugin/0.0.1", got[1].UserAgent)
}

func TestAmpParse_UsesCLIProcessMetadata(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))

	logsDir := filepath.Join(home, ".cache", "amp", "logs")
	transcriptDir := filepath.Join(logsDir, "threads")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	projectDir := filepath.Join(home, "current-project")
	staleProjectDir := filepath.Join(home, "stale-project")
	futureProjectDir := filepath.Join(home, "future-project")
	otherProjectDir := filepath.Join(home, "other-project")

	const (
		pid        = 101
		otherPID   = 202
		threadID   = "T-current"
		toolCallID = "TU-first-action-patch"
	)

	cliLog := strings.Join([]string{
		mustJSONLine(t, map[string]interface{}{
			"@timestamp": "2026-06-21T10:00:00Z", "message": "Starting Amp CLI.",
			"version": "0.1.0", "pid": pid,
		}),
		mustJSONLine(t, map[string]interface{}{
			"@timestamp": "2026-06-21T10:00:01Z", "message": "Using settings file",
			"workspaceRootPath": staleProjectDir, "pid": pid,
		}),
		mustJSONLine(t, map[string]interface{}{
			"@timestamp": "2026-06-21T12:00:00Z", "message": "Starting Amp CLI.",
			"version": "2.3.4", "pid": pid,
		}),
		mustJSONLine(t, map[string]interface{}{
			"@timestamp": "2026-06-21T12:00:01Z", "message": "Using settings file",
			"workspaceRoot": "file://" + filepath.ToSlash(projectDir), "pid": pid,
		}),
		mustJSONLine(t, map[string]interface{}{
			"@timestamp": "2026-06-21T12:00:02Z", "message": "Using settings file",
			"workspaceRootPath": otherProjectDir, "pid": otherPID,
		}),
		mustJSONLine(t, map[string]interface{}{
			"@timestamp": "2026-06-21T14:00:00Z", "message": "Starting Amp CLI.",
			"version": "9.9.9", "pid": pid,
		}),
		mustJSONLine(t, map[string]interface{}{
			"@timestamp": "2026-06-21T14:00:01Z", "message": "Using settings file",
			"workspaceRootPath": futureProjectDir, "pid": pid,
		}),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(logsDir, "cli.log"), []byte(cliLog), 0o644))

	transcript := strings.Join([]string{
		mustJSONLine(t, map[string]interface{}{
			"@timestamp": "2026-06-21T12:01:00Z", "message": "connected",
			"threadId": threadID, "pid": pid,
		}),
		mustJSONLine(t, map[string]interface{}{
			"@timestamp": "2026-06-21T12:01:01Z", "message": "onToolLease",
			"threadId": threadID, "pid": pid,
			"data": map[string]interface{}{
				"toolCallId": toolCallID,
				"toolName":   "apply_patch",
				"args": map[string]interface{}{
					"patchText": "*** Begin Patch\n*** Update File: README.md\n-old\n+new\n*** End Patch",
				},
			},
		}),
		mustJSONLine(t, map[string]interface{}{
			"@timestamp": "2026-06-21T12:01:02Z", "message": "websocket message",
			"threadId": threadID, "pid": pid, "type": "executor_tool_result",
			"toolCallId": toolCallID, "runStatus": "done", "hasRunError": false,
		}),
	}, "\n") + "\n"

	transcriptPath := filepath.Join(transcriptDir, threadID+".log")
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	modifiedAt := time.Date(2026, 6, 21, 12, 2, 0, 0, time.UTC)
	require.NoError(t, os.Chtimes(transcriptPath, modifiedAt, modifiedAt))

	got, err := (ai.Amp{
		After:             time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, filepath.Join(projectDir, "README.md"), got[0].Entity)
	assert.Equal(t, threadID, got[0].AISession)
	assert.Equal(t, "amp/2.3.4 plugin/0.0.1", got[0].UserAgent)
}

func TestAmpParse_NoThreadLogsDir(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))

	got, err := ai.Amp{After: time.Now()}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}
