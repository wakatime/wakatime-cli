package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestNermerle_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/nemerle_if.n")
	assert.NoError(t, err)

	l := lexer.Nemerle{}.Lexer()

	assert.Equal(t, float32(0.1), l.AnalyseText(string(data)))
}
