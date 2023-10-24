package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestTeraTerm_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/teraterm_commands.ttl")
	assert.NoError(t, err)

	l := lexer.TeraTerm{}.Lexer()

	assert.Equal(t, float32(0.01), l.AnalyseText(string(data)))
}
