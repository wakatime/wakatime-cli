package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForth_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/forth_command.frt")
	assert.NoError(t, err)

	l := lexers.Get(heartbeat.LanguageForth.StringChroma())
	require.NotNil(t, l)

	assert.Equal(t, float32(0.3), l.AnalyseText(string(data)))
}
