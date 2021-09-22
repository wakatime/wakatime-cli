package config_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/config"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"
	"gopkg.in/ini.v1"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadInConfig(t *testing.T) {
	v := viper.New()
	v.Set("config", "./testdata/wakatime.cfg")
	filePath, err := config.FilePath(v)
	require.NoError(t, err)
	err = config.ReadInConfig(v, filePath)
	require.NoError(t, err)

	assert.Equal(t, "b9485572-74bf-419a-916b-22056ca3a24c", vipertools.GetString(v, "settings.api_key"))
	assert.Equal(t, "true", vipertools.GetString(v, "settings.debug"))
	assert.Equal(t, "true", vipertools.GetString(v, "test.pandemia"))
}

func TestReadInConfig_Multiline(t *testing.T) {
	multilineOption := viper.IniLoadOptions(ini.LoadOptions{AllowPythonMultilineValues: true})
	v := viper.NewWithOptions(multilineOption)
	v.Set("config", "./testdata/wakatime-multiline.cfg")
	filePath, err := config.FilePath(v)
	require.NoError(t, err)
	err = config.ReadInConfig(v, filePath)
	require.NoError(t, err)

	ignoreConfig := strings.ReplaceAll(vipertools.GetString(v, "settings.ignore"), "\r", "")
	assert.Equal(t, "\nCOMMIT_EDITMSG$\nPULLREQ_EDITMSG$\nMERGE_MSG$\nTAG_EDITMSG$", ignoreConfig)

	gitConfig := strings.ReplaceAll(vipertools.GetString(v, "git.submodules_disabled"), "\r", "")
	assert.Equal(t, "\n.*secret.*\nfix.*", gitConfig)
}

func TestReadInConfig_DoesNotExit_NoError(t *testing.T) {
	v := viper.New()
	v.Set("config", "./testdata/any.cfg")
	filePath, err := config.FilePath(v)
	require.NoError(t, err)
	err = config.ReadInConfig(v, filePath)
	require.NoError(t, err)
}

func TestReadInConfigMissing(t *testing.T) {
	v := viper.New()
	err := config.ReadInConfig(v, "not-exists")
	assert.NoError(t, err)
}

func TestReadInConfigMalformed(t *testing.T) {
	v := viper.New()
	v.Set("config", "./testdata/malformed.cfg")
	filePath, err := config.FilePath(v)
	require.NoError(t, err)
	err = config.ReadInConfig(v, filePath)
	assert.Error(t, err)
}

func TestFilePath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := map[string]struct {
		ViperValue string
		EnvVar     string
		Expected   string
	}{
		"default": {
			Expected: filepath.Join(home, ".wakatime.cfg"),
		},
		"viper": {
			ViperValue: "~/path/.wakatime.cfg",
			Expected:   filepath.Join(home, "path", ".wakatime.cfg"),
		},
		"surrounding double quotes": {
			ViperValue: `"~/path/.wakatime.cfg"`,
			Expected:   filepath.Join(home, "path", ".wakatime.cfg"),
		},
		"env trailing slash": {
			EnvVar:   "~/path2/",
			Expected: filepath.Join(home, "path2", ".wakatime.cfg"),
		},
		"env without trailing slash": {
			EnvVar:   "~/path2",
			Expected: filepath.Join(home, "path2", ".wakatime.cfg"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("config", test.ViperValue)
			err := os.Setenv("WAKATIME_HOME", test.EnvVar)
			require.NoError(t, err)

			configFilepath, err := config.FilePath(v)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, configFilepath)
		})
	}
}

func TestInternalFilePath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := map[string]struct {
		ViperValue string
		EnvVar     string
		Expected   string
	}{
		"default": {
			Expected: filepath.Join(home, ".wakatime-internal.cfg"),
		},
		"env trailing slash": {
			EnvVar:   "~/path2/",
			Expected: filepath.Join(home, "path2", ".wakatime-internal.cfg"),
		},
		"env without trailing slash": {
			EnvVar:   "~/path2",
			Expected: filepath.Join(home, "path2", ".wakatime-internal.cfg"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("internal-config", test.ViperValue)
			err := os.Setenv("WAKATIME_HOME", test.EnvVar)
			require.NoError(t, err)

			configFilepath, err := config.InternalFilePath(v)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, configFilepath)
		})
	}
}

func TestNewIniWriter(t *testing.T) {
	v := viper.New()
	w, err := config.NewIniWriter(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return "testdata/wakatime.cfg", nil
	})
	require.NoError(t, err)

	assert.Equal(t, "testdata/wakatime.cfg", w.ConfigFilepath)
	assert.NotNil(t, w.File)
}

func TestNewIniWriterErr(t *testing.T) {
	v := viper.New()
	_, err := config.NewIniWriter(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return "", errors.New("error")
	})
	require.Error(t, err)

	assert.Contains(t, err.Error(), "error getting filepath")
}

func TestWrite(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	tests := map[string]struct {
		Value   map[string]string
		Section string
	}{
		"single_value": {
			Value: map[string]string{
				"debug": "true",
			},
			Section: "settings",
		},
		"double_value": {
			Value: map[string]string{
				"debug":   "true",
				"api_key": "b9485572-74bf-419a-916b-22056ca3a24c",
			},
			Section: "settings",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			w := config.IniWriter{
				File:           ini.Empty(),
				ConfigFilepath: tmpFile.Name(),
			}

			err := w.Write(test.Section, test.Value)

			require.NoError(t, err)
		})
	}
}

func TestWriteErr(t *testing.T) {
	w := config.IniWriter{}

	err := w.Write("settings", map[string]string{"debug": "true"})
	require.Error(t, err)

	assert.Equal(t, "got undefined wakatime config file instance", err.Error())
}
