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

func TestCodexParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	now := time.Now()
	transcriptDir := filepath.Join(home, ".codex", "sessions", now.Format("2006"), now.Format("01"), now.Format("02"))
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "rollout-2026-03-28T07-33-13-019d3438-39ae-7fb2-8526-d6c02ba3577c.jsonl")
	copyFile(t, "testdata/codex.jsonl", transcriptPath)

	parser := ai.Codex{
		After:             time.Date(2026, 3, 28, 04, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			"/home/user/projects/wakatime-cli/pkg/ai/claude.go": heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)

	for _, h := range got {
		assert.Equal(t, "plus", h.AISubscriptionPlan)
	}

	assert.Equal(t, "Codex rollout-2026-03-28T07-33-13-019d3438-39ae-7fb2-8526-d6c02ba3577c", got[0].Entity)
	assert.Equal(t, "019d3438-39ae-7fb2-8526-d6c02ba3577c", got[0].AISession)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[0].Category)
	assert.Nil(t, got[0].AILineChanges)
	assert.Equal(t, len([]rune("Please implement the code as described by the comment.")), got[0].AIPromptLength)
	assert.Equal(t, "/root/wakatime-cli", got[0].ProjectPathOverride)
	require.NotNil(t, got[0].IsWrite)
	assert.False(t, *got[0].IsWrite)
	assert.Equal(t, float64(time.Date(2026, 3, 28, 11, 33, 14, 289, time.UTC).Unix()), got[0].Time)
	assert.NotContains(t, got[0].UserAgent, "Codex/")
	assert.Contains(t, got[0].UserAgent, "vscode-wakatime/unknown")
	assert.True(
		t,
		strings.Index(got[0].UserAgent, "vscode-wakatime/unknown") <
			strings.Index(got[0].UserAgent, "plugin/0.0.1"),
	)
	assert.Contains(t, got[0].UserAgent, "plugin/0.0.1")

	assert.Equal(t, "Codex rollout-2026-03-28T07-33-13-019d3438-39ae-7fb2-8526-d6c02ba3577c", got[1].Entity)
	assert.Equal(t, heartbeat.AppType, got[1].EntityType)
	assert.Equal(t, "019d3438-39ae-7fb2-8526-d6c02ba3577c", got[1].AISession)
	assert.Nil(t, got[1].AILineChanges)
	assert.Zero(t, got[1].AIPromptLength)
	assert.Equal(t, "/root/wakatime-cli", got[1].ProjectPathOverride)
	require.NotNil(t, got[1].IsWrite)
	assert.False(t, *got[1].IsWrite)
	assert.Zero(t, got[1].AIInputTokens)
	assert.Zero(t, got[1].AIOutputTokens)
	assert.Equal(t, float64(time.Date(2026, 3, 28, 11, 33, 18, 535000000, time.UTC).Unix()), got[1].Time)
	assert.NotContains(t, got[1].UserAgent, "Codex/")
	assert.Contains(t, got[1].UserAgent, "vscode-wakatime/unknown")
	assert.True(
		t,
		strings.Index(got[1].UserAgent, "vscode-wakatime/unknown") <
			strings.Index(got[1].UserAgent, "plugin/0.0.1"),
	)
	assert.Contains(t, got[1].UserAgent, "plugin/0.0.1")
}

func TestCodexParse_ParsesTranscriptFromPreviousDayFolder(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".codex", "sessions", "2026", "03", "27")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "rollout-2026-03-28T07-33-13-019d3438-39ae-7fb2-8526-d6c02ba3577c.jsonl")
	copyFile(t, "testdata/codex.jsonl", transcriptPath)

	now := time.Now()
	require.NoError(t, os.Chtimes(transcriptPath, now, now))

	parser := ai.Codex{
		After: time.Date(2026, 3, 28, 11, 0, 0, 0, time.UTC),
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)
}

