package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestPawn_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/pawn_tagof.pwn")
	assert.NoError(t, err)

	l := lexer.Pawn{}.Lexer()

	assert.Equal(t, float32(0.01), l.AnalyseText(string(data)))
}
