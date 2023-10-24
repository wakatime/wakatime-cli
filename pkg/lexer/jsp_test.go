package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestJSP_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/jsp_basic.jsp")
	assert.NoError(t, err)

	l := lexer.JSP{}.Lexer()

	assert.Equal(t, float32(0.49), l.AnalyseText(string(data)))
}
