package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
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

			l := lexer.FSharp{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
