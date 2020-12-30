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

func TestParserKotlin_Parse(t *testing.T) {
	lexer := lexers.Get(heartbeat.LanguageKotlin.StringChroma())

	f, err := os.Open("testdata/kotlin.kt")
	require.NoError(t, err)

	parser := deps.ParserKotlin{}

	dependencies, err := parser.Parse(f, lexer)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"alpha.time",
		"bravo.charlie",
		"delta.io",
		"echo.Foxtrot",
		"h",
	}, dependencies)
}
