package params

import (
	"regexp"
	"testing"
	"time"

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
