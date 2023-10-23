package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestProlog_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/prolog.ecl")
	assert.NoError(t, err)

	l := lexer.Prolog{}.Lexer()

	assert.Equal(t, float32(1.0), l.AnalyseText(string(data)))
}
