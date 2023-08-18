package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestTADS3_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"GameMainDef": {
			Filepath: "testdata/tads3_game_main_def.t",
			Expected: 0.2,
		},
		"__TADS keyword": {
			Filepath: "testdata/tads3_tads_keyword.t",
			Expected: 0.2,
		},
		"version info": {
			Filepath: "testdata/tads3_version_info.t",
			Expected: 0.1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.TADS3{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
