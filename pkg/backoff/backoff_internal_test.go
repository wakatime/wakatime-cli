package backoff

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/config"
	"github.com/wakatime/wakatime-cli/pkg/ini"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldBackoff(t *testing.T) {
	at := time.Now().Add(time.Second * -1)

	should := shouldBackoff(1, at)

	assert.True(t, should)
}

func TestShouldBackoff_AfterResetTime(t *testing.T) {
	at := time.Now().Add((resetAfter + 1) * time.Second)

	should := shouldBackoff(0, at)

	assert.False(t, should)
}

func TestShouldBackoff_NegateBackoff(t *testing.T) {
	should := shouldBackoff(0, time.Time{})

	assert.False(t, should)
}

func TestUpdateBackoffSettings(t *testing.T) {
	v := viper.New()

	tmpFile, err := os.CreateTemp(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	v.Set("config", tmpFile.Name())
	v.Set("internal-config", tmpFile.Name())

	at := time.Now().Add(time.Second * -1)

	err = updateBackoffSettings(v, 2, at)
	require.NoError(t, err)

	writer, err := config.NewIniWriter(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return tmpFile.Name(), nil
	})
	require.NoError(t, err)

	backoffAt := writer.File.Section("internal").Key("backoff_at").MustTimeFormat(config.DateFormat)

	assert.WithinDuration(t, time.Now(), backoffAt, 15*time.Second)
	assert.Equal(t, "2", writer.File.Section("internal").Key("backoff_retries").String())
}

func TestUpdateBackoffSettings_NotInBackoff(t *testing.T) {
	v := viper.New()

	tmpFile, err := os.CreateTemp(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	v.Set("config", tmpFile.Name())
	v.Set("internal-config", tmpFile.Name())

	err = updateBackoffSettings(v, 0, time.Time{})
	require.NoError(t, err)

	writer, err := config.NewIniWriter(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return tmpFile.Name(), nil
	})
	require.NoError(t, err)

	assert.Empty(t, writer.File.Section("internal").Key("backoff_at").String())
	assert.Equal(t, "0", writer.File.Section("internal").Key("backoff_retries").String())
}

func TestUpdateBackoffSettings_NoMultilineSideEffects(t *testing.T) {
	v := viper.New()

	tmpFile, err := os.CreateTemp(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	v.Set("config", tmpFile.Name())
	v.Set("internal-config", tmpFile.Name())

	copyFile(t, "testdata/multiline.cfg", tmpFile.Name())

	err = updateBackoffSettings(v, 0, time.Time{})
	require.NoError(t, err)

	writer, err := config.NewIniWriter(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return tmpFile.Name(), nil
	})
	require.NoError(t, err)

	assert.Equal(t, "\none\ntwo", writer.File.Section("settings").Key("ignore").String())
	assert.Empty(t, writer.File.Section("internal").Key("backoff_at").String())
	assert.Equal(t, "0", writer.File.Section("internal").Key("backoff_retries").String())

	value := ini.GetKey(tmpFile.Name(), ini.Key{Section: "settings", Name: "ignore"})
	assert.Equal(t, "\n one\n two", value)

	actual, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)

	expected, err := os.ReadFile("testdata/multiline_expected.cfg")
	require.NoError(t, err)

	assert.Equal(t, strings.ReplaceAll(string(expected), "\r", ""), string(actual))
}

func copyFile(t *testing.T, source, destination string) {
	input, err := os.ReadFile(source)
	require.NoError(t, err)

	err = os.WriteFile(destination, input, 0600)
	require.NoError(t, err)
}
