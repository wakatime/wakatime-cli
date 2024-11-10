package backoff_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/backoff"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithBackoff(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	v := viper.New()
	v.Set("internal-config", tmpFile.Name())

	opt := backoff.WithBackoff(backoff.Config{
		V: v,
	})

	handle := opt(func(_ context.Context, _ []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	_, err = handle(context.Background(), []heartbeat.Heartbeat{})
	require.NoError(t, err)

	err = ini.ReadInConfig(v, tmpFile.Name())
	require.NoError(t, err)

	// make sure backoff settings were not written
	assert.Empty(t, v.GetString("internal.backoff_at"))
	assert.Empty(t, v.GetString("internal.backoff_retries"))
}

func TestWithBackoff_BeforeNextBackoff(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	ctx := context.Background()

	v := viper.New()
	v.Set("internal-config", tmpFile.Name())

	at := time.Now()

	// first, cause backoff to be set
	opt := backoff.WithBackoff(backoff.Config{
		V:       v,
		Retries: 0,
		At:      at,
	})

	handle := opt(func(_ context.Context, _ []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{}, errors.New("error")
	})

	_, err = handle(ctx, []heartbeat.Heartbeat{})
	require.Error(t, err)

	assert.Equal(t, "error", err.Error())

	err = ini.ReadInConfig(v, tmpFile.Name())
	require.NoError(t, err)

	// make sure backoff settings written
	assert.Equal(t, at.Format(ini.DateFormat), v.GetString("internal.backoff_at"))
	assert.Equal(t, "1", v.GetString("internal.backoff_retries"))

	// then, make sure it's reset when max backoff reached
	opt = backoff.WithBackoff(backoff.Config{
		V:       v,
		Retries: 1,
		At:      at.Add(time.Second * 15),
	})

	handle = opt(func(_ context.Context, _ []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	_, err = handle(ctx, []heartbeat.Heartbeat{})
	require.Error(t, err)

	assert.Equal(t, "won't send heartbeat due to backoff without proxy", err.Error())

	err = ini.ReadInConfig(v, tmpFile.Name())
	require.NoError(t, err)

	// make sure backoff settings not written
	assert.Equal(t, at.Format(ini.DateFormat), v.GetString("internal.backoff_at"))
	assert.Equal(t, "1", v.GetString("internal.backoff_retries"))
}

func TestWithBackoff_BeforeNextBackoffWithProxy(t *testing.T) {
	backoffAt := time.Now().Add(time.Second * -1)

	opt := backoff.WithBackoff(backoff.Config{
		At:       backoffAt,
		Retries:  1,
		HasProxy: true,
	})

	handle := opt(func(_ context.Context, _ []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	_, err := handle(context.Background(), []heartbeat.Heartbeat{})
	require.Error(t, err)

	assert.Equal(t, "won't send heartbeat due to backoff with proxy", err.Error())
}

func TestWithBackoff_ApiError(t *testing.T) {
	v := viper.New()

	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	v.Set("internal-config", tmpFile.Name())

	opt := backoff.WithBackoff(backoff.Config{
		V: v,
	})

	handle := opt(func(_ context.Context, _ []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{}, errors.New("error")
	})

	_, err = handle(context.Background(), []heartbeat.Heartbeat{})
	require.Error(t, err)

	assert.Equal(t, "error", err.Error())

	err = ini.ReadInConfig(v, tmpFile.Name())
	require.NoError(t, err)

	// make sure backoff settings written
	assert.NotEmpty(t, v.GetString("internal.backoff_at"))
	assert.Equal(t, "1", v.GetString("internal.backoff_retries"))
}

func TestWithBackoff_BackoffAndNotReset(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	v := viper.New()
	v.Set("internal-config", tmpFile.Name())

	opt := backoff.WithBackoff(backoff.Config{
		V:       v,
		Retries: 1,
		At:      time.Now().Add(time.Second * -1),
	})

	handle := opt(func(_ context.Context, _ []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	_, err = handle(context.Background(), []heartbeat.Heartbeat{})
	require.Error(t, err)

	var errbackoff api.ErrBackoff

	assert.ErrorAs(t, err, &errbackoff)

	err = ini.ReadInConfig(v, tmpFile.Name())
	require.NoError(t, err)

	// make sure backoff settings were not written because we didn't attempt sending
	assert.Empty(t, v.GetString("internal.backoff_at"))
	assert.Empty(t, v.GetString("internal.backoff_retries"))
}

func TestWithBackoff_BackoffMaxReached(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	ctx := context.Background()

	v := viper.New()
	v.Set("internal-config", tmpFile.Name())

	// first, cause backoff to be set
	opt := backoff.WithBackoff(backoff.Config{
		V: v,
	})

	handle := opt(func(_ context.Context, _ []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{}, errors.New("error")
	})

	_, err = handle(ctx, []heartbeat.Heartbeat{})
	require.Error(t, err)

	err = ini.ReadInConfig(v, tmpFile.Name())
	require.NoError(t, err)

	// make sure backoff settings written
	assert.NotEmpty(t, v.GetString("internal.backoff_at"))
	assert.Equal(t, "1", v.GetString("internal.backoff_retries"))

	// then, make sure it's reset when max backoff reached
	opt = backoff.WithBackoff(backoff.Config{
		V:       v,
		Retries: 8,
		At:      time.Now().Add(time.Second * -1),
	})

	handle = opt(func(_ context.Context, _ []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	_, err = handle(ctx, []heartbeat.Heartbeat{})
	require.NoError(t, err)

	err = ini.ReadInConfig(v, tmpFile.Name())
	require.NoError(t, err)

	// make sure backoff settings reset
	assert.Empty(t, v.GetString("internal.backoff_at"))
	assert.Equal(t, "0", v.GetString("internal.backoff_retries"))
}

func TestWithBackoff_BackoffMaxReachedWithZeroRetries(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	ctx := context.Background()

	v := viper.New()
	v.Set("internal-config", tmpFile.Name())

	// first, cause backoff to be set
	opt := backoff.WithBackoff(backoff.Config{
		V: v,
	})

	handle := opt(func(_ context.Context, _ []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{}, errors.New("error")
	})

	_, err = handle(ctx, []heartbeat.Heartbeat{})
	require.Error(t, err)

	err = ini.ReadInConfig(v, tmpFile.Name())
	require.NoError(t, err)

	// make sure backoff settings written
	assert.NotEmpty(t, v.GetString("internal.backoff_at"))
	assert.Equal(t, "1", v.GetString("internal.backoff_retries"))

	// then, make sure it's reset when max backoff reached
	opt = backoff.WithBackoff(backoff.Config{
		V:       v,
		Retries: 0,
		At:      time.Now().Add(time.Hour + 1*time.Second),
	})

	handle = opt(func(_ context.Context, _ []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	_, err = handle(ctx, []heartbeat.Heartbeat{})
	require.NoError(t, err)

	err = ini.ReadInConfig(v, tmpFile.Name())
	require.NoError(t, err)

	// make sure backoff settings reset
	assert.Empty(t, v.GetString("internal.backoff_at"))
	assert.Equal(t, "0", v.GetString("internal.backoff_retries"))
}

func TestWithBackoff_ShouldRetry(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	ctx := context.Background()

	v := viper.New()
	v.Set("internal-config", tmpFile.Name())

	opt := backoff.WithBackoff(backoff.Config{
		V: v,
	})

	handle := opt(func(_ context.Context, _ []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{}, errors.New("error")
	})

	_, err = handle(ctx, []heartbeat.Heartbeat{})
	require.Error(t, err)

	assert.Equal(t, "error", err.Error())

	err = ini.ReadInConfig(v, tmpFile.Name())
	require.NoError(t, err)

	// make sure backoff settings written
	assert.NotEmpty(t, v.GetString("internal.backoff_at"))
	assert.Equal(t, "1", v.GetString("internal.backoff_retries"))

	at := time.Now()

	// then, make sure we retry if after rate limit
	opt = backoff.WithBackoff(backoff.Config{
		V:       v,
		Retries: 1,
		At:      at.Add(time.Second * -60),
	})

	handle = opt(func(_ context.Context, _ []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	_, err = handle(ctx, []heartbeat.Heartbeat{})
	require.NoError(t, err)

	err = ini.ReadInConfig(v, tmpFile.Name())
	require.NoError(t, err)

	// make sure backoff settings reset
	assert.Empty(t, v.GetString("internal.backoff_at"))
	assert.Equal(t, "0", v.GetString("internal.backoff_retries"))
}
