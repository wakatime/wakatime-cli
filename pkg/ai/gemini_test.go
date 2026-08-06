package ai_test

import (
	"context"
	"encoding/json"
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

func TestGeminiParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	projectDir := filepath.Join(home, "wakatime-cli")
	require.NoError(t, os.MkdirAll(projectDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".gemini"), 0o755))

	projectSlugDir := filepath.Join(home, ".gemini", "tmp", "wakatime-cli")
	sessionDir := filepath.Join(projectSlugDir, "chats")
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(projectSlugDir, ".project_root"),
		[]byte(projectDir),
		0o644,
	))

	logs := []map[string]any{
		{
			"sessionId": "gem-session-1",
			"type":      "user",
			"message":   "fallback prompt",
			"timestamp": "2026-04-21T11:59:59Z",
		},
	}
	logContents, err := json.Marshal(logs)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(
		filepath.Join(projectSlugDir, "logs.json"),
		logContents,
		0o644,
	))

	mainFile := filepath.Join(projectDir, "pkg", "ai", "gemini.go")
	notesFile := filepath.Join(projectDir, "notes.txt")

	session := map[string]any{
		"sessionId":   "gem-session-1",
		"projectHash": "hash",
		"startTime":   "2026-04-21T12:00:00Z",
		"lastUpdated": "2026-04-21T12:00:07Z",
		"messages": []map[string]any{
			{
				"id":        "m1",
				"type":      "user",
				"timestamp": "2026-04-21T12:00:01Z",
				"content": []map[string]any{
					{"text": "look for bugs and fix any you find"},
				},
			},
			{
				"id":        "m2",
				"type":      "gemini",
				"timestamp": "2026-04-21T12:00:02Z",
				"content":   "I will inspect the parser files first.",
				"tokens": map[string]any{
					"input":  100,
					"output": 20,
					"cached": 0,
					"tool":   0,
					"total":  120,
				},
				"model": "gemini-3-flash-preview",
				"toolCalls": []map[string]any{
					{
						"id":        "list-1",
						"name":      "list_directory",
						"args":      map[string]any{"dir_path": "pkg/ai"},
						"status":    "success",
						"timestamp": "2026-04-21T12:00:02Z",
					},
					{
						"id":        "read-1",
						"name":      "read_file",
						"args":      map[string]any{"file_path": "pkg/ai/ai.go"},
						"status":    "success",
						"timestamp": "2026-04-21T12:00:02Z",
					},
				},
			},
			{
				"id":        "m3",
				"type":      "gemini",
				"timestamp": "2026-04-21T12:00:03Z",
				"content":   "I found one issue and patched it.",
				"tokens": map[string]any{
					"input":  160,
					"output": 34,
					"cached": 30,
					"tool":   0,
					"total":  194,
				},
				"model": "gemini-3-flash-preview",
				"toolCalls": []map[string]any{
					{
						"id":     "replace-1",
						"name":   "replace",
						"status": "success",
						"args": map[string]any{
							"file_path":  "pkg/ai/gemini.go",
							"old_string": "old line",
							"new_string": "new line\nextra line",
						},
						"timestamp": "2026-04-21T12:00:03Z",
						"resultDisplay": map[string]any{
							"filePath": mainFile,
							"diffStat": map[string]any{
								"model_added_lines":   3,
								"model_removed_lines": 1,
							},
						},
					},
				},
			},
			{
				"id":        "m4",
				"type":      "gemini",
				"timestamp": "2026-04-21T12:00:04Z",
				"content":   "This replace was cancelled.",
				"tokens": map[string]any{
					"input":  170,
					"output": 40,
					"cached": 30,
					"tool":   0,
					"total":  210,
				},
				"model": "gemini-3-flash-preview",
				"toolCalls": []map[string]any{
					{
						"id":     "replace-2",
						"name":   "replace",
						"status": "cancelled",
						"args": map[string]any{
							"file_path":  "pkg/ai/ai.go",
							"old_string": "before",
							"new_string": "after",
						},
						"timestamp": "2026-04-21T12:00:04Z",
						"resultDisplay": map[string]any{
							"filePath": filepath.Join(projectDir, "pkg", "ai", "ai.go"),
						},
					},
				},
			},
			{
				"id":        "m5",
				"type":      "gemini",
				"timestamp": "2026-04-21T12:00:05Z",
				"content":   "I also wrote a notes file.",
				"tokens": map[string]any{
					"input":  190,
					"output": 45,
					"cached": 30,
					"tool":   0,
					"total":  235,
				},
				"model": "gemini-3-flash-preview",
				"toolCalls": []map[string]any{
					{
						"id":     "write-1",
						"name":   "write_file",
						"status": "success",
						"args": map[string]any{
							"file_path": "notes.txt",
							"content":   "alpha\nbeta",
						},
						"timestamp":     "2026-04-21T12:00:05Z",
						"resultDisplay": "",
					},
				},
			},
		},
	}

	sessionContents, err := json.Marshal(session)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionDir, "session-2026-04-21T12-00-00-test.json"),
		sessionContents,
		0o644,
	))

	parser := ai.Gemini{
		After:             time.Date(2026, 4, 21, 11, 59, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			mainFile: heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 7)

	assert.Equal(t, "Gemini gem-session-1", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "gem-session-1", got[0].AISession)
	assert.Equal(t, projectDir, got[0].ProjectPathOverride)
	assert.Equal(t, len([]rune("look for bugs and fix any you find")), got[0].AIPromptLength)
	assert.Zero(t, got[0].AIInputTokens)
	assert.Zero(t, got[0].AIOutputTokens)
	assert.Equal(t, "plugin/0.0.1", got[0].UserAgent)

	assert.Equal(t, "Gemini gem-session-1", got[1].Entity)
	assert.Equal(t, heartbeat.AppType, got[1].EntityType)
	assert.Zero(t, got[1].AIPromptLength)
	assert.EqualValues(t, 100, got[1].AIInputTokens)
	assert.EqualValues(t, 20, got[1].AIOutputTokens)
	assert.Equal(t, projectDir, got[1].ProjectPathOverride)
	assert.Contains(t, got[1].UserAgent, "gemini/3-flash-preview")
	assert.NotContains(t, got[1].UserAgent, "Gemini/")

	assert.Equal(t, "Gemini gem-session-1", got[2].Entity)
	assert.Equal(t, heartbeat.AppType, got[2].EntityType)
	assert.Zero(t, got[2].AIPromptLength)
	assert.EqualValues(t, 30, got[2].AIInputTokens)
	assert.EqualValues(t, 30, got[2].AICachedInputTokens)
	assert.EqualValues(t, 14, got[2].AIOutputTokens)

	assert.Equal(t, mainFile, got[3].Entity)
	assert.Equal(t, heartbeat.FileType, got[3].EntityType)
	require.NotNil(t, got[3].AILineChanges)
	assert.Equal(t, 2, *got[3].AILineChanges)
	require.NotNil(t, got[3].IsWrite)
	assert.True(t, *got[3].IsWrite)
	assert.Zero(t, got[3].AIInputTokens)
	assert.Zero(t, got[3].AIOutputTokens)
	assert.Contains(t, got[3].UserAgent, "gemini/3-flash-preview")
	assert.NotContains(t, got[3].UserAgent, "Gemini/")
	assert.Contains(t, got[3].UserAgent, "editor/1.2.3")

	assert.Equal(t, "Gemini gem-session-1", got[4].Entity)
	assert.Equal(t, heartbeat.AppType, got[4].EntityType)
	assert.Zero(t, got[4].AIPromptLength)
	assert.EqualValues(t, 10, got[4].AIInputTokens)
	assert.EqualValues(t, 6, got[4].AIOutputTokens)

	assert.Equal(t, "Gemini gem-session-1", got[5].Entity)
	assert.Equal(t, heartbeat.AppType, got[5].EntityType)
	assert.Zero(t, got[5].AIPromptLength)
	assert.EqualValues(t, 20, got[5].AIInputTokens)
	assert.EqualValues(t, 5, got[5].AIOutputTokens)

	assert.Equal(t, notesFile, got[6].Entity)
	assert.Equal(t, heartbeat.FileType, got[6].EntityType)
	require.NotNil(t, got[6].AILineChanges)
	assert.Equal(t, 2, *got[6].AILineChanges)
	require.NotNil(t, got[6].IsWrite)
	assert.True(t, *got[6].IsWrite)
	assert.Zero(t, got[6].AIInputTokens)
	assert.Zero(t, got[6].AIOutputTokens)
}

