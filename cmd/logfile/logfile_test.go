package logfile_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/cmd/logfile"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadParams(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	ctx := context.Background()

	v := viper.New()
	v.Set("log-file", tmpFile.Name())
	v.Set("log-to-stdout", true)
	v.Set("metrics", true)
	v.Set("verbose", true)
	v.Set("send-diagnostics-on-errors", true)

	params, err := logfile.LoadParams(ctx, v)
	require.NoError(t, err)

	assert.True(t, params.Verbose)
	assert.True(t, params.Metrics)
	assert.True(t, params.ToStdout)
	assert.True(t, params.SendDiagsOnErrors)
	assert.Equal(t, tmpFile.Name(), params.File)
}

func TestLoadParams_LogFile_FlagDeprecated(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	ctx := context.Background()

	v := viper.New()
	v.Set("logfile", tmpFile.Name())

	params, err := logfile.LoadParams(ctx, v)
	require.NoError(t, err)

	assert.Equal(t, tmpFile.Name(), params.File)
}

func TestLoadParams_LogFile_FromConfig(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	ctx := context.Background()

	v := viper.New()
	v.Set("settings.log_file", tmpFile.Name())

	params, err := logfile.LoadParams(ctx, v)
	require.NoError(t, err)

	assert.Equal(t, tmpFile.Name(), params.File)
}

func TestLoadParams_LogFile_FromEnvVar(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	dir, _ := filepath.Split(tmpFile.Name())

	ctx := context.Background()

	v := viper.New()

	t.Setenv("WAKATIME_HOME", dir)

	params, err := logfile.LoadParams(ctx, v)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(dir, "wakatime.log"), params.File)
}

func TestLoadParams_LogFile_FlagTakesPrecedence(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	ctx := context.Background()

	v := viper.New()
	v.Set("log-file", tmpFile.Name())
	v.Set("settings.log_file", "otherfolder/wakatime.config.log")

	params, err := logfile.LoadParams(ctx, v)
	require.NoError(t, err)

	assert.Equal(t, tmpFile.Name(), params.File)
}

func TestLoadParams_Metrics_FromConfig(t *testing.T) {
	ctx := context.Background()

	v := viper.New()
	v.Set("settings.metrics", true)

	params, err := logfile.LoadParams(ctx, v)
	require.NoError(t, err)

	assert.True(t, params.Metrics)
}

func TestLoadParams_Metrics_FlagTakesPrecedence(t *testing.T) {
	ctx := context.Background()

	v := viper.New()
	v.Set("metrics", false)
	v.Set("settings.metrics", true)

	params, err := logfile.LoadParams(ctx, v)
	require.NoError(t, err)

	assert.False(t, params.Metrics)
}

func TestLoadParams_Verbose_FromConfig(t *testing.T) {
	ctx := context.Background()

	v := viper.New()
	v.Set("settings.debug", true)

	params, err := logfile.LoadParams(ctx, v)
	require.NoError(t, err)

	assert.True(t, params.Verbose)
}

func TestLoadParams_Verbose_FlagTakesPrecedence(t *testing.T) {
	ctx := context.Background()

	v := viper.New()
	v.Set("verbose", false)
	v.Set("settings.debug", true)

	params, err := logfile.LoadParams(ctx, v)
	require.NoError(t, err)

	assert.False(t, params.Verbose)
}
