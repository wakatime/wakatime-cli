package deps_test

import (
	"context"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserHaxe_Parse(t *testing.T) {
	parser := deps.ParserHaxe{}

	dependencies, err := parser.Parse(context.Background(), "testdata/haxe.hx")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"alpha",
		"bravo",
		"Math",
		"charlie",
		"delta",
	}, dependencies)
}
