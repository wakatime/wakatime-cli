package deps_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserSwift_Parse(t *testing.T) {
	lexer := lexers.Get(heartbeat.LanguageSwift.StringChroma())

	f, err := os.Open("testdata/swift.swift")
	require.NoError(t, err)

	parser := deps.ParserSwift{}

	dependencies, err := parser.Parse(f, lexer)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"UIKit",
		"PromiseKit",
	}, dependencies)
}
