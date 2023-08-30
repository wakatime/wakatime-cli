package lexer_test

import (
	"os"
	"testing"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestC_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"include": {
			Filepath: "testdata/c_include.c",
			Expected: 0.1,
		},
		"ifdef": {
			Filepath: "testdata/c_ifdef.c",
			Expected: 0.1,
		},
		"ifndef": {
			Filepath: "testdata/c_ifndef.c",
			Expected: 0.1,
		},
	}

	l := lexers.Get(heartbeat.LanguageC.StringChroma())
	require.NotNil(t, l)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