func TestCodexParse_NoCodexSessionsDir(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := ai.Codex{After: time.Now()}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestCodexParse_UserAgentUsesTranscriptSource(t *testing.T) {
	tests := map[string]struct {
		Source            string
		Version           string
		FallbackUserAgent string
		ExpectedUserAgent string
	}{
		"vscode": {
			Source:            "vscode",
			Version:           "0.131.0-alpha.9",
			FallbackUserAgent: "zoom.us/6.7.7(76486)-6.7.7.76486 macos-wakatime/5.28.4",
			ExpectedUserAgent: "vscode-wakatime/unknown " +
				"zoom.us/6.7.7(76486)-6.7.7.76486 macos-wakatime/5.28.4",
		},
		"cli": {
			Source:            "cli",
			Version:           "0.134.0",
			FallbackUserAgent: "claude-code/2.1.142 claude-code-wakatime/3.1.6",
			ExpectedUserAgent: "codex-cli/0.134.0 " +
				"claude-code/2.1.142 claude-code-wakatime/3.1.6",
		},
		"empty source preserves codex fallback": {
			Version:           "0.134.0",
			FallbackUserAgent: "codex/0.134.0 codex-wakatime/1.0.0",
			ExpectedUserAgent: "codex/0.134.0 codex-wakatime/1.0.0",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)

			transcriptDir := filepath.Join(home, ".codex", "sessions", "2026", "05", "27")
			require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

			transcriptPath := filepath.Join(transcriptDir, "session.jsonl")

			sessionMeta := map[string]interface{}{
				"timestamp": "2026-05-27T12:00:00Z",
				"type":      "session_meta",
				"payload": map[string]interface{}{
					"id":          "019e6b07-17ac-7070-bde0-864ecbda2cac",
					"cwd":         "/workspace/project",
					"cli_version": tt.Version,
				},
			}
			if tt.Source != "" {
				sessionMeta["payload"].(map[string]interface{})["source"] = tt.Source
			}

			firstLine, err := json.Marshal(sessionMeta)
			require.NoError(t, err)

			transcript := string(firstLine) + "\n" + strings.Join([]string{
				strings.Join([]string{
					`{"timestamp":"2026-05-27T12:00:01Z","type":"event_msg",`,
					`"payload":{"type":"agent_message","message":"I will make the change."}}`,
				}, ""),
			}, "\n") + "\n"
			require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

			parser := ai.Codex{
				After:             time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC),
				FallbackUserAgent: tt.FallbackUserAgent,
			}

			got, err := parser.Parse(ctx)
			require.NoError(t, err)
			require.Len(t, got, 1)
			assert.Equal(t, tt.ExpectedUserAgent, got[0].UserAgent)
		})
	}
}

