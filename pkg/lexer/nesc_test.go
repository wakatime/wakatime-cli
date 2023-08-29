package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestNesc_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"include": {
			Filepath: "testdata/nesc_include.nc",
			Expected: 0.1,
		},
		"ifdef": {
			Filepath: "testdata/nesc_ifdef.nc",
			Expected: 0.1,
		},
		"ifndef": {
			Filepath: "testdata/nesc_ifndef.nc",
			Expected: 0.1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.NesC{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
