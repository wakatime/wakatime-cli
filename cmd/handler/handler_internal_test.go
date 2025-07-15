package handler

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	paramspkg "github.com/wakatime/wakatime-cli/pkg/params"

	viperini "github.com/go-viper/encoding/ini"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	iniv1 "gopkg.in/ini.v1"
)

func TestNoopSendHeartbeats(t *testing.T) {
	noop := noop{}

	res, err := noop.SendHeartbeats(t.Context(), []heartbeat.Heartbeat{
		{APIKey: "test-api-key", Entity: "test-entity", EntityType: heartbeat.FileType, IsUnsavedEntity: false},
		{APIKey: "test-api-key-2", Entity: "test-entity-2", EntityType: heartbeat.AppType, IsUnsavedEntity: true},
	})
	require.NoError(t, err)

	require.Len(t, res, 2)
	assert.EqualValues(
		t,
		heartbeat.Heartbeat{
			APIKey:          "test-api-key",
			Entity:          "test-entity",
			EntityType:      heartbeat.FileType,
			IsUnsavedEntity: false,
		},
		res[0].Heartbeat,
	)
	assert.EqualValues(
		t,
		heartbeat.Heartbeat{
			APIKey:          "test-api-key-2",
			Entity:          "test-entity-2",
			EntityType:      heartbeat.AppType,
			IsUnsavedEntity: true,
		},
		res[1].Heartbeat,
	)
}

func TestFindProjectConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	dir := filepath.Join(tmpDir, "src", "folder-1", "folder-2", "folder-3")

	err := os.MkdirAll(dir, os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.OpenFile(filepath.Join(dir, "some-entity.go"), os.O_RDONLY|os.O_CREATE, 0700)
	require.NoError(t, err)

	tmpProjectFile, err := os.OpenFile(filepath.Join(tmpDir, "src", ".wakatime"), os.O_RDONLY|os.O_CREATE, 0700)
	require.NoError(t, err)

	defer func() {
		tmpFile.Close()
		tmpProjectFile.Close()
	}()

	v := setupViper(t)

	params, ok, err := findProjectConfigFile(
		t.Context(), v, dir, heartbeat.FileType, false,
		func(_ context.Context, _ *viper.Viper) (paramspkg.Params, error) {
			return paramspkg.Params{}, nil
		},
	)
	require.NoError(t, err)

	assert.True(t, ok)
	assert.Equal(t, paramspkg.Params{}, params)
}

func TestFindProjectConfigFile_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	dir := filepath.Join(tmpDir, "src", "folder-1", "folder-2", "folder-3")

	err := os.MkdirAll(dir, os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.OpenFile(filepath.Join(dir, "some-entity.go"), os.O_RDONLY|os.O_CREATE, 0700)
	require.NoError(t, err)

	defer tmpFile.Close()

	v := setupViper(t)

	params, ok, err := findProjectConfigFile(
		t.Context(), v, dir, heartbeat.FileType, false,
		func(_ context.Context, _ *viper.Viper) (paramspkg.Params, error) {
			return paramspkg.Params{}, nil
		},
	)
	require.NoError(t, err)

	assert.False(t, ok)
	assert.Equal(t, paramspkg.Params{}, params)
}

func TestFindProjectConfigFile_NotFileType(t *testing.T) {
	params, ok, err := findProjectConfigFile(t.Context(), nil, "", heartbeat.AppType, false, nil)
	require.NoError(t, err)

	assert.False(t, ok)
	assert.Equal(t, paramspkg.Params{}, params)
}

func TestFindProjectConfigFile_IsUnsavedEntity(t *testing.T) {
	params, ok, err := findProjectConfigFile(t.Context(), nil, "", heartbeat.FileType, true, nil)
	require.NoError(t, err)

	assert.False(t, ok)
	assert.Equal(t, paramspkg.Params{}, params)
}

func TestFindProjectConfigFile_ParamsLoader_Err(t *testing.T) {
	tmpDir := t.TempDir()

	dir := filepath.Join(tmpDir, "src", "folder-1", "folder-2", "folder-3")

	err := os.MkdirAll(dir, os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.OpenFile(filepath.Join(dir, "some-entity.go"), os.O_RDONLY|os.O_CREATE, 0700)
	require.NoError(t, err)

	tmpProjectFile, err := os.OpenFile(filepath.Join(tmpDir, "src", ".wakatime"), os.O_RDONLY|os.O_CREATE, 0700)
	require.NoError(t, err)

	defer func() {
		tmpFile.Close()
		tmpProjectFile.Close()
	}()

	v := setupViper(t)

	params, ok, err := findProjectConfigFile(
		t.Context(), v, dir, heartbeat.FileType, false,
		func(_ context.Context, _ *viper.Viper) (paramspkg.Params, error) {
			return paramspkg.Params{}, errors.New("fail")
		},
	)
	require.EqualError(t, err, "failed to load project-level parameters: fail")

	assert.False(t, ok)
	assert.Equal(t, paramspkg.Params{}, params)
}

func setupViper(t *testing.T) *viper.Viper {
	multilineOption := iniv1.LoadOptions{AllowPythonMultilineValues: true}
	iniCodec := viperini.Codec{LoadOptions: multilineOption}

	codecRegistry := viper.NewCodecRegistry()
	err := codecRegistry.RegisterCodec("ini", iniCodec)
	require.NoError(t, err)

	v := viper.NewWithOptions(viper.WithCodecRegistry(codecRegistry))

	return v
}
