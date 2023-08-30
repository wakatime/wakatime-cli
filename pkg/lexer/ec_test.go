package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestEC_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"include": {
			Filepath: "testdata/ec_include.ec",
			Expected: 0.1,
		},
		"ifdef": {
			Filepath: "testdata/ec_ifdef.ec",
			Expected: 0.1,
		},
		"ifndef": {
			Filepath: "testdata/ec_ifndef.ec",
			Expected: 0.1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.EC{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
