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

func TestParserJavaScript_Parse(t *testing.T) {
	tests := map[string]struct {
		Lexer    chroma.Lexer
		Filepath string
		Expected []string
	}{
		"js": {
			Lexer:    lexers.Get(heartbeat.LanguageJavaScript.StringChroma()),
			Filepath: "testdata/es6.js",
			Expected: []string{
				"bravo",
				"foxtrot",
				"india",
				"kilo",
				"november",
				"oscar",
				"quebec",
				"tango",
				"uniform",
				"victor",
				"whiskey",
			},
		},
		"typescript": {
			Lexer:    lexers.Get(heartbeat.LanguageTypeScript.StringChroma()),
			Filepath: "testdata/typescript.ts",
			Expected: []string{
				"bravo",
				"foxtrot",
				"india",
				"kilo",
				"november",
				"oscar",
				"quebec",
				"tango",
				"uniform",
				"victor",
				"whiskey",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parser := deps.ParserJavaScript{}

			dependencies, err := parser.Parse(test.Filepath)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, dependencies)
		})
	}
}
