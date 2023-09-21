package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCa65Assembler_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/ca65assembler_comment.s")
	assert.NoError(t, err)

	l := lexers.Get(heartbeat.LanguageCa65Assembler.StringChroma())
	require.NotNil(t, l)

	assert.Equal(t, float32(0.9), l.AnalyseText(string(data)))
}
