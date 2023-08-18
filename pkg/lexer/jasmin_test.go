package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestJasmin_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"class": {
			Filepath: "testdata/jasmin_class.j",
			Expected: 0.5,
		},
		"instruction": {
			Filepath: "testdata/jasmin_instruction.j",
			Expected: 0.8,
		},
		"keyword": {
			Filepath: "testdata/jasmin_keyword.j",
			Expected: 0.6,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.Jasmin{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
