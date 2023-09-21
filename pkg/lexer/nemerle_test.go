package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNermerle_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/nemerle_if.n")
	assert.NoError(t, err)

	l := lexers.Get(heartbeat.LanguageNemerle.StringChroma())
	require.NotNil(t, l)

	assert.Equal(t, float32(0.1), l.AnalyseText(string(data)))
}
