package filter_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/filter"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/regex"

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
	second.Time++

	opt := filter.WithFiltering(filter.Config{})
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

func TestWithFiltering_AbortAllFiltered(t *testing.T) {
	opt := filter.WithFiltering(filter.Config{})
	h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{}, errors.New("this will should never be called")
	})

	result, err := h([]heartbeat.Heartbeat{testHeartbeat()})
	require.NoError(t, err)

	assert.Equal(t, result, []heartbeat.Result{})
}

func TestFilter(t *testing.T) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	tmpFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	h := testHeartbeat()
	h.Entity = tmpFile.Name()

	err = filter.Filter(h, filter.Config{})
	require.NoError(t, err)
}

func TestFilter_NonFileTypeEmptyEntity(t *testing.T) {
	h := testHeartbeat()
	h.Entity = ""
	h.EntityType = heartbeat.AppType

	err := filter.Filter(h, filter.Config{
		ExcludeUnknownProject: true,
	})
	require.NoError(t, err)
}

func TestFilter_IncludeMatchOverwritesExcludeMatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	tmpFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	h := testHeartbeat()
	h.Entity = tmpFile.Name()

	err = filter.Filter(h, filter.Config{
		Exclude: []regex.Regex{
			regex.MustCompile(".*main.go$"),
		},
		Include: []regex.Regex{
			regex.MustCompile(".*/tmp/.*"),
		},
	})
	require.NoError(t, err)
}

func TestFilter_ErrMatchesExcludePattern(t *testing.T) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	tmpFile, err := os.CreateTemp(tmpDir, "exclude-this-file")
	require.NoError(t, err)

	h := testHeartbeat()
	h.Entity = tmpFile.Name()

	err = filter.Filter(h, filter.Config{
		Exclude: []regex.Regex{
			regex.MustCompile("^.*exclude-this-file.*$"),
		},
	})

	assert.EqualError(t, err, "filter by pattern: skipping because matches exclude pattern \"^.*exclude-this-file.*$\"")
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

			h := testHeartbeat()
			h.Entity = tmpFile.Name()
			h.Project = projectValue

			err = filter.Filter(h, filter.Config{
				ExcludeUnknownProject: true,
			})

			assert.EqualError(t, err, "skipping because of unknown project")
		})
	}
}

func TestFilter_ErrNonExistingFile(t *testing.T) {
	h := testHeartbeat()

	err := filter.Filter(h, filter.Config{})

	assert.EqualError(t, err, "filter file: skipping because of non-existing file \"/tmp/main.go\"")
}

func TestFilter_ExistingProjectFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	tmpFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	_, err = os.Create(filepath.Join(tmpDir, ".wakatime-project"))
	require.NoError(t, err)

	h := testHeartbeat()
	h.Entity = tmpFile.Name()

	err = filter.Filter(h, filter.Config{
		IncludeOnlyWithProjectFile: true,
	})
	require.NoError(t, err)
}

func TestFilter_ErrNonExistingProjectFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	tmpFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	h := testHeartbeat()
	h.Entity = tmpFile.Name()

	err = filter.Filter(h, filter.Config{
		IncludeOnlyWithProjectFile: true,
	})

	assert.EqualError(t, err, "filter file: skipping because missing .wakatime-project file in parent path")
}

func testHeartbeat() heartbeat.Heartbeat {
	return heartbeat.Heartbeat{
		Branch:         heartbeat.String("heartbeat"),
		Category:       heartbeat.CodingCategory,
		CursorPosition: heartbeat.Int(12),
		Dependencies:   []string{"dep1", "dep2"},
		Entity:         "/tmp/main.go",
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
