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
	tmpDir, err := os.MkdirTemp(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	tmpFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

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
				Branch:         heartbeat.String("heartbeat"),
				Category:       heartbeat.CodingCategory,
				CursorPosition: heartbeat.Int(12),
				Dependencies:   []string{"dep1", "dep2"},
				Entity:         tmpFile.Name(),
				EntityType:     heartbeat.FileType,
				IsWrite:        heartbeat.Bool(true),
				Language:       heartbeat.String("Go"),
				LineNumber:     heartbeat.Int(42),
				Lines:          heartbeat.Int(100),
				Project:        heartbeat.String("wakatime"),
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
		"empty string": heartbeat.String(""),
	}

	for name, projectValue := range tests {
		t.Run(name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp(os.TempDir(), "wakatime")
			require.NoError(t, err)

			defer os.RemoveAll(tmpDir)

			tmpFile, err := os.CreateTemp(tmpDir, "")
			require.NoError(t, err)

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
		Branch:         heartbeat.String("heartbeat"),
		Category:       heartbeat.CodingCategory,
		CursorPosition: heartbeat.Int(12),
		Dependencies:   []string{"dep1", "dep2"},
		EntityType:     heartbeat.FileType,
		IsWrite:        heartbeat.Bool(true),
		Language:       heartbeat.String("Go"),
		LineNumber:     heartbeat.Int(42),
		Lines:          heartbeat.Int(100),
		Project:        heartbeat.String("wakatime"),
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}
}
