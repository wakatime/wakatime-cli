package ai_test

import (
	"encoding/json"
	"net/url"
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

func TestCopilotParseJSONSessionAndEditState(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	workspaceDir := filepath.Join(home, "Library", "Application Support", "Code", "User", "workspaceStorage", "workspace-1")
	require.NoError(t, os.MkdirAll(filepath.Join(workspaceDir, "chatSessions"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(workspaceDir, "chatEditingSessions", "session-1"), 0o755))

	mainFile := filepath.Join(home, "project", "main.go")
	secondFile := filepath.Join(home, "project", "util.go")

	require.NoError(t, os.MkdirAll(filepath.Dir(mainFile), 0o755))
	require.NoError(t, os.WriteFile(mainFile, []byte("package main\n"), 0o644))
	require.NoError(t, os.WriteFile(secondFile, []byte("package main\nfunc util() {}\n"), 0o644))

	sessionPath := filepath.Join(workspaceDir, "chatSessions", "session-1.json")
	session := map[string]any{
		"version":         3,
		"creationDate":    int64(1770000000000),
		"lastMessageDate": int64(1770000100000),
		"sessionId":       "session-1",
		"requests": []any{
			map[string]any{
				"requestId": "request-1",
				"timestamp": int64(1770000001000),
				"agent": map[string]any{
					"extensionVersion": "0.42.3",
				},
				"message": map[string]any{
					"text": "Update the file and explain what changed",
				},
				"result": map[string]any{
					"metadata": map[string]any{
						"promptTokens": 9,
						"outputTokens": 4,
					},
				},
				"modelState": map[string]any{
					"value":       1,
					"completedAt": int64(1770000009000),
				},
				"variableData": map[string]any{
					"variables": []any{
						map[string]any{
							"kind": "file",
							"id":   "file://" + strings.ReplaceAll(mainFile, " ", "%20"),
							"value": map[string]any{
								"fsPath": mainFile,
							},
						},
					},
				},
				"response": []any{
					map[string]any{
						"kind": "thinking",
					},
					map[string]any{
						"kind": "toolInvocationSerialized",
						"invocationMessage": map[string]any{
							"value": "Reading files",
							"uris": map[string]any{
								"file://" + strings.ReplaceAll(secondFile, " ", "%20"): map[string]any{
									"fsPath": secondFile,
								},
							},
						},
					},
				},
			},
		},
	}
	data, err := json.Marshal(session)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(sessionPath, data, 0o644))

	statePath := filepath.Join(workspaceDir, "chatEditingSessions", "session-1", "state.json")
	state := map[string]any{
		"version": 2,
		"timeline": map[string]any{
			"fileBaselines": []any{
				[]any{
					"file://" + mainFile + "::request-1",
					map[string]any{
						"content": "package main\n",
					},
				},
			},
			"operations": []any{
				map[string]any{
					"type":      "textEdit",
					"requestId": "request-1",
					"uri": map[string]any{
						"fsPath": mainFile,
					},
					"epoch": 1,
					"edits": []any{
						map[string]any{
							"text": "package main\n\nfunc main() {}\n",
							"range": map[string]any{
								"startLineNumber": 1,
								"startColumn":     1,
								"endLineNumber":   2,
								"endColumn":       1,
							},
						},
					},
				},
			},
		},
	}
	data, err = json.Marshal(state)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(statePath, data, 0o644))

	got, err := ai.Copilot{
		After:             time.Unix(1769999990, 0),
		FallbackUserAgent: "editor/1.0.0",
	}.Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 5)

	assert.Equal(t, "Copilot session-1", got[0].Entity)
	assert.Equal(t, "session-1", got[0].AISession)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, int64(9), got[0].AIInputTokens)
	assert.Equal(t, int64(4), got[0].AIOutputTokens)
	assert.Equal(t, len([]rune("Update the file and explain what changed")), got[0].AIPromptLength)
	require.NotNil(t, got[0].IsWrite)
	assert.False(t, *got[0].IsWrite)
	assert.Equal(t, filepath.Dir(mainFile), got[0].ProjectPathOverride)
	assert.Contains(t, got[0].UserAgent, "github-copilot/0.42.3")
	assert.NotContains(t, got[0].UserAgent, "Copilot/")

	assert.Equal(t, mainFile, got[1].Entity)
	assert.Equal(t, "session-1", got[1].AISession)
	assert.Equal(t, heartbeat.FileType, got[1].EntityType)
	require.NotNil(t, got[1].IsWrite)
	assert.False(t, *got[1].IsWrite)

	assert.Equal(t, secondFile, got[2].Entity)
	assert.Equal(t, "session-1", got[2].AISession)
	require.NotNil(t, got[2].IsWrite)
	assert.False(t, *got[2].IsWrite)

	assert.Equal(t, mainFile, got[3].Entity)
	assert.Equal(t, "session-1", got[3].AISession)
	require.NotNil(t, got[3].AILineChanges)
	assert.Equal(t, 2, *got[3].AILineChanges)
	require.NotNil(t, got[3].IsWrite)
	assert.True(t, *got[3].IsWrite)
	assert.Contains(t, got[3].UserAgent, "github-copilot/0.42.3")
	assert.NotContains(t, got[3].UserAgent, "Copilot/")

	assert.Equal(t, "Copilot session-1", got[4].Entity)
	assert.Equal(t, "session-1", got[4].AISession)
	assert.Equal(t, heartbeat.AppType, got[4].EntityType)
	assert.Zero(t, got[4].AIPromptLength)
}

