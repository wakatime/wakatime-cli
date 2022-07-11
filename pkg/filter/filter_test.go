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
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	first := testHeartbeat()
	first.Entity = tmpFile.Name()

	second := testHeartbeat()
	second.Time++

	opt := filter.WithFiltering(filter.Config{})
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

func TestWithLengthValidator(t *testing.T) {
	opt := filter.WithLengthValidator()
	h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{}, errors.New("this will should never be called")
	})

	result, err := h([]heartbeat.Heartbeat{})
	require.NoError(t, err)

	assert.Equal(t, result, []heartbeat.Result{})
}

func TestFilter(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	h := testHeartbeat()
	h.Entity = tmpFile.Name()

	err = filter.Filter(h, filter.Config{})
	require.NoError(t, err)
}

func TestFilter_NonFileTypeEmptyEntity(t *testing.T) {
	h := testHeartbeat()
	h.Entity = ""
	h.EntityType = heartbeat.AppType

	err := filter.Filter(h, filter.Config{})
	require.NoError(t, err)
}

func TestFilter_IsUnsavedEntity(t *testing.T) {
	h := testHeartbeat()
	h.Entity = "nonexisting"
	h.EntityType = heartbeat.FileType
	h.IsUnsavedEntity = true

	err := filter.Filter(h, filter.Config{})
	require.NoError(t, err)
}

func TestFilter_IncludeMatchOverwritesExcludeMatch(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

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
	tmpFile, err := os.CreateTemp(t.TempDir(), "exclude-this-file")
	require.NoError(t, err)

	defer tmpFile.Close()

	h := testHeartbeat()
	h.Entity = tmpFile.Name()

	err = filter.Filter(h, filter.Config{
		Exclude: []regex.Regex{
			regex.MustCompile("^.*exclude-this-file.*$"),
		},
	})

	assert.EqualError(t, err, "filter by pattern: skipping because matches exclude pattern \"^.*exclude-this-file.*$\"")
}

func TestFilter_ErrNonExistingFile(t *testing.T) {
	h := testHeartbeat()

	err := filter.Filter(h, filter.Config{})

	assert.EqualError(t, err, "filter file: skipping because of non-existing file \"/tmp/main.go\"")
}

func TestFilter_ExistingProjectFile(t *testing.T) {
	tmpDir := t.TempDir()

	tmpFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer tmpFile.Close()

	tmpFile2, err := os.Create(filepath.Join(tmpDir, ".wakatime-project"))
	require.NoError(t, err)

	defer tmpFile2.Close()

	h := testHeartbeat()
	h.Entity = tmpFile.Name()

	err = filter.Filter(h, filter.Config{
		IncludeOnlyWithProjectFile: true,
	})
	require.NoError(t, err)
}

func TestFilter_RemoteFileSkipsFiltering(t *testing.T) {
	h := testHeartbeat()
	h.LocalFile = h.Entity
	h.Entity = "ssh://wakatime:1234@192.168.1.1/path/to/remote/main.go"

	err := filter.Filter(h, filter.Config{})
	require.NoError(t, err)
}

func TestFilter_ErrNonExistingProjectFile(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	h := testHeartbeat()
	h.Entity = tmpFile.Name()

	err = filter.Filter(h, filter.Config{
		IncludeOnlyWithProjectFile: true,
	})

	assert.EqualError(t, err, "filter file: skipping because missing .wakatime-project file in parent path")
}

func testHeartbeat() heartbeat.Heartbeat {
	return heartbeat.Heartbeat{
		Branch:          heartbeat.PointerTo("heartbeat"),
		Category:        heartbeat.CodingCategory,
		CursorPosition:  heartbeat.PointerTo(12),
		Dependencies:    []string{"dep1", "dep2"},
		Entity:          "/tmp/main.go",
		EntityType:      heartbeat.FileType,
		IsWrite:         heartbeat.PointerTo(true),
		Language:        heartbeat.PointerTo("Go"),
		LineNumber:      heartbeat.PointerTo(42),
		Lines:           heartbeat.PointerTo(100),
		Project:         heartbeat.PointerTo("wakatime"),
		Time:            1585598060,
		UserAgent:       "wakatime/13.0.7",
		IsUnsavedEntity: false,
	}
}