func TestCodexParse_UserAgentUsesModelAndReasoningEffort(t *testing.T) {
	ctx := context.Background()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".codex", "sessions", "2026", "06", "19")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))
	transcriptPath := filepath.Join(transcriptDir, "rollout-2026-06-19T19-32-45-session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-06-19T23:33:06.518Z","type":"session_meta",`,
			`"payload":{"id":"session","cwd":"/workspace/project",`,
			`"cli_version":"0.141.0","source":"cli"}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-06-19T23:33:06.520Z","type":"turn_context",`,
			`"payload":{"model":"gpt-5.5","collaboration_mode":{"settings":`,
			`{"model":"gpt-5.5","reasoning_effort":"medium"}},"effort":"medium"}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-06-19T23:33:07.000Z","type":"event_msg",`,
			`"payload":{"type":"agent_message","message":"I will make the change."}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Codex{
		After:             time.Date(2026, 6, 19, 23, 33, 0, 0, time.UTC),
		FallbackUserAgent: "antigravity-cli/1.0.10 antigravity-cli-wakatime/1.0.0",
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t,
		"gpt/5.5-medium codex-cli/0.141.0 antigravity-cli/1.0.10 antigravity-cli-wakatime/1.0.0",
		got[0].UserAgent,
	)
}

func TestCodexParse_UpdatesReasoningEffortForFutureHeartbeats(t *testing.T) {
	ctx := context.Background()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".codex", "sessions", "2026", "06", "19")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))
	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	turnContext := func(timestamp string, effort string) string {
		return strings.Join([]string{
			`{"timestamp":"` + timestamp + `","type":"turn_context",`,
			`"payload":{"model":"gpt-5.5","collaboration_mode":{"settings":`,
			`{"model":"gpt-5.5","reasoning_effort":"` + effort + `"}},`,
			`"effort":"` + effort + `"}}`,
		}, "")
	}
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-06-20T00:34:30.000Z","type":"session_meta",`,
			`"payload":{"id":"session","cwd":"/workspace/project",`,
			`"cli_version":"0.141.0","source":"cli"}}`,
		}, ""),
		turnContext("2026-06-20T00:34:31.000Z", "high"),
		`{"timestamp":"2026-06-20T00:34:32.000Z","type":"event_msg",` +
			`"payload":{"type":"agent_message","message":"First response."}}`,
		turnContext("2026-06-20T00:34:39.244Z", "medium"),
		`{"timestamp":"2026-06-20T00:34:40.000Z","type":"event_msg",` +
			`"payload":{"type":"agent_message","message":"Second response."}}`,
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	got, err := (ai.Codex{
		After:             time.Date(2026, 6, 20, 0, 34, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "gpt/5.5-high codex-cli/0.141.0 plugin/0.0.1", got[0].UserAgent)
	assert.Equal(t, "gpt/5.5-medium codex-cli/0.141.0 plugin/0.0.1", got[1].UserAgent)
}

func TestCodexParse_AttributesTokenCountsToPreviousHeartbeat(t *testing.T) {
	ctx := context.Background()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".codex", "sessions", "2026", "06", "20")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))
	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		`{"timestamp":"2026-06-20T11:59:58Z","type":"session_meta",` +
			`"payload":{"id":"session","cwd":"/workspace/project"}}`,
		`{"timestamp":"2026-06-20T11:59:59Z","type":"event_msg",` +
			`"payload":{"type":"token_count","info":{"total_token_usage":` +
			`{"input_tokens":100,"output_tokens":10}}}}`,
		`{"timestamp":"2026-06-20T12:00:01Z","type":"response_item",` +
			`"payload":{"type":"message","role":"user","content":` +
			`[{"type":"input_text","text":"Make the change"}]}}`,
		`{"timestamp":"2026-06-20T12:00:02Z","type":"event_msg",` +
			`"payload":{"type":"token_count","info":{"total_token_usage":` +
			`{"input_tokens":110,"output_tokens":12}}}}`,
		`{"timestamp":"2026-06-20T12:00:03Z","type":"event_msg",` +
			`"payload":{"type":"agent_message","message":"The change is complete."}}`,
		`{"timestamp":"2026-06-20T12:00:04Z","type":"event_msg",` +
			`"payload":{"type":"token_count","info":{"total_token_usage":` +
			`{"input_tokens":150,"output_tokens":20}}}}`,
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	got, err := (ai.Codex{After: time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.EqualValues(t, 10, got[0].AIInputTokens)
	assert.EqualValues(t, 2, got[0].AIOutputTokens)
	assert.EqualValues(t, 40, got[1].AIInputTokens)
	assert.EqualValues(t, 8, got[1].AIOutputTokens)
}

func TestCodexParse_EmitsOnlySuccessfulPatches(t *testing.T) {
	ctx := context.Background()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".codex", "sessions", "2026", "06", "20")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))
	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		`{"timestamp":"2026-06-20T12:00:00Z","type":"session_meta",` +
			`"payload":{"id":"session","cwd":"/workspace/project"}}`,
		`{"timestamp":"2026-06-20T12:00:01Z","type":"response_item",` +
			`"payload":{"type":"custom_tool_call","call_id":"failed-direct",` +
			`"name":"apply_patch","input":"*** Update File: failed.go\n+one"}}`,
		`{"timestamp":"2026-06-20T12:00:02Z","type":"event_msg",` +
			`"payload":{"type":"patch_apply_end","call_id":"failed-direct",` +
			`"success":false,"status":"failed"}}`,
		`{"timestamp":"2026-06-20T12:00:03Z","type":"response_item",` +
			`"payload":{"type":"custom_tool_call","call_id":"successful-direct",` +
			`"name":"apply_patch","input":"*** Update File: successful.go\n+one"}}`,
		`{"timestamp":"2026-06-20T12:00:04Z","type":"event_msg",` +
			`"payload":{"type":"patch_apply_end","call_id":"successful-direct",` +
			`"success":true,"status":"completed"}}`,
		`{"timestamp":"2026-06-20T12:00:05Z","type":"response_item",` +
			`"payload":{"type":"custom_tool_call","call_id":"failed-exec","name":"exec",` +
			`"input":"const patch = \"*** Begin Patch\\n*** Update File: failed-exec.go` +
			`\\n+one\\n*** End Patch\"; tools.apply_patch(patch);"}}`,
		`{"timestamp":"2026-06-20T12:00:06Z","type":"response_item",` +
			`"payload":{"type":"custom_tool_call_output","call_id":"failed-exec",` +
			`"output":[{"type":"input_text","text":"Failed to apply patch"}]}}`,
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	got, err := (ai.Codex{After: time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, filepath.Join("/workspace/project", "successful.go"), got[0].Entity)
	assert.Equal(t, float64(time.Date(2026, 6, 20, 12, 0, 4, 0, time.UTC).Unix()), got[0].Time)
}

