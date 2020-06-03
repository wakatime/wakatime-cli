package logfile_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/cmd/legacy/logfile"

	jww "github.com/spf13/jwalterweatherman"
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
		EnvVar             string
		Expected           logfile.Params
	}{
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
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("log-file", test.ViperLogFile)
			v.Set("logfile", test.ViperLogFileOld)
			v.Set("settings.log_file", test.ViperLogFileConfig)

			err := os.Setenv("WAKATIME_HOME", test.EnvVar)
			require.NoError(t, err)

			params, err := logfile.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params)
		})
	}
}

func TestSet(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "wakatime.log")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	v := viper.New()
	v.Set("log-file", tmpFile.Name())

	logfile.Set(v)

	jww.DEBUG.Println("some log")

	data, err := ioutil.ReadFile(tmpFile.Name())
	require.NoError(t, err)

	assert.Contains(t, "some log", string(data))
}
