package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestLogos_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/logos_basic.xm")
	assert.NoError(t, err)

	l := lexer.Logos{}.Lexer()

	assert.Equal(t, float32(1.0), l.AnalyseText(string(data)))
}
