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

func TestParserCSharp_Parse(t *testing.T) {
	lexer := lexers.Get(heartbeat.LanguageCSharp.StringChroma())

	f, err := os.Open("testdata/csharp.cs")
	require.NoError(t, err)

	parser := deps.ParserCSharp{}

	dependencies, err := parser.Parse(f, lexer)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"WakaTime",
		"Math",
		"Fart",
		"Proper",
	}, dependencies)
}
