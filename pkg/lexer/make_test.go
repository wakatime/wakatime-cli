package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestMakefile_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/makefile")
	assert.NoError(t, err)

	l := lexer.Makefile{}.Lexer()

	assert.Equal(t, float32(0.1), l.AnalyseText(string(data)))
}
