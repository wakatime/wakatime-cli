package ai

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestOpenClaudeParsePreservesStandaloneCaller(t *testing.T) {
	const native = "openclaude/2.1.45 openclaude-wakatime/3"

	tests := map[string]struct {
		Fallback   string
		UserAgents map[string]string
		Expected   string
	}{
		"native caller without file context": {
			Fallback: native,
			Expected: "claude/4.6 " + native,
		},
		"native caller with unrelated file context": {
			Fallback:   native,
			UserAgents: map[string]string{"/tmp/main.go": "Cursor/2 macos-wakatime/3"},
			Expected:   "claude/4.6 " + native,
		},
		"native caller with CLI header": {
			Fallback: "wakatime/13.0.7 (Windows NT-10.0-amd64) go1.26 " + native,
			Expected: "wakatime/13.0.7 (Windows NT-10.0-amd64) go1.26 claude/4.6 " + native,
		},
		"native caller preserves plugin spelling": {
			Fallback: "OpenClaude/2.1.45 OpenClaude-WakaTime/3",
			Expected: "claude/4.6 openclaude/2.1.45 OpenClaude-WakaTime/3",
		},
		"native caller with mixed case and CLI header": {
			Fallback: "wakatime/13.0.7 (Windows NT-10.0-amd64) go1.26 OPENCLAUDE/2.1.45 openclaude-wakatime/3",
			Expected: "wakatime/13.0.7 (Windows NT-10.0-amd64) go1.26 claude/4.6 " + native,
		},
		"foreground Cursor": {
			Fallback: "Cursor/2 macos-wakatime/3",
			Expected: "claude/4.6 openclaude/2.1.45",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)

			root := filepath.Join(home, ".openclaude", "projects", "project")
			require.NoError(t, os.MkdirAll(root, 0o755))

			transcript := `{"timestamp":"2026-10-16T12:00:00Z","sessionId":"standalone","version":"2.1.45",` +
				`"type":"user","message":{"role":"user","model":"claude-4.6","content":"please continue"}}`
			require.NoError(t, os.WriteFile(filepath.Join(root, "standalone.jsonl"), []byte(transcript+"\n"), 0o600))

			got, err := (OpenClaude{FallbackUserAgent: tt.Fallback, UserAgents: tt.UserAgents}).Parse(t.Context())
			require.NoError(t, err)
			require.Len(t, got, 1)
			assert.Equal(t, heartbeat.AppType, got[0].EntityType)
			assert.Equal(t, "openclaude:standalone", got[0].AISession)
			assert.Equal(t, tt.Expected, got[0].UserAgent)
		})
	}
}

func TestOpenClaudeParseKeepsUsageAndActivityAttributionConsistent(t *testing.T) {
	tests := map[string]struct {
		Fallback   string
		UserAgents map[string]string
		Prompt     string
		Expected   string
	}{
		"standalone foreground Cursor": {
			Fallback: "Cursor/2 macos-wakatime/3",
			Prompt:   "please continue",
			Expected: "claude/4.6 openclaude/2.1.45",
		},
		"standalone native caller": {
			Fallback:   "openclaude/2.1.45 openclaude-wakatime/3",
			UserAgents: map[string]string{"/tmp/main.go": "Cursor/2 macos-wakatime/3"},
			Prompt:     "please continue",
			Expected:   "claude/4.6 openclaude/2.1.45 openclaude-wakatime/3",
		},
		"IDE context before usage record": {
			Fallback: "Cursor/2 macos-wakatime/3",
			Prompt:   "<ide_opened_file><path>/tmp/main.go</path></ide_opened_file>please continue",
			Expected: "claude/4.6 Cursor/2 openclaude/2.1.45 macos-wakatime/3",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)

			root := filepath.Join(home, ".openclaude", "projects", "project")
			require.NoError(t, os.MkdirAll(root, 0o755))

			transcript := `{"timestamp":"2026-10-16T12:00:00Z","sessionId":"usage","version":"2.1.45",` +
				`"type":"user","message":{"role":"user","model":"claude-4.6","content":"` + tt.Prompt + `"}}` + "\n" +
				`{"timestamp":"2026-10-16T12:01:00Z","sessionId":"usage","version":"2.1.45",` +
				`"type":"assistant","message":{"role":"assistant","model":"claude-4.6",` +
				`"usage":{"input_tokens":3,"output_tokens":5},"content":"done"}}` + "\n"
			require.NoError(t, os.WriteFile(filepath.Join(root, "usage.jsonl"), []byte(transcript), 0o600))

			got, err := (OpenClaude{FallbackUserAgent: tt.Fallback, UserAgents: tt.UserAgents}).Parse(t.Context())
			require.NoError(t, err)
			require.Len(t, got, 2)
			assert.Positive(t, got[0].AIPromptLength)
			assert.Equal(t, int64(3), got[1].AIInputTokens)
			assert.Equal(t, int64(5), got[1].AIOutputTokens)

			for _, h := range got {
				assert.Equal(t, tt.Expected, h.UserAgent)
			}
		})
	}
}
