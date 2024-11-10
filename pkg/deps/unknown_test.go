package deps_test

import (
	"context"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserUnknown_Parse(t *testing.T) {
	ctx := context.Background()

	tests := map[string]struct {
		Filepath string
		Expected []string
	}{
		"bower": {
			Filepath: "testdata/bower.json",
			Expected: []string{
				"bower",
			},
		},
		"grunt": {
			Filepath: "testdata/Gruntfile",
			Expected: []string{
				"grunt",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parser := deps.ParserUnknown{}

			dependencies, err := parser.Parse(ctx, test.Filepath)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, dependencies)
		})
	}
}
