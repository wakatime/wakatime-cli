package deps_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/deps"
)

func TestParserCPP_Parse(t *testing.T) {
	parser := deps.ParserCPP{}

	dependencies, err := parser.Parse("testdata/cpp.cpp")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"openssl",
		"wakatime",
	}, dependencies)
}
