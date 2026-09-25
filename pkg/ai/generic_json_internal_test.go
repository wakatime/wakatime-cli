package ai

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenericAIKeyMatches(t *testing.T) {
	keys := []string{
		"", "_ -", "input_tokens", "InputTokens", "INPUT-TOKENS", "input tokens",
		"input", "inputtokensextra", "output_tokens", "Édit_文件", "édit文件", "\xff", "\ufffd",
	}
	for _, raw := range keys {
		for _, lookup := range keys {
			key := genericAIKey(lookup)
			assert.Equal(t, genericAIKey(raw) == key, genericAIKeyMatches(raw, key),
				"raw=%q lookup=%q", raw, lookup)
		}
	}
}

func TestGenericAIFieldLookupNestedContainers(t *testing.T) {
	for _, child := range []any{
		map[string]any{"Input-Tokens": 7},
		[]any{nil, map[string]any{"Input-Tokens": 7}},
		` {"Input-Tokens":7}`,
		` [{"Input-Tokens":7}]`,
		genericAICacheJSONText(`{"Input-Tokens":7}`),
	} {
		value := map[string]any{
			"a_scalar": "plain text", "b_invalid": "{invalid", "c_number": 123,
			"d_nil": nil, "e_bool": true, "f_first": child,
			"z_last": map[string]any{"input_tokens": 99},
		}
		// Nested matches retain alphabetical container precedence, including
		// objects and arrays encoded as JSON strings in SQLite columns.
		assert.Equal(t, int64(7), genericAIInt64(genericAIFind(value, "input_tokens")))
		number, ok := genericAINumericField(value, "input_tokens")
		require.True(t, ok)
		assert.Equal(t, int64(7), number)
		assert.Nil(t, genericAIFind(value, "missing"))
		_, ok = genericAINumericField(value, "missing")
		assert.False(t, ok)
	}
}

func TestGenericAITranscriptPathsPreferMessagesWithUsage(t *testing.T) {
	tests := []struct {
		name     string
		messages string
		wantBoth bool
	}{
		{name: "input", messages: `{"messages":[{"usage":{"input_tokens":12}}]}`},
		{name: "cached input", messages: `{"messages":[{"usage":{"cache_read_input_tokens":12}}]}`},
		{name: "cache creation", messages: `{"messages":[{"usage":{"cache_creation_input_tokens":12}}]}`},
		{name: "output", messages: `{"messages":[{"usage":{"output_tokens":12}}]}`},
		{name: "string count", messages: `{"messages":[{"usage":{"input_tokens":" 12 "}}]}`},
		{name: "later message", messages: `{"messages":[{"text":"hello"},{"usage":{"output_tokens":12}}]}`},
		{name: "no usage", messages: `{"messages":[{"text":"hello"}]}`, wantBoth: true},
		{name: "zero usage", messages: `{"messages":[{"input_tokens":0,"output_tokens":0}]}`, wantBoth: true},
		{name: "negative usage", messages: `{"messages":[{"input_tokens":-1}]}`, wantBoth: true},
		{name: "invalid count", messages: `{"messages":[{"input_tokens":"unknown"}]}`, wantBoth: true},
		{name: "malformed JSON", messages: `{"messages":`, wantBoth: true},
		{name: "non-object JSON", messages: `[null,42,"text"]`, wantBoth: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			original := filepath.Join(root, "session.json")
			messages := filepath.Join(root, "session.messages.json")

			require.NoError(t, os.WriteFile(original, []byte(`{"input_tokens":12}`), 0o600))
			require.NoError(t, os.WriteFile(messages, []byte(tt.messages), 0o600))

			provider := genericAIProvider{
				parser: ClineCLI{}, roots: []string{root},
				extensions: map[string]bool{".json": true}, preferMessagesFile: true,
			}
			got, err := provider.transcriptPaths()
			require.NoError(t, err)

			want := []string{messages}
			if tt.wantBoth {
				want = append(want, original)
			}

			assert.ElementsMatch(t, want, got)

			// Providers without this preference must retain both transcripts.
			provider.preferMessagesFile = false
			got, err = provider.transcriptPaths()
			require.NoError(t, err)
			assert.ElementsMatch(t, []string{original, messages}, got)
		})
	}
}

func TestGenericAITranscriptHasTokensMissingFile(t *testing.T) {
	assert.False(t, genericAITranscriptHasTokens(filepath.Join(t.TempDir(), "missing.json")))
}

func TestGenericAITranscriptPathsFiltersAndDeduplicates(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	require.NoError(t, os.Mkdir(nested, 0o700))

	cutoff := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)

	for _, name := range []string{"session.JSONL", "messages.data", "ignored.txt", "old.jsonl"} {
		path := filepath.Join(nested, name)
		require.NoError(t, os.WriteFile(path, []byte("{}"), 0o600))

		modified := cutoff.Add(time.Hour)
		if name == "old.jsonl" {
			modified = cutoff.Add(-time.Hour)
		}

		require.NoError(t, os.Chtimes(path, modified, modified))
	}

	provider := genericAIProvider{
		parser: Droid{}, config: ParserConfig{After: cutoff},
		roots:      []string{"", filepath.Join(root, "missing"), root, nested},
		fileNames:  map[string]bool{"messages.data": true},
		extensions: map[string]bool{".jsonl": true},
	}
	got, err := provider.transcriptPaths()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{
		filepath.Join(nested, "session.JSONL"), filepath.Join(nested, "messages.data"),
	}, got)

	// Explicit file roots obey the same extension and modification-time filters.
	provider.roots = []string{
		filepath.Join(nested, "session.JSONL"), filepath.Join(nested, "ignored.txt"),
		filepath.Join(nested, "old.jsonl"),
	}
	got, err = provider.transcriptPaths()
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(nested, "session.JSONL")}, got)
}

