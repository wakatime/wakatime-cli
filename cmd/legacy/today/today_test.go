package today_test

import (
	"errors"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/cmd/legacy/today"
	"github.com/wakatime/wakatime-cli/pkg/api"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadParams_APIKey(t *testing.T) {
	tests := map[string]struct {
		ViperAPIKey          string
		ViperAPIKeyConfig    string
		ViperAPIKeyConfigOld string
		Expected             today.Params
	}{
		"api key flag takes preceedence": {
			ViperAPIKey:          "00000000-0000-4000-8000-000000000000",
			ViperAPIKeyConfig:    "10000000-0000-4000-8000-000000000000",
			ViperAPIKeyConfigOld: "20000000-0000-4000-8000-000000000000",
			Expected: today.Params{
				APIKey: "00000000-0000-4000-8000-000000000000",
			},
		},
		"api from config takes preceedence": {
			ViperAPIKeyConfig:    "00000000-0000-4000-8000-000000000000",
			ViperAPIKeyConfigOld: "10000000-0000-4000-8000-000000000000",
			Expected: today.Params{
				APIKey: "00000000-0000-4000-8000-000000000000",
			},
		},
		"api key from config deprecated": {
			ViperAPIKeyConfigOld: "00000000-0000-4000-8000-000000000000",
			Expected: today.Params{
				APIKey: "00000000-0000-4000-8000-000000000000",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", test.ViperAPIKey)
			v.Set("settings.api_key", test.ViperAPIKeyConfig)
			v.Set("settings.apikey", test.ViperAPIKeyConfigOld)

			params, err := today.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params)
		})
	}
}

func TestLoadParams_APIUrl(t *testing.T) {
	tests := map[string]struct {
		ViperAPIUrl       string
		ViperAPIUrlConfig string
		ViperAPIUrlOld    string
		Expected          today.Params
	}{
		"api url flag takes preceedence": {
			ViperAPIUrl:       "http://localhost:8080",
			ViperAPIUrlConfig: "http://localhost:8081",
			ViperAPIUrlOld:    "http://localhost:8082",
			Expected: today.Params{
				APIUrl: "http://localhost:8080",
			},
		},
		"api url deprecated flag takes preceedence": {
			ViperAPIUrlConfig: "http://localhost:8081",
			ViperAPIUrlOld:    "http://localhost:8082",
			Expected: today.Params{
				APIUrl: "http://localhost:8082",
			},
		},
		"api url from config": {
			ViperAPIUrlConfig: "http://localhost:8081",
			Expected: today.Params{
				APIUrl: "http://localhost:8081",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("api-url", test.ViperAPIUrl)
			v.Set("apiurl", test.ViperAPIUrlOld)
			v.Set("settings.api_url", test.ViperAPIUrlConfig)

			params, err := today.LoadParams(v)
			require.NoError(t, err)

			test.Expected.APIKey = "00000000-0000-4000-8000-000000000000"
			assert.Equal(t, test.Expected, params)
		})
	}
}

func TestLoadParams_Plugin(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("plugin", "plugin/10.0.0")

	params, err := today.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, today.Params{
		APIKey: "00000000-0000-4000-8000-000000000000",
		Plugin: "plugin/10.0.0",
	}, params)
}

func TestLoadParams_Timeout(t *testing.T) {
	tests := map[string]struct {
		ViperTimeout       int
		ViperTimeoutConfig int
		Expected           today.Params
	}{
		"timeout flag takes preceedence": {
			ViperTimeout:       5,
			ViperTimeoutConfig: 10,
			Expected: today.Params{
				Timeout: 5 * time.Second,
			},
		},
		"timeout from config": {
			ViperTimeoutConfig: 10,
			Expected: today.Params{
				Timeout: 10 * time.Second,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("timeout", test.ViperTimeout)
			v.Set("settings.timeout", test.ViperTimeoutConfig)

			params, err := today.LoadParams(v)
			require.NoError(t, err)

			test.Expected.APIKey = "00000000-0000-4000-8000-000000000000"
			assert.Equal(t, test.Expected, params)
		})
	}
}

func TestLoadParamsErr_InvalidAPIKey(t *testing.T) {
	tests := map[string]string{
		"unset":            "",
		"invalid format 1": "not-uuid",
		"invalid format 2": "00000000-0000-0000-8000-000000000000",
		"invalid format 3": "00000000-0000-4000-0000-000000000000",
	}

	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", value)

			_, err := today.LoadParams(v)
			require.Error(t, err)

			var errauth api.ErrAuth
			require.True(t, errors.As(err, &errauth))
		})
	}
}
