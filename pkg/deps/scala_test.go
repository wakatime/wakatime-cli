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

func TestParserScala_Parse(t *testing.T) {
	lexer := lexers.Get(heartbeat.LanguageScala.StringChroma())

	f, err := os.Open("testdata/scala.scala")
	require.NoError(t, err)

	parser := deps.ParserScala{}

	dependencies, err := parser.Parse(f, lexer)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"com.alpha.SomeClass",
		"com.bravo.something",
		"com.charlie",
		"golf",
		"com.hotel.india",
		"juliett.kilo",
	}, dependencies)
}
