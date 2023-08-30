package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestR_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/r_expression.r")
	assert.NoError(t, err)

	l := lexer.R{}.Lexer()

	assert.Equal(t, float32(0.11), l.AnalyseText(string(data)))
}
