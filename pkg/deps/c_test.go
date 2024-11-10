package deps_test

import (
	"context"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserC_Parse(t *testing.T) {
	parser := deps.ParserC{}

	dependencies, err := parser.Parse(context.Background(), "testdata/c.c")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"math",
		"openssl",
	}, dependencies)
}
