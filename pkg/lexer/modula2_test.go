package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestModula2_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"pascal flavour": {
			Filepath: "testdata/modula2_pascal.def",
			Expected: 0,
		},
		"pascal flavour with function": {
			Filepath: "testdata/modula2_pascal_function.def",
			Expected: 0,
		},
		"basic": {
			Filepath: "testdata/modula2_basic.def",
			Expected: 0.6,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.Modula2{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
