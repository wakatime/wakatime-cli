package backoff_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/backoff"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithBackoff(t *testing.T) {
	v := viper.New()

	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	v.Set("internal-config", tmpFile.Name())

	opt := backoff.WithBackoff(backoff.Config{
		V: v,
	})

	handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	_, err = handle([]heartbeat.Heartbeat{})
	require.NoError(t, err)
}

func TestWithBackoff_BeforeNextBackoff(t *testing.T) {
	backoffAt := time.Now().Add(time.Second * -1)

	opt := backoff.WithBackoff(backoff.Config{
		At:      backoffAt,
		Retries: 1,
	})

	handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	_, err := handle([]heartbeat.Heartbeat{})
	require.Error(t, err)

	assert.Equal(t, "won't send heartbeat due to backoff", err.Error())
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

	handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{}, errors.New("error")
	})

	_, err = handle([]heartbeat.Heartbeat{})
	require.Error(t, err)

	assert.Equal(t, "error", err.Error())
}
