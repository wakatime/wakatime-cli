package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestLimbo_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/limbo_basic.b")
	assert.NoError(t, err)

	l := lexer.Limbo{}.Lexer()

	assert.Equal(t, float32(0.7), l.AnalyseText(string(data)))
}
