package deps_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserC_Parse(t *testing.T) {
	tests := map[string]struct {
		Lexer    chroma.Lexer
		Filepath string
		Expected []string
	}{
		"c": {
			Lexer:    lexers.Get(heartbeat.LanguageC.StringChroma()),
			Filepath: "testdata/c.c",
			Expected: []string{
				"math",
				"openssl",
			},
		},
		"cpp": {
			Lexer:    lexers.Get(heartbeat.LanguageCPP.StringChroma()),
			Filepath: "testdata/cpp.cpp",
			Expected: []string{
				"iostream",
				"openssl",
				"wakatime",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parser := deps.ParserC{}

			dependencies, err := parser.Parse(test.Filepath)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, dependencies)
		})
	}
}
