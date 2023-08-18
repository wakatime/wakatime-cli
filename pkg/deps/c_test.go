package deps_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserC_Parse(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected []string
	}{
		"c": {
			Filepath: "testdata/c.c",
			Expected: []string{
				"math",
				"openssl",
			},
		},
		"cpp": {
			Filepath: "testdata/cpp.cpp",
			Expected: []string{
				"iostream",
				"openssl",
				"wakatime",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parser := deps.ParserC{}

			dependencies, err := parser.Parse(test.Filepath)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, dependencies)
		})
	}
}
