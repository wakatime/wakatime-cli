package deps_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserScala_Parse(t *testing.T) {
	parser := deps.ParserScala{}

	dependencies, err := parser.Parse("testdata/scala.scala")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"com.alpha.SomeClass",
		"com.bravo.something",
		"com.charlie",
		"golf",
		"com.hotel.india",
		"juliett.kilo",
	}, dependencies)
}
