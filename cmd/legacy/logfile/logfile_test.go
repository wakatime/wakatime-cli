package logfile_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/cmd/legacy/logfile"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadParams(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "wakatime.log")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	dir, _ := filepath.Split(tmpFile.Name())

	_, err = os.Create(filepath.Join(dir, ".wakatime.log"))
	require.NoError(t, err)

	defer os.Remove(filepath.Join(dir, ".wakatime.log"))

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := map[string]struct {
		ViperLogFile       string
		ViperLogFileConfig string
		ViperLogFileOld    string
		ViperToStdout      bool
		EnvVar             string
		ViperDebug         bool
		ViperDebugConfig   bool
		Expected           logfile.Params
	}{
		"log file and verbose set": {
			ViperDebug: true,
			Expected: logfile.Params{
				File:    filepath.Join(home, ".wakatime.log"),
				Verbose: true,
			},
		},
		"log file and verbose from config": {
			ViperDebugConfig: true,
			Expected: logfile.Params{
				File:    filepath.Join(home, ".wakatime.log"),
				Verbose: true,
			},
		},
		"log file flag takes preceedence": {
			ViperLogFile:       tmpFile.Name(),
			ViperLogFileConfig: "otherfolder/wakatime.config.log",
			ViperLogFileOld:    "otherfolder/wakatime.old.log",
			Expected: logfile.Params{
				File: tmpFile.Name(),
			},
		},
		"log file deprecated flag takes preceedence": {
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
				File: filepath.Join(dir, ".wakatime.log"),
			},
		},
		"log file from home dir": {
			Expected: logfile.Params{
				File: filepath.Join(home, ".wakatime.log"),
			},
		},
		"log to stdout": {
			ViperToStdout: true,
			Expected: logfile.Params{
				File:     filepath.Join(home, ".wakatime.log"),
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
			v.Set("settings.log_file", test.ViperLogFileConfig)
			v.Set("settings.debug", test.ViperDebug)
			v.Set("verbose", test.ViperDebugConfig)

			err := os.Setenv("WAKATIME_HOME", test.EnvVar)
			require.NoError(t, err)

			params, err := logfile.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params)
		})
	}
}
