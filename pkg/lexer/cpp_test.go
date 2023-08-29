package lexer_test

import (
	"os"
	"testing"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCpp_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"include": {
			Filepath: "testdata/cpp_include.cpp",
			Expected: 0.2,
		},
		"namespace": {
			Filepath: "testdata/cpp_namespace.cpp",
			Expected: 0.4,
		},
	}

	l := lexers.Get(heartbeat.LanguageCPP.StringChroma())
	require.NotNil(t, l)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