func TestCopilotParseJSONLSession(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	workspaceDir := filepath.Join(home, "Library", "Application Support", "Code", "User", "workspaceStorage", "workspace-2")
	require.NoError(t, os.MkdirAll(filepath.Join(workspaceDir, "chatSessions"), 0o755))

	readFile := filepath.Join(home, "project-jsonl", "page.astro")
	require.NoError(t, os.MkdirAll(filepath.Dir(readFile), 0o755))
	require.NoError(t, os.WriteFile(readFile, []byte("---\n---\n"), 0o644))

	sessionPath := filepath.Join(workspaceDir, "chatSessions", "session-2.jsonl")
	readFileURI := (&url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(readFile),
	}).String()
	readFileURIJSON, err := json.Marshal(readFileURI)
	require.NoError(t, err)
	readFileJSON, err := json.Marshal(readFile)
	require.NoError(t, err)

	requestLine := strings.Join([]string{
		`{"kind":2,"k":["requests"],"v":[{"requestId":"request-a",`,
		`"timestamp":1771000005000,`,
		`"agent":{"extensionVersion":"0.42.3"},`,
		`"message":{"text":"Please inspect the Astro page"},`,
		`"result":{"metadata":{"usage":{"promptTokens":12,"completionTokens":5}}},`,
		`"modelState":{"value":1,"completedAt":1771000009000},`,
		`"response":[]}],"i":0}`,
	}, "")
	responseLine := strings.Join([]string{
		`{"kind":2,"k":["requests",0,"response"],"v":[{`,
		`"kind":"toolInvocationSerialized",`,
		`"invocationMessage":{"value":"Reading file","uris":{`,
		string(readFileURIJSON),
		`:{"fsPath":`,
		string(readFileJSON),
		`}}}}],"i":0}`,
	}, "")

	lines := []string{
		`{"kind":0,"v":{"version":3,"creationDate":1771000000000,"sessionId":"session-2","requests":[]}}`,
		requestLine,
		responseLine,
	}
	require.NoError(t, os.WriteFile(sessionPath, []byte(strings.Join(lines, "\n")+"\n"), 0o644))

	got, err := ai.Copilot{
		After:             time.Unix(1771000000, 0),
		FallbackUserAgent: "editor/1.0.0",
	}.Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 3)

	assert.Equal(t, "Copilot session-2", got[0].Entity)
	assert.Equal(t, "session-2", got[0].AISession)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, int64(12), got[0].AIInputTokens)
	assert.Equal(t, int64(5), got[0].AIOutputTokens)
	assert.Equal(t, len([]rune("Please inspect the Astro page")), got[0].AIPromptLength)
	assert.Equal(t, readFile, got[1].Entity)
	assert.Equal(t, "session-2", got[1].AISession)
	assert.Equal(t, heartbeat.FileType, got[1].EntityType)
	require.NotNil(t, got[1].IsWrite)
	assert.False(t, *got[1].IsWrite)
	assert.Equal(t, "Copilot session-2", got[2].Entity)
	assert.Equal(t, "session-2", got[2].AISession)
	assert.Equal(t, heartbeat.AppType, got[2].EntityType)
	assert.Zero(t, got[2].AIPromptLength)
}

