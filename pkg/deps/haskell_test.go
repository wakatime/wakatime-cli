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

func TestParserHaskell_Parse(t *testing.T) {
	lexer := lexers.Get(heartbeat.LanguageHaskell.StringChroma())

	f, err := os.Open("testdata/haskell.hs")
	require.NoError(t, err)

	parser := deps.ParserHaskell{}

	dependencies, err := parser.Parse(f, lexer)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"Control",
		"Data",
		"Network",
		"System",
	}, dependencies)
}
