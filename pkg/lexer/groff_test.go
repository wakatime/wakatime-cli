package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestGroff_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"comment": {
			Filepath: "testdata/groff_comment.man",
			Expected: 1.0,
		},
		"title head": {
			Filepath: "testdata/groff_title_head.man",
			Expected: 1.0,
		},
		"macro": {
			Filepath: "testdata/groff_macro.man",
			Expected: 0.9,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.Groff{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
