package params

import (
	"regexp"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/regex"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBoolOrRegexList(t *testing.T) {
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
			Expected: []regex.Regex{regexp.MustCompile("a^")},
		},
		"true string": {
			Input:    "true",
			Expected: []regex.Regex{regexp.MustCompile(".*")},
		},
		"valid regex": {
			Input: "\t.?\n\t\n \n\t\twakatime.? \t\n",
			Expected: []regex.Regex{
				regexp.MustCompile(".?"),
				regexp.MustCompile("wakatime.?"),
			},
		},
		"valid regex with windows style": {
			Input: "\t.?\r\n\t\t\twakatime.? \t\r\n",
			Expected: []regex.Regex{
				regexp.MustCompile(".?"),
				regexp.MustCompile("wakatime.?"),
			},
		},
		"valid regex with old mac style": {
			Input: "\t.?\r\t\t\twakatime.? \t\r",
			Expected: []regex.Regex{
				regexp.MustCompile(".?"),
				regexp.MustCompile("wakatime.?"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			regex, err := parseBoolOrRegexList(test.Input)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, regex)
		})
	}
}
