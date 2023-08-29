package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestUcode_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"endsuspend": {
			Filepath: "testdata/ucode_endsuspend.u",
			Expected: 0.1,
		},
		"endrepeat": {
			Filepath: "testdata/ucode_endrepeat.u",
			Expected: 0.1,
		},
		"variable set": {
			Filepath: "testdata/ucode_varset.u",
			Expected: 0.01,
		},
		"procedure": {
			Filepath: "testdata/ucode_procedure.u",
			Expected: 0.01,
		},
		"self": {
			Filepath: "testdata/ucode_self.u",
			Expected: 0.5,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.Ucode{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
