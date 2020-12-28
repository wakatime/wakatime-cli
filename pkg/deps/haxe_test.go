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

func TestParserHaxe_Parse(t *testing.T) {
	lexer := lexers.Get(heartbeat.LanguageHaxe.StringChroma())

	f, err := os.Open("testdata/haxe.hx")
	require.NoError(t, err)

	parser := deps.ParserHaxe{}

	dependencies, err := parser.Parse(f, lexer)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"alpha",
		"bravo",
		"Math",
		"charlie",
		"delta",
	}, dependencies)
}
