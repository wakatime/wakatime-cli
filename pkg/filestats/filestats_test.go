package filestats_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/filestats"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithDetection(t *testing.T) {
	opt := filestats.WithDetection()
	handle := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
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

	result, err := handle(t.Context(), []heartbeat.Heartbeat{
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
	handle := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
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

	result, err := handle(t.Context(), []heartbeat.Heartbeat{
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

	b := bytes.NewBuffer(make([]byte, 5*1024*1024+1))
	_, err = f.Write(b.Bytes())
	require.NoError(t, err)

	opt := filestats.WithDetection()
	handle := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, hh, []heartbeat.Heartbeat{
			{
				EntityType: heartbeat.FileType,
				Entity:     f.Name(),
				Lines:      nil,
			},
		})

		return []heartbeat.Result{}, nil
	})

	_, err = handle(t.Context(), []heartbeat.Heartbeat{
		{
			EntityType: heartbeat.FileType,
			Entity:     f.Name(),
		},
	})
	require.NoError(t, err)
}

func TestWithDetection_SkipsUnsupportedInputs(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "existing.txt")
	require.NoError(t, os.WriteFile(existing, []byte("one\n"), 0600))

	opt := filestats.WithDetection()
	handle := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		require.Len(t, hh, 6)
		assert.Nil(t, hh[0].Lines)
		assert.Nil(t, hh[1].Lines)
		assert.Equal(t, 9, *hh[2].Lines)
		assert.Nil(t, hh[3].Lines)
		assert.Nil(t, hh[4].Lines)
		assert.Nil(t, hh[5].Lines)

		return []heartbeat.Result{{Status: 42}}, nil
	})

	result, err := handle(t.Context(), []heartbeat.Heartbeat{
		{EntityType: heartbeat.AppType, Entity: "app"},
		{EntityType: heartbeat.FileType, Entity: existing, IsUnsavedEntity: true},
		{EntityType: heartbeat.FileType, Entity: existing, Lines: heartbeat.PointerTo(9)},
		{EntityType: heartbeat.FileType, Entity: filepath.Join(dir, "missing.txt")},
		{EntityType: heartbeat.FileType, Entity: dir},
		{EntityType: heartbeat.FileType, Entity: "remote.go", LocalFile: filepath.Join(dir, "missing-local.txt")},
	})
	require.NoError(t, err)
	assert.Equal(t, []heartbeat.Result{{Status: 42}}, result)
}
