package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestSmali_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"class": {
			Filepath: "testdata/smali_class.smali",
			Expected: 0.5,
		},
		"class with keyword": {
			Filepath: "testdata/smali_class_keyword.smali",
			Expected: 0.8,
		},
		"keyword": {
			Filepath: "testdata/smali_keyword.smali",
			Expected: 0.6,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.Smali{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
