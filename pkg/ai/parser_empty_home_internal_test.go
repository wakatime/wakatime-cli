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

	after := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	parsers := []Parser{
		Amp{After: after},
		Claude{After: after},
		Cline{After: after},
		Codex{After: after},
		Cody{After: after},
		Continue{After: after},
		Copilot{After: after},
		Cursor{After: after},
		Gemini{After: after},
		Goose{After: after},
		GrokBuild{After: after},
		Kiro{After: after},
		OpenCode{After: after},
		Pi{After: after},
		Qoder{After: after},
		QwenCode{After: after},
		RooCode{After: after},
		Windsurf{After: after},
	}

	for _, parser := range parsers {
		t.Run(parser.Name(), func(t *testing.T) {
			heartbeats, err := parser.Parse(t.Context())

			require.NoError(t, err)
			assert.Empty(t, heartbeats)
		})
	}
}
