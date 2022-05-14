package backoff

import (
	"os"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/ini"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldBackoff(t *testing.T) {
	at := time.Now().Add(time.Second * -1)

	should, reset := shouldBackoff(1, at)

	assert.True(t, should)
	assert.False(t, reset)
}

func TestShouldBackoff_AfterResetTime(t *testing.T) {
	at := time.Now().Add(time.Second * -1)

	should, reset := shouldBackoff(8, at)

	assert.False(t, should)
	assert.True(t, reset)
}

func TestShouldBackoff_AfterResetTime_ZeroRetries(t *testing.T) {
	at := time.Now().Add((resetAfter + 1) * time.Second)

	should, reset := shouldBackoff(0, at)

	assert.False(t, should)
	assert.True(t, reset)
}

func TestShouldBackoff_NegateBackoff(t *testing.T) {
	should, reset := shouldBackoff(0, time.Time{})

	assert.False(t, should)
	assert.True(t, reset)
}

func TestUpdateBackoffSettings(t *testing.T) {
	v := viper.New()

	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	v.Set("config", tmpFile.Name())
	v.Set("internal-config", tmpFile.Name())

	at := time.Now().Add(time.Second * -1)

	err = updateBackoffSettings(v, 2, at)
	require.NoError(t, err)

	writer, err := ini.NewWriter(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return tmpFile.Name(), nil
	})
	require.NoError(t, err)

	backoffAt := writer.File.Section("internal").Key("backoff_at").MustTimeFormat(ini.DateFormat)

	assert.WithinDuration(t, time.Now(), backoffAt, 15*time.Second)
	assert.Equal(t, "2", writer.File.Section("internal").Key("backoff_retries").String())
}

func TestUpdateBackoffSettings_NotInBackoff(t *testing.T) {
	v := viper.New()

	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	v.Set("config", tmpFile.Name())
	v.Set("internal-config", tmpFile.Name())

	err = updateBackoffSettings(v, 0, time.Time{})
	require.NoError(t, err)

	writer, err := ini.NewWriter(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return tmpFile.Name(), nil
	})
	require.NoError(t, err)

	assert.Empty(t, writer.File.Section("internal").Key("backoff_at").String())
	assert.Equal(t, "0", writer.File.Section("internal").Key("backoff_retries").String())
}
