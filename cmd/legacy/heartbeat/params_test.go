package heartbeat_test

import (
	"errors"
	"os"
	"testing"
	"time"

	cmd "github.com/wakatime/wakatime-cli/cmd/legacy/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadParams_APIKey_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.api_key", "ignored")
	v.Set("settings.apikey", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-4000-8000-000000000000", params.APIKey)
}

func TestLoadParams_APIKey_FromConfigTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("settings.api_key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.apikey", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-4000-8000-000000000000", params.APIKey)
}

func TestLoadParams_APIKey_FromConfigDeprecated(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("settings.apikey", "00000000-0000-4000-8000-000000000000")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-4000-8000-000000000000", params.APIKey)
}

func TestLoadParams_InvalidAPIKey(t *testing.T) {
	tests := map[string]string{
		"unset":            "",
		"invalid format 1": "not-uuid",
		"invalid format 2": "00000000-0000-0000-8000-000000000000",
		"invalid format 3": "00000000-0000-4000-0000-000000000000",
	}

	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", "/path/to/file")
			v.Set("key", value)

			_, err := cmd.LoadParams(v)
			require.Error(t, err)

			var errauth api.ErrAuth
			require.True(t, errors.As(err, &errauth))
		})
	}
}

func TestLoadParams_APIUrl_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("api-url", "http://localhost:8080")
	v.Set("apiurl", "ignored")
	v.Set("settings.api_url", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:8080", params.APIUrl)
}

func TestLoadParams_APIUrl_FlagDeprecatedTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("apiurl", "http://localhost:8080")
	v.Set("settings.api_url", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:8080", params.APIUrl)
}

func TestLoadParams_APIUrl_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.api_url", "http://localhost:8080")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:8080", params.APIUrl)
}

func TestLoadParams_APIUrl_Default(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, api.BaseURL, params.APIUrl)
}

func TestLoadParams_Category(t *testing.T) {
	tests := map[string]heartbeat.Category{
		"coding":         heartbeat.CodingCategory,
		"browsing":       heartbeat.BrowsingCategory,
		"building":       heartbeat.BuildingCategory,
		"code reviewing": heartbeat.CodeReviewingCategory,
		"debugging":      heartbeat.DebuggingCategory,
		"designing":      heartbeat.DesigningCategory,
		"indexing":       heartbeat.IndexingCategory,
		"manual testing": heartbeat.ManualTestingCategory,
		"running tests":  heartbeat.RunningTestsCategory,
		"writing tests":  heartbeat.WritingTestsCategory,
	}

	for name, category := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("category", name)

			params, err := cmd.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, category, params.Category)
		})
	}
}

func TestLoadParams_Category_Default(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, heartbeat.CodingCategory, params.Category)
}

func TestLoadParams_Category_Invalid(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("category", "invalid")

	_, err := cmd.LoadParams(v)
	require.Error(t, err)

	assert.Equal(t, "failed to parse category: invalid category \"invalid\"", err.Error())
}

func TestLoadParams_Entity_EntityFlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("file", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "/path/to/file", params.Entity)
}

func TestLoadParams_Entity_FileFlag(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("file", "/path/to/file")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "/path/to/file", params.Entity)
}

func TestLoadParams_Entity_Unset(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	_, err := cmd.LoadParams(v)
	require.Error(t, err)

	assert.Equal(t, "failed to retrieve entity", err.Error())
}

func TestLoadParams_EntityType(t *testing.T) {
	tests := map[string]heartbeat.EntityType{
		"file":   heartbeat.FileType,
		"domain": heartbeat.DomainType,
		"app":    heartbeat.AppType,
	}

	for name, entityType := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("entity-type", name)

			params, err := cmd.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, entityType, params.EntityType)
		})
	}
}

func TestLoadParams_EntityType_Default(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, heartbeat.FileType, params.EntityType)
}

func TestLoadParams_EntityType_Invalid(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("entity-type", "invalid")

	_, err := cmd.LoadParams(v)
	require.Error(t, err)

	assert.Equal(t, "failed to parse entity type: invalid entity type \"invalid\"", err.Error())
}

func TestLoadParams_Hostname_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hostname", "my-machine")
	v.Set("settings.hostname", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "my-machine", params.Hostname)
}

func TestLoadParams_Hostname_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hostname", "my-machine")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "my-machine", params.Hostname)
}

func TestLoadParams_Hostname_DefaultFromSystem(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	expected, err := os.Hostname()
	require.NoError(t, err)

	assert.Equal(t, expected, params.Hostname)
}

func TestLoadParams_IsWrite(t *testing.T) {
	tests := map[string]bool{
		"is write":    true,
		"is no write": false,
	}

	for name, isWrite := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("write", isWrite)

			params, err := cmd.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, isWrite, *params.IsWrite)
		})
	}
}

func TestLoadParams_IsWrite_Unset(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Nil(t, params.IsWrite)
}

func TestLoadParams_Plugin(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("plugin", "plugin/10.0.0")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "plugin/10.0.0", params.Plugin)
}

func TestLoadParams_Plugin_Unset(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Empty(t, params.Plugin)
}

func TestLoadParams_Timeout_FlagTakesPreceedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("timeout", 5)
	v.Set("settings.timeout", 10)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, 5*time.Second, params.Timeout)
}

func TestLoadParams_Timeout_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.timeout", 10)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, 10*time.Second, params.Timeout)
}

func TestLoadParams_Time(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("time", 1590609206.1)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, 1590609206.1, params.Time)
}

func TestLoadParams_Time_Default(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	now := time.Now().Unix()
	assert.GreaterOrEqual(t, float64(now), params.Time)
	assert.GreaterOrEqual(t, params.Time, float64(now-60))
}

func TestLoadParams_Network_DisableSSLVerify_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("no-ssl-verify", true)
	v.Set("settings.no_ssl_verify", false)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.True(t, params.Network.DisableSSLVerify)
}

func TestLoadParams_Network_DisableSSLVerify_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.no_ssl_verify", true)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.True(t, params.Network.DisableSSLVerify)
}

func TestLoadParams_Network_DisableSSLVerify_Default(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.False(t, params.Network.DisableSSLVerify)
}

func TestLoadParams_Network_ProxyURL(t *testing.T) {
	tests := map[string]string{
		"https":  "https://john:secret@example.org:8888",
		"http":   "http://john:secret@example.org:8888",
		"socks5": "socks5://john:secret@example.org:8888",
	}

	for name, proxyURL := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("proxy", proxyURL)

			params, err := cmd.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, proxyURL, params.Network.ProxyURL)
		})
	}
}

func TestLoadParams_Network_ProxyURL_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("proxy", "https://john:secret@example.org:8888")
	v.Set("settings.proxy", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "https://john:secret@example.org:8888", params.Network.ProxyURL)
}

func TestLoadParams_Network_ProxyURL_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.proxy", "https://john:secret@example.org:8888")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "https://john:secret@example.org:8888", params.Network.ProxyURL)
}

func TestLoadParams_Network_ProxyURL_InvalidFormat(t *testing.T) {
	proxyURL := "ftp://john:secret@example.org:8888"

	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("proxy", proxyURL)

	_, err := cmd.LoadParams(v)
	require.Error(t, err)
}

func TestLoadParams_Network_SSLCertFilepath_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("ssl-certs-file", "/path/to/cert.pem")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "/path/to/cert.pem", params.Network.SSLCertFilepath)
}

func TestLoadParams_Network_SSLCertFilepath_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.ssl_certs_file", "/path/to/cert.pem")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "/path/to/cert.pem", params.Network.SSLCertFilepath)
}
