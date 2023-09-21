package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrainfuck_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"plus minus": {
			Filepath: "testdata/brainfuck_plus_minus.bf",
			Expected: 1.0,
		},
		"greater less": {
			Filepath: "testdata/brainfuck_greater_less.bf",
			Expected: 1.0,
		},
		"minus only": {
			Filepath: "testdata/brainfuck_minus.bf",
			Expected: 0.5,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexers.Get(heartbeat.LanguageBrainfuck.StringChroma())
			require.NotNil(t, l)

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
