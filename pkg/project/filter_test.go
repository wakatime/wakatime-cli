package project_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithFiltering(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	first := testHeartbeat()
	first.Entity = tmpFile.Name()

	second := testHeartbeat()
	second.Project = nil

	opt := project.WithFiltering(project.FilterConfig{
		ExcludeUnknownProject: true,
	})
	h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Branch:         heartbeat.PointerTo("heartbeat"),
				Category:       heartbeat.CodingCategory,
				CursorPosition: heartbeat.PointerTo(12),
				Dependencies:   []string{"dep1", "dep2"},
				Entity:         tmpFile.Name(),
				EntityType:     heartbeat.FileType,
				IsWrite:        heartbeat.PointerTo(true),
				Language:       heartbeat.PointerTo("Go"),
				LineNumber:     heartbeat.PointerTo(42),
				Lines:          heartbeat.PointerTo(100),
				Project:        heartbeat.PointerTo("wakatime"),
				Time:           1585598060,
				UserAgent:      "wakatime/13.0.7",
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := h([]heartbeat.Heartbeat{first, second})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestFilter_ErrUnknownProject(t *testing.T) {
	tests := map[string]*string{
		"nil":          nil,
		"empty string": heartbeat.PointerTo(""),
	}

	for name, projectValue := range tests {
		t.Run(name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp(t.TempDir(), "")
			require.NoError(t, err)

			defer tmpFile.Close()

			h := heartbeat.Heartbeat{
				Entity:  tmpFile.Name(),
				Project: projectValue,
			}

			err = project.Filter(h, project.FilterConfig{
				ExcludeUnknownProject: true,
			})

			assert.EqualError(t, err, "skipping because of unknown project")
		})
	}
}

func testHeartbeat() heartbeat.Heartbeat {
	return heartbeat.Heartbeat{
		Branch:         heartbeat.PointerTo("heartbeat"),
		Category:       heartbeat.CodingCategory,
		CursorPosition: heartbeat.PointerTo(12),
		Dependencies:   []string{"dep1", "dep2"},
		EntityType:     heartbeat.FileType,
		IsWrite:        heartbeat.PointerTo(true),
		Language:       heartbeat.PointerTo("Go"),
		LineNumber:     heartbeat.PointerTo(42),
		Lines:          heartbeat.PointerTo(100),
		Project:        heartbeat.PointerTo("wakatime"),
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}
}
