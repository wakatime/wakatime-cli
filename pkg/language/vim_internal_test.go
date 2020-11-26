package language

import (
	"fmt"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectVimModeline(t *testing.T) {
	tests := map[string]struct {
		Text     string
		Language heartbeat.Language
		Error    error
	}{
		"ft": {
			Text:     `/* vim: ft=python tw=60 ts=2: */`,
			Language: heartbeat.LanguagePython,
		},
		"filetype": {
			Text:     "/* vim: filetype=python tw=60 ts=2: */",
			Language: heartbeat.LanguagePython,
		},
		"syn": {
			Text:     "/* vim: syn=python tw=60 ts=2: */",
			Language: heartbeat.LanguagePython,
		},
		"syntax": {
			Text:     "/* vim: syntax=python tw=60 ts=2: */",
			Language: heartbeat.LanguagePython,
		},
		"different order": {
			Text:     "/* vim: tw=60 ft=python ts=2: */",
			Language: heartbeat.LanguagePython,
		},
		"multiline": {
			Text: `
			/* vim: tw=60 ft=python ts=2: */
			`,
			Language: heartbeat.LanguagePython,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			lang, weight, ok := detectVimModeline(test.Text)
			require.True(t, ok)

			assert.Equal(t, float32(0), weight)
			assert.Equal(t, test.Language, lang, fmt.Sprintf("Got: %q, want: %q", lang, test.Language))
		})
	}
}
