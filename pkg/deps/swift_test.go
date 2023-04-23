package deps_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserSwift_Parse(t *testing.T) //TestParserSwift_Parse tests the Parse function of the ParserSwift struct.
func TestParserSwift_Parse(t *testing.T) {
	t.Helper()
{
	parser := deps.ParserSwift{}

	dependencies, err := parser.Parse("testdata/swift.swift")
	require.NoError(t, err)

	require.ElementsMatch(t, []string{
		"UIKit",
		"PromiseKit",
	}, dependencies)
}
