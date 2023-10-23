package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestJCL_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/jcl_job_header.jcl")
	assert.NoError(t, err)

	l := lexer.JCL{}.Lexer()

	assert.Equal(t, float32(1.0), l.AnalyseText(string(data)))
}
