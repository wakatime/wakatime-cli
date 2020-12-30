package deps_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserKotlin_Parse(t *testing.T) {
	parser := deps.ParserKotlin{}

	dependencies, err := parser.Parse("testdata/kotlin.kt")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"alpha.time",
		"bravo.charlie",
		"delta.io",
		"echo.Foxtrot",
		"h",
	}, dependencies)
}
