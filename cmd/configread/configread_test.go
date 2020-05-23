package configread_test

import (
	"errors"
	"testing"

	"github.com/wakatime/wakatime-cli/cmd/configread"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadParams(t *testing.T) {
	v := viper.New()
	v.Set("config-section", "settings")
	v.Set("config-read", "api_key")

	params, err := configread.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, configread.Params{
		Key:     "api_key",
		Section: "settings",
	}, params)
}

func TestLoadParamsErr(t *testing.T) {
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

			_, err := configread.LoadParams(v)

			var cfrerr configread.ErrFileRead
			assert.True(t, errors.As(err, &cfrerr))
		})
	}
}

func TestRun(t *testing.T) {
	v := viper.New()
	v.Set("config-section", "settings")
	v.Set("config-read", "api_key")
	v.Set("settings.api_key", "b9485572-74bf-419a-916b-22056ca3a24c")

	err := configread.Run(v)
	require.NoError(t, err)
}

func TestRunErr(t *testing.T) {
	v := viper.New()
	v.Set("config-section", "settings")
	v.Set("config-read", "unset_key")

	err := configread.Run(v)

	var cfrerr configread.ErrFileRead

	assert.True(t, errors.As(err, &cfrerr))
}
