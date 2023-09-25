package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCBMBasicV2_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/cbmbasicv2_basic.bas")
	assert.NoError(t, err)

	l := lexers.Get(heartbeat.LanguageCBMBasicV2.StringChroma())
	require.NotNil(t, l)

	assert.Equal(t, float32(0.2), l.AnalyseText(string(data)))
}
