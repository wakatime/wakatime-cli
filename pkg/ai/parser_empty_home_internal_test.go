package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsersEmptyHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("CODEX_HOME", "")
	t.Setenv("DSH_HOME", "")
	t.Setenv("GROK_HOME", "")

	after := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	parsers := []Parser{
		Amp{After: after},
		Claude{After: after},
		Cline{After: after},
		ClineCLI{After: after},
		Codebuff{After: after},
		CodeWhale{After: after},
		Codex{After: after},
		Cody{After: after},
		Continue{After: after},
		Copilot{After: after},
		Crush{After: after},
		Cursor{After: after},
		CursorAgent{After: after},
		Devin{After: after},
		Droid{After: after},
		DeepSeek{After: after},
		Forge{After: after},
		Gemini{After: after},
		Goose{After: after},
		GrokBuild{After: after},
		Hermes{After: after},
		IBMBob{After: after},
		Kiro{After: after},
		KiloCode{After: after},
		Kimi{After: after},
		KimiCode{After: after},
		LingTaiTUI{After: after},
		MistralVibe{After: after},
		Mux{After: after},
		OMP{After: after},
		OpenClaw{After: after},
		OpenCode{After: after},
		OpenDesign{After: after},
		Pi{After: after},
		Qoder{After: after},
		QwenCode{After: after},
		QuickDesk{After: after},
		RooCode{After: after},
		Warp{After: after},
		Windsurf{After: after},
		ZCode{After: after},
		Zed{After: after},
		Zerostack{After: after},
	}

	for _, parser := range parsers {
		t.Run(parser.Name(), func(t *testing.T) {
			heartbeats, err := parser.Parse(t.Context())

			require.NoError(t, err)
			assert.Empty(t, heartbeats)
		})
	}
}
