package deps_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserSwift_Parse(t *testing.T) {
	parser := deps.ParserSwift{}

	dependencies, err := parser.Parse("testdata/swift.swift")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"UIKit",
		"PromiseKit",
	}, dependencies)
}
