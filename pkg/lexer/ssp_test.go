package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestSSP_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/ssp_basic.ssp")
	assert.NoError(t, err)

	l := lexer.SSP{}.Lexer()

	assert.Equal(t, float32(0.9), l.AnalyseText(string(data)))
}
