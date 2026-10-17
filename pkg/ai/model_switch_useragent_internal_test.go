package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAIUserAgentWithModelAndEditorReplacesPreviousModel(t *testing.T) {
	tests := map[string]struct {
		UserAgent string
		Fallback  string
		Model     string
		Editor    string
		Expected  string
	}{
		"file editor": {
			UserAgent: "gpt/5.4 cursor/2 codex-vscode/1 plugin/3",
			Model:     "claude-4.6",
			Editor:    "codex-vscode/1",
			Expected:  "claude/4.6 cursor/2 codex-vscode/1 plugin/3",
		},
		"CLI header with spaces": {
			UserAgent: "wakatime/13.0.7 (Windows NT-10.0-amd64) go1.26 " +
				"gpt/5.4 cursor/2 codex-vscode/1 plugin/3",
			Model:  "claude-4.6",
			Editor: "codex-vscode/1",
			Expected: "wakatime/13.0.7 (Windows NT-10.0-amd64) go1.26 " +
				"claude/4.6 cursor/2 codex-vscode/1 plugin/3",
		},
		"configured editor": {
			UserAgent: "gpt/5.4 future-editor/2 codex-vscode/1 plugin/3",
			Fallback:  "future-editor/2 plugin/3",
			Model:     "claude-4.6",
			Editor:    "codex-vscode/1",
			Expected:  "claude/4.6 future-editor/2 codex-vscode/1 plugin/3",
		},
		"same model family": {
			UserAgent: "gpt/5.4 cursor/2 codex-vscode/1 plugin/3",
			Model:     "gpt-5.5",
			Editor:    "codex-vscode/1",
			Expected:  "gpt/5.5 cursor/2 codex-vscode/1 plugin/3",
		},
		"Claude to GPT": {
			UserAgent: "claude/4.6 cursor/2 codex-vscode/1 plugin/3",
			Fallback:  "cursor/2 plugin/3",
			Model:     "gpt-5.5",
			Editor:    "codex-vscode/1",
			Expected:  "gpt/5.5 cursor/2 codex-vscode/1 plugin/3",
		},
		"Gemini to GPT": {
			UserAgent: "gemini/3.1-pro cursor/2 codex-vscode/1 plugin/3",
			Fallback:  "cursor/2 plugin/3",
			Model:     "gpt-5.5",
			Editor:    "codex-vscode/1",
			Expected:  "gpt/5.5 cursor/2 codex-vscode/1 plugin/3",
		},
		"unknown model family": {
			UserAgent: "future-model/8.1 cursor/2 codex-vscode/1 plugin/3",
			Model:     "claude-4.6",
			Editor:    "codex-vscode/1",
			Expected:  "claude/4.6 cursor/2 codex-vscode/1 plugin/3",
		},
		"standalone agent": {
			UserAgent: "gpt/5.4 codex-cli/1 plugin/3",
			Fallback:  "codex-cli/1 plugin/3",
			Model:     "claude-4.6",
			Editor:    "codex-cli/1",
			Expected:  "claude/4.6 codex-cli/1 plugin/3",
		},
		"configured editor without previous model": {
			UserAgent: "future-editor/2 codex-vscode/1 plugin/3",
			Model:     "claude-4.6",
			Editor:    "codex-vscode/1",
			Expected:  "claude/4.6 future-editor/2 codex-vscode/1 plugin/3",
		},
		"configured editor before known IDE": {
			UserAgent: "future-editor/2 cursor/2 plugin/3",
			Fallback:  "future-editor/2 cursor/2 plugin/3",
			Model:     "gpt-5.5",
			Editor:    "codex-vscode/1",
			Expected:  "gpt/5.5 future-editor/2 codex-vscode/1 cursor/2 plugin/3",
		},
		"configured editor with auxiliary plugin": {
			UserAgent: "future-editor/2 auxiliary/1 codex-vscode/1 plugin/3",
			Model:     "claude-4.6",
			Editor:    "codex-vscode/1",
			Expected:  "claude/4.6 future-editor/2 codex-vscode/1 auxiliary/1 plugin/3",
		},
		"missing new model": {
			UserAgent: "gpt/5.4 cursor/2 codex-vscode/1 plugin/3",
			Editor:    "codex-vscode/1",
			Expected:  "gpt/5.4 cursor/2 codex-vscode/1 plugin/3",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			const entity = "/tmp/main.go"

			userAgents := map[string]string{entity: tt.UserAgent}
			got := aiUserAgentWithModelAndEditor(entity, userAgents, tt.Fallback, tt.Model, "", tt.Editor)
			assert.Equal(t, tt.Expected, got)

			userAgents[entity] = got
			assert.Equal(t, tt.Expected,
				aiUserAgentWithModelAndEditor(entity, userAgents, tt.Fallback, tt.Model, "", tt.Editor),
			)
		})
	}
}

func TestAIUserAgentWithModelAndEditorPreservesFallbackProducts(t *testing.T) {
	const (
		entity   = "/tmp/main.go"
		fallback = "future-editor/2 cursor/2 plugin/3"
		expected = "gpt/5.5 cursor/2 codex-vscode/1 future-editor/2 plugin/3"
	)

	tests := map[string]map[string]string{
		"nil map":      nil,
		"missing file": {"/tmp/other.go": "other-editor/4 plugin/3"},
	}

	for name, userAgents := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, expected,
				aiUserAgentWithModelAndEditor(entity, userAgents, fallback, "gpt-5.5", "", "codex-vscode/1"),
			)
		})
	}
}

func TestAIUserAgentWithModelReplacesPreviousModel(t *testing.T) {
	tests := map[string]struct {
		UserAgent string
		Fallback  string
		Expected  string
	}{
		"file editor": {
			UserAgent: "gpt/5.4 future-editor/2 plugin/3",
			Fallback:  "future-editor/2 plugin/3",
			Expected:  "claude/4.6 future-editor/2 plugin/3",
		},
		"CLI header with spaces": {
			UserAgent: "wakatime/13.0.7 (Linux (Ubuntu)-6-amd64) go1.26 gpt/5.4 future-editor/2 plugin/3",
			Fallback:  "future-editor/2 plugin/3",
			Expected:  "wakatime/13.0.7 (Linux (Ubuntu)-6-amd64) go1.26 claude/4.6 future-editor/2 plugin/3",
		},
		"standalone agent": {
			UserAgent: "gpt/5.4 codex-cli/1 plugin/3",
			Fallback:  "codex-cli/1 plugin/3",
			Expected:  "claude/4.6 codex-cli/1 plugin/3",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			const entity = "/tmp/main.go"

			userAgents := map[string]string{entity: tt.UserAgent}
			got := aiUserAgentWithModel(entity, userAgents, tt.Fallback, "claude-4.6", "")
			assert.Equal(t, tt.Expected, got)

			userAgents[entity] = got
			assert.Equal(t, tt.Expected, aiUserAgentWithModel(entity, userAgents, tt.Fallback, "claude-4.6", ""))
		})
	}
}