func TestCopilotParseJSONLSession_IgnoresStringVariableValues(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	workspaceDir := filepath.Join(home, "Library", "Application Support", "Code", "User", "workspaceStorage", "workspace-3")
	require.NoError(t, os.MkdirAll(filepath.Join(workspaceDir, "chatSessions"), 0o755))

	sessionPath := filepath.Join(workspaceDir, "chatSessions", "session-3.jsonl")
	lines := []string{
		strings.Join([]string{
			`{"kind":0,"v":{"version":3,"creationDate":1771000000000,`,
			`"sessionId":"session-3","requests":[{"requestId":"request-b",`,
			`"timestamp":1771000005000,`,
			`"agent":{"extensionVersion":"0.43.0"},`,
			`"message":{"text":"Find bugs in this repo"},`,
			`"variableData":{"variables":[{"id":"vscode.customizations.index",`,
			`"kind":"promptText","value":"<skills>...</skills>"}]},`,
			`"response":[]}]}}`,
		}, ""),
		`{"kind":1,"k":["requests",0,"modelState"],"v":{"value":1,"completedAt":1771000009000}}`,
	}
	require.NoError(t, os.WriteFile(sessionPath, []byte(strings.Join(lines, "\n")+"\n"), 0o644))

	got, err := ai.Copilot{
		After:             time.Unix(1771000000, 0),
		FallbackUserAgent: "editor/1.0.0",
	}.Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "Copilot session-3", got[0].Entity)
	assert.Equal(t, "session-3", got[0].AISession)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, len([]rune("Find bugs in this repo")), got[0].AIPromptLength)
	assert.Equal(t, "Copilot session-3", got[1].Entity)
	assert.Equal(t, "session-3", got[1].AISession)
	assert.Equal(t, heartbeat.AppType, got[1].EntityType)
	assert.Zero(t, got[1].AIPromptLength)
}

