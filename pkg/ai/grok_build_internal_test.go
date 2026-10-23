package ai

import (
	"bufio"
	"context"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGrokBuildReadJSONLLineContinuesAfterOversizedRecord(t *testing.T) {
	contents := "one\r\n" + strings.Repeat("x", 32) + "\ntwo"
	reader := bufio.NewReaderSize(strings.NewReader(contents), 16)

	line, err := grokBuildReadJSONLLine(reader, 4)
	require.NoError(t, err)
	assert.Equal(t, "one", string(line))

	line, err = grokBuildReadJSONLLine(reader, 4)
	assert.Nil(t, line)
	assert.ErrorIs(t, err, errGrokBuildLineTooLong)

	line, err = grokBuildReadJSONLLine(reader, 4)
	assert.Equal(t, "two", string(line))
	assert.ErrorIs(t, err, io.EOF)

	line, err = grokBuildReadJSONLLine(reader, 4)
	assert.Nil(t, line)
	assert.True(t, errors.Is(err, io.EOF))
}

func TestGrokBuildUserAgentProduct(t *testing.T) {
	assert.Equal(t, "grok-build/unknown", grokBuildUserAgentProduct(""))
	assert.Equal(t, "grok-build/1.2.3", grokBuildUserAgentProduct(" 1.2.3 "))
}

func TestGrokBuildUserAgentEditorPreference(t *testing.T) {
	const entity = "/tmp/main.go"

	tests := []struct {
		name      string
		fallback  string
		fileAgent string
		model     string
		version   string
		defaultUA string
		preferred string
	}{
		{
			name:      "host IDE",
			fallback:  "cursor/1.128.0 vscode-wakatime/30.2.1",
			model:     "grok-4.7",
			version:   "1.0.67",
			defaultUA: "grok/4.7 cursor/1.128.0 grok-build/1.0.67 vscode-wakatime/30.2.1",
			preferred: "grok/4.7 grok-build/1.0.67 cursor/1.128.0 vscode-wakatime/30.2.1",
		},
		{
			name:      "file editor overrides fallback",
			fallback:  "cursor/1.128.0 vscode-wakatime/30.2.1",
			fileAgent: "neovim/0.12 wakatime.nvim/12.0.0",
			model:     "grok-4.7",
			version:   "1.0.67",
			defaultUA: "grok/4.7 neovim/0.12 grok-build/1.0.67 wakatime.nvim/12.0.0",
			preferred: "grok/4.7 grok-build/1.0.67 neovim/0.12 wakatime.nvim/12.0.0",
		},
		{
			name:      "missing model and version",
			fallback:  "cursor/1.128.0 vscode-wakatime/30.2.1",
			defaultUA: "cursor/1.128.0 grok-build/unknown vscode-wakatime/30.2.1",
			preferred: "grok-build/unknown cursor/1.128.0 vscode-wakatime/30.2.1",
		},
		{
			name:      "without host IDE",
			fallback:  "plugin/1",
			model:     "grok-4.7",
			version:   "1.0.67",
			defaultUA: "grok/4.7 grok-build/1.0.67 plugin/1",
			preferred: "grok/4.7 grok-build/1.0.67 plugin/1",
		},
		{
			name:      "full user agent",
			fallback:  "wakatime/1 (Linux (Ubuntu)-6-amd64) go1.27.2 cursor/1 vscode-wakatime/2",
			model:     "grok-4.7",
			version:   "1.0.67",
			defaultUA: "wakatime/1 (Linux (Ubuntu)-6-amd64) go1.27.2 grok/4.7 cursor/1 grok-build/1.0.67 vscode-wakatime/2",
			preferred: "wakatime/1 (Linux (Ubuntu)-6-amd64) go1.27.2 grok/4.7 grok-build/1.0.67 cursor/1 vscode-wakatime/2",
		},
		{
			name:      "existing product version and duplicates",
			fallback:  "cursor/1 grok-build/1.0.13 grok-build/1.0.13 vscode-wakatime/2",
			model:     "grok-4.7",
			version:   "1.0.67",
			defaultUA: "grok/4.7 cursor/1 grok-build/1.0.13 vscode-wakatime/2",
			preferred: "grok/4.7 grok-build/1.0.13 cursor/1 vscode-wakatime/2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for name, prefer := range map[string]bool{"default": false, "preferred": true} {
				t.Run(name, func(t *testing.T) {
					parser := GrokBuild{FallbackUserAgent: tt.fallback, GrokBuildAsEditor: prefer}
					if tt.fileAgent != "" {
						parser.UserAgents = map[string]string{entity: tt.fileAgent}
					}

					expected := tt.defaultUA
					if prefer {
						expected = tt.preferred
					}

					got := parser.userAgent(entity, tt.model, tt.version)
					assert.Equal(t, expected, got)
					parser.UserAgents = map[string]string{entity: got}
					assert.Equal(t, expected, parser.userAgent(entity, tt.model, tt.version))
				})
			}
		})
	}
}

