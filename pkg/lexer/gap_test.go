package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGap_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"declaration": {
			Filepath: "testdata/gap_declaration.g",
			Expected: 0.7,
		},
		"implementation": {
			Filepath: "testdata/gap_implementation.g",
			Expected: 0.7,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexers.Get(heartbeat.LanguageGap.StringChroma())
			require.NotNil(t, l)

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
