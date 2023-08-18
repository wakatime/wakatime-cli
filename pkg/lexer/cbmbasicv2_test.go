package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestCBMBasicV2_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/cbmbasicv2_basic.bas")
	assert.NoError(t, err)

	l := lexer.CBMBasicV2{}.Lexer()

	assert.Equal(t, float32(0.2), l.AnalyseText(string(data)))
}
