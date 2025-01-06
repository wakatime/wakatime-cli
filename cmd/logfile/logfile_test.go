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

	dir, _ := filepath.Split(tmpFile.Name())

	logFile, err := os.Create(filepath.Join(dir, "wakatime.log"))
	require.NoError(t, err)

	defer logFile.Close()

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	ctx := context.Background()

	tests := map[string]struct {
		EnvVar             string
		ViperDebug         bool
		ViperDebugConfig   bool
		ViperLogFile       string
		ViperLogFileConfig string
		ViperLogFileOld    string
		ViperMetrics       bool
		ViperMetricsConfig bool
		ViperToStdout      bool
		Expected           logfile.Params
	}{
		"verbose set": {
			ViperDebug: true,
			Expected: logfile.Params{
				File:    filepath.Join(home, ".wakatime", "wakatime.log"),
				Verbose: true,
			},
		},
		"verbose from config": {
			ViperDebugConfig: true,
			Expected: logfile.Params{
				File:    filepath.Join(home, ".wakatime", "wakatime.log"),
				Verbose: true,
			},
		},
		"log file flag takes precedence": {
			ViperLogFile:       tmpFile.Name(),
			ViperLogFileConfig: "otherfolder/wakatime.config.log",
			ViperLogFileOld:    "otherfolder/wakatime.old.log",
			Expected: logfile.Params{
				File: tmpFile.Name(),
			},
		},
		"log file deprecated flag takes precedence": {
			ViperLogFileConfig: "otherfolder/wakatime.config.log",
			ViperLogFileOld:    tmpFile.Name(),
			Expected: logfile.Params{
				File: tmpFile.Name(),
			},
		},
		"log file from config": {
			ViperLogFileConfig: tmpFile.Name(),
			Expected: logfile.Params{
				File: tmpFile.Name(),
			},
		},
		"log file from WAKATIME_HOME": {
			EnvVar: dir,
			Expected: logfile.Params{
				File: filepath.Join(dir, "wakatime.log"),
			},
		},
		"log file from home dir": {
			Expected: logfile.Params{
				File: filepath.Join(home, ".wakatime", "wakatime.log"),
			},
		},
		"metrics set": {
			ViperMetrics: true,
			Expected: logfile.Params{
				File:    filepath.Join(home, ".wakatime", "wakatime.log"),
				Metrics: true,
			},
		},
		"metrics from config": {
			ViperMetricsConfig: true,
			Expected: logfile.Params{
				File:    filepath.Join(home, ".wakatime", "wakatime.log"),
				Metrics: true,
			},
		},
		"metrics flag takes precedence": {
			ViperMetrics:       true,
			ViperMetricsConfig: false,
			Expected: logfile.Params{
				File:    filepath.Join(home, ".wakatime", "wakatime.log"),
				Metrics: true,
			},
		},
		"log to stdout": {
			ViperToStdout: true,
			Expected: logfile.Params{
				File:     filepath.Join(home, ".wakatime", "wakatime.log"),
				ToStdout: true,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("log-file", test.ViperLogFile)
			v.Set("logfile", test.ViperLogFileOld)
			v.Set("log-to-stdout", test.ViperToStdout)
			v.Set("metrics", test.ViperMetrics)
			v.Set("settings.metrics", test.ViperMetricsConfig)
			v.Set("settings.log_file", test.ViperLogFileConfig)
			v.Set("settings.debug", test.ViperDebug)
			v.Set("verbose", test.ViperDebugConfig)

			t.Setenv("WAKATIME_HOME", test.EnvVar)

			params, err := logfile.LoadParams(ctx, v)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params)
		})
	}
}
