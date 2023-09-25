package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerilog_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"reg": {
			Filepath: "testdata/verilog_reg.v",
			Expected: 0.1,
		},
		"wire": {
			Filepath: "testdata/verilog_wire.v",
			Expected: 0.1,
		},
		"assign": {
			Filepath: "testdata/verilog_assign.v",
			Expected: 0.1,
		},
		"all": {
			Filepath: "testdata/verilog_all.v",
			Expected: 0.3,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexers.Get(heartbeat.LanguageVerilog.StringChroma())
			require.NotNil(t, l)

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
