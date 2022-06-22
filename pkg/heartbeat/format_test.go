package heartbeat_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/windows"
	"github.com/yookoala/realpath"
)

func TestWithFormatting(t *testing.T) {
	opt := heartbeat.WithFormatting()

	handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		entity, err := filepath.Abs(hh[0].Entity)
		require.NoError(t, err)

		entity, err = realpath.Realpath(entity)
		require.NoError(t, err)

		if runtime.GOOS == "windows" {
			entity = windows.FormatFilePath(entity)
		}

		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Entity: entity,
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := handle([]heartbeat.Heartbeat{{
		Entity:     "testdata/main.go",
		EntityType: heartbeat.FileType,
	}})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestFormat_NetworkMount(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping because OS is not windows.")
	}

	h := heartbeat.Heartbeat{
		Entity:     `\\192.168.1.1\apilibrary.sl`,
		EntityType: heartbeat.FileType,
	}

	r := heartbeat.Format(h)

	assert.Equal(t, heartbeat.Heartbeat{
		Entity:     `\\192.168.1.1/apilibrary.sl`,
		EntityType: heartbeat.FileType,
	}, r)
}
