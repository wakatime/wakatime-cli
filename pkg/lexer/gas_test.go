package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGas_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"data directive": {
			Filepath: "testdata/gas_data_directive.S",
			Expected: 1.0,
		},
		"other directive": {
			Filepath: "testdata/gas_other_directive.S",
			Expected: 0.1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexers.Get(heartbeat.LanguageGas.StringChroma())
			require.NotNil(t, l)

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
