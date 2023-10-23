package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestRSL_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/raise.rsl")
	assert.NoError(t, err)

	l := lexer.RSL{}.Lexer()

	assert.Equal(t, float32(1.0), l.AnalyseText(string(data)))
}
