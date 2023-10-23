package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestSingularity_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"only header": {
			Filepath: "testdata/singularity_only_header.def",
			Expected: 0.5,
		},
		"only section": {
			Filepath: "testdata/singularity_only_section.def",
			Expected: 0.49,
		},
		"full": {
			Filepath: "testdata/singularity_full.def",
			Expected: 0.99,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.Singularity{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
