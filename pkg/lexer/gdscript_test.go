package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestGdSript_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"func": {
			Filepath: "testdata/gdscript_func.gd",
			Expected: 0.8,
		},
		"keyword first group": {
			Filepath: "testdata/gdscript_keyword.gd",
			Expected: 0.4,
		},
		"keyword second group": {
			Filepath: "testdata/gdscript_keyword2.gd",
			Expected: 0.2,
		},
		"full": {
			Filepath: "testdata/gdscript_full.gd",
			Expected: 1.0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.GDScript{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
