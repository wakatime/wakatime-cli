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

func TestParserPython_Parse(t *testing.T) {
	lexer := lexers.Get(heartbeat.LanguagePython.StringChroma())

	f, err := os.Open("testdata/python.py")
	require.NoError(t, err)

	parser := deps.ParserPython{}

	dependencies, err := parser.Parse(f, lexer)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"first",
		"second",
		"django",
		"app",
		"flask",
		"simplejson",
		"jinja",
		"pygments",
		"sqlalchemy",
		"mock",
		"unittest",
	}, dependencies)
}
