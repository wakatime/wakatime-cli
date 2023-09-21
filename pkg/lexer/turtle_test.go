package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTurtle_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/turtle_basic.ttl")
	assert.NoError(t, err)

	l := lexers.Get(heartbeat.LanguageTurtle.StringChroma())
	require.NotNil(t, l)

	assert.Equal(t, float32(0.8), l.AnalyseText(string(data)))
}
