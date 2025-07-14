package ratelimit_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/ratelimit"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRateLimited(t *testing.T) {
	p := ratelimit.Params{
		Timeout:    time.Duration(offline.RateLimitDefaultSeconds) * time.Second,
		LastSentAt: time.Now(),
	}

	assert.True(t, ratelimit.IsRateLimited(p))
}

func TestIsRateLimited_NotLimited(t *testing.T) {
	p := ratelimit.Params{
		LastSentAt: time.Now().Add(time.Duration(-offline.RateLimitDefaultSeconds*2) * time.Second),
		Timeout:    time.Duration(offline.RateLimitDefaultSeconds) * time.Second,
	}

	assert.False(t, ratelimit.IsRateLimited(p))
}

func TestIsRateLimited_Disabled(t *testing.T) {
	p := ratelimit.Params{
		Disabled: true,
	}

	assert.False(t, ratelimit.IsRateLimited(p))
}

func TestIsRateLimited_TimeoutZero(t *testing.T) {
	p := ratelimit.Params{
		Timeout: 0,
	}

	assert.False(t, ratelimit.IsRateLimited(p))
}

func TestIsRateLimited_LastSentAtZero(t *testing.T) {
	p := ratelimit.Params{
		LastSentAt: time.Time{},
	}

	assert.False(t, ratelimit.IsRateLimited(p))
}

func TestRateLimitReset(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	tmpFileInternal, err := os.CreateTemp(t.TempDir(), "wakatime-internal")
	require.NoError(t, err)

	defer tmpFileInternal.Close()

	ctx := t.Context()

	v := viper.New()
	v.Set("config", tmpFileInternal.Name())
	v.Set("internal-config", tmpFileInternal.Name())

	writer, err := ini.NewWriter(ctx, v, func(_ context.Context, vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)

		return tmpFileInternal.Name(), nil
	})
	require.NoError(t, err)

	err = ratelimit.Reset(ctx, v)
	require.NoError(t, err)

	err = writer.File.Reload()
	require.NoError(t, err)

	lastSentAt, err := writer.File.Section("internal").Key("heartbeats_last_sent_at").TimeFormat(ini.DateFormat)
	require.NoError(t, err)

	assert.WithinDuration(t, time.Now(), lastSentAt, 1*time.Second)
}
