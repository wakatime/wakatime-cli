package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestActionScript3_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"basic": {
			Filepath: "testdata/actionscript3.as",
			Expected: 0.3,
		},
		"capital letters": {
			Filepath: "testdata/actionscript3_capital_letter.as",
			Expected: 0.3,
		},
		"spaces": {
			Filepath: "testdata/actionscript3_spaces.as",
			Expected: 0.3,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.ActionScript3{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
