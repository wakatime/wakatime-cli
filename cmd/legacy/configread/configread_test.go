package configread_test

import (
	"fmt"
	"testing"

	"github.com/wakatime/wakatime-cli/cmd/legacy/configread"

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
			assert.Error(t, err)
		})
	}
}

func TestRead(t *testing.T) {
	v := viper.New()
	v.Set("settings.api_key", "b9485572-74bf-419a-916b-22056ca3a24c")
	v.Set("config-section", "settings")
	v.Set("config-read", "api_key")

	output, err := configread.Read(v)
	require.NoError(t, err)

	assert.Equal(t, "b9485572-74bf-419a-916b-22056ca3a24c", output)
}

func TestReadErr(t *testing.T) {
	tests := map[string]struct {
		Key      string
		Section  string
		Value    string
		ErrorMsg string
	}{
		"empty_value": {
			Key:      "api_key",
			Section:  "settings",
			Value:    "",
			ErrorMsg: fmt.Sprintf("given section and key \"%s.%s\" returned an empty string", "settings", "api_key"),
		},
		"section_missing": {
			Key:      "api_key",
			Value:    "b9485572-74bf-419a-916b-22056ca3a24c",
			ErrorMsg: "failed reading wakatime config file. neither section nor key can be empty",
		},
		"key_missing": {
			Section:  "settings",
			Value:    "b9485572-74bf-419a-916b-22056ca3a24c",
			ErrorMsg: "failed reading wakatime config file. neither section nor key can be empty",
		},
		"all_missing": {
			ErrorMsg: "failed reading wakatime config file. neither section nor key can be empty",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set(test.Section+"."+test.Key, test.Value)
			v.Set("config-section", test.Section)
			v.Set("config-read", test.Key)

			output, err := configread.Read(v)
			require.Error(t, err)

			assert.Equal(
				t,
				err.Error(),
				test.ErrorMsg,
				fmt.Sprintf("error %q differs from the string set", err),
			)

			assert.Empty(t, output)
		})
	}
}
