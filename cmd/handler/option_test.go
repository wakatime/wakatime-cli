package handler_test

import (
	"context"
	"testing"

	"github.com/wakatime/wakatime-cli/cmd/handler"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/params"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithFormatting(t *testing.T) {
	opt := handler.WithFormatting()

	chain := heartbeat.NewHandle(noopMock{})
	hdl := opt(params.Params{})(chain)

	res, err := hdl(context.Background(), nil)
	require.NoError(t, err)

	assert.Len(t, res, 0)
}

func TestWithEntityModifier(t *testing.T) {
	opt := handler.WithEntityModifier()

	chain := heartbeat.NewHandle(noopMock{})
	hdl := opt(params.Params{})(chain)

	res, err := hdl(context.Background(), nil)
	require.NoError(t, err)

	assert.Len(t, res, 0)
}

func TestWithHeartbeatFiltering(t *testing.T) {
	opt := handler.WithHeartbeatFiltering()

	chain := heartbeat.NewHandle(noopMock{})
	hdl := opt(params.Params{})(chain)

	res, err := hdl(context.Background(), nil)
	require.NoError(t, err)

	assert.Len(t, res, 0)
}

func TestWithRemoteDetection(t *testing.T) {
	opt := handler.WithRemoteDetection()

	chain := heartbeat.NewHandle(noopMock{})
	hdl := opt(params.Params{})(chain)

	res, err := hdl(context.Background(), nil)
	require.NoError(t, err)

	assert.Len(t, res, 0)
}

func TestWithAPIKeyReplacing(t *testing.T) {
	opt := handler.WithAPIKeyReplacing()

	chain := heartbeat.NewHandle(noopMock{})
	hdl := opt(params.Params{})(chain)

	res, err := hdl(context.Background(), nil)
	require.NoError(t, err)

	assert.Len(t, res, 0)
}

func TestWithFileStatsDetection(t *testing.T) {
	opt := handler.WithFileStatsDetection()

	chain := heartbeat.NewHandle(noopMock{})
	hdl := opt(params.Params{})(chain)

	res, err := hdl(context.Background(), nil)
	require.NoError(t, err)

	assert.Len(t, res, 0)
}

func TestWithLanguageDetection(t *testing.T) {
	opt := handler.WithLanguageDetection()

	chain := heartbeat.NewHandle(noopMock{})
	hdl := opt(params.Params{})(chain)

	res, err := hdl(context.Background(), nil)
	require.NoError(t, err)

	assert.Len(t, res, 0)
}

func TestWithDependencyDetection(t *testing.T) {
	opt := handler.WithDependencyDetection()

	chain := heartbeat.NewHandle(noopMock{})
	hdl := opt(params.Params{})(chain)

	res, err := hdl(context.Background(), nil)
	require.NoError(t, err)

	assert.Len(t, res, 0)
}

func TestWithProjectDetection(t *testing.T) {
	opt := handler.WithProjectDetection()

	chain := heartbeat.NewHandle(noopMock{})
	hdl := opt(params.Params{})(chain)

	res, err := hdl(context.Background(), nil)
	require.NoError(t, err)

	assert.Len(t, res, 0)
}

func TestWithProjectFiltering(t *testing.T) {
	opt := handler.WithProjectFiltering()

	chain := heartbeat.NewHandle(noopMock{})
	hdl := opt(params.Params{})(chain)

	res, err := hdl(context.Background(), nil)
	require.NoError(t, err)

	assert.Len(t, res, 0)
}

func TestWithHeartbeatSanitization(t *testing.T) {
	opt := handler.WithHeartbeatSanitization()

	chain := heartbeat.NewHandle(noopMock{})
	hdl := opt(params.Params{})(chain)

	res, err := hdl(context.Background(), nil)
	require.NoError(t, err)

	assert.Len(t, res, 0)
}

func TestWithFileExpertsValidation(t *testing.T) {
	opt := handler.WithFileExpertsValidation()

	chain := heartbeat.NewHandle(noopMock{})
	hdl := opt(params.Params{})(chain)

	res, err := hdl(context.Background(), nil)
	require.NoError(t, err)

	assert.Len(t, res, 0)
}

func TestWithRemoteCleanup(t *testing.T) {
	opt := handler.WithRemoteCleanup()

	chain := heartbeat.NewHandle(noopMock{})
	hdl := opt(params.Params{})(chain)

	res, err := hdl(context.Background(), nil)
	require.NoError(t, err)

	assert.Len(t, res, 0)
}

func TestWithLengthValidator(t *testing.T) {
	opt := handler.WithLengthValidator()

	chain := heartbeat.NewHandle(noopMock{})
	hdl := opt(params.Params{})(chain)

	res, err := hdl(context.Background(), nil)
	require.NoError(t, err)

	assert.Len(t, res, 0)
}
