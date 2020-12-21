package deps_test

import (
	"os"
	"testing"

	"github.com/alecthomas/chroma/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/deps"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestParserElm_Parse(t *testing.T) {
	lexer := lexers.Get(heartbeat.LanguageElm.StringChroma())

	f, err := os.Open("testdata/elm.elm")
	require.NoError(t, err)

	parser := deps.ParserElm{}

	dependencies, err := parser.Parse(f, lexer)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"Color",
		"Dict",
		"TempFontAwesome",
		"Html",
		"Html",
		"Markdown",
		"String",
	}, dependencies)
}
