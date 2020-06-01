package config_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/config"
	"gopkg.in/ini.v1"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadInConfig(t *testing.T) {
	v := viper.New()
	err := config.ReadInConfig(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return "testdata/wakatime.cfg", nil
	})
	require.NoError(t, err)

	assert.Equal(t, "b9485572-74bf-419a-916b-22056ca3a24c", v.GetString("settings.api_key"))
	assert.Equal(t, "true", v.GetString("settings.debug"))
	assert.Equal(t, "true", v.GetString("test.pandemia"))
}

func TestReadInConfigErr(t *testing.T) {
	v := viper.New()
	err := config.ReadInConfig(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return "", errors.New("error")
	})

	var cfperr config.ErrFileParse

	assert.True(t, errors.As(err, &cfperr))
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
			Expected: path.Join(home, ".wakatime.cfg"),
		},
		"viper": {
			ViperValue: "~/path/.wakatime.cfg",
			Expected:   path.Join(home, "/path/.wakatime.cfg"),
		},
		"env_trailling_slash": {
			EnvVar:   "~/path2/",
			Expected: path.Join(home, "/path2/.wakatime.cfg"),
		},
		"env_without_trailling_slash": {
			EnvVar:   "~/path2",
			Expected: path.Join(home, "/path2/.wakatime.cfg"),
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

	var cfperr config.ErrFileParse

	assert.True(t, errors.As(err, &cfperr))
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

	var cfwerr config.ErrFileWrite

	assert.True(t, errors.As(err, &cfwerr))
	assert.Equal(t, "got undefined wakatime config file instance", err.Error())
}