func TestCodexParse_UserAgentUsesSourceFromRealSessionMetaLines(t *testing.T) {
	tests := map[string]struct {
		SessionMeta       string
		TaskStarted       string
		FallbackUserAgent string
		ExpectedUserAgent string
	}{
		"vscode": {
			SessionMeta: strings.Join([]string{
				`{"timestamp":"2026-05-26T16:50:38.014Z","type":"session_meta",`,
				`"payload":{"id":"019e6532-026f-71c1-9e09-edd195cd23c8",`,
				`"timestamp":"2026-05-26T16:50:36.783Z","cwd":"/Users/user/git/wakatime",`,
				`"originator":"codex_vscode","cli_version":"0.131.0-alpha.9","source":"vscode",`,
				`"thread_source":"user","model_provider":"openai","base_instructions":{},`,
				`"git":{"commit_hash":"7c953e3f1f4cbb768da2f23f140a2655663600b5",`,
				`"branch":"master","repository_url":"git@github.com:wakatime/wakatime.git"}}}`,
			}, ""),
			TaskStarted: strings.Join([]string{
				`{"timestamp":"2026-05-26T16:50:38.015Z","type":"event_msg",`,
				`"payload":{"type":"task_started","turn_id":"019e6532-02cd-7a72-bdfd-730d72df8689",`,
				`"started_at":1779814236,"model_context_window":258400,`,
				`"collaboration_mode_kind":"default"}}`,
			}, ""),
			FallbackUserAgent: "zoom.us/6.7.7(76486)-6.7.7.76486 macos-wakatime/5.28.4",
			ExpectedUserAgent: "vscode-wakatime/unknown " +
				"zoom.us/6.7.7(76486)-6.7.7.76486 macos-wakatime/5.28.4",
		},
		"cli": {
			SessionMeta: strings.Join([]string{
				`{"timestamp":"2026-05-27T20:05:36.264Z","type":"session_meta",`,
				`"payload":{"id":"019e6b07-17ac-7070-bde0-864ecbda2cac",`,
				`"timestamp":"2026-05-27T20:01:27.478Z","cwd":"/Users/user/git/wakatime",`,
				`"originator":"codex-tui","cli_version":"0.134.0","source":"cli",`,
				`"thread_source":"user","model_provider":"openai","base_instructions":{},`,
				`"git":{"commit_hash":"f158dd9429b54fbd2ec6a70fca7f705bcc7915db",`,
				`"branch":"master","repository_url":"git@github.com:wakatime/wakatime.git"}}}`,
			}, ""),
			TaskStarted: strings.Join([]string{
				`{"timestamp":"2026-05-27T20:05:36.265Z","type":"event_msg",`,
				`"payload":{"type":"task_started","turn_id":"019e6b0a-e353-7963-953d-dc34e114af8e",`,
				`"started_at":1779912336,"model_context_window":258400,`,
				`"collaboration_mode_kind":"default"}}`,
			}, ""),
			FallbackUserAgent: "claude-code/2.1.142 claude-code-wakatime/3.1.6",
			ExpectedUserAgent: "codex-cli/0.134.0 " +
				"claude-code/2.1.142 claude-code-wakatime/3.1.6",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)

			transcriptDir := filepath.Join(home, ".codex", "sessions", "2026", "05", "27")
			require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

			transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
			transcript := strings.Join([]string{
				tt.SessionMeta,
				tt.TaskStarted,
				strings.Join([]string{
					`{"timestamp":"2026-05-27T21:04:07.000Z","type":"event_msg",`,
					`"payload":{"type":"agent_message","message":"I will make the change."}}`,
				}, ""),
			}, "\n") + "\n"
			require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

			parser := ai.Codex{
				After:             time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC),
				FallbackUserAgent: tt.FallbackUserAgent,
			}

			got, err := parser.Parse(ctx)
			require.NoError(t, err)
			require.Len(t, got, 1)
			assert.Equal(t, tt.ExpectedUserAgent, got[0].UserAgent)
		})
	}
}

