package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestSourcesList_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"standard": {
			Filepath: "testdata/sources.list",
			Expected: 1.0,
		},
		"indented": {
			Filepath: "testdata/sources-indented.list",
			Expected: 1.0,
		},
		"invalid": {
			Filepath: "testdata/sources-invalid.list",
			Expected: 0.0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.SourcesList{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
