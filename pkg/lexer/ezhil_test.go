package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEzhil_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/ezhil_basic.n")
	assert.NoError(t, err)

	l := lexers.Get(heartbeat.LanguageEzhil.StringChroma())
	require.NotNil(t, l)

	assert.Equal(t, float32(0.25), l.AnalyseText(string(data)))
}
