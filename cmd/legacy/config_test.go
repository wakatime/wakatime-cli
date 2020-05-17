package legacy_test

import (
	"errors"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wakatime/wakatime-cli/cmd/legacy"
)

func TestConfig_LoadConfigReadParams(t *testing.T) {
	v := viper.New()
	v.Set("config-section", "settings")
	v.Set("config-read", "api_key")

	params, err := legacy.LoadConfigReadParams(v)
	require.NoError(t, err)

	assert.Equal(t, legacy.ConfigReadParams{
		Key:     "api_key",
		Section: "settings",
	}, params)
}

func TestConfig_LoadConfigReadParamsErr(t *testing.T) {
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

			_, err := legacy.LoadConfigReadParams(v)

			var cfrerr legacy.ErrConfigFileRead
			assert.True(t, errors.As(err, &cfrerr))
		})
	}
}
