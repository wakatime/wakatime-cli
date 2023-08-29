package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestTASM_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/tasm.asm")
	assert.NoError(t, err)

	l := lexer.TASM{}.Lexer()

	assert.Equal(t, float32(1.0), l.AnalyseText(string(data)))
}
