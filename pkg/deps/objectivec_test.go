package deps_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserObjectiveC_Parse(t *testing.T) {
	parser := deps.ParserObjectiveC{}

	dependencies, err := parser.Parse("testdata/objective_c.m")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"SomeViewController",
		"OtherViewController",
		"UIKit",
		"PromiseKit",
	}, dependencies)
}
