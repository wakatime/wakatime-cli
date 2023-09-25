package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLimbo_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/limbo_basic.b")
	assert.NoError(t, err)

	l := lexers.Get(heartbeat.LanguageLimbo.StringChroma())
	require.NotNil(t, l)

	assert.Equal(t, float32(0.7), l.AnalyseText(string(data)))
}
