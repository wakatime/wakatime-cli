package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestAspxVBNet_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"page language": {
			Filepath: "testdata/aspxvbnet_page_language.aspx",
			Expected: 0.2,
		},
		"script language": {
			Filepath: "testdata/aspxvbnet_script_language.aspx",
			Expected: 0.15,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.AspxVBNet{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
