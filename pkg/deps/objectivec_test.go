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

func TestParserObjectiveC_Parse(t *testing.T) {
	lexer := lexers.Get(heartbeat.LanguageObjectiveC.StringChroma())

	f, err := os.Open("testdata/objective_c.m")
	require.NoError(t, err)

	parser := deps.ParserObjectiveC{}

	dependencies, err := parser.Parse(f, lexer)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"SomeViewController",
		"OtherViewController",
		"UIKit",
		"PromiseKit",
	}, dependencies)
}
