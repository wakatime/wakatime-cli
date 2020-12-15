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

func TestParserGo_Parse_Minimal(t *testing.T) {
	lexer := lexers.Get(heartbeat.LanguageGo.StringChroma())

	f, err := os.Open("testdata/golang_minimal.go")
	require.NoError(t, err)

	parser := deps.ParserGo{}

	dependencies, err := parser.Parse(f, lexer)
	require.NoError(t, err)

	assert.Equal(t, []string{
		`"os"`,
		`"github.com/wakatime/wakatime-cli/pkg/heartbeat"`,
	}, dependencies)
}

func TestParserGo_Parse_All(t *testing.T) {
	lexer := lexers.Get(heartbeat.LanguageGo.StringChroma())

	f, err := os.Open("testdata/golang.go")
	require.NoError(t, err)

	parser := deps.ParserGo{}

	dependencies, err := parser.Parse(f, lexer)
	require.NoError(t, err)

	assert.Equal(t, []string{
		`"os"`,
		`"compress/gzip"`,
		`"github.com/golang/example/stringutil"`,
		`"log"`,
		`"os"`,
		`"oldname"`,
		`"direct"`,
		`"suppress"`,
		`"foobar"`,
		`"image/gif"`,
		`"math"`,
	}, dependencies)
}
