package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestTurtle_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/turtle_basic.ttl")
	assert.NoError(t, err)

	l := lexer.Turtle{}.Lexer()

	assert.Equal(t, float32(0.8), l.AnalyseText(string(data)))
}
