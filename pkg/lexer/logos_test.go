package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogos_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/logos_basic.xm")
	assert.NoError(t, err)

	l := lexers.Get(heartbeat.LanguageLogos.StringChroma())
	require.NotNil(t, l)

	assert.Equal(t, float32(1.0), l.AnalyseText(string(data)))
}
