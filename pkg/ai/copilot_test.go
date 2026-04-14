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
	require.NotNil(t, got[0].IsWrite)
	assert.False(t, *got[0].IsWrite)
	assert.Equal(t, filepath.Dir(mainFile), got[0].ProjectPathOverride)
	assert.Contains(t, got[0].UserAgent, "Copilot/0.42.3")

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
	assert.Contains(t, got[3].UserAgent, "Copilot/0.42.3")

	assert.Equal(t, "Copilot session-1", got[4].Entity)
	assert.Equal(t, "session-1", got[4].AISession)
	assert.Equal(t, heartbeat.AppType, got[4].EntityType)
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
	assert.Equal(t, readFile, got[1].Entity)
	assert.Equal(t, "session-2", got[1].AISession)
	assert.Equal(t, heartbeat.FileType, got[1].EntityType)
	require.NotNil(t, got[1].IsWrite)
	assert.False(t, *got[1].IsWrite)
	assert.Equal(t, "Copilot session-2", got[2].Entity)
	assert.Equal(t, "session-2", got[2].AISession)
	assert.Equal(t, heartbeat.AppType, got[2].EntityType)
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
	assert.Equal(t, "Copilot session-3", got[1].Entity)
	assert.Equal(t, "session-3", got[1].AISession)
	assert.Equal(t, heartbeat.AppType, got[1].EntityType)
}
