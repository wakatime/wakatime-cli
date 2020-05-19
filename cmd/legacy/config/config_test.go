package config_test

import (
	"errors"
	"testing"

	"github.com/wakatime/wakatime-cli/cmd/legacy/config"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_LoadReadParams(t *testing.T) {
	v := viper.New()
	v.Set("config-section", "settings")
	v.Set("config-read", "api_key")

	params, err := config.LoadReadParams(v)
	require.NoError(t, err)

	assert.Equal(t, config.ReadParams{
		Key:     "api_key",
		Section: "settings",
	}, params)
}

func TestConfig_LoadReadParamsErr(t *testing.T) {
	tests := map[string]struct {
		Key     string
		Section string
	}{
		"section_missing": {
			Key: "api_key",
		},
		"key_missing": {
			Section: "settings",
		},
		"all_missing": {},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("config-section", test.Section)
			v.Set("config-read", test.Key)

			_, err := config.LoadReadParams(v)

			var cfrerr config.ErrFileRead
			assert.True(t, errors.As(err, &cfrerr))
		})
	}
}

func TestConfig_RunRead(t *testing.T) {
	v := viper.New()
	v.Set("config-section", "settings")
	v.Set("config-read", "api_key")
	v.Set("settings.api_key", "b9485572-74bf-419a-916b-22056ca3a24c")

	err := config.RunRead(v)
	require.NoError(t, err)
}

func TestConfig_RunReadErr(t *testing.T) {
	v := viper.New()
	v.Set("config-section", "settings")
	v.Set("config-read", "unset_key")

	err := config.RunRead(v)

	var cfrerr config.ErrFileRead

	assert.True(t, errors.As(err, &cfrerr))
}
