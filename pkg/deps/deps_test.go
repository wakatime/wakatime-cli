package deps_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithDetection(t *testing.T) {
	opt := deps.WithDetection()

	h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Dependencies: []string{`"os"`, `"github.com/wakatime/wakatime-cli/pkg/heartbeat"`},
				Entity:       "testdata/golang_minimal.go",
				EntityType:   heartbeat.FileType,
				Language:     heartbeat.LanguageGo,
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := h([]heartbeat.Heartbeat{{
		Entity:     "testdata/golang_minimal.go",
		EntityType: heartbeat.FileType,
		Language:   heartbeat.LanguageGo,
	}})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestWithDetection_LocalFile(t *testing.T) {
	opt := deps.WithDetection()

	h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Dependencies: []string{`"os"`, `"github.com/wakatime/wakatime-cli/pkg/heartbeat"`},
				Entity:       "testdata/golang.go",
				EntityType:   heartbeat.FileType,
				Language:     heartbeat.LanguageGo,
				LocalFile:    "testdata/golang_minimal.go",
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := h([]heartbeat.Heartbeat{{
		Entity:     "testdata/golang.go",
		EntityType: heartbeat.FileType,
		Language:   heartbeat.LanguageGo,
		LocalFile:  "testdata/golang_minimal.go",
	}})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestWithDetection_NonFileType(t *testing.T) {
	opt := deps.WithDetection()

	h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Entity:     "testdata/codefiles/golang.go",
				EntityType: heartbeat.AppType,
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := h([]heartbeat.Heartbeat{{
		Entity:     "testdata/codefiles/golang.go",
		EntityType: heartbeat.AppType,
	}})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}
