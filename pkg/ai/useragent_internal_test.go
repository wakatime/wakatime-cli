package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestAIUserAgentWithModelAndEditorPreservesConfiguredEditor(t *testing.T) {
	tests := map[string]struct {
		UserAgent string
		Expected  string
	}{
		"unknown editor": {
			UserAgent: "future-editor/2 plugin/3",
			Expected:  "gpt/5.6 future-editor/2 codex-vscode/1.2.3 plugin/3",
		},
		"desktop Cursor": {
			UserAgent: "Cursor/2 macos-wakatime/3",
			Expected:  "gpt/5.6 Cursor/2 codex-vscode/1.2.3 macos-wakatime/3",
		},
		"unrelated desktop app": {
			UserAgent: "zoom.us/6 macos-wakatime/3",
			Expected:  "gpt/5.6 codex-vscode/1.2.3 zoom.us/6 macos-wakatime/3",
		},
		"antigravity IDE": {
			UserAgent: "antigravity-ide/2 vscode-wakatime/3",
			Expected:  "gpt/5.6 antigravity-ide/2 codex-vscode/1.2.3 vscode-wakatime/3",
		},
		"nested OS description": {
			UserAgent: "wakatime/13.0.7 (Linux (Ubuntu)-6-amd64) go1.26 future-editor/2 plugin/3",
			Expected: "wakatime/13.0.7 (Linux (Ubuntu)-6-amd64) go1.26 gpt/5.6 " +
				"future-editor/2 codex-vscode/1.2.3 plugin/3",
		},
		"full user agent": {
			UserAgent: "wakatime/13.0.7 (Windows NT-10.0-amd64) go1.26 future-editor/2 plugin/3",
			Expected: "wakatime/13.0.7 (Windows NT-10.0-amd64) go1.26 gpt/5.6 " +
				"future-editor/2 codex-vscode/1.2.3 plugin/3",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			const entity = "/tmp/main.go"

			userAgents := map[string]string{entity: tt.UserAgent}
			got := aiUserAgentWithModelAndEditor(
				entity, userAgents, tt.UserAgent, "gpt-5.6", "", "codex-vscode/1.2.3",
			)
			assert.Equal(t, tt.Expected, got)

			userAgents[entity] = got
			assert.Equal(t, tt.Expected, aiUserAgentWithModelAndEditor(
				entity, userAgents, tt.UserAgent, "gpt-5.6", "", "codex-vscode/1.2.3",
			))
		})
	}
}

func TestAIUserAgentCanonicalizesParserWhitespace(t *testing.T) {
	tests := map[string]struct {
		Parser   string
		Expected string
	}{
		"tab": {
			Parser:   "Cursor\tAgent",
			Expected: "cursor-agent",
		},
		"newline": {
			Parser:   "Grok\nBot",
			Expected: "grok-bot",
		},
		"versioned parser": {
			Parser:   "Grok\tBuild/1.2.3",
			Expected: "grok-build/1.2.3",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := aiUserAgent("/tmp/main.go", nil, "", tt.Parser)
			assert.Equal(t, tt.Expected, got)
			assert.Equal(t, tt.Expected, aiUserAgent("/tmp/main.go", nil, got, tt.Parser))
		})
	}
}

func TestAIUserAgentReordersExistingAgentProducts(t *testing.T) {
	assert.Equal(t,
		"cursor/2 codex-vscode/1 plugin/3",
		userAgentWithAgentProduct("codex-vscode/1 cursor/2 plugin/3", "codex-vscode/2", false),
	)
	assert.Equal(t,
		"cursor/2 github-copilot-cli/1 copilot/2 plugin/3",
		userAgentWithAgentProduct("cursor/2 plugin/3 github-copilot-cli/1", "github-copilot-cli/2 copilot/2", false),
	)
	assert.Equal(t, "gpt/5.6 cursor/2 plugin/3",
		userAgentWithPrependedAgent("cursor/2 gpt/5.6 plugin/3 gpt/5.6", "gpt/5.6"),
	)
}

func TestAIUserAgentRepeatedParserComposition(t *testing.T) {
	const (
		entity = "/tmp/main.go"
		plugin = "cursor/2 plugin/3"
	)

	copilot := Copilot{FallbackUserAgent: plugin}
	copilotAgent := copilot.cliUserAgent(entity, "gpt-5.6", "1", "2")
	assert.Equal(t, "gpt/5.6 cursor/2 github-copilot-cli/1 copilot/2 plugin/3", copilotAgent)
	copilot.FallbackUserAgent = copilotAgent
	assert.Equal(t, copilotAgent, copilot.cliUserAgent(entity, "gpt-5.6", "1", "2"))

	cursor := Cursor{FallbackUserAgent: plugin}
	cursorAgent := cursor.userAgent(entity, "grok-4.7")
	assert.Equal(t, "grok/4.7 cursor/2 plugin/3", cursorAgent)
	cursor.FallbackUserAgent = cursorAgent
	assert.Equal(t, cursorAgent, cursor.userAgent(entity, "grok-4.7"))

	event := genericAIEvent{model: "grok-4.7", filePath: entity}
	config := ParserConfig{FallbackUserAgent: plugin}
	generic := genericAIEventHeartbeats(CursorAgent{}, config, event)
	assert.Equal(t, "grok/4.7 cursor/2 cursor-agent plugin/3", generic[1].UserAgent)
	config.FallbackUserAgent = generic[1].UserAgent
	assert.Equal(t, generic[1].UserAgent, genericAIEventHeartbeats(CursorAgent{}, config, event)[1].UserAgent)
}

func TestAIUserAgentUsesFileContextForFallbackEditor(t *testing.T) {
	const plugin = "configured-editor/2 editor-wakatime/3"

	userAgents := entityUserAgents([]heartbeat.Heartbeat{
		{Entity: "/tmp/current.go", UserAgent: plugin, EntityType: heartbeat.FileType},
		{Entity: "example.com", UserAgent: "browser/2 browser-wakatime/3", EntityType: heartbeat.DomainType},
		{Entity: "unrelated-app", UserAgent: "app/2 macos-wakatime/3", EntityType: heartbeat.AppType},
	})
	assert.Equal(t, map[string]string{"/tmp/current.go": plugin}, userAgents)
	assert.Equal(t, "gpt/5.6 configured-editor/2 codex-vscode/1 editor-wakatime/3",
		aiUserAgentWithModelAndEditor("/tmp/other.go", userAgents, plugin, "gpt-5.6", "", "codex-vscode/1"),
	)
	assert.Equal(t, "gpt/5.6 codex-vscode/1 browser/2 browser-wakatime/3",
		aiUserAgentWithModelAndEditor("/tmp/other.go", userAgents,
			"browser/2 browser-wakatime/3", "gpt-5.6", "", "codex-vscode/1"),
	)
	assert.Equal(t, "gpt/5.6 codex-cli/1 claude-code/2 claude-code-wakatime/3",
		aiUserAgentWithModelAndEditor("/tmp/current.go", map[string]string{
			"/tmp/current.go": "claude-code/2 claude-code-wakatime/3",
		}, "", "gpt-5.6", "", "codex-cli/1"),
	)
}
