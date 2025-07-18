package handler_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/cmd/handler"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/params"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerNew(t *testing.T) {
	tmpDir := t.TempDir()

	dir := filepath.Join(tmpDir, "src", "folder-1", "folder-2", "folder-3")

	err := os.MkdirAll(dir, os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpEntity, err := os.OpenFile(filepath.Join(dir, "some-entity.go"), os.O_RDONLY|os.O_CREATE, 0700)
	require.NoError(t, err)

	err = tmpEntity.Close()
	require.NoError(t, err)

	wakatimeProjectFile, err := os.OpenFile(filepath.Join(tmpDir, "src", ".wakatime"), os.O_RDONLY|os.O_CREATE, 0700)
	require.NoError(t, err)

	err = wakatimeProjectFile.Close()
	require.NoError(t, err)

	v := vipertools.MustNew()

	sender := newHandle(noopMock{})
	hdl := handler.New(v, handler.Config{
		Params: params.Params{},
		ParamsLoader: func(_ context.Context, _ *viper.Viper, _ params.FlagReadOrder) (params.Params, error) {
			return params.Params{
				// this simulates the project-level parameters loaded from the .wakatime file.
				API: params.API{Key: "00000000-0000-4000-8000-000000000002"},
			}, nil
		},
		Opts: []handler.Preprocessor{
			// we're only testing the API key replacement, so we need to add this option.
			handler.WithAPIKeyReplacing(),
		},
	})(sender)

	res, err := hdl(t.Context(), []heartbeat.Heartbeat{
		{
			APIKey:          "00000000-0000-4000-8000-000000000001",
			Entity:          filepath.Join(dir, "some-entity.go"),
			EntityType:      heartbeat.FileType,
			IsUnsavedEntity: false,
		},
	})
	require.NoError(t, err)

	assert.Len(t, res, 1)
	assert.Equal(t, res[0].Heartbeat.APIKey, "00000000-0000-4000-8000-000000000002")
}

func newHandle(sender heartbeat.Sender) heartbeat.Handle {
	return func(ctx context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		var handle heartbeat.Handle = sender.SendHeartbeats
		return handle(ctx, hh)
	}
}

type noopMock struct{}

// SendHeartbeats always returns results containing formatted heartbeats and no error.
func (noopMock) SendHeartbeats(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
	res := make([]heartbeat.Result, len(hh))
	for i, h := range hh {
		res[i] = heartbeat.Result{
			Heartbeat: h,
		}
	}

	return res, nil
}
