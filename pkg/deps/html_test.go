package deps_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserHTML_Parse(t *testing.T) {
	parser := deps.ParserHTML{}

	dependencies, err := parser.Parse("testdata/html.html")
	require.NoError(t, err)

	assert.Equal(t, []string{
		`"wakatime.js"`,
		`"../scripts/wakatime.js"`,
		`"https://www.wakatime.com/scripts/my.js"`,
		"\"this is a\n multiline value\"",
	}, dependencies)
}

func TestParserHTML_Parse_Django(t *testing.T) {
	parser := deps.ParserHTML{}

	dependencies, err := parser.Parse("testdata/html_django.html")
	require.NoError(t, err)

	assert.Equal(t, []string{
		`"libs/json2.js"`,
	}, dependencies)
}

func TestParserHTML_Parse_WithPHP(t *testing.T) {
	parser := deps.ParserHTML{}

	dependencies, err := parser.Parse("testdata/html_with_php.html")
	require.NoError(t, err)

	assert.Equal(t, []string{
		`"https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/js/bootstrap.min.js"`,
	}, dependencies)
}
