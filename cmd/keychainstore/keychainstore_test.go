package keychainstore_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/cmd/keychainstore"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/keychain"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
	iniv1 "gopkg.in/ini.v1"
)

func TestRun(t *testing.T) {
	keyring.MockInit()

	configFile := filepath.Join(t.TempDir(), ".wakatime.cfg")

	contents := []byte("[settings]\napi_key = 00000000-0000-4000-8000-000000000000\ndebug = true\n")

	err := os.WriteFile(configFile, contents, 0600)
	require.NoError(t, err)

	v := viper.New()
	v.Set("config", configFile)

	err = ini.ReadInConfig(v, configFile)
	require.NoError(t, err)

	code, err := keychainstore.Run(t.Context(), v)
	require.NoError(t, err)

	assert.Equal(t, exitcode.Success, code)

	stored, err := keyring.Get(keychain.Service, keychain.User)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-4000-8000-000000000000", stored)

	cfg, err := iniv1.Load(configFile)
	require.NoError(t, err)

	assert.Empty(t, cfg.Section("settings").Key("api_key").String())
	assert.Equal(t, "keychain", cfg.Section("settings").Key("api_key_storage").String())
	assert.Equal(t, "true", cfg.Section("settings").Key("debug").String())
}

func TestRun_KeyFlag(t *testing.T) {
	keyring.MockInit()

	configFile := filepath.Join(t.TempDir(), ".wakatime.cfg")

	v := viper.New()
	v.Set("config", configFile)
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	code, err := keychainstore.Run(t.Context(), v)
	require.NoError(t, err)

	assert.Equal(t, exitcode.Success, code)

	stored, err := keyring.Get(keychain.Service, keychain.User)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-4000-8000-000000000000", stored)

	cfg, err := iniv1.Load(configFile)
	require.NoError(t, err)

	assert.False(t, cfg.Section("settings").HasKey("api_key"))
	assert.Equal(t, "keychain", cfg.Section("settings").Key("api_key_storage").String())
}

func TestRun_NoAPIKey(t *testing.T) {
	keyring.MockInit()

	v := viper.New()
	v.Set("config", filepath.Join(t.TempDir(), ".wakatime.cfg"))

	code, err := keychainstore.Run(t.Context(), v)

	assert.Equal(t, exitcode.ErrAuth, code)
	assert.EqualError(t, err, "failed to load api key: api key not found or empty")
}

func TestRun_KeychainErr(t *testing.T) {
	keyring.MockInitWithError(errors.New("keychain locked"))

	configFile := filepath.Join(t.TempDir(), ".wakatime.cfg")
	contents := []byte("[settings]\napi_key = 00000000-0000-4000-8000-000000000000\n")

	err := os.WriteFile(configFile, contents, 0600)
	require.NoError(t, err)

	v := viper.New()
	v.Set("config", configFile)

	err = ini.ReadInConfig(v, configFile)
	require.NoError(t, err)

	code, err := keychainstore.Run(t.Context(), v)

	assert.Equal(t, exitcode.ErrGeneric, code)
	assert.EqualError(t, err, "failed to store api key in keychain: keychain locked")

	after, err := os.ReadFile(configFile)
	require.NoError(t, err)

	assert.Equal(t, contents, after)
}
