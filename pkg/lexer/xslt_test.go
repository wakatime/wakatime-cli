package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestXSLT_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/xslt.xsl")
	assert.NoError(t, err)

	l := lexer.XSLT{}.Lexer()

	assert.Equal(t, float32(0.8), l.AnalyseText(string(data)))
}
