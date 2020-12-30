package deps_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserCSharp_Parse(t *testing.T) {
	parser := deps.ParserCSharp{}

	dependencies, err := parser.Parse("testdata/csharp.cs")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"WakaTime",
		"Math",
		"Fart",
		"Proper",
	}, dependencies)
}
