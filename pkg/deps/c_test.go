package deps_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserC_Parse(t *testing.T) {
	parser := deps.ParserC{}

	dependencies, err := parser.Parse("testdata/c.c")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"math",
		"openssl",
	}, dependencies)
}
