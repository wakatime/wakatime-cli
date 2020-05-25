package config_test

import (
	"errors"
	"os"
	"path"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/config"

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

func TestLoadIni(t *testing.T) {
	v := viper.New()
	cfg, err := config.LoadIni(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return "testdata/wakatime.cfg", nil
	})
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "b9485572-74bf-419a-916b-22056ca3a24c", cfg.Section("settings").Key("api_key").String())
	assert.Equal(t, "true", cfg.Section("settings").Key("debug").String())
	assert.Equal(t, "true", cfg.Section("test").Key("pandemia").String())
}

func TestLoadIniErr(t *testing.T) {
	v := viper.New()
	_, err := config.LoadIni(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return "", errors.New("error")
	})

	var cfperr config.ErrFileParse

	assert.True(t, errors.As(err, &cfperr))
}
