package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestObjectiveC_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"keyword_end": {
			Filepath: "testdata/objectivec_keyword_end.m",
			Expected: 1.0,
		},
		"keyword_implementation": {
			Filepath: "testdata/objectivec_keyword_implementation.m",
			Expected: 1.0,
		},
		"keyword_protocol": {
			Filepath: "testdata/objectivec_keyword_protocol.m",
			Expected: 1.0,
		},
		"nsstring": {
			Filepath: "testdata/objectivec_nsstring.m",
			Expected: 0.8,
		},
		"nsnumber": {
			Filepath: "testdata/objectivec_nsnumber.m",
			Expected: 0.7,
		},
		"message": {
			Filepath: "testdata/objectivec_message.m",
			Expected: 0.8,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.ObjectiveC{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
