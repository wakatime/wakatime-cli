package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFSharp_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"pipeline operator": {
			Filepath: "testdata/fsharp_pipeline_operator.fs",
			Expected: 0.1,
		},
		"forward pipeline operator": {
			Filepath: "testdata/fsharp_forward_pipeline_operator.fs",
			Expected: 0.05,
		},
		"backward pipeline operator": {
			Filepath: "testdata/fsharp_backward_pipeline_operator.fs",
			Expected: 0.05,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexers.Get(heartbeat.LanguageFSharp.StringChroma())
			require.NotNil(t, l)

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
