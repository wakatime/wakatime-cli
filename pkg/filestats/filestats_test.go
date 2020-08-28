package filestats_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/filestats"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithDetection(t *testing.T) {
	opt := filestats.WithDetection()
	handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Len(t, hh, 2)
		assert.Contains(t, hh, heartbeat.Heartbeat{
			EntityType: heartbeat.FileType,
			Entity:     "testdata/first.txt",
			Lines:      heartbeat.Int(1),
		})
		assert.Contains(t, hh, heartbeat.Heartbeat{
			EntityType: heartbeat.FileType,
			Entity:     "testdata/second.txt",
			Lines:      heartbeat.Int(2),
		})

		return []heartbeat.Result{
			{
				Status: 42,
			},
		}, nil
	})

	result, err := handle([]heartbeat.Heartbeat{
		{
			EntityType: heartbeat.FileType,
			Entity:     "testdata/first.txt",
		},
		{
			EntityType: heartbeat.FileType,
			Entity:     "testdata/second.txt",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 42,
		},
	}, result)
}
