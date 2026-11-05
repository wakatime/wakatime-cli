package ai

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

func TestCodexPromptAttributionIsolation(t *testing.T) {
	logger := log.New(io.Discard)
	parser := Codex{
		FallbackUserAgent: "webstorm/2026.2.3 webstorm-wakatime/16.1.2",
		UserAgents:        map[string]string{"/workspace/project/main.go": "webstorm/2026.2.3 webstorm-wakatime/16.1.2"},
	}
	sessionA := codexSessionState{id: "a", entity: "Codex A", source: "vscode", version: "0.165.0"}
	sessionB := codexSessionState{id: "b", entity: "Codex B", source: "cli", version: "0.141.0"}

	var a, b codexParseState

	prompt := []byte(`{"timestamp":"2026-10-29T01:18:32.472Z","type":"event_msg",` +
		`"payload":{"type":"user_message","message":"Hello"}}`)
	context := []byte(`{"timestamp":"2026-10-29T01:18:32.474Z","type":"turn_context",` +
		`"payload":{"model":"gpt-6.1-terra","effort":"xhigh"}}`)

	parser.handleTranscriptLine(logger, "a", prompt, &sessionA, &a)
	parser.handleTranscriptLine(logger, "b", prompt, &sessionB, &b)
	parser.handleTranscriptLine(logger, "a", prompt, &sessionA, &a)
	parser.handleTranscriptLine(logger, "a", context, &sessionA, &a)
	require.Len(t, a.heartbeats, 2)
	require.Len(t, b.heartbeats, 1)

	for _, h := range a.heartbeats {
		assert.Equal(t, "a", h.AISession)
		assert.Equal(t, "gpt/6.1-terra-xhigh webstorm/2026.2.3 codex-vscode/0.165.0 webstorm-wakatime/16.1.2", h.UserAgent)
	}

	assert.Equal(t, "b", b.heartbeats[0].AISession)
	assert.Equal(t, "webstorm/2026.2.3 codex-cli/0.141.0 webstorm-wakatime/16.1.2", b.heartbeats[0].UserAgent)
}

func TestCodexPromptAttributionMalformedBoundary(t *testing.T) {
	logger := log.New(io.Discard)
	parser := Codex{FallbackUserAgent: "plugin/1"}
	session := codexSessionState{id: "a", entity: "Codex A", source: "cli", version: "1"}

	var state codexParseState
	for _, line := range []string{
		`{"timestamp":"2026-10-29T01:18:32Z","type":"event_msg",` +
			`"payload":{"type":"user_message","message":"Hello"}}`,
		`{broken JSON`,
		`{"timestamp":"2026-10-29T01:18:34Z","type":"turn_context",` +
			`"payload":{"model":"gpt-5.5","effort":"high"}}`,
	} {
		parser.handleTranscriptLine(logger, "a", []byte(line), &session, &state)
	}

	require.Len(t, state.heartbeats, 1)
	assert.Equal(t, "codex-cli/1 plugin/1", state.heartbeats[0].UserAgent)
}

// Includes transcript decoding and heartbeat construction, not filesystem I/O.
func BenchmarkCodexPromptAttribution(b *testing.B) {
	logger := log.New(io.Discard)
	parser := Codex{
		FallbackUserAgent: "webstorm/2026.2.3 webstorm-wakatime/16.1.2",
		UserAgents:        map[string]string{"/workspace/project/main.go": "webstorm/2026.2.3 webstorm-wakatime/16.1.2"},
	}
	lines := [][]byte{
		[]byte(`{"timestamp":"2026-10-29T01:18:32Z","type":"event_msg",` +
			`"payload":{"type":"user_message","message":"Hello"}}`),
		[]byte(`{"timestamp":"2026-10-29T01:18:33Z","type":"turn_context",` +
			`"payload":{"model":"gpt-5.5","effort":"high"}}`),
		[]byte(`{"timestamp":"2026-10-29T01:18:34Z","type":"event_msg",` +
			`"payload":{"type":"agent_message","message":"Hi"}}`),
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		session := codexSessionState{id: "a", entity: "Codex A", source: "vscode", version: "1"}

		var state codexParseState
		for _, line := range lines {
			parser.handleTranscriptLine(logger, "a", line, &session, &state)
		}
	}
}
