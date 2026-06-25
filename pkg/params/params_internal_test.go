package params

import (
	"net/url"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/regex"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBoolOrRegexList(t *testing.T) {
	ctx := t.Context()

	tests := map[string]struct {
		Input    string
		Expected []regex.Regex
	}{
		"string empty": {
			Input:    " ",
			Expected: nil,
		},
		"false string": {
			Input:    "false",
			Expected: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile("a^"))},
		},
		"true string": {
			Input:    "true",
			Expected: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
		},
		"valid regex": {
			Input: "\t.?\n\t\n \n\t\twakatime.? \t\n",
			Expected: []regex.Regex{
				regex.NewRegexpWrap(regexp.MustCompile("(?i).?")),
				regex.NewRegexpWrap(regexp.MustCompile("(?i)wakatime.?")),
			},
		},
		"valid regex with windows style": {
			Input: "\t.?\r\n\t\t\twakatime.? \t\r\n",
			Expected: []regex.Regex{
				regex.NewRegexpWrap(regexp.MustCompile("(?i).?")),
				regex.NewRegexpWrap(regexp.MustCompile("(?i)wakatime.?")),
			},
		},
		"valid regex with old mac style": {
			Input: "\t.?\r\t\t\twakatime.? \t\r",
			Expected: []regex.Regex{
				regex.NewRegexpWrap(regexp.MustCompile("(?i).?")),
				regex.NewRegexpWrap(regexp.MustCompile("(?i)wakatime.?")),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			regex, err := parseBoolOrRegexList(ctx, test.Input)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, regex)
		})
	}
}

func TestSafeTimeParse(t *testing.T) {
	parsed, err := vipertools.SafeTimeParse(ini.DateFormat, "2024-01-13T13:35:58Z")
	require.NoError(t, err)

	assert.Equal(t, time.Date(2024, 1, 13, 13, 35, 58, 0, time.UTC), parsed)
}

func TestSafeTimeParse_Err(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected string
	}{
		"empty string": {
			Input:    "",
			Expected: `parsing time "" as "2006-01-02T15:04:05Z07:00": cannot parse "" as "2006"`,
		},
		"invalid time": {
			Input:    "invalid",
			Expected: `parsing time "invalid" as "2006-01-02T15:04:05Z07:00": cannot parse "invalid" as "2006"`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parsed, err := vipertools.SafeTimeParse(ini.DateFormat, test.Input)
			require.Equal(t, time.Time{}, parsed)

			assert.EqualError(t, err, test.Expected)
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected string
	}{
		"already normalized": {
			Input:    "https://api.wakatime.com/api/v1",
			Expected: "https://api.wakatime.com/api/v1",
		},
		"trailing slash": {
			Input:    "https://api.wakatime.com/api/v1/",
			Expected: "https://api.wakatime.com/api/v1",
		},
		"with heartbeat endpoint": {
			Input:    "https://api.wakatime.com/api/v1/heartbeat",
			Expected: "https://api.wakatime.com/api/v1",
		},
		"with heartbeats endpoint": {
			Input:    "https://api.wakatime.com/api/v1/heartbeats",
			Expected: "https://api.wakatime.com/api/v1",
		},
		"with users current heartbeats endpoint": {
			Input:    "https://api.wakatime.com/api/v1/users/current/heartbeats",
			Expected: "https://api.wakatime.com/api/v1",
		},
		"with bulk suffix": {
			Input:    "https://api.wakatime.com/api/v1/users/current/heartbeats.bulk",
			Expected: "https://api.wakatime.com/api/v1",
		},
		"with trailing slash and bulk suffix": {
			Input:    "https://api.wakatime.com/api/v1/users/current/heartbeats.bulk/",
			Expected: "https://api.wakatime.com/api/v1",
		},
		"custom domain": {
			Input:    "https://custom.example.com/api/v1",
			Expected: "https://custom.example.com/api/v1",
		},
		"custom domain with endpoint": {
			Input:    "https://custom.example.com/api/v1/users/current/heartbeats.bulk",
			Expected: "https://custom.example.com/api/v1",
		},
		"http scheme": {
			Input:    "http://localhost:8080/api/v1",
			Expected: "http://localhost:8080/api/v1",
		},
		"empty string": {
			Input:    "",
			Expected: "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := normalizeURL(test.Input)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, result)
		})
	}
}

