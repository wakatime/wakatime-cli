package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestSlurm_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/slurm.sl")
	assert.NoError(t, err)

	l := lexer.Slurm{}.Lexer()

	assert.Equal(t, float32(1.0), l.AnalyseText(string(data)))
}
