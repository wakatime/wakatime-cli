package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestEzhil_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/ezhil_basic.n")
	assert.NoError(t, err)

	l := lexer.Ezhil{}.Lexer()

	assert.Equal(t, float32(0.25), l.AnalyseText(string(data)))
}
