package setup_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/log/setup"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/alecthomas/assert"
	"github.com/stretchr/testify/require"
)

func TestLogging(t *testing.T) {
	logFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer logFile.Close()

	v := vipertools.MustNew()
	v.Set("log-file", logFile.Name())
	v.Set("send-diagnostics-on-errors", true)
	v.Set("metrics", true)
	v.Set("verbose", true)

	logger, err := setup.Logging(t.Context(), v)
	require.NoError(t, err)

	assert.NotNil(t, logger)

	assert.True(t, logger.SendDiagsOnErrors())
	assert.True(t, logger.IsMetricsEnabled())
	assert.True(t, logger.IsVerboseEnabled())
}