func TestCodexParse_RetainsCwdFromSkippedSessionMetaLine(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	now := time.Now()
	transcriptDir := filepath.Join(home, ".codex", "sessions", now.Format("2006"), now.Format("01"), now.Format("02"))
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-03-28T11:00:00Z","type":"session_meta",`,
			`"payload":{"cwd":"/workspace/project","cli_version":"0.116.0-alpha.10"}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-03-28T11:33:18Z","type":"message",`,
			`"payload":{"type":"message","role":"assistant",`,
			`"content":[{"type":"output_text","text":"I am on it"}]}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Codex{
		After: time.Date(2026, 3, 28, 11, 30, 0, 0, time.UTC),
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "/workspace/project", got[0].ProjectPathOverride)
}

func TestCodexParse_SkipsHarnessInputWhenCalculatingPromptLength(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	now := time.Now()
	transcriptDir := filepath.Join(home, ".codex", "sessions", now.Format("2006"), now.Format("01"), now.Format("02"))
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-04-15T04:14:05Z","type":"session_meta",`,
			`"payload":{"cwd":"/workspace/project","cli_version":"0.119.0-alpha.28"}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-04-15T04:14:05Z","type":"response_item",`,
			`"payload":{"type":"message","role":"user","content":[`,
			`{"type":"input_text","text":"<environment_context>\n  <cwd>/workspace/project</cwd>\n</environment_context>"}`,
			`]}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-04-15T04:14:06Z","type":"response_item",`,
			`"payload":{"type":"message","role":"user","content":[`,
			`{"type":"input_text","text":"Add a unit test for this line"}`,
			`]}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Codex{
		After: time.Date(2026, 4, 15, 4, 0, 0, 0, time.UTC),
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, len([]rune("Add a unit test for this line")), got[0].AIPromptLength)
	assert.Equal(t, "/workspace/project", got[0].ProjectPathOverride)
}

func TestCodexParse_StripsBundledHarnessPrefixBeforeCountingPromptLength(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	now := time.Now()
	transcriptDir := filepath.Join(home, ".codex", "sessions", now.Format("2006"), now.Format("01"), now.Format("02"))
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	userPrompt := strings.Join([]string{
		"<environment_context>",
		"  <cwd>/Users/user/git/vim-wakatime</cwd>",
		"</environment_context>",
		"# Context from my IDE setup:",
		"",
		"## Active file: lua/wakatime/init.lua",
		"",
		"## Open tabs:",
		"- init.lua: lua/wakatime/init.lua",
		"",
		"## My request for Codex:",
		"look for any divergences in the new Lua plugin vs the old VimL plugin, " +
			"or any bugs in the Lua plugin since it's not as tested as the older plugin.",
		"",
	}, "\n")
	expectedPrompt := "look for any divergences in the new Lua plugin vs the old VimL plugin, " +
		"or any bugs in the Lua plugin since it's not as tested as the older plugin."
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-04-21T10:57:40Z","type":"session_meta",`,
			`"payload":{"cwd":"/Users/user/git/vim-wakatime","cli_version":"0.120.0-alpha.1"}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-04-21T10:57:40Z","type":"response_item",`,
			`"payload":{"type":"message","role":"developer","content":[`,
			`{"type":"input_text","text":"<permissions instructions>\nblocked\n</permissions instructions>"}`,
			`]}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-04-21T10:57:40Z","type":"response_item",`,
			`"payload":{"type":"message","role":"user","content":[`,
			`{"type":"input_text","text":` + jsonString(userPrompt) + `}`,
			`]}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Codex{
		After: time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC),
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, len([]rune(expectedPrompt)), got[0].AIPromptLength)
	assert.Equal(t, "/Users/user/git/vim-wakatime", got[0].ProjectPathOverride)
}

func TestCodexParse_StripsVSCodePrefixFromUserMessage(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	now := time.Now()
	transcriptDir := filepath.Join(home, ".codex", "sessions", now.Format("2006"), now.Format("01"), now.Format("02"))
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2026-04-16T00:02:25Z","type":"session_meta",`,
			`"payload":{"cwd":"/workspace/project","cli_version":"0.119.0-alpha.28"}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2026-04-16T00:02:25Z","type":"event_msg",`,
			`"payload":{"type":"user_message","message":"# Context from my IDE setup:\n\n` +
				`## Active file: wakatime-cli/static/css/index.less\n\n## Open tabs:\n` +
				`- index.less: wakatime-cli/static/css/index.less\n\n## My request for Codex:\n` +
				`remove scrolling horizontally\n"}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Codex{
		After: time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC),
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, len([]rune("remove scrolling horizontally")), got[0].AIPromptLength)
	assert.Equal(t, "/workspace/project", got[0].ProjectPathOverride)
}

func TestCodexParse_ParsesLegacyAgentMessage(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".codex", "sessions", "2025", "11", "17")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "rollout-2025-11-17T08-27-28-019a91c8-d06a-7a80-9417-51ed9fef0508.jsonl")
	transcript := strings.Join([]string{
		strings.Join([]string{
			`{"timestamp":"2025-11-17T12:27:28.000Z","type":"session_meta",`,
			`"payload":{"id":"019a91c8-d06a-7a80-9417-51ed9fef0508",`,
			`"cwd":"/workspace/project","cli_version":"0.50.0"}}`,
		}, ""),
		strings.Join([]string{
			`{"timestamp":"2025-11-17T12:27:30.000Z","type":"event_msg",`,
			`"payload":{"type":"agent_message","message":"I will inspect the code and make the change."}}`,
		}, ""),
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	parser := ai.Codex{
		After:             time.Date(2025, 2, 24, 0, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "Codex rollout-2025-11-17T08-27-28-019a91c8-d06a-7a80-9417-51ed9fef0508", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "019a91c8-d06a-7a80-9417-51ed9fef0508", got[0].AISession)
	assert.Equal(t, heartbeat.AICodingCategory.String(), got[0].Category)
	assert.Nil(t, got[0].AILineChanges)
	assert.Zero(t, got[0].AIPromptLength)
	assert.Equal(t, "/workspace/project", got[0].ProjectPathOverride)
	require.NotNil(t, got[0].IsWrite)
	assert.False(t, *got[0].IsWrite)
	assert.Equal(t, float64(time.Date(2025, 11, 17, 12, 27, 30, 0, time.UTC).Unix()), got[0].Time)
	assert.NotContains(t, got[0].UserAgent, "Codex/")
	assert.Contains(t, got[0].UserAgent, "plugin/0.0.1")
}

func TestCodexParse_RolloutFixtureIncludesExpectedHeartbeatAttributes(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	now := time.Now()
	transcriptDir := filepath.Join(home, ".codex", "sessions", now.Format("2026"), now.Format("04"), now.Format("15"))
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcript := "rollout-2026-04-15T18-46-36-019d9353-333c-7c41-a909-26f5c6221a5a.jsonl"
	transcriptPath := filepath.Join(transcriptDir, transcript)
	copyFile(t, filepath.Join("testdata", transcript), transcriptPath)

	parser := ai.Codex{
		After:             time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}

	heartbeats, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, heartbeats, 9)

	for _, h := range heartbeats {
		assert.Equal(t, "plus", h.AISubscriptionPlan)
	}

	i := 0
	assert.EqualValues(t, 1776297745, heartbeats[i].Time)
	assert.Equal(t, heartbeat.AppType, heartbeats[i].EntityType)
	assert.Equal(t, "Codex rollout-2026-04-15T18-46-36-019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.False(t, *heartbeats[i].IsWrite)
	assert.EqualValues(t, 1, heartbeats[i].AIInputTokens)
	assert.EqualValues(t, 2, heartbeats[i].AIOutputTokens)
	assert.Equal(t, 29, heartbeats[i].AIPromptLength)
	assert.Nil(t, heartbeats[i].AILineChanges)
	assert.Equal(t, "/Users/user/git/wakatime-cli", heartbeats[i].ProjectPathOverride)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)

	// 1
	i++
	assert.EqualValues(t, 1776297755, heartbeats[i].Time)
	assert.Equal(t, heartbeat.AppType, heartbeats[i].EntityType)
	assert.Equal(t, "Codex rollout-2026-04-15T18-46-36-019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.False(t, *heartbeats[i].IsWrite)
	assert.EqualValues(t, 105133, heartbeats[i].AIInputTokens)
	assert.EqualValues(t, 477, heartbeats[i].AIOutputTokens)
	assert.Zero(t, heartbeats[i].AIPromptLength)
	assert.Nil(t, heartbeats[i].AILineChanges)
	assert.Equal(t, "/Users/user/git/wakatime-cli", heartbeats[i].ProjectPathOverride)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)

	// 2
	i++
	assert.EqualValues(t, 1776297759, heartbeats[i].Time)
	assert.Equal(t, heartbeat.AppType, heartbeats[i].EntityType)
	assert.Equal(t, "Codex rollout-2026-04-15T18-46-36-019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.False(t, *heartbeats[i].IsWrite)
	assert.EqualValues(t, 217197, heartbeats[i].AIInputTokens)
	assert.EqualValues(t, 599, heartbeats[i].AIOutputTokens)
	assert.Zero(t, heartbeats[i].AIPromptLength)
	assert.Nil(t, heartbeats[i].AILineChanges)
	assert.Equal(t, "/Users/user/git/wakatime-cli", heartbeats[i].ProjectPathOverride)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)

	// 3
	i++
	assert.EqualValues(t, 1776297781, heartbeats[i].Time)
	assert.Equal(t, heartbeat.AppType, heartbeats[i].EntityType)
	assert.Equal(t, "Codex rollout-2026-04-15T18-46-36-019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.False(t, *heartbeats[i].IsWrite)
	assert.EqualValues(t, 109309, heartbeats[i].AIInputTokens)
	assert.EqualValues(t, 843, heartbeats[i].AIOutputTokens)
	assert.Zero(t, heartbeats[i].AIPromptLength)
	assert.Nil(t, heartbeats[i].AILineChanges)
	assert.Equal(t, "/Users/user/git/wakatime-cli", heartbeats[i].ProjectPathOverride)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)

	// 4
	i++
	assert.EqualValues(t, 1776297784, heartbeats[i].Time)
	assert.Equal(t, heartbeat.FileType, heartbeats[i].EntityType)
	assert.Equal(t, "/Users/user/git/wakatime-cli/templates/index.html", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.True(t, *heartbeats[i].IsWrite)
	assert.Zero(t, heartbeats[i].AIInputTokens)
	assert.Zero(t, heartbeats[i].AIOutputTokens)
	assert.Zero(t, heartbeats[i].AIPromptLength)
	assert.EqualValues(t, -1, *heartbeats[i].AILineChanges)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)

	// 5
	i++
	assert.EqualValues(t, 1776297787, heartbeats[i].Time)
	assert.Equal(t, heartbeat.AppType, heartbeats[i].EntityType)
	assert.Equal(t, "Codex rollout-2026-04-15T18-46-36-019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.False(t, *heartbeats[i].IsWrite)
	assert.Zero(t, heartbeats[i].AIInputTokens)
	assert.Zero(t, heartbeats[i].AIOutputTokens)
	assert.Zero(t, heartbeats[i].AIPromptLength)
	assert.Nil(t, heartbeats[i].AILineChanges)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)

	// 6
	i++
	assert.EqualValues(t, 1776297787, heartbeats[i].Time)
	assert.Equal(t, heartbeat.AppType, heartbeats[i].EntityType)
	assert.Equal(t, "Codex rollout-2026-04-15T18-46-36-019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.False(t, *heartbeats[i].IsWrite)
	assert.EqualValues(t, 110200, heartbeats[i].AIInputTokens)
	assert.EqualValues(t, 223, heartbeats[i].AIOutputTokens)
	assert.Equal(t, 10, heartbeats[i].AIPromptLength)
	assert.Nil(t, heartbeats[i].AILineChanges)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)

	// 7
	i++
	assert.EqualValues(t, 1776297789, heartbeats[i].Time)
	assert.Equal(t, heartbeat.FileType, heartbeats[i].EntityType)
	assert.Equal(t, "/Users/user/git/wakatime-cli/static/css/index.less", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.True(t, *heartbeats[i].IsWrite)
	assert.Zero(t, heartbeats[i].AIInputTokens)
	assert.Zero(t, heartbeats[i].AIOutputTokens)
	assert.Zero(t, heartbeats[i].AIPromptLength)
	assert.Equal(t, 23, *heartbeats[i].AILineChanges)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)

	// 8
	i++
	assert.EqualValues(t, 1776297792, heartbeats[i].Time)
	assert.Equal(t, heartbeat.AppType, heartbeats[i].EntityType)
	assert.Equal(t, "Codex rollout-2026-04-15T18-46-36-019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].Entity)
	assert.Equal(t, "ai coding", heartbeats[i].Category)
	assert.False(t, *heartbeats[i].IsWrite)
	assert.EqualValues(t, 110472, heartbeats[i].AIInputTokens)
	assert.EqualValues(t, 156, heartbeats[i].AIOutputTokens)
	assert.Zero(t, heartbeats[i].AIPromptLength)
	assert.Nil(t, heartbeats[i].AILineChanges)
	assert.Equal(t, "/Users/user/git/wakatime-cli", heartbeats[i].ProjectPathOverride)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", heartbeats[i].AISession)
}

func copyFile(t *testing.T, source, destination string) {
	input, err := os.ReadFile(source)
	require.NoError(t, err)

	err = os.WriteFile(destination, input, 0600)
	require.NoError(t, err)
}

func jsonString(value string) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	return string(encoded)
}