func TestCopilotParseCLISessionEvents(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	sessionID := "cli-session-1"
	projectDir := filepath.Join(home, "project")
	readmePath := filepath.Join(projectDir, "README.md")
	planPath := filepath.Join(home, ".copilot", "session-state", "other-session", "plan.md")
	windowsPlanPath := `C:\Users\user\.copilot\session-state\windows-session\plan.md`
	eventsPath := filepath.Join(home, ".copilot", "session-state", sessionID, "events.jsonl")

	require.NoError(t, os.MkdirAll(filepath.Dir(readmePath), 0o755))
	require.NoError(t, os.WriteFile(readmePath, []byte("# Project\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Dir(eventsPath), 0o755))

	writeCopilotCLIEvents(t, eventsPath, []map[string]any{
		{
			"type":      "session.start",
			"timestamp": "2026-06-14T12:00:00Z",
			"data": map[string]any{
				"sessionId":      sessionID,
				"copilotVersion": "1.0.62",
				"context": map[string]any{
					"cwd":     projectDir,
					"gitRoot": projectDir,
				},
			},
		},
		{
			"type":      "session.model_change",
			"timestamp": "2026-06-14T12:00:00.500Z",
			"data": map[string]any{
				"newModel": "gpt-5.4",
			},
		},
		{
			"type":      "user.message",
			"timestamp": "2026-06-14T12:00:01Z",
			"data": map[string]any{
				"content": "Update README",
			},
		},
		{
			"type":      "assistant.message",
			"timestamp": "2026-06-14T12:00:02Z",
			"data": map[string]any{
				"model":        "gpt-5.4",
				"outputTokens": 9,
			},
		},
		{
			"type":      "tool.execution_start",
			"timestamp": "2026-06-14T12:00:02.100Z",
			"data": map[string]any{
				"toolCallId": "tool-1",
				"toolName":   "apply_patch",
			},
		},
		{
			"type":      "tool.execution_complete",
			"timestamp": "2026-06-14T12:00:03Z",
			"data": map[string]any{
				"toolCallId": "tool-1",
				"success":    true,
				"toolTelemetry": map[string]any{
					"properties": map[string]any{
						"codeBlocks": jsonEncodedString(t, []map[string]any{
							{
								"fileExt":      ".md",
								"languageId":   "markdown",
								"linesAdded":   1,
								"linesRemoved": 1,
							},
						}),
					},
					"restrictedProperties": map[string]any{
						"filePaths":    jsonEncodedString(t, []string{readmePath}),
						"addedPaths":   jsonEncodedString(t, []string{}),
						"deletedPaths": jsonEncodedString(t, []string{}),
					},
					"metrics": map[string]any{
						"linesAdded":   1,
						"linesRemoved": 1,
					},
				},
			},
		},
		{
			"type":      "tool.execution_start",
			"timestamp": "2026-06-14T12:00:03.100Z",
			"data": map[string]any{
				"toolCallId": "tool-plan",
				"toolName":   "apply_patch",
			},
		},
		{
			"type":      "tool.execution_complete",
			"timestamp": "2026-06-14T12:00:03.200Z",
			"data": map[string]any{
				"toolCallId": "tool-plan",
				"success":    true,
				"toolTelemetry": map[string]any{
					"restrictedProperties": map[string]any{
						"filePaths": jsonEncodedString(t, []string{planPath, windowsPlanPath}),
					},
					"metrics": map[string]any{
						"linesAdded":   4,
						"linesRemoved": 0,
					},
				},
			},
		},
		{
			"type":      "session.shutdown",
			"timestamp": "2026-06-14T12:00:04Z",
			"data": map[string]any{
				"currentModel": "gpt-5.4",
				"tokenDetails": map[string]any{
					"input": map[string]any{
						"tokenCount": 17,
					},
					"output": map[string]any{
						"tokenCount": 11,
					},
				},
				"codeChanges": map[string]any{
					"linesAdded":    5,
					"linesRemoved":  1,
					"filesModified": []string{readmePath, planPath},
				},
			},
		},
	})

	got, err := ai.Copilot{
		After:             time.Date(2026, time.June, 14, 11, 59, 0, 0, time.UTC),
		FallbackUserAgent: "editor/1.2.3",
	}.Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 4)

	assert.Equal(t, "Copilot "+sessionID, got[0].Entity)
	assert.Equal(t, sessionID, got[0].AISession)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, len([]rune("Update README")), got[0].AIPromptLength)
	assert.Equal(t, projectDir, got[0].ProjectPathOverride)
	require.NotNil(t, got[0].IsWrite)
	assert.False(t, *got[0].IsWrite)
	assert.Contains(t, got[0].UserAgent, "gpt/5.4")
	assert.True(t, strings.Index(got[0].UserAgent, "gpt/5.4") <
		strings.Index(got[0].UserAgent, "github-copilot-cli/1.0.62"))
	assert.Contains(t, got[0].UserAgent, "github-copilot-cli/1.0.62 copilot/1.0.62")
	assert.Contains(t, got[0].UserAgent, "editor/1.2.3")

	assert.Equal(t, "Copilot "+sessionID, got[1].Entity)
	assert.Equal(t, heartbeat.AppType, got[1].EntityType)
	assert.Equal(t, int64(0), got[1].AIInputTokens)
	assert.Equal(t, int64(9), got[1].AIOutputTokens)

	assert.Equal(t, readmePath, got[2].Entity)
	assert.Equal(t, sessionID, got[2].AISession)
	assert.Equal(t, heartbeat.FileType, got[2].EntityType)
	require.NotNil(t, got[2].IsWrite)
	assert.True(t, *got[2].IsWrite)
	require.NotNil(t, got[2].AILineChanges)
	assert.Equal(t, 0, *got[2].AILineChanges)
	assert.Contains(t, got[2].UserAgent, "github-copilot-cli/1.0.62 copilot/1.0.62")

	assert.Equal(t, "Copilot "+sessionID, got[3].Entity)
	assert.Equal(t, heartbeat.AppType, got[3].EntityType)
	assert.Equal(t, int64(17), got[3].AIInputTokens)
	assert.Equal(t, int64(2), got[3].AIOutputTokens)

	for _, h := range got {
		assert.NotEqual(t, planPath, h.Entity)
		assert.NotEqual(t, windowsPlanPath, h.Entity)
	}
}

func TestCopilotParseCLISessionEvents_IgnoresSessionWithoutEvents(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	sessionDir := filepath.Join(home, ".copilot", "session-state", "cli-session-empty")
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "workspace.yaml"), []byte("id: cli-session-empty\n"), 0o644))

	got, err := ai.Copilot{
		After:             time.Date(2026, time.June, 14, 11, 59, 0, 0, time.UTC),
		FallbackUserAgent: "editor/1.2.3",
	}.Parse(t.Context())
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestCopilotParseCLISessionEvents_IgnoresNoopTranscript(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	sessionID := "cli-session-noop"
	eventsPath := filepath.Join(home, ".copilot", "session-state", sessionID, "events.jsonl")
	require.NoError(t, os.MkdirAll(filepath.Dir(eventsPath), 0o755))

	writeCopilotCLIEvents(t, eventsPath, []map[string]any{
		{
			"type":      "session.start",
			"timestamp": "2026-06-14T12:00:00Z",
			"data": map[string]any{
				"sessionId":      sessionID,
				"copilotVersion": "1.0.62",
			},
		},
		{
			"type":      "session.shutdown",
			"timestamp": "2026-06-14T12:00:01Z",
			"data": map[string]any{
				"currentModel": "gpt-5.4",
			},
		},
	})

	got, err := ai.Copilot{
		After:             time.Date(2026, time.June, 14, 11, 59, 0, 0, time.UTC),
		FallbackUserAgent: "editor/1.2.3",
	}.Parse(t.Context())
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestCopilotParseCLISessionEvents_IgnoresPreCutoffActivityWithoutShutdownDelta(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	sessionID := "cli-session-before-cutoff"
	eventsPath := filepath.Join(home, ".copilot", "session-state", sessionID, "events.jsonl")
	require.NoError(t, os.MkdirAll(filepath.Dir(eventsPath), 0o755))

	writeCopilotCLIEvents(t, eventsPath, []map[string]any{
		{
			"type":      "session.start",
			"timestamp": "2026-06-14T12:00:00Z",
			"data": map[string]any{
				"sessionId":      sessionID,
				"copilotVersion": "1.0.62",
			},
		},
		{
			"type":      "user.message",
			"timestamp": "2026-06-14T12:00:01Z",
			"data": map[string]any{
				"content": "Update README",
			},
		},
		{
			"type":      "assistant.message",
			"timestamp": "2026-06-14T12:00:02Z",
			"data": map[string]any{
				"model":        "gpt-5.4",
				"outputTokens": 9,
			},
		},
		{
			"type":      "session.shutdown",
			"timestamp": "2026-06-14T12:00:04Z",
			"data": map[string]any{
				"currentModel": "gpt-5.4",
				"tokenDetails": map[string]any{
					"output": map[string]any{
						"tokenCount": 9,
					},
				},
			},
		},
	})

	got, err := ai.Copilot{
		After:             time.Date(2026, time.June, 14, 12, 0, 3, 0, time.UTC),
		FallbackUserAgent: "editor/1.2.3",
	}.Parse(t.Context())
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestCopilotParseCLISessionEvents_DoesNotFallbackWriteForPreCutoffTool(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	sessionID := "cli-session-pre-cutoff-tool"
	eventsPath := filepath.Join(home, ".copilot", "session-state", sessionID, "events.jsonl")
	require.NoError(t, os.MkdirAll(filepath.Dir(eventsPath), 0o755))

	writeCopilotCLIEvents(t, eventsPath, []map[string]any{
		{
			"type":      "session.start",
			"timestamp": "2026-06-14T12:00:00Z",
			"data": map[string]any{
				"sessionId":      sessionID,
				"copilotVersion": "1.0.62",
			},
		},
		{
			"type":      "tool.execution_start",
			"timestamp": "2026-06-14T12:00:01Z",
			"data": map[string]any{
				"toolCallId": "tool-1",
				"toolName":   "apply_patch",
			},
		},
		{
			"type":      "tool.execution_complete",
			"timestamp": "2026-06-14T12:00:02Z",
			"data": map[string]any{
				"toolCallId": "tool-1",
				"success":    true,
				"toolTelemetry": map[string]any{
					"restrictedProperties": map[string]any{
						"filePaths": jsonEncodedString(t, []string{`C:\project\README.md`}),
					},
					"metrics": map[string]any{
						"linesAdded":   1,
						"linesRemoved": 0,
					},
				},
			},
		},
		{
			"type":      "session.shutdown",
			"timestamp": "2026-06-14T12:00:04Z",
			"data": map[string]any{
				"currentModel": "gpt-5.4",
				"codeChanges": map[string]any{
					"linesAdded":    1,
					"linesRemoved":  0,
					"filesModified": []string{"C:/project/README.md"},
				},
			},
		},
	})

	got, err := ai.Copilot{
		After:             time.Date(2026, time.June, 14, 12, 0, 3, 0, time.UTC),
		FallbackUserAgent: "editor/1.2.3",
	}.Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
}

func TestCopilotParseCLISessionEvents_TailScannerKeepsStartMetadata(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	sessionID := "cli-session-large"
	projectDir := filepath.Join(home, "project")
	eventsPath := filepath.Join(home, ".copilot", "session-state", sessionID, "events.jsonl")
	require.NoError(t, os.MkdirAll(filepath.Dir(eventsPath), 0o755))

	start, err := json.Marshal(map[string]any{
		"type":      "session.start",
		"timestamp": "2026-06-14T12:00:00Z",
		"data": map[string]any{
			"sessionId":      sessionID,
			"copilotVersion": "1.0.62",
			"context": map[string]any{
				"cwd":     projectDir,
				"gitRoot": projectDir,
			},
		},
	})
	require.NoError(t, err)

	message, err := json.Marshal(map[string]any{
		"type":      "user.message",
		"timestamp": "2026-06-14T12:00:01Z",
		"data": map[string]any{
			"content": "Update README",
		},
	})
	require.NoError(t, err)

	filler := `{"type":"noop","timestamp":"2026-06-14T12:00:00Z","data":{"message":"` +
		strings.Repeat("x", 900) + `"}}` + "\n"

	var transcript strings.Builder
	transcript.Write(start)
	transcript.WriteByte('\n')

	for transcript.Len() <= 11*1024*1024 {
		transcript.WriteString(filler)
	}

	transcript.Write(message)
	transcript.WriteByte('\n')
	require.NoError(t, os.WriteFile(eventsPath, []byte(transcript.String()), 0o644))

	got, err := ai.Copilot{
		After:             time.Date(2026, time.June, 14, 12, 0, 0, 0, time.UTC),
		FallbackUserAgent: "editor/1.2.3",
	}.Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, projectDir, got[0].ProjectPathOverride)
	assert.Contains(t, got[0].UserAgent, "github-copilot-cli/1.0.62 copilot/1.0.62")
}

func writeCopilotCLIEvents(t *testing.T, path string, events []map[string]any) {
	t.Helper()

	var lines []string

	for _, event := range events {
		data, err := json.Marshal(event)
		require.NoError(t, err)

		lines = append(lines, string(data))
	}

	require.NoError(t, os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644))
}

func jsonEncodedString(t *testing.T, value any) string {
	t.Helper()

	data, err := json.Marshal(value)
	require.NoError(t, err)

	return string(data)
}
