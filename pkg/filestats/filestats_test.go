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
	opt := filestats.WithDetection(filestats.Config{})
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

func TestWithDetection_LinesInFile(t *testing.T) {
	opt := filestats.WithDetection(filestats.Config{
		LinesInFile: heartbeat.Int(158),
	})
	handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Len(t, hh, 1)
		assert.Contains(t, hh, heartbeat.Heartbeat{
			EntityType: heartbeat.FileType,
			Entity:     "/path/to/remote",
			LocalFile:  "testdata/first.txt",
			Lines:      heartbeat.Int(158),
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
			Entity:     "/path/to/remote",
			LocalFile:  "testdata/first.txt",
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
	f, err := os.CreateTemp(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(f.Name())

	b := bytes.NewBuffer(make([]byte, 2*1024*1024+1))
	_, err = f.Write(b.Bytes())
	require.NoError(t, err)

	opt := filestats.WithDetection(filestats.Config{})
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