func TestGrokBuildUserAgentEditorPreferenceRecomposesPreviousModel(t *testing.T) {
	const (
		entity   = "/tmp/main.go"
		fallback = "cursor/2 vscode-wakatime/3"
		previous = "wakatime/1 (Linux-6-amd64) go1.27.2 grok/4.7 grok-build/1 cursor/2 vscode-wakatime/3"
	)

	for _, model := range []string{"", "grok-4.8"} {
		t.Run(model, func(t *testing.T) {
			parser := GrokBuild{
				FallbackUserAgent: fallback,
				UserAgents:        map[string]string{entity: previous},
				GrokBuildAsEditor: true,
			}

			expected := previous
			if model != "" {
				expected = strings.Replace(previous, "grok/4.7", "grok/4.8", 1)
			}

			got := parser.userAgent(entity, model, "1")
			assert.Equal(t, expected, got)
			parser.UserAgents[entity] = got
			assert.Equal(t, expected, parser.userAgent(entity, model, "1"))
		})
	}
}

func TestGrokBuildUserAgentEditorPreferencePreservesCallerProducts(t *testing.T) {
	const entity = "/tmp/main.go"

	tests := []struct {
		name     string
		fallback string
		existing string
		expected string
	}{
		{
			name:     "custom editor before Grok Build",
			fallback: "cursor/2",
			existing: "custom-editor/4 grok-build/1 cursor/2",
			expected: "grok-build/1 custom-editor/4 cursor/2",
		},
		{
			name:     "caller editor named Grok",
			fallback: "grok/2 editor-wakatime/3",
			existing: "grok/2 grok-build/1 editor-wakatime/3",
			expected: "grok-build/1 grok/2 editor-wakatime/3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := GrokBuild{
				FallbackUserAgent: tt.fallback,
				UserAgents:        map[string]string{entity: tt.existing},
				GrokBuildAsEditor: true,
			}
			assert.Equal(t, tt.expected, parser.userAgent(entity, "", "1"))
		})
	}
}

