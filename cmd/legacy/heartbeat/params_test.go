package heartbeat_test

import (
	"errors"
	"os"
	"regexp"
	"testing"
	"time"

	cmd "github.com/wakatime/wakatime-cli/cmd/legacy/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/spf13/pflag"
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

func TestLoadParams_CursorPosition(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("cursorpos", 42)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, 42, *params.CursorPosition)
}

func TestLoadParams_CursorPosition_Zero(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("cursorpos", 0)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, 0, *params.CursorPosition)
}

func TestLoadParams_CursorPosition_Unset(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Nil(t, params.CursorPosition)
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

func TestLoadParams_LineNumber(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("lineno", 42)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, 42, *params.LineNumber)
}

func TestLoadParams_LineNumber_Zero(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("lineno", 0)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, 0, *params.LineNumber)
}

func TestLoadParams_LineNumber_Unset(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Nil(t, params.LineNumber)
}

func TestLoadParams_OfflineDisabled_ConfigTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("disable-offline", false)
	v.Set("disableoffline", false)
	v.Set("settings.offline", false)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.True(t, params.OfflineDisabled)
}

func TestLoadParams_OfflineDisabled_FlagDeprecatedTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("disable-offline", false)
	v.Set("disableoffline", true)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.True(t, params.OfflineDisabled)
}

func TestLoadParams_OfflineDisabled_FromFlag(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("disable-offline", true)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.True(t, params.OfflineDisabled)
}

func TestLoadParams_OfflineSyncMax(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("sync-offline-activity", 42)
	v.SetDefault("sync-offline-activity", 100)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, 42, params.OfflineSyncMax)
}

func TestLoadParams_OfflineSyncMax_NoEntity(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("sync-offline-activity", 42)
	v.SetDefault("sync-offline-activity", 100)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, 42, params.OfflineSyncMax)
}

func TestLoadParams_OfflineSyncMax_NoEntity_DefaultNotAccepted(t *testing.T) {
	v := viper.New()

	flags := pflag.FlagSet{}
	flags.String("sync-offline-activity", "100", "")

	err := v.BindPFlags(&flags)
	require.NoError(t, err)

	v.Set("key", "00000000-0000-4000-8000-000000000000")

	_, err = cmd.LoadParams(v)
	require.Error(t, err)

	assert.Equal(t, "failed to retrieve entity", err.Error())
}

func TestLoadParams_OfflineSyncMax_None(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("sync-offline-activity", "none")
	v.SetDefault("sync-offline-activity", 100)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, 0, params.OfflineSyncMax)
}

func TestLoadParams_OfflineSyncMax_Default(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.SetDefault("sync-offline-activity", 100)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, 100, params.OfflineSyncMax)
}

func TestLoadParams_OfflineSyncMax_NegativeNumber(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("sync-offline-activity", -1)
	v.SetDefault("sync-offline-activity", 100)

	_, err := cmd.LoadParams(v)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "--sync-offline-activity")
}

func TestLoadParams_OfflineSyncMax_NonIntegerValue(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("sync-offline-activity", "invalid")
	v.SetDefault("sync-offline-activity", 100)

	_, err := cmd.LoadParams(v)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "--sync-offline-activity")
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

	now := float64(time.Now().UnixNano()) / 1000000000
	assert.GreaterOrEqual(t, now, params.Time)
	assert.GreaterOrEqual(t, params.Time, now-60)
}

func TestLoadParams_Filter_Exclude(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("exclude", []string{".*", "wakatime.*"})
	v.Set("settings.exclude", []string{".+", "wakatime.+"})
	v.Set("settings.ignore", []string{".?", "wakatime.?"})

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, []*regexp.Regexp{
		regexp.MustCompile(".*"),
		regexp.MustCompile("wakatime.*"),
		regexp.MustCompile(".+"),
		regexp.MustCompile("wakatime.+"),
		regexp.MustCompile(".?"),
		regexp.MustCompile("wakatime.?"),
	}, params.Filter.Exclude)
}

func TestLoadParams_Filter_Exclude_IgnoresInvalidRegex(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("exclude", []string{".*", "["})

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, []*regexp.Regexp{regexp.MustCompile(".*")}, params.Filter.Exclude)
}

func TestLoadParams_Filter_ExcludeUnknownProject(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("exclude-unknown-project", true)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, true, params.Filter.ExcludeUnknownProject)
}

func TestLoadParams_Filter_ExcludeUnknownProject_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("exclude-unknown-project", false)
	v.Set("settings.exclude_unknown_project", true)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, true, params.Filter.ExcludeUnknownProject)
}

func TestLoadParams_Filter_Include(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("include", []string{".*", "wakatime.*"})
	v.Set("settings.include", []string{".+", "wakatime.+"})

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, []*regexp.Regexp{
		regexp.MustCompile(".*"),
		regexp.MustCompile("wakatime.*"),
		regexp.MustCompile(".+"),
		regexp.MustCompile("wakatime.+"),
	}, params.Filter.Include)
}