func TestGenericAIJSONLinesCanceled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	require.NoError(t, os.WriteFile(path, []byte("{}\n"), 0o600))

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	got, err := genericAIJSONLines(ctx, "Droid", path)
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, got)

	provider := genericAIProvider{parser: Droid{}, roots: []string{path}, extensions: map[string]bool{".jsonl": true}}
	_, err = parseGenericAIProvider(ctx, provider)
	require.ErrorIs(t, err, context.Canceled)
}

func TestParseGenericAITranscriptMissingFile(t *testing.T) {
	for _, extension := range []string{".json", ".jsonl"} {
		t.Run(extension, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "missing"+extension)
			got, err := parseGenericAITranscript(t.Context(), genericAIProvider{parser: Droid{}}, path)
			require.Error(t, err)
			assert.Nil(t, got)
		})
	}
}

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

func TestGenericAIHeartbeatsSingleValueTimestampFallback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.json")
	require.NoError(t, os.WriteFile(path, []byte("{}"), 0o600))

	info, err := os.Stat(path)
	require.NoError(t, err)

	value := map[string]any{"session_id": "session-1", "input_tokens": 10}
	provider := genericAIProvider{parser: Droid{}}

	got, err := genericAIHeartbeats(t.Context(), provider, path, []any{value})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, heartbeatTimestamp(info.ModTime()), got[0].Time)

	// A shared file timestamp must not manufacture dates for multiple events.
	got, err = genericAIHeartbeats(t.Context(), provider, path, []any{value, value})
	require.NoError(t, err)
	assert.Empty(t, got)

	got, err = genericAIHeartbeats(t.Context(), provider, path, nil)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGenericAITimeFormats(t *testing.T) {
	want := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		value any
		want  time.Time
	}{
		{name: "time value", value: want, want: want},
		{name: "RFC3339", value: "2026-08-06T12:00:00Z", want: want},
		{name: "timezone offset", value: "2026-08-06 14:00:00+02:00", want: want},
		{name: "SQL timestamp", value: "2026-08-06 12:00:00", want: want},
		{name: "seconds", value: float64(want.Unix()), want: want},
		{name: "integer seconds", value: int(want.Unix()), want: want},
		{name: "int64 seconds", value: want.Unix(), want: want},
		{name: "milliseconds", value: float64(want.UnixMilli()), want: want},
		{name: "microseconds", value: float64(want.UnixMicro()), want: want},
		{name: "nanoseconds", value: float64(want.UnixNano()), want: want},
		{name: "numeric string", value: "1786017600", want: want},
		{name: "fractional seconds", value: 1.25, want: time.Unix(1, 250000000)},
		{name: "negative seconds", value: -1.25, want: time.Unix(-2, 750000000)},
		{name: "negative milliseconds", value: -float64(want.UnixMilli()), want: time.Unix(-want.Unix(), 0)},
		{name: "negative microseconds", value: -float64(want.UnixMicro()), want: time.Unix(-want.Unix(), 0)},
		{name: "negative nanoseconds", value: -float64(want.UnixNano()), want: time.Unix(-want.Unix(), 0)},
		{name: "invalid string", value: "yesterday"},
		{name: "nil", value: nil},
		{name: "unsupported type", value: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.want.Equal(genericAITime(tt.value)), "input: %v", tt.value)
		})
	}
}

func TestGenericAIPromptUserInputFlags(t *testing.T) {
	tests := []struct {
		name string
		flag any
		want string
	}{
		{name: "boolean true", flag: true, want: "first\nsecond"},
		{name: "boolean false", flag: false},
		{name: "string true", flag: " true ", want: "first\nsecond"},
		{name: "string false", flag: "false"},
		{name: "invalid string", flag: "unknown"},
		{name: "float true", flag: float64(1), want: "first\nsecond"},
		{name: "float false", flag: float64(0)},
		{name: "sqlite true", flag: int64(1), want: "first\nsecond"},
		{name: "sqlite false", flag: int64(0)},
		{name: "missing flag", flag: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := map[string]any{
				"is_user_input": tt.flag,
				"content": []any{
					map[string]any{"text": " first "}, nil,
					map[string]any{"text": "", "content": "second"},
				},
			}
			assert.Equal(t, tt.want, genericAIPrompt(value))
		})
	}
}

func TestGenericAIJSONLinesReadsTailOfLargeTranscript(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large.jsonl")
	contents := `{"text":"` + strings.Repeat("x", maxTranscriptLineSize) + "\"}\n\n" +
		"{\"session_id\":\"recent\",\"input_tokens\":7}\n"
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))

	got, err := genericAIJSONLines(t.Context(), "Droid", path)
	require.NoError(t, err)
	assert.Equal(t, []any{map[string]any{"session_id": "recent", "input_tokens": float64(7)}}, got)
}
