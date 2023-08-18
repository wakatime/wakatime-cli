package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestCoq_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/coq_reserved_keyword.v")
	assert.NoError(t, err)

	l := lexer.Coq{}.Lexer()

	assert.Equal(t, float32(1.0), l.AnalyseText(string(data)))
}
