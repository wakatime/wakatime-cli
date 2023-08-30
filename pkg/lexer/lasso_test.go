package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestLasso_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"shebang": {
			Filepath: "testdata/lasso_shebang.lasso",
			Expected: 0.8,
		},
		"delimiter": {
			Filepath: "testdata/lasso_delimiter.lasso",
			Expected: 0.4,
		},
		"local": {
			Filepath: "testdata/lasso_local.lasso",
			Expected: 0.4,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.Lasso{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
