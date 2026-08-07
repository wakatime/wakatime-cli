package ai

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenericAIJSONLinesContinueAfterMalformedRecordAndDiffCumulativeCounters(t *testing.T) {
	transcript := filepath.Join(t.TempDir(), "session.jsonl")
	contents := strings.Join([]string{
		`{"timestamp":"2026-08-06T11:00:00Z","session_id":"session-1","cwd":"/workspace/project",` +
			`"usage":{"input_tokens":100,"cached_input_tokens":20,"output_tokens":40},` +
			`"tool":{"tool_name":"edit_file","file_path":"pkg/main.go","lines_added":7,"lines_removed":3}}`,
		`{"timestamp":"broken"`,
		`{"timestamp":"2026-08-06T11:45:00Z","unexpected":{"shape":true}}`,
		`{"timestamp":"2026-08-06T12:00:00Z","session_id":"session-1",` +
			`"usage":{"input_tokens":145,"cached_input_tokens":25,"output_tokens":60},` +
			`"tool":{"tool_name":"edit_file","file_path":"pkg/main.go","lines_added":9,"lines_removed":5}}`,
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcript, []byte(contents), 0o600))

	provider := genericAIProvider{
		parser:           Droid{},
		config:           ParserConfig{After: time.Date(2026, 8, 6, 11, 30, 0, 0, time.UTC)},
		tokenCounterMode: genericAICumulativeCounters,
		lineCounterMode:  genericAICumulativeCounters,
	}

	got, err := parseGenericAITranscript(t.Context(), provider, transcript)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, int64(45), got[0].AIInputTokens)
	assert.Equal(t, int64(5), got[0].AICachedInputTokens)
	assert.Equal(t, int64(20), got[0].AIOutputTokens)
	assert.Equal(t, "Droid session-1", got[0].Entity)
	assert.Equal(t, filepath.ToSlash(filepath.Join("/workspace/project", "pkg/main.go")), got[1].Entity)
	assert.Equal(t, 4, *got[1].AILineChanges)
}

func TestGenericAIJSONLinesUsePerEventTokenCountsByDefault(t *testing.T) {
	transcript := filepath.Join(t.TempDir(), "session.jsonl")
	contents := strings.Join([]string{
		`{"timestamp":"2026-08-06T12:00:00Z","session_id":"session-1","input_tokens":10,"output_tokens":4}`,
		`{"timestamp":"2026-08-06T12:01:00Z","session_id":"session-1","input_tokens":6,"output_tokens":3}`,
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcript, []byte(contents), 0o600))

	provider := genericAIProvider{parser: Droid{}, config: ParserConfig{}}
	got, err := parseGenericAITranscript(t.Context(), provider, transcript)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, int64(10), got[0].AIInputTokens)
	assert.Equal(t, int64(4), got[0].AIOutputTokens)
	assert.Equal(t, int64(6), got[1].AIInputTokens)
	assert.Equal(t, int64(3), got[1].AIOutputTokens)
}
