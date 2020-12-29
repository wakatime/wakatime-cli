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

func TestParserC_Parse(t *testing.T) {
	lexer := lexers.Get(heartbeat.LanguageC.StringChroma())

	f, err := os.Open("testdata/c.c")
	require.NoError(t, err)

	parser := deps.ParserC{}

	dependencies, err := parser.Parse(f, lexer)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"math",
		"openssl",
	}, dependencies)
}
