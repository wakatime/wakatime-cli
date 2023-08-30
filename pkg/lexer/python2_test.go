package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestPython2_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/python2_shebang.py")
	assert.NoError(t, err)

	l := lexer.Python2{}.Lexer()

	assert.Equal(t, float32(1.0), l.AnalyseText(string(data)))
}
