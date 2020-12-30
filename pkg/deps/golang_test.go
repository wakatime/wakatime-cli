package deps_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserGo_Parse(t *testing.T) {
	parser := deps.ParserGo{}

	dependencies, err := parser.Parse("testdata/golang.go")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"compress/gzip",
		"github.com/golang/example/stringutil",
		"log",
		"os",
		"oldname",
		"direct",
		"suppress",
		"foobar",
		"image/gif",
		"math",
	}, dependencies)
}
