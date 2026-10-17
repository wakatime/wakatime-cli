package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestLeadingAppHeartbeatUsesFileEditorForSameEvent(t *testing.T) {
	const filePath = "/workspace/main.go"

	config := ParserConfig{
		FallbackUserAgent: "cursor/1.128.0 vscode-wakatime/30.2.1",
		UserAgents: map[string]string{
			filePath: "neovim/0.12 wakatime.nvim/12.0.0",
		},
	}
	tests := map[string]struct {
		model     string
		userAgent string
	}{
		"known model": {
			model:     "grok-4.5",
			userAgent: "grok/4.5 neovim/0.12 cursor-agent wakatime.nvim/12.0.0",
		},
		"unknown model": {
			userAgent: "neovim/0.12 cursor-agent wakatime.nvim/12.0.0",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			event := genericAIEvent{
				timestamp:   time.Date(2026, 8, 6, 12, 0, 0, 123000000, time.UTC),
				sessionID:   "session-1",
				model:       tt.model,
				filePath:    filePath,
				isWrite:     true,
				lineChanges: heartbeat.PointerTo(3),
				prompt:      "change me",
				input:       10,
				cachedInput: 2,
				output:      4,
			}

			got := genericAIEventHeartbeats(CursorAgent{}, config, event)
			require.Len(t, got, 2)
			assert.Equal(t, tt.userAgent, got[0].UserAgent)
			assert.Equal(t, tt.userAgent, got[1].UserAgent)

			got = replaceAppHeartbeats(got)
			require.Len(t, got, 2)

			var (
				input, cachedInput, output int64
				promptLength, lineChanges  int
			)

			for _, h := range got {
				assert.Equal(t, heartbeat.FileType, h.EntityType)
				assert.Equal(t, filePath, h.Entity)
				assert.Equal(t, event.sessionID, h.AISession)
				assert.Equal(t, heartbeatTimestamp(event.timestamp), h.Time)
				assert.Equal(t, tt.userAgent, h.UserAgent)
				input += h.AIInputTokens
				cachedInput += h.AICachedInputTokens
				output += h.AIOutputTokens
				promptLength += h.AIPromptLength

				if h.AILineChanges != nil {
					lineChanges += *h.AILineChanges
				}
			}

			assert.Equal(t, int64(10), input)
			assert.Equal(t, int64(2), cachedInput)
			assert.Equal(t, int64(4), output)
			assert.Equal(t, 9, promptLength)
			assert.Equal(t, 3, lineChanges)
			assert.Equal(t, heartbeat.PointerTo(true), got[1].IsWrite)
		})
	}
}

func TestAppOnlyHeartbeatUsesFallbackEditor(t *testing.T) {
	got := genericAIEventHeartbeats(CursorAgent{}, ParserConfig{
		FallbackUserAgent: "cursor/1.128.0 vscode-wakatime/30.2.1",
		UserAgents: map[string]string{
			"/workspace/main.go": "neovim/0.12 wakatime.nvim/12.0.0",
		},
	}, genericAIEvent{
		timestamp: time.Date(2026, 8, 6, 12, 0, 0, 123000000, time.UTC),
		sessionID: "session-1",
		model:     "grok-4.5",
		input:     10,
	})

	require.Len(t, got, 1)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "grok/4.5 cursor/1.128.0 cursor-agent vscode-wakatime/30.2.1", got[0].UserAgent)
	assert.Equal(t, int64(10), got[0].AIInputTokens)
}
