package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlurm_AnalyseText(t *testing.T) {
	data, err := os.ReadFile("testdata/slurm.sl")
	assert.NoError(t, err)

	l := lexers.Get(heartbeat.LanguageSlurm.StringChroma())
	require.NotNil(t, l)

	assert.Equal(t, float32(1.0), l.AnalyseText(string(data)))
}
