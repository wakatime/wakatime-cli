package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestOpenEdge_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"end": {
			Filepath: "testdata/openedge_end.p",
			Expected: 0.05,
		},
		"end procedure": {
			Filepath: "testdata/openedge_end_procedure.p",
			Expected: 0.05,
		},
		"else do": {
			Filepath: "testdata/openedge_else_do.p",
			Expected: 0.05,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.OpenEdgeABL{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
