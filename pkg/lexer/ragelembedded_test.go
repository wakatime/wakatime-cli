package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestRagelEmbedded_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/ragel.rl")
	assert.NoError(t, err)

	l := lexer.RagelEmbedded{}.Lexer()

	assert.Equal(t, float32(1.0), l.AnalyseText(string(data)))
}
