package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestXML_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/xml_doctype_html.xml")
	assert.NoError(t, err)

	l := lexer.XML{}.Lexer()

	assert.Equal(t, float32(0.45), l.AnalyseText(string(data)))
}
