package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestHTML_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/html_doctype.html")
	assert.NoError(t, err)

	l := lexer.HTML{}.Lexer()

	assert.Equal(t, float32(0.5), l.AnalyseText(string(data)))
}
