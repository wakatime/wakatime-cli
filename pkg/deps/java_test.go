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

func TestParserJava_Parse(t *testing.T) {
	lexer := lexers.Get(heartbeat.LanguageJava.StringChroma())

	f, err := os.Open("testdata/java.java")
	require.NoError(t, err)

	parser := deps.ParserJava{}

	dependencies, err := parser.Parse(f, lexer)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"googlecode.javacv",
		"colorfulwolf.webcamapplet",
		"foobar",
		"apackage.something",
		"anamespace.other",
	}, dependencies)
}
