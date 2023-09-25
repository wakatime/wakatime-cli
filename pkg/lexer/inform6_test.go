package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInform6_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/inform6_basic.inf")
	assert.NoError(t, err)

	l := lexers.Get(heartbeat.LanguageInform6.StringChroma())
	require.NotNil(t, l)

	assert.Equal(t, float32(0.05), l.AnalyseText(string(data)))
}
