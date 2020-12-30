package deps_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserJavaScript_Parse(t *testing.T) {
	lexer := lexers.Get(heartbeat.LanguageJavaScript.StringChroma())

	f, err := os.Open("testdata/es6.js")
	require.NoError(t, err)

	parser := deps.ParserJavaScript{}

	dependencies, err := parser.Parse(f, lexer)
	require.NoError(t, err)

	assert.Equal(t, []string{
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
	}, dependencies)
}
