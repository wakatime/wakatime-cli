package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestBBUGS_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/bugs_basic.bug")
	assert.NoError(t, err)

	l := lexer.BUGS{}.Lexer()

	assert.Equal(t, float32(0.7), l.AnalyseText(string(data)))
}
