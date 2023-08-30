package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestERB_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/erb_basic.erb")
	assert.NoError(t, err)

	l := lexer.ERB{}.Lexer()

	assert.Equal(t, float32(0.4), l.AnalyseText(string(data)))
}
