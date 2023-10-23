package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestVBNet_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"module": {
			Filepath: "testdata/vb_module.vb",
			Expected: 0.5,
		},
		"namespace": {
			Filepath: "testdata/vb_namespace.vb",
			Expected: 0.5,
		},
		"if": {
			Filepath: "testdata/vb_if.vb",
			Expected: 0.5,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.VBNet{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
