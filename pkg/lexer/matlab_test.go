package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestMatlab_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"function": {
			Filepath: "testdata/matlab_function.m",
			Expected: 1.0,
		},
		"comment": {
			Filepath: "testdata/matlab_comment.m",
			Expected: 0.2,
		},
		"systemcmd": {
			Filepath: "testdata/matlab_systemcmd.m",
			Expected: 0.2,
		},
		"windows": {
			Filepath: "testdata/matlab_windows.m",
			Expected: 1.0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.Matlab{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
