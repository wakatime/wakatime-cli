package offline_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cmdheartbeat "github.com/wakatime/wakatime-cli/cmd/heartbeat"
	cmdoffline "github.com/wakatime/wakatime-cli/cmd/offline"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/params"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveHeartbeats(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	ctx := t.Context()

	v := viper.New()
	v.Set("config", tmpFile.Name())
	v.Set("category", "debugging")
	v.Set("cursorpos", 42)
	v.Set("entity", "testdata/main.go")
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("language", "Go")
	v.Set("alternate-language", "Golang")
	v.Set("hide-branch-names", true)
	v.Set("project", "wakatime-cli")
	v.Set("lineno", 13)
	v.Set("time", 1585598059.1)
	v.Set("timeout", 5)
	v.Set("write", true)

	hh, err := testBuildHeartbeats(ctx, v)
	require.NoError(t, err)

	err = cmdoffline.SaveHeartbeats(ctx, v, offlineQueueFile.Name(), hh)
	require.NoError(t, err)

	offlineCount, err := offline.CountHeartbeats(ctx, offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Equal(t, 1, offlineCount)
}

func TestSaveHeartbeatsWithParams(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	ctx := t.Context()

	v := viper.New()
	v.Set("config", tmpFile.Name())
	v.Set("entity", "testdata/main.go")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("project", "wakatime-cli")

	hh, err := testBuildHeartbeats(ctx, v)
	require.NoError(t, err)

	apiParams, err := params.LoadAPIParams(ctx, v, params.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	heartbeatParams, err := params.LoadHeartbeatParams(ctx, v, params.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	err = cmdoffline.SaveHeartbeatsWithParams(ctx, v, offlineQueueFile.Name(), hh, params.Params{
		API:       apiParams,
		Heartbeat: heartbeatParams,
		Offline:   params.LoadOfflineParams(ctx, v, params.FlagReadOrderFlagPrecedence),
	})
	require.NoError(t, err)

	offlineCount, err := offline.CountHeartbeats(ctx, offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Equal(t, 1, offlineCount)
}

func TestSaveHeartbeats_ExtraHeartbeats(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	ctx := t.Context()

	data, err := os.ReadFile("testdata/extra_heartbeats.json")
	require.NoError(t, err)

	var hh []heartbeat.Heartbeat

	err = json.Unmarshal(data, &hh)
	require.NoError(t, err)

	v := viper.New()
	v.Set("config", tmpFile.Name())
	v.Set("entity", "testdata/main.go")
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	err = cmdoffline.SaveHeartbeats(ctx, v, offlineQueueFile.Name(), hh)
	require.NoError(t, err)

	offlineCount, err := offline.CountHeartbeats(ctx, offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Equal(t, 25, offlineCount)
}

func TestSaveHeartbeats_OfflineDisabled(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	ctx := t.Context()

	v := viper.New()
	v.Set("config", tmpFile.Name())
	v.Set("disable-offline", true)
	v.Set("entity", "testdata/main.go")
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	hh, err := testBuildHeartbeats(ctx, v)
	require.NoError(t, err)

	err = cmdoffline.SaveHeartbeats(ctx, v, offlineQueueFile.Name(), hh)

	assert.EqualError(t, err, "saving to offline db disabled")
}

func TestSaveHeartbeats_NilViper(t *testing.T) {
	err := cmdoffline.SaveHeartbeats(
		t.Context(),
		nil,
		filepath.Join(t.TempDir(), "offline.bdb"),
		[]heartbeat.Heartbeat{{Entity: "main.go"}},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load command parameters")
	assert.Contains(t, err.Error(), "viper instance unset")
}

func TestSaveHeartbeats_NoHeartbeats(t *testing.T) {
	queueFile, err := os.CreateTemp(t.TempDir(), "offline-queue")
	require.NoError(t, err)
	require.NoError(t, queueFile.Close())

	v := viper.New()
	v.Set("entity", "testdata/main.go")
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	err = cmdoffline.SaveHeartbeats(t.Context(), v, queueFile.Name(), nil)
	require.NoError(t, err)

	count, err := offline.CountHeartbeats(t.Context(), queueFile.Name())
	require.NoError(t, err)
	assert.Zero(t, count)
}

func testBuildHeartbeats(ctx context.Context, v *viper.Viper) ([]heartbeat.Heartbeat, error) {
	apiParams, err := params.LoadAPIParams(ctx, v, params.FlagReadOrderFlagPrecedence)
	if err != nil {
		return nil, fmt.Errorf("failed to load API parameters: %w", err)
	}

	heartbeatParams, err := params.LoadHeartbeatParams(ctx, v, params.FlagReadOrderFlagPrecedence)
	if err != nil {
		return nil, fmt.Errorf("failed to load heartbeat params: %w", err)
	}

	return cmdheartbeat.BuildHeartbeats(ctx, apiParams.Plugin, heartbeatParams), nil
}
