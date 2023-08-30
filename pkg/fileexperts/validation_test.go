package fileexperts_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/fileexperts"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithValidation(t *testing.T) {
	opt := fileexperts.WithValidation()
	h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Entity:           "/path/to/file",
				Project:          heartbeat.PointerTo("wakatime"),
				ProjectRootCount: heartbeat.PointerTo(3),
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := h([]heartbeat.Heartbeat{
		{
			Entity:           "/path/to/file",
			Project:          heartbeat.PointerTo("wakatime"),
			ProjectRootCount: heartbeat.PointerTo(3),
		},
	})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestValidate_EmptyEntity(t *testing.T) {
	h := heartbeat.Heartbeat{}

	result := fileexperts.Validate(h)

	assert.False(t, result)
}

func TestValidate_NilProject(t *testing.T) {
	h := heartbeat.Heartbeat{
		Entity: "/path/to/file",
	}

	result := fileexperts.Validate(h)

	assert.False(t, result)
}

func TestValidate_EmptyProject(t *testing.T) {
	h := heartbeat.Heartbeat{
		Entity:  "/path/to/file",
		Project: heartbeat.PointerTo(""),
	}

	result := fileexperts.Validate(h)

	assert.False(t, result)
}

func TestValidate_NilProjectRootCount(t *testing.T) {
	h := heartbeat.Heartbeat{
		Entity:  "/path/to/file",
		Project: heartbeat.PointerTo("waktime"),
	}

	result := fileexperts.Validate(h)

	assert.False(t, result)
}

func TestValidate_ZeroProjectRootCount(t *testing.T) {
	h := heartbeat.Heartbeat{
		Entity:           "/path/to/file",
		Project:          heartbeat.PointerTo("waktime"),
		ProjectRootCount: heartbeat.PointerTo(0),
	}

	result := fileexperts.Validate(h)

	assert.False(t, result)
}
