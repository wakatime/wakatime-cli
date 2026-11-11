package ai

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
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

func TestGrokBuildLegacyUsageCutoff(t *testing.T) {
	ctx := t.Context()
	base := time.Date(2026, 7, 24, 1, 0, 0, 0, time.UTC)
	session := grokBuildSession{id: "legacy"}
	prompt := grokBuildPrompt{timestamp: base.Add(time.Second), length: 10, model: "grok-4.5"}
	completedAt := base.Add(3 * time.Second)
	savedAt := base.Add(4 * time.Second)
	v := viper.New()
	v.Set("internal-config", filepath.Join(t.TempDir(), "wakatime-internal.cfg"))

	legacy, err := loadSyncCheckpoints(ctx, v, base)
	require.NoError(t, err)

	g := GrokBuild{}
	oldPrompt := g.promptHeartbeat(session, prompt)
	oldPrompt.AIInputTokens = 100
	oldPrompt.AIOutputTokens = 10
	oldParser := legacy.parser(g.Name(), base)
	oldParser.filter(Heartbeats{oldPrompt})
	require.NoError(t, legacy.save())
	require.NoError(t, os.Chtimes(legacy.path, savedAt, savedAt))

	// The legacy completion is later than the stored prompt cutoff. Keep it
	// suppressed across repeated scans while allowing completions after the upgrade.
	for range 2 {
		state, err := loadSyncCheckpoints(ctx, v, base)
		require.NoError(t, err)

		g.checkpoint = state.parser(g.Name(), base)
		g.initLegacyUsageCutoff()
		assert.True(t, g.legacyTurnReported(session, completedAt, prompt))
		assert.False(t, g.legacyTurnReported(session, savedAt.Add(time.Second), prompt))
		require.NoError(t, state.save())
	}
}

func TestGrokBuildSmallHelpers(t *testing.T) {
	zero, one, anotherOne := 0, 1, 1

	assert.True(t, grokBuildSamePromptIndex(nil, nil))
	assert.False(t, grokBuildSamePromptIndex(nil, &zero))
	assert.False(t, grokBuildSamePromptIndex(&zero, &one))
	assert.True(t, grokBuildSamePromptIndex(&one, &anotherOne))

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
