package offline_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	cmdoffline "github.com/wakatime/wakatime-cli/cmd/offline"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/offline"

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

	ctx := context.Background()

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

	err = cmdoffline.SaveHeartbeats(ctx, v, nil, offlineQueueFile.Name())
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

	ctx := context.Background()

	data, err := os.ReadFile("testdata/extra_heartbeats.json")
	require.NoError(t, err)

	var hh []heartbeat.Heartbeat

	err = json.Unmarshal(data, &hh)
	require.NoError(t, err)

	v := viper.New()
	v.Set("config", tmpFile.Name())
	v.Set("entity", "testdata/main.go")
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	err = cmdoffline.SaveHeartbeats(ctx, v, hh, offlineQueueFile.Name())
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

	ctx := context.Background()

	v := viper.New()
	v.Set("config", tmpFile.Name())
	v.Set("disable-offline", true)
	v.Set("entity", "testdata/main.go")
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	err = cmdoffline.SaveHeartbeats(ctx, v, nil, offlineQueueFile.Name())

	assert.EqualError(t, err, "saving to offline db disabled")
}