func TestLoadParams_Filter_Include_IgnoresInvalidRegex(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("include", []string{".*", "["})

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, []*regexp.Regexp{regexp.MustCompile(".*")}, params.Filter.Include)
}

func TestLoadParams_Filter_IncludeOnlyWithProjectFile(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("include-only-with-project-file", true)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, true, params.Filter.IncludeOnlyWithProjectFile)
}

func TestLoadParams_Filter_IncludeOnlyWithProjectFile_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("include-only-with-project-file", false)
	v.Set("settings.include_only_with_project_file", true)

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, true, params.Filter.IncludeOnlyWithProjectFile)
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

func TestLoadParams_SanitizeParams_HideBranchNames_True(t *testing.T) {
	tests := map[string]string{
		"lowercase":       "true",
		"uppercase":       "TRUE",
		"first uppercase": "True",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-branch-names", viperValue)

			params, err := cmd.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, cmd.SanitizeParams{
				HideBranchNames: []*regexp.Regexp{regexp.MustCompile(".*")},
			}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideBranchNames_False(t *testing.T) {
	tests := map[string]string{
		"lowercase":       "false",
		"uppercase":       "FALSE",
		"first uppercase": "False",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-branch-names", viperValue)

			params, err := cmd.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, cmd.SanitizeParams{}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideBranchNames_List(t *testing.T) {
	tests := map[string]struct {
		ViperValue string
		Expected   []*regexp.Regexp
	}{
		"regex": {
			ViperValue: "fix.*",
			Expected: []*regexp.Regexp{
				regexp.MustCompile("fix.*"),
			},
		},
		"regex list": {
			ViperValue: ".*secret.*\nfix.*",
			Expected: []*regexp.Regexp{
				regexp.MustCompile(".*secret.*"),
				regexp.MustCompile("fix.*"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-branch-names", test.ViperValue)

			params, err := cmd.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, cmd.SanitizeParams{
				HideBranchNames: test.Expected,
			}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideBranchNames_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hide-branch-names", "true")
	v.Set("settings.hide_branch_names", "ignored")
	v.Set("settings.hide_branchnames", "ignored")
	v.Set("settings.hidebranchnames", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, cmd.SanitizeParams{
		HideBranchNames: []*regexp.Regexp{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideBranchNames_ConfigTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_branch_names", "true")
	v.Set("settings.hide_branchnames", "ignored")
	v.Set("settings.hidebranchnames", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, cmd.SanitizeParams{
		HideBranchNames: []*regexp.Regexp{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideBranchNames_ConfigDeprecatedOneTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_branchnames", "true")
	v.Set("settings.hidebranchnames", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, cmd.SanitizeParams{
		HideBranchNames: []*regexp.Regexp{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideBranchNames_ConfigDeprecatedTwo(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hidebranchnames", "true")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, cmd.SanitizeParams{
		HideBranchNames: []*regexp.Regexp{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideBranchNames_InvalidRegex(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hide-branch-names", ".*secret.*\n[0-9+")

	_, err := cmd.LoadParams(v)
	require.Error(t, err)

	assert.Equal(t, errors.New(
		"failed to load sanitize params:"+
			" failed to parse regex hide branch names param \".*secret.*\\n[0-9+\":"+
			" failed to compile regex \"[0-9+\":"+
			" error parsing regexp: missing closing ]: `[0-9+`",
	), err)
}

func TestLoadParams_SanitizeParams_HideProjectNames_True(t *testing.T) {
	tests := map[string]string{
		"lowercase":       "true",
		"uppercase":       "TRUE",
		"first uppercase": "True",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-project-names", viperValue)

			params, err := cmd.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, cmd.SanitizeParams{
				HideProjectNames: []*regexp.Regexp{regexp.MustCompile(".*")},
			}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideProjectNames_False(t *testing.T) {
	tests := map[string]string{
		"lowercase":       "false",
		"uppercase":       "FALSE",
		"first uppercase": "False",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-project-names", viperValue)

			params, err := cmd.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, cmd.SanitizeParams{}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideProjecthNames_List(t *testing.T) {
	tests := map[string]struct {
		ViperValue string
		Expected   []*regexp.Regexp
	}{
		"regex": {
			ViperValue: "fix.*",
			Expected: []*regexp.Regexp{
				regexp.MustCompile("fix.*"),
			},
		},
		"regex list": {
			ViperValue: ".*secret.*\nfix.*",
			Expected: []*regexp.Regexp{
				regexp.MustCompile(".*secret.*"),
				regexp.MustCompile("fix.*"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-project-names", test.ViperValue)

			params, err := cmd.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, cmd.SanitizeParams{
				HideProjectNames: test.Expected,
			}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideProjectNames_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hide-project-names", "true")
	v.Set("settings.hide_project_names", "ignored")
	v.Set("settings.hide_projectnames", "ignored")
	v.Set("settings.hideprojectnames", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, cmd.SanitizeParams{
		HideProjectNames: []*regexp.Regexp{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideProjectNames_ConfigTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_project_names", "true")
	v.Set("settings.hide_projectnames", "ignored")
	v.Set("settings.hideprojectnames", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, cmd.SanitizeParams{
		HideProjectNames: []*regexp.Regexp{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideProjectNames_ConfigDeprecatedOneTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_projectnames", "true")
	v.Set("settings.hideprojectnames", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, cmd.SanitizeParams{
		HideProjectNames: []*regexp.Regexp{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideProjectNames_ConfigDeprecatedTwo(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hideprojectnames", "true")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, cmd.SanitizeParams{
		HideProjectNames: []*regexp.Regexp{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideProjectNames_InvalidRegex(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hide-project-names", ".*secret.*\n[0-9+")

	_, err := cmd.LoadParams(v)
	require.Error(t, err)

	assert.Equal(t, errors.New(
		"failed to load sanitize params:"+
			" failed to parse regex hide project names param \".*secret.*\\n[0-9+\":"+
			" failed to compile regex \"[0-9+\":"+
			" error parsing regexp: missing closing ]: `[0-9+`",
	), err)
}

func TestLoadParams_SanitizeParams_HideFileNames_True(t *testing.T) {
	tests := map[string]string{
		"lowercase":       "true",
		"uppercase":       "TRUE",
		"first uppercase": "True",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-file-names", viperValue)

			params, err := cmd.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, cmd.SanitizeParams{
				HideFileNames: []*regexp.Regexp{regexp.MustCompile(".*")},
			}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideFileNames_False(t *testing.T) {
	tests := map[string]string{
		"lowercase":       "false",
		"uppercase":       "FALSE",
		"first uppercase": "False",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-file-names", viperValue)

			params, err := cmd.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, cmd.SanitizeParams{}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideFilehNames_List(t *testing.T) {
	tests := map[string]struct {
		ViperValue string
		Expected   []*regexp.Regexp
	}{
		"regex": {
			ViperValue: "fix.*",
			Expected: []*regexp.Regexp{
				regexp.MustCompile("fix.*"),
			},
		},
		"regex list": {
			ViperValue: ".*secret.*\nfix.*",
			Expected: []*regexp.Regexp{
				regexp.MustCompile(".*secret.*"),
				regexp.MustCompile("fix.*"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-file-names", test.ViperValue)

			params, err := cmd.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, cmd.SanitizeParams{
				HideFileNames: test.Expected,
			}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideFileNames_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hide-file-names", "true")
	v.Set("hide-filenames", "ignored")
	v.Set("hidefilenames", "ignored")
	v.Set("settings.hide_file_names", "ignored")
	v.Set("settings.hide_filenames", "ignored")
	v.Set("settings.hidefilenames", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, cmd.SanitizeParams{
		HideFileNames: []*regexp.Regexp{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_FlagDeprecatedOneTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hide-filenames", "true")
	v.Set("hidefilenames", "ignored")
	v.Set("settings.hide_file_names", "ignored")
	v.Set("settings.hide_filenames", "ignored")
	v.Set("settings.hidefilenames", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, cmd.SanitizeParams{
		HideFileNames: []*regexp.Regexp{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_FlagDeprecatedTwoTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hidefilenames", "true")
	v.Set("settings.hide_file_names", "ignored")
	v.Set("settings.hide_filenames", "ignored")
	v.Set("settings.hidefilenames", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, cmd.SanitizeParams{
		HideFileNames: []*regexp.Regexp{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_ConfigTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_file_names", "true")
	v.Set("settings.hide_filenames", "ignored")
	v.Set("settings.hidefilenames", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, cmd.SanitizeParams{
		HideFileNames: []*regexp.Regexp{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_ConfigDeprecatedOneTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_filenames", "true")
	v.Set("settings.hidefilenames", "ignored")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, cmd.SanitizeParams{
		HideFileNames: []*regexp.Regexp{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_ConfigDeprecatedTwo(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hidefilenames", "true")

	params, err := cmd.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, cmd.SanitizeParams{
		HideFileNames: []*regexp.Regexp{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_InvalidRegex(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hide-file-names", ".*secret.*\n[0-9+")

	_, err := cmd.LoadParams(v)
	require.Error(t, err)

	assert.Equal(t, errors.New(
		"failed to load sanitize params:"+
			" failed to parse regex hide file names param \".*secret.*\\n[0-9+\":"+
			" failed to compile regex \"[0-9+\":"+
			" error parsing regexp: missing closing ]: `[0-9+`",
	), err)
}
