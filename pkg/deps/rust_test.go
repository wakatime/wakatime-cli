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

func TestParserRust_Parse(t *testing.T) {
	lexer := lexers.Get(heartbeat.LanguageRust.StringChroma())

	f, err := os.Open("testdata/rust.rs")
	require.NoError(t, err)

	parser := deps.ParserRust{}

	dependencies, err := parser.Parse(f, lexer)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"proc_macro",
		"phrases",
		"syn",
		"quote",
	}, dependencies)
}
