package deps_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserElm_Parse(t *testing.T) {
	parser := deps.ParserElm{}

	dependencies, err := parser.Parse("testdata/elm.elm")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"Color",
		"Dict",
		"TempFontAwesome",
		"Html",
		"Html",
		"Markdown",
		"String",
	}, dependencies)
}
