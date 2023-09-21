package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQBasic_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"dynamic_cmd": {
			Filepath: "testdata/qbasic_dynamiccmd.bas",
			Expected: 0.9,
		},
		"static_cmd": {
			Filepath: "testdata/qbasic_staticcmd.bas",
			Expected: 0.9,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexers.Get(heartbeat.LanguageQBasic.StringChroma())
			require.NotNil(t, l)

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
