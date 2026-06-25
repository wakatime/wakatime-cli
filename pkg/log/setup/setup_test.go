package setup_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/log/setup"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/stretchr/testify/assert"
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

func TestLogging_ToStdout(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("log-to-stdout", true)

	logger, err := setup.Logging(t.Context(), v)
	require.NoError(t, err)

	assert.NotNil(t, logger)
	assert.False(t, logger.SendDiagsOnErrors())
	assert.False(t, logger.IsMetricsEnabled())
	assert.False(t, logger.IsVerboseEnabled())
}

func TestLogging_CreatesLogDirectory(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "nested", "wakatime.log")

	v := vipertools.MustNew()
	v.Set("log-file", logFile)

	logger, err := setup.Logging(t.Context(), v)
	require.NoError(t, err)

	assert.NotNil(t, logger)

	info, err := os.Stat(filepath.Dir(logFile))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}
