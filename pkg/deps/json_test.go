package deps_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserJSON_Parse(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected []string
	}{
		"bower": {
			Filepath: "testdata/bower.json",
			Expected: []string{
				"bower",
				"animate.css",
				"bootstrap",
				"bootstrap-daterangepicker",
				"moment",
				"moment-timezone",
			},
		},
		"component": {
			Filepath: "testdata/component.json",
			Expected: []string{
				"bower",
				"component/emitter",
				"component/jquery",
			},
		},
		"package": {
			Filepath: "testdata/package.json",
			Expected: []string{
				"npm",
				"wakatime",
				"another_dep",
				"test_framework",
				"another_dev_dep",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parser := deps.ParserJSON{}

			dependencies, err := parser.Parse(test.Filepath)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, dependencies)
		})
	}
}
