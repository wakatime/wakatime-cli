package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestJAGS_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"model only": {
			Filepath: "testdata/jags_model.jag",
			Expected: 0.3,
		},
		"model and data": {
			Filepath: "testdata/jags_data.jag",
			Expected: 0.9,
		},
		"model and var": {
			Filepath: "testdata/jags_var.jag",
			Expected: 0.9,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.JAGS{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
