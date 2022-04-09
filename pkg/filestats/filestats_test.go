package filestats_test

import (
	"bytes"
	"os"
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
			Lines:      heartbeat.PointerTo(1),
		})
		assert.Contains(t, hh, heartbeat.Heartbeat{
			EntityType: heartbeat.FileType,
			Entity:     "testdata/second.txt",
			Lines:      heartbeat.PointerTo(2),
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

func TestWithDetection_RemoteFile(t *testing.T) {
	opt := filestats.WithDetection()
	handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Len(t, hh, 1)
		assert.Contains(t, hh, heartbeat.Heartbeat{
			EntityType: heartbeat.FileType,
			Entity:     "ssh://192.168.1.1/path/to/remote/main.go",
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
			Entity:     "ssh://192.168.1.1/path/to/remote/main.go",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 42,
		},
	}, result)
}

func TestWithDetection_MaxFileSizeExceeded(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	b := bytes.NewBuffer(make([]byte, 2*1024*1024+1))
	_, err = f.Write(b.Bytes())
	require.NoError(t, err)

	opt := filestats.WithDetection()
	handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, hh, []heartbeat.Heartbeat{
			{
				EntityType: heartbeat.FileType,
				Entity:     f.Name(),
				Lines:      nil,
			},
		})

		return []heartbeat.Result{}, nil
	})

	_, err = handle([]heartbeat.Heartbeat{
		{
			EntityType: heartbeat.FileType,
			Entity:     f.Name(),
		},
	})
	require.NoError(t, err)
}