func TestLoadAPIKeyPatternsBranches(t *testing.T) {
	v := viper.New()
	v.Set("project_api_key.(?", "00000000-0000-4000-8000-000000000000")
	v.Set("project_api_key./same", "00000000-0000-4000-8000-000000000000")
	v.Set("project_api_key./custom", "11111111-1111-4111-8111-111111111111")

	patterns, err := loadAPIKeyPatterns(t.Context(), v, "00000000-0000-4000-8000-000000000000")
	require.NoError(t, err)
	require.Len(t, patterns, 1)
	assert.Equal(t, "11111111-1111-4111-8111-111111111111", patterns[0].APIKey)

	v = viper.New()
	v.Set("project_api_key./bad", "invalid")

	_, err = loadAPIKeyPatterns(t.Context(), v, "00000000-0000-4000-8000-000000000000")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid api key format")
}

func TestLoadAPIURLPatternsBranches(t *testing.T) {
	defaultURL, err := url.Parse("https://api.wakatime.com/api/v1")
	require.NoError(t, err)

	v := viper.New()
	v.Set("api_urls.(?", "https://ignored.example|11111111-1111-4111-8111-111111111111")
	v.Set("api_urls./default", "|00000000-0000-4000-8000-000000000000")
	v.Set(
		"api_urls./custom",
		"https://custom.example/api/v1/users/current/heartbeats.bulk|11111111-1111-4111-8111-111111111111",
	)
	v.Set("api_urls./invalidurl", "%|11111111-1111-4111-8111-111111111111")

	patterns, err := loadAPIURLPatterns(t.Context(), v, defaultURL, "00000000-0000-4000-8000-000000000000")
	require.NoError(t, err)
	require.Len(t, patterns, 2)

	var urls []string
	for _, pattern := range patterns {
		urls = append(urls, pattern.APIURL)
	}

	assert.ElementsMatch(t, []string{
		"https://api.wakatime.com/api/v1",
		"https://custom.example/api/v1",
	}, urls)

	v = viper.New()
	v.Set("api_urls./badkey", "https://custom.example|invalid")

	_, err = loadAPIURLPatterns(t.Context(), v, defaultURL, "00000000-0000-4000-8000-000000000000")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid api key format in api_urls")

	emptyURL := &url.URL{}
	v = viper.New()
	v.Set("api_urls./empty", "|00000000-0000-4000-8000-000000000000")
	patterns, err = loadAPIURLPatterns(t.Context(), v, emptyURL, "00000000-0000-4000-8000-000000000000")
	require.NoError(t, err)
	assert.Empty(t, patterns)
}

