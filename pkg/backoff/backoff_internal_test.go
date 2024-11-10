package backoff

import (
	"context"
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

	should := shouldBackoff(context.Background(), 1, at)

	assert.True(t, should)
}

func TestShouldBackoff_AfterResetTime(t *testing.T) {
	at := time.Now().Add(time.Second * -1)

	should := shouldBackoff(context.Background(), 8, at)

	assert.False(t, should)
}

func TestShouldBackoff_AfterResetTime_ZeroRetries(t *testing.T) {
	at := time.Now().Add(maxBackoffSecs + 1*time.Second)

	should := shouldBackoff(context.Background(), 0, at)

	assert.False(t, should)
}

func TestShouldBackoff_NegateBackoff(t *testing.T) {
	should := shouldBackoff(context.Background(), 0, time.Time{})

	assert.False(t, should)
}

func TestUpdateBackoffSettings(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	ctx := context.Background()

	v := viper.New()
	v.Set("config", tmpFile.Name())
	v.Set("internal-config", tmpFile.Name())

	at := time.Now().Add(time.Second * -1)

	err = updateBackoffSettings(ctx, v, 2, at)
	require.NoError(t, err)

	writer, err := ini.NewWriter(ctx, v, func(_ context.Context, vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return tmpFile.Name(), nil
	})
	require.NoError(t, err)

	backoffAt := writer.File.Section("internal").Key("backoff_at").MustTimeFormat(ini.DateFormat)

	assert.WithinDuration(t, time.Now(), backoffAt, 15*time.Second)
	assert.Equal(t, "2", writer.File.Section("internal").Key("backoff_retries").String())
}

func TestUpdateBackoffSettings_NotInBackoff(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	ctx := context.Background()

	v := viper.New()
	v.Set("config", tmpFile.Name())
	v.Set("internal-config", tmpFile.Name())

	err = updateBackoffSettings(ctx, v, 0, time.Time{})
	require.NoError(t, err)

	writer, err := ini.NewWriter(ctx, v, func(_ context.Context, vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return tmpFile.Name(), nil
	})
	require.NoError(t, err)

	assert.Empty(t, writer.File.Section("internal").Key("backoff_at").String())
	assert.Equal(t, "0", writer.File.Section("internal").Key("backoff_retries").String())
}
