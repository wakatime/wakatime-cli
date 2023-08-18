package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestTransactSQL_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"declare": {
			Filepath: "testdata/transactsql_declare.sql",
			Expected: 1.0,
		},
		"bracket": {
			Filepath: "testdata/transactsql_bracket.sql",
			Expected: 0.5,
		},
		"variable": {
			Filepath: "testdata/transactsql_variable.sql",
			Expected: 0.1,
		},
		"go": {
			Filepath: "testdata/transactsql_go.sql",
			Expected: 0.1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.TransactSQL{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