func TestGrokBuildDecodeUpdate(t *testing.T) {
	t.Run("envelope", func(t *testing.T) {
		event, err := grokBuildDecodeUpdate([]byte(`{
			"timestamp": 123,
			"method": "_x.ai/session/update",
			"params": {"sessionId": "session", "update": {"sessionUpdate": "turn_completed"}}
		}`))
		require.NoError(t, err)
		assert.EqualValues(t, 123, event.timestamp)
		assert.True(t, event.isXAI)
		assert.Equal(t, "session", event.params.SessionID)
	})

	t.Run("raw params", func(t *testing.T) {
		event, err := grokBuildDecodeUpdate([]byte(
			`{"sessionId":"session","update":{"sessionUpdate":"user_message_chunk"}}`,
		))
		require.NoError(t, err)
		assert.False(t, event.isXAI)
		assert.Equal(t, "user_message_chunk", event.params.Update.SessionUpdate)
	})

	t.Run("tool update with array content", func(t *testing.T) {
		event, err := grokBuildDecodeUpdate([]byte(`{
			"params": {
				"sessionId": "session",
				"update": {
					"sessionUpdate": "tool_call_update",
					"content": [{"type": "content", "content": "result"}]
				}
			}
		}`))
		require.NoError(t, err)
		assert.Equal(t, "tool_call_update", event.params.Update.SessionUpdate)
		assert.Nil(t, event.params.Update.Content)
	})

	for name, input := range map[string]string{
		"malformed envelope": `{`,
		"malformed params":   `{"params":{"update":`,
		"missing type":       `{"params":{"sessionId":"session","update":{}}}`,
		"raw missing type":   `{"sessionId":"session","update":{}}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := grokBuildDecodeUpdate([]byte(input))
			assert.Error(t, err)
		})
	}
}

func TestGrokBuildEventTime(t *testing.T) {
	metaTime := time.Date(2026, 7, 24, 1, 2, 3, 4_000_000, time.UTC)
	secondsTime := time.Date(2026, 7, 24, 1, 2, 3, 0, time.UTC)

	tests := []struct {
		name  string
		event grokBuildUpdateEvent
		want  time.Time
	}{
		{
			name: "params metadata takes precedence",
			event: grokBuildUpdateEvent{timestamp: secondsTime.Unix(), params: grokBuildUpdateParams{
				Meta: &grokBuildUpdateMeta{AgentTimestampMS: metaTime.UnixMilli()},
			}},
			want: metaTime,
		},
		{
			name: "update metadata fallback",
			event: grokBuildUpdateEvent{params: grokBuildUpdateParams{Update: grokBuildUpdate{
				Meta: &grokBuildUpdateMeta{AgentTimestampMS: metaTime.UnixMilli()},
			}}},
			want: metaTime,
		},
		{name: "milliseconds", event: grokBuildUpdateEvent{timestamp: metaTime.UnixMilli()}, want: metaTime},
		{name: "seconds", event: grokBuildUpdateEvent{timestamp: secondsTime.Unix()}, want: secondsTime},
		{name: "missing", event: grokBuildUpdateEvent{}, want: time.Time{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, grokBuildEventTime(tt.event))
		})
	}
}

func TestGrokBuildSmallHelpers(t *testing.T) {
	zero, one, anotherOne := 0, 1, 1

	assert.True(t, grokBuildSamePromptIndex(nil, nil))
	assert.False(t, grokBuildSamePromptIndex(nil, &zero))
	assert.False(t, grokBuildSamePromptIndex(&zero, &one))
	assert.True(t, grokBuildSamePromptIndex(&one, &anotherOne))

	negative, positive := int64(-1), int64(7)

	assert.Zero(t, grokBuildNonNegativeTokenCount(nil))
	assert.Zero(t, grokBuildNonNegativeTokenCount(&negative))
	assert.EqualValues(t, 7, grokBuildNonNegativeTokenCount(&positive))

	assert.Empty(t, grokBuildNativePath("  "))

	wantPath := filepath.Join("project", "file.go")
	assert.Equal(t, wantPath, grokBuildNativePath(" project/file.go "))
}

func TestGrokBuildHomeDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	t.Run("absolute", func(t *testing.T) {
		configured := filepath.Join(home, "custom", "..", "grok")
		t.Setenv("GROK_HOME", configured)

		got, err := grokBuildHomeDir(context.Background())
		require.NoError(t, err)
		assert.Equal(t, filepath.Clean(configured), got)
	})

	t.Run("home", func(t *testing.T) {
		t.Setenv("GROK_HOME", "~")

		got, err := grokBuildHomeDir(context.Background())
		require.NoError(t, err)
		assert.Equal(t, home, got)
	})

	t.Run("home child", func(t *testing.T) {
		t.Setenv("GROK_HOME", "~/.custom-grok")

		got, err := grokBuildHomeDir(context.Background())
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(home, ".custom-grok"), got)
	})

	t.Run("relative", func(t *testing.T) {
		t.Setenv("GROK_HOME", filepath.Join("testdata", "grok"))

		got, err := grokBuildHomeDir(context.Background())
		require.NoError(t, err)
		assert.True(t, filepath.IsAbs(got))
		assert.Equal(t, filepath.Join("testdata", "grok"), filepath.Join(
			filepath.Base(filepath.Dir(got)),
			filepath.Base(got),
		))
	})
}
