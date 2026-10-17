package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestClaudeStandaloneUserAgentIgnoresForegroundEditor(t *testing.T) {
	const plugin = "Cursor/2 macos-wakatime/3"

	appUserAgents := entityUserAgents([]heartbeat.Heartbeat{
		{Entity: "Cursor", UserAgent: plugin, EntityType: heartbeat.AppType},
	})
	require.Empty(t, appUserAgents)

	tests := map[string]struct {
		UserAgents map[string]string
		Fallback   string
		Version    string
		Model      string
		Expected   string
	}{
		"filtered foreground app": {
			UserAgents: appUserAgents,
			Fallback:   plugin,
			Version:    "2.1.45",
			Model:      "claude-4.6",
			Expected:   "claude/4.6 claude-code/2.1.45",
		},
		"nil file context": {
			Fallback: plugin,
			Version:  "2.1.45",
			Model:    "claude-4.6",
			Expected: "claude/4.6 claude-code/2.1.45",
		},
		"empty file context": {
			UserAgents: map[string]string{},
			Fallback:   plugin,
			Version:    "2.1.45",
			Model:      "claude-4.6",
			Expected:   "claude/4.6 claude-code/2.1.45",
		},
		"matching file context": {
			UserAgents: map[string]string{"/tmp/main.go": plugin},
			Fallback:   plugin,
			Version:    "2.1.45",
			Model:      "claude-4.6",
			Expected:   "claude/4.6 claude-code/2.1.45",
		},
		"unrelated file context": {
			UserAgents: map[string]string{"/tmp/other.go": plugin},
			Fallback:   plugin,
			Version:    "2.1.45",
			Model:      "claude-4.6",
			Expected:   "claude/4.6 claude-code/2.1.45",
		},
		"unknown version": {
			Fallback: plugin,
			Model:    "claude-4.6",
			Expected: "claude/4.6 claude-code/unknown",
		},
		"no model": {
			Fallback: plugin,
			Version:  "2.1.45",
			Expected: "claude-code/2.1.45",
		},
		"unknown version and no model": {
			Fallback: plugin,
			Expected: "claude-code/unknown",
		},
		"full CLI header": {
			Fallback: "wakatime/13.0.7 (Windows NT-10.0-amd64) go1.26 " + plugin,
			Version:  "2.1.45",
			Model:    "claude-4.6",
			Expected: "claude/4.6 claude-code/2.1.45",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			parser := Claude(ParserConfig{UserAgents: tt.UserAgents, FallbackUserAgent: tt.Fallback})
			for _, entity := range []string{"claude-code", "/tmp/main.go"} {
				assert.Equal(t, tt.Expected, parser.userAgent(entity, tt.Version, tt.Model, "", false), entity)
			}
		})
	}
}

func TestClaudeIDESessionUserAgentPreservesEditor(t *testing.T) {
	tests := map[string]struct {
		UserAgents map[string]string
		Fallback   string
		Entity     string
		Expected   string
	}{
		"Cursor fallback": {
			Fallback: "Cursor/2 macos-wakatime/3",
			Entity:   "claude-code",
			Expected: "claude/4.6 Cursor/2 claude-code/2.1.45 macos-wakatime/3",
		},
		"Cursor file context": {
			UserAgents: map[string]string{"/tmp/main.go": "Cursor/2 macos-wakatime/3"},
			Entity:     "/tmp/main.go",
			Expected:   "claude/4.6 Cursor/2 claude-code/2.1.45 macos-wakatime/3",
		},
		"configured editor": {
			UserAgents: map[string]string{"/tmp/main.go": "configured-editor/2 editor-wakatime/3"},
			Entity:     "/tmp/main.go",
			Expected:   "claude/4.6 configured-editor/2 claude-code/2.1.45 editor-wakatime/3",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			parser := Claude(ParserConfig{UserAgents: tt.UserAgents, FallbackUserAgent: tt.Fallback})
			assert.Equal(t, tt.Expected, parser.userAgent(tt.Entity, "2.1.45", "claude-4.6", "", true))
		})
	}
}

func TestClaudeStandaloneUserAgentPreservesNativeHookMetadata(t *testing.T) {
	const plugin = "claude-code/2.1.45 claude-code-wakatime/3"

	tests := map[string]map[string]string{
		"nil file context":       nil,
		"empty file context":     {},
		"matching file context":  {"/tmp/main.go": "Cursor/2 macos-wakatime/3"},
		"unrelated file context": {"/tmp/other.go": "Cursor/2 macos-wakatime/3"},
	}

	for name, userAgents := range tests {
		t.Run(name, func(t *testing.T) {
			parser := Claude(ParserConfig{UserAgents: userAgents, FallbackUserAgent: plugin})
			for _, entity := range []string{"claude-code", "/tmp/main.go"} {
				assert.Equal(t, "claude/4.6 "+plugin,
					parser.userAgent(entity, "2.1.45", "claude-4.6", "", false), entity,
				)
			}
		})
	}
}
