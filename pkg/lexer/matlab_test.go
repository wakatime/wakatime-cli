package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatlab_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"function": {
			Filepath: "testdata/matlab_function.m",
			Expected: 1.0,
		},
		"comment": {
			Filepath: "testdata/matlab_comment.m",
			Expected: 0.2,
		},
		"systemcmd": {
			Filepath: "testdata/matlab_systemcmd.m",
			Expected: 0.2,
		},
		"windows": {
			Filepath: "testdata/matlab_windows.m",
			Expected: 1.0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexers.Get(heartbeat.LanguageMatlab.StringChroma())
			require.NotNil(t, l)

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
