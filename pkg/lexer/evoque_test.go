package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestEvoque_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/evoque_basic.evoque")
	assert.NoError(t, err)

	l := lexer.Evoque{}.Lexer()

	assert.Equal(t, float32(1.0), l.AnalyseText(string(data)))
}
