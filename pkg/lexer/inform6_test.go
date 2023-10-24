package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestInform6_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/inform6_basic.inf")
	assert.NoError(t, err)

	l := lexer.Inform6{}.Lexer()

	assert.Equal(t, float32(0.05), l.AnalyseText(string(data)))
}
