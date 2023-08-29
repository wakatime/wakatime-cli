package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestCa65Assembler_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/ca65assembler_comment.s")
	assert.NoError(t, err)

	l := lexer.Ca65Assembler{}.Lexer()

	assert.Equal(t, float32(0.9), l.AnalyseText(string(data)))
}
