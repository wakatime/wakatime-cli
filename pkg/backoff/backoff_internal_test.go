package backoff

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/config"

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

	tmpFile, err := ioutil.TempFile(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	v.Set("config", tmpFile.Name())

	at := time.Now().Add(time.Second * -1)

	err = updateBackoffSettings(v, 2, at)
	require.NoError(t, err)

	ini, err := config.NewIniWriter(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return tmpFile.Name(), nil
	})
	require.NoError(t, err)

	backoffAt := ini.File.Section("internal").Key("backoff_at").MustTimeFormat(config.DateFormat)

	assert.WithinDuration(t, time.Now(), backoffAt, 15*time.Second)
	assert.Equal(t, "2", ini.File.Section("internal").Key("backoff_retries").String())
}

func TestUpdateBackoffSettings_NotInBackoff(t *testing.T) {
	v := viper.New()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	v.Set("config", tmpFile.Name())

	err = updateBackoffSettings(v, 0, time.Time{})
	require.NoError(t, err)

	ini, err := config.NewIniWriter(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return tmpFile.Name(), nil
	})
	require.NoError(t, err)

	assert.Empty(t, ini.File.Section("internal").Key("backoff_at").String())
	assert.Equal(t, "0", ini.File.Section("internal").Key("backoff_retries").String())
}
