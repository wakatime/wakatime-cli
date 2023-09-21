package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSP_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/ssp_basic.ssp")
	assert.NoError(t, err)

	l := lexers.Get(heartbeat.LanguageSSP.StringChroma())
	require.NotNil(t, l)

	assert.Equal(t, float32(0.9), l.AnalyseText(string(data)))
}
