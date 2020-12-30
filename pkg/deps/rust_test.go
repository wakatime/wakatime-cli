package deps_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserRust_Parse(t *testing.T) {
	parser := deps.ParserRust{}

	dependencies, err := parser.Parse("testdata/rust.rs")
	require.NoError(t, err)

	assert.Equal(t, []string{
		"proc_macro",
		"phrases",
		"syn",
		"quote",
	}, dependencies)
}