func TestParseExtraHeartbeatErrorBranches(t *testing.T) {
	base := ExtraHeartbeat{
		Category: "coding",
		Entity:   "main.go",
		Time:     float64(1),
		Type:     "file",
	}

	tests := map[string]struct {
		Heartbeat ExtraHeartbeat
		Contains  string
	}{
		"category": {
			Heartbeat: withExtraHeartbeat(base, func(h *ExtraHeartbeat) { h.Category = "bad" }),
			Contains:  "failed to parse category",
		},
		"entity type": {
			Heartbeat: withExtraHeartbeat(base, func(h *ExtraHeartbeat) { h.Type = "bad" }),
			Contains:  "invalid entity type",
		},
		"cursor": {
			Heartbeat: withExtraHeartbeat(base, func(h *ExtraHeartbeat) { h.CursorPosition = "bad" }),
			Contains:  "failed to convert cursorpos to int",
		},
		"is write": {
			Heartbeat: withExtraHeartbeat(base, func(h *ExtraHeartbeat) { h.IsWrite = "bad" }),
			Contains:  "failed to convert is write to bool",
		},
		"line number": {
			Heartbeat: withExtraHeartbeat(base, func(h *ExtraHeartbeat) { h.LineNumber = "bad" }),
			Contains:  "failed to convert lineno to int",
		},
		"lines": {
			Heartbeat: withExtraHeartbeat(base, func(h *ExtraHeartbeat) { h.Lines = "bad" }),
			Contains:  "failed to convert lines to int",
		},
		"time": {
			Heartbeat: withExtraHeartbeat(base, func(h *ExtraHeartbeat) { h.Time = "bad" }),
			Contains:  "failed to convert time to float64",
		},
		"timestamp": {
			Heartbeat: withExtraHeartbeat(base, func(h *ExtraHeartbeat) {
				h.Time = nil
				h.Timestamp = "bad"
			}),
			Contains: "failed to convert timestamp to float64",
		},
		"missing timestamp": {
			Heartbeat: withExtraHeartbeat(base, func(h *ExtraHeartbeat) { h.Time = nil }),
			Contains:  "no valid timestamp",
		},
		"is unsaved entity": {
			Heartbeat: withExtraHeartbeat(base, func(h *ExtraHeartbeat) { h.IsUnsavedEntity = "bad" }),
			Contains:  "failed to convert is_unsaved_entity to bool",
		},
		"ai line changes": {
			Heartbeat: withExtraHeartbeat(base, func(h *ExtraHeartbeat) { h.AILineChanges = "bad" }),
			Contains:  "failed to convert ai_line_changes to int",
		},
		"human line changes": {
			Heartbeat: withExtraHeartbeat(base, func(h *ExtraHeartbeat) { h.HumanLineChanges = "bad" }),
			Contains:  "failed to convert human_line_changes to int",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := parseExtraHeartbeat(test.Heartbeat)
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.Contains)
		})
	}
}

func TestParseExtraHeartbeatsBranches(t *testing.T) {
	heartbeats, err := parseExtraHeartbeats(t.Context(), "")
	require.NoError(t, err)
	assert.Nil(t, heartbeats)

	_, err = parseExtraHeartbeats(t.Context(), "{")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to json decode")

	_, err = parseExtraHeartbeats(t.Context(), `[{"category":"bad"}]`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse category")
}

func TestParamsStringAndHelperBranches(t *testing.T) {
	assert.Contains(t, (Params{}).String(), "api params:")
	assert.Equal(t, "unknown", FlagReadOrder(99).String())
	assert.Nil(t, mustParseIntegerNumber(t, nil))
	assert.Equal(t, 3, *mustParseIntegerNumber(t, float64(3)))
	assert.Equal(t, "first", firstNonEmptyString("", "first", "second"))
	assert.Equal(t, "", firstNonEmptyString("", ""))
}

func withExtraHeartbeat(base ExtraHeartbeat, update func(*ExtraHeartbeat)) ExtraHeartbeat {
	update(&base)

	return base
}

func mustParseIntegerNumber(t *testing.T, value any) *int {
	t.Helper()

	parsed, err := parseIntegerNumber(value)
	require.NoError(t, err)

	return parsed
}

func TestReadAPIKeyFromCommandBranches(t *testing.T) {
	key, err := readAPIKeyFromCommand("  ")
	require.NoError(t, err)
	assert.Empty(t, key)

	if runtime.GOOS == "windows" {
		return
	}

	key, err = readAPIKeyFromCommand(`printf 'waka_key'`)
	require.NoError(t, err)
	assert.Equal(t, "waka_key", strings.TrimSpace(key))

	_, err = readAPIKeyFromCommand("exit 7")
	require.Error(t, err)
}
