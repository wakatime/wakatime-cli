package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVelocity_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"macro": {
			Filepath: "testdata/velocity_macro.vm",
			Expected: 0.26,
		},
		"if": {
			Filepath: "testdata/velocity_if.vm",
			Expected: 0.16,
		},
		"foreach": {
			Filepath: "testdata/velocity_foreach.vm",
			Expected: 0.16,
		},
		"reference": {
			Filepath: "testdata/velocity_reference.vm",
			Expected: 0.01,
		},
		"all": {
			Filepath: "testdata/velocity_all.vm",
			Expected: 0.16,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexers.Get(heartbeat.LanguageVelocity.StringChroma())
			require.NotNil(t, l)

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
