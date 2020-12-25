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

func TestParserPHP_Parse(t *testing.T) {
	lexer := lexers.Get(heartbeat.LanguagePHP.StringChroma())

	f, err := os.Open("testdata/php.php")
	require.NoError(t, err)

	parser := deps.ParserPHP{}

	dependencies, err := parser.Parse(f, lexer)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"Interop",
		"'ServiceLocator.php'",
		"'ServiceLocatorTwo.php'",
		"FooBarOne",
		"FooBarTwo",
		"ArrayObject",
		"FooBarThree",
		"FooBarFour",
		"FooBarSeven",
		"FooBarEight",
	}, dependencies)
}
