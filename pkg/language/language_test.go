package language_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/language"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithDetection(t *testing.T) {
	tests := map[string]struct {
		Alternate string
		Override  string
		Expected  heartbeat.Language
	}{
		"alternate": {
			Alternate: "Go",
			Expected:  heartbeat.LanguageGo,
		},
		"override": {
			Alternate: "Go",
			Override:  "Python",
			Expected:  heartbeat.LanguagePython,
		},
		"empty": {
			Expected: heartbeat.LanguageUnknown,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			opt := language.WithDetection(language.Config{
				Alternate: test.Alternate,
				Override:  test.Override,
			})
			h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
				assert.Equal(t, []heartbeat.Heartbeat{
					{
						Language: test.Expected,
					},
				}, hh)

				return []heartbeat.Result{
					{
						Status: 201,
					},
				}, nil
			})

			result, err := h([]heartbeat.Heartbeat{{}})
			require.NoError(t, err)

			assert.Equal(t, []heartbeat.Result{
				{
					Status: 201,
				},
			}, result)
		})
	}
}
