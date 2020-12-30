package deps_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserJava_Parse(t *testing.T) {
	parser := deps.ParserJava{}

	dependencies, err := parser.Parse("testdata/java.java")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"googlecode.javacv",
		"colorfulwolf.webcamapplet",
		"foobar",
		"apackage.something",
		"anamespace.other",
	}, dependencies)
}