func TestGeminiParse_JSONLSubagent(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	projectDir := filepath.Join(home, "jsonl-project")
	projectSlugDir := filepath.Join(home, ".gemini", "tmp", "jsonl-project")
	sessionDir := filepath.Join(projectSlugDir, "chats", "parent-session")
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(projectSlugDir, ".project_root"),
		[]byte(projectDir),
		0o644,
	))

	records := []any{
		map[string]any{
			"sessionId":   "gem-subagent-1",
			"projectHash": "hash",
			"startTime":   "2026-07-01T12:00:00Z",
			"lastUpdated": "2026-07-01T12:00:04Z",
			"kind":        "subagent",
		},
		map[string]any{
			"id":        "m1",
			"type":      "user",
			"timestamp": "2026-07-01T12:00:01Z",
			"content":   "Inspect this parser",
		},
		map[string]any{
			"id":        "m2",
			"type":      "gemini",
			"timestamp": "2026-07-01T12:00:02Z",
			"content":   "I found the issue.",
			"model":     "gemini-3-flash-preview",
		},
		// Gemini appends a replacement record when token metadata arrives.
		map[string]any{
			"id":        "m2",
			"type":      "gemini",
			"timestamp": "2026-07-01T12:00:02Z",
			"content":   "I found the issue.",
			"model":     "gemini-3-flash-preview",
			"tokens": map[string]any{
				"input":  120,
				"output": 15,
				"cached": 30,
				"total":  135,
			},
		},
		map[string]any{
			"id":        "m3",
			"type":      "gemini",
			"timestamp": "2026-07-01T12:00:03Z",
			"content":   "This message is rewound.",
			"model":     "gemini-3-flash-preview",
		},
		map[string]any{"$rewindTo": "m3"},
		map[string]any{"$set": map[string]any{"lastUpdated": "2026-07-01T12:00:05Z"}},
	}

	var lines []string

	for _, record := range records {
		line, err := json.Marshal(record)
		require.NoError(t, err)

		lines = append(lines, string(line))
	}

	lines = append(lines, "{") // Current Gemini ignores malformed individual records.

	require.NoError(t, os.WriteFile(
		filepath.Join(sessionDir, "gem-subagent-1.jsonl"),
		[]byte(strings.Join(lines, "\n")+"\n"),
		0o644,
	))

	legacy, err := json.Marshal(map[string]any{
		"sessionId": "gem-subagent-1",
		"startTime": "2026-07-01T12:00:00Z",
		"messages": []map[string]any{
			{
				"id":        "legacy-only",
				"type":      "user",
				"timestamp": "2026-07-01T12:00:00Z",
				"content":   "This migrated legacy copy must not be parsed too",
			},
		},
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionDir, "gem-subagent-1.json"),
		legacy,
		0o644,
	))

	got, err := (ai.Gemini{
		After:             time.Date(2026, 7, 1, 11, 59, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "Gemini gem-subagent-1", got[0].Entity)
	assert.Equal(t, projectDir, got[0].ProjectPathOverride)
	assert.Equal(t, len([]rune("Inspect this parser")), got[0].AIPromptLength)

	assert.EqualValues(t, 90, got[1].AIInputTokens)
	assert.EqualValues(t, 30, got[1].AICachedInputTokens)
	assert.EqualValues(t, 15, got[1].AIOutputTokens)
}

func TestGeminiParse_NoGeminiTmpDir(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := ai.Gemini{After: time.Now()}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGeminiParse_AntigravityProducts(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	type productFixture struct {
		name             string
		appDir           string
		entityName       string
		userAgentProduct string
		version          string
		sessionID        string
		agentToken       string
		settingsModel    string
		prompt           string
		workspace        string
		relativeFile     string
		filePath         string
		startTime        time.Time
		diff             string
		lineChanges      int
	}

	products := []productFixture{
		{
			name:             "Desktop",
			appDir:           "antigravity",
			entityName:       "Antigravity Desktop",
			userAgentProduct: "antigravity-desktop",
			version:          "2.1.4",
			sessionID:        "2488a220-aa94-48fe-ad91-5ca34453e4d1",
			agentToken:       "gemini/3.5-flash-medium",
			settingsModel:    "Gemini 3.5 Flash (Medium)",
			prompt:           "because you are Antigravity 2.0, add yourself as author in the LICENSE file",
			workspace:        filepath.Join(home, "desktop-project"),
			relativeFile:     filepath.Join("plugins", "codex-cli-wakatime", "LICENSE"),
			startTime:        time.Date(2026, 6, 19, 19, 51, 0, 0, time.UTC),
			diff: "@@ -1,6 +1,7 @@\n" +
				" BSD 3-Clause License\n \n" +
				" Copyright (c) 2026, WakaTime\n" +
				"+Copyright (c) 2026, Antigravity 2.0\n" +
				" All rights reserved.",
			lineChanges: 1,
		},
		{
			name:             "IDE",
			appDir:           "antigravity-ide",
			entityName:       "Antigravity IDE",
			userAgentProduct: "antigravity-ide",
			version:          "1.107.0",
			sessionID:        "cc955ef1-cf5d-4a38-8d93-3a6ceb2bd2e8",
			agentToken:       "gemini/3.5-flash-medium",
			settingsModel:    "Gemini 3.5 Flash (Medium)",
			prompt:           "because you are Antigravity IDE, add yourself to the license",
			workspace:        filepath.Join(home, "ide-project"),
			relativeFile:     filepath.Join("plugins", "codex-cli-wakatime", "LICENSE"),
			startTime:        time.Date(2026, 6, 19, 19, 52, 0, 0, time.UTC),
			diff: "@@ -1,6 +1,6 @@\n" +
				" BSD 3-Clause License\n \n" +
				"-Copyright (c) 2026, WakaTime\n" +
				"+Copyright (c) 2026, WakaTime and Antigravity IDE\n" +
				" All rights reserved.",
			lineChanges: 0,
		},
		{
			name:             "CLI",
			appDir:           "antigravity-cli",
			entityName:       "Antigravity CLI",
			userAgentProduct: "antigravity-cli",
			version:          "unknown",
			sessionID:        "f7da61a0-c935-43b4-9425-eb08ba98231d",
			agentToken:       "gemini/3.5-flash-medium",
			settingsModel:    "Gemini 3.5 Flash (Medium)",
			prompt:           "pretend you wrote this repo, add a made with love to the end of the repo's readme file",
			workspace:        filepath.Join(home, "cli-project"),
			relativeFile:     "README.md",
			startTime:        time.Date(2026, 6, 19, 19, 53, 20, 0, time.UTC),
			diff: "@@ -29,4 +29,8 @@\n" +
				" [wakatime]: https://wakatime.com/\n" +
				" [codex-cli]: https://developers.openai.com/codex/cli\n" +
				" [wakatime-cli]: https://github.com/wakatime/wakatime-cli\n" +
				"+\n" +
				"+---\n" +
				"+Made with ❤️ by Antigravity\n" +
				"+",
			lineChanges: 4,
		},
	}

	conversationCaches := map[string]map[string]string{}
	userAgents := map[string]string{}

	for i := range products {
		product := &products[i]
		product.filePath = filepath.Join(product.workspace, product.relativeFile)
		userAgents[product.filePath] = heartbeat.UserAgent(ctx, "editor/1.2.3")
		appRoot := filepath.Join(home, ".gemini", product.appDir)
		require.NoError(t, os.MkdirAll(appRoot, 0o755))

		switch product.userAgentProduct {
		case "antigravity-desktop":
			bundle := filepath.Join(home, "Applications", "Antigravity.app")
			plistDir := filepath.Join(bundle, "Contents")
			require.NoError(t, os.MkdirAll(plistDir, 0o755))
			require.NoError(t, os.WriteFile(
				filepath.Join(plistDir, "Info.plist"),
				[]byte("<plist><dict><key>CFBundleShortVersionString</key><string>"+
					product.version+"</string></dict></plist>"),
				0o644,
			))
			binDir := filepath.Join(appRoot, "bin")
			require.NoError(t, os.MkdirAll(binDir, 0o755))
			require.NoError(t, os.WriteFile(
				filepath.Join(binDir, "agentapi"),
				[]byte("#!/bin/sh\nexec \""+
					filepath.Join(bundle, "Contents", "Resources", "bin", "language_server")+
					"\" agentapi \"$@\"\n"),
				0o755,
			))
		case "antigravity-ide":
			appDir := filepath.Join(home, "Applications", "Antigravity IDE.app", "Contents", "Resources", "app")
			require.NoError(t, os.MkdirAll(appDir, 0o755))
			writeJSON(t, filepath.Join(appDir, "product.json"), map[string]any{"version": product.version})
			binDir := filepath.Join(appRoot, "bin")
			require.NoError(t, os.MkdirAll(binDir, 0o755))
			require.NoError(t, os.WriteFile(
				filepath.Join(binDir, "agentapi"),
				[]byte("#!/bin/sh\nexec \""+filepath.Join(appDir, "extensions", "antigravity", "bin", "language_server")+
					"\" agentapi \"$@\"\n"),
				0o755,
			))
		}

		userContent := "<USER_REQUEST>\n" + product.prompt +
			"\n</USER_REQUEST>\n<ADDITIONAL_METADATA>ignore this</ADDITIONAL_METADATA>"
		if product.settingsModel != "" {
			userContent += "\n<USER_SETTINGS_CHANGE>\nThe user changed setting `Model Selection` from None to " +
				product.settingsModel + ". No need to comment on this change.\n</USER_SETTINGS_CHANGE>"
		}

		logsDir := filepath.Join(
			home,
			".gemini",
			product.appDir,
			"brain",
			product.sessionID,
			".system_generated",
			"logs",
		)
		require.NoError(t, os.MkdirAll(logsDir, 0o755), product.name)

		writeJSONL(t, filepath.Join(logsDir, "transcript.jsonl"), []map[string]any{
			{
				"step_index": 0,
				"source":     "USER_EXPLICIT",
				"type":       "USER_INPUT",
				"status":     "DONE",
				"created_at": product.startTime.Format(time.RFC3339),
				"content":    userContent,
			},
			{
				"step_index": 1,
				"source":     "MODEL",
				"type":       "PLANNER_RESPONSE",
				"status":     "DONE",
				"created_at": product.startTime.Add(time.Second).Format(time.RFC3339),
				"tool_calls": []map[string]any{
					{
						"name": "list_dir",
						"args": map[string]any{
							"DirectoryPath": "\"" + product.workspace + "\"",
						},
					},
				},
			},
			{
				"step_index": 2,
				"source":     "MODEL",
				"type":       "CODE_ACTION",
				"status":     "DONE",
				"created_at": product.startTime.Add(2 * time.Second).Format(time.RFC3339),
				"content": "The following changes were made by the replace_file_content tool to: " +
					product.filePath + ". If relevant, proactively run tests.\n" +
					"[diff_block_start]\n" +
					product.diff + "\n" +
					"[diff_block_end]",
			},
			{
				"step_index": 3,
				"source":     "MODEL",
				"type":       "PLANNER_RESPONSE",
				"status":     "DONE",
				"created_at": product.startTime.Add(3 * time.Second).Format(time.RFC3339),
				"content":    "I updated the requested file.",
			},
		})

		if product.userAgentProduct == "antigravity-cli" {
			conversationCaches[product.appDir] = map[string]string{
				product.workspace: product.sessionID,
			}
		}
	}

	for appDir, conversations := range conversationCaches {
		cacheDir := filepath.Join(home, ".gemini", appDir, "cache")
		require.NoError(t, os.MkdirAll(cacheDir, 0o755))
		writeJSON(t, filepath.Join(cacheDir, "last_conversations.json"), conversations)
	}

	legacyLogsDir := filepath.Join(
		home,
		".gemini",
		"antigravity-backup",
		"brain",
		"legacy-session",
		".system_generated",
		"logs",
	)
	require.NoError(t, os.MkdirAll(legacyLogsDir, 0o755))
	writeJSONL(t, filepath.Join(legacyLogsDir, "transcript.jsonl"), []map[string]any{
		{
			"source":     "USER_EXPLICIT",
			"type":       "USER_INPUT",
			"status":     "DONE",
			"created_at": "2026-06-19T19:55:00Z",
			"content":    "this legacy backup must not be parsed",
		},
	})

	got, err := ai.Gemini{
		After:             time.Date(2026, 6, 19, 19, 50, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents:        userAgents,
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 9)

	heartbeatsBySession := map[string][]heartbeat.Heartbeat{}
	for _, h := range got {
		heartbeatsBySession[h.AISession] = append(heartbeatsBySession[h.AISession], h)
	}

	for _, product := range products {
		t.Run(product.name, func(t *testing.T) {
			heartbeats := heartbeatsBySession[product.sessionID]
			require.Len(t, heartbeats, 3)

			for _, h := range heartbeats {
				assert.NotContains(t, strings.Fields(h.UserAgent), product.userAgentProduct)
			}

			userAgentProduct := product.userAgentProduct + "/" + product.version

			promptHeartbeat := heartbeats[0]
			assert.Equal(t, product.entityName+" "+product.sessionID, promptHeartbeat.Entity)
			assert.Equal(t, heartbeat.AppType, promptHeartbeat.EntityType)
			assert.Equal(t, len([]rune(product.prompt)), promptHeartbeat.AIPromptLength)
			assert.Equal(t, product.workspace, promptHeartbeat.ProjectPathOverride)
			assert.Contains(t, promptHeartbeat.UserAgent, product.agentToken)
			assert.Contains(t, promptHeartbeat.UserAgent, userAgentProduct)

			fileHeartbeat := heartbeats[1]
			assert.Equal(t, product.filePath, fileHeartbeat.Entity)
			assert.Equal(t, heartbeat.FileType, fileHeartbeat.EntityType)
			require.NotNil(t, fileHeartbeat.AILineChanges)
			assert.Equal(t, product.lineChanges, *fileHeartbeat.AILineChanges)
			require.NotNil(t, fileHeartbeat.IsWrite)
			assert.True(t, *fileHeartbeat.IsWrite)
			assert.Contains(t, fileHeartbeat.UserAgent, product.agentToken)
			assert.Contains(t, fileHeartbeat.UserAgent, userAgentProduct)
			assert.Contains(t, fileHeartbeat.UserAgent, "editor/1.2.3")

			responseHeartbeat := heartbeats[2]
			assert.Equal(t, product.entityName+" "+product.sessionID, responseHeartbeat.Entity)
			assert.Equal(t, heartbeat.AppType, responseHeartbeat.EntityType)
			assert.Zero(t, responseHeartbeat.AIPromptLength)
			assert.Equal(t, product.workspace, responseHeartbeat.ProjectPathOverride)
			assert.Contains(t, responseHeartbeat.UserAgent, product.agentToken)
			assert.Contains(t, responseHeartbeat.UserAgent, userAgentProduct)
		})
	}
}
