package deps_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserJavaScript_Parse(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected []string
	}{
		"js": {
			Filepath: "testdata/es6.js",
			Expected: []string{
				"bravo",
				"foxtrot",
				"india",
				"kilo",
				"november",
				"oscar",
				"quebec",
				"tango",
				"uniform",
				"victor",
				"whiskey",
			},
		},
		"typescript": {
			Filepath: "testdata/typescript.ts",
			Expected: []string{
				"bravo",
				"foxtrot",
				"india",
				"kilo",
				"november",
				"oscar",
				"quebec",
				"tango",
				"uniform",
				"victor",
				"whiskey",
			},
		},
		"react js": {
			Filepath: "testdata/react.jsx",
			Expected: []string{
				"react",
				"react-dom",
			},
		},
		"react typescript": {
			Filepath: "testdata/react.tsx",
			Expected: []string{
				"head",
				"react",
				"contants",
				"Footer",
				"Nav",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parser := deps.ParserJavaScript{}

			dependencies, err := parser.Parse(test.Filepath)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, dependencies)
		})
	}
}
