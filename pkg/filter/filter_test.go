package filter_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/filter"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithFiltering(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

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
				Language:       heartbeat.String("golang"),
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

func TestFilter(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

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
	tmpFile, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	h := testHeartbeat()
	h.Entity = tmpFile.Name()

	err = filter.Filter(h, filter.Config{
		Exclude: []*regexp.Regexp{
			regexp.MustCompile(".*main.go$"),
		},
		Include: []*regexp.Regexp{
			regexp.MustCompile(".*/tmp/.*"),
		},
	})
	require.NoError(t, err)
}

func TestFilter_ErrMatchesExcludePattern(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	h := testHeartbeat()
	h.Entity = tmpFile.Name()

	err = filter.Filter(h, filter.Config{
		Exclude: []*regexp.Regexp{
			regexp.MustCompile("^" + tmpFile.Name() + "$"),
		},
	})

	var errv filter.Err

	assert.True(t, errors.As(err, &errv))
	assert.Equal(t, filter.Err(fmt.Sprintf("skipping because matching exclude pattern \"^%s$\"", tmpFile.Name())), errv)
}

func TestFilter_ErrUnknownLanguage(t *testing.T) {
	tests := map[string]*string{
		"nil":          nil,
		"empty string": heartbeat.String(""),
	}

	for name, languageValue := range tests {
		t.Run(name, func(t *testing.T) {
			tmpFile, err := ioutil.TempFile(os.TempDir(), "")
			require.NoError(t, err)

			defer os.Remove(tmpFile.Name())

			h := testHeartbeat()
			h.Entity = tmpFile.Name()
			h.Language = languageValue

			err = filter.Filter(h, filter.Config{
				ExcludeUnknownProject: true,
			})
			var errv filter.Err
			assert.True(t, errors.As(err, &errv))

			assert.Equal(t, filter.Err("skipping because of unknown language"), errv)
		})
	}
}

func TestFilter_ErrUnknownProject(t *testing.T) {
	tests := map[string]*string{
		"nil":          nil,
		"empty string": heartbeat.String(""),
	}

	for name, projectValue := range tests {
		t.Run(name, func(t *testing.T) {
			tmpFile, err := ioutil.TempFile(os.TempDir(), "")
			require.NoError(t, err)

			defer os.Remove(tmpFile.Name())

			h := testHeartbeat()
			h.Entity = tmpFile.Name()
			h.Project = projectValue

			err = filter.Filter(h, filter.Config{
				ExcludeUnknownProject: true,
			})
			var errv filter.Err
			assert.True(t, errors.As(err, &errv))

			assert.Equal(t, filter.Err("skipping because of unknown project"), errv)
		})
	}
}

func TestFilter_ErrNonExistingFile(t *testing.T) {
	h := testHeartbeat()

	err := filter.Filter(h, filter.Config{})

	var errv filter.Err

	assert.True(t, errors.As(err, &errv))
	assert.Equal(t, filter.Err("skipping because of non-existing file \"/tmp/main.go\""), errv)
}

func TestFilter_ExistingProjectFile(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	projectFile, err := os.Create(path.Join(os.TempDir(), ".wakatime-project"))
	require.NoError(t, err)

	defer os.Remove(projectFile.Name())

	h := testHeartbeat()
	h.Entity = tmpFile.Name()

	err = filter.Filter(h, filter.Config{
		IncludeOnlyWithProjectFile: true,
	})
	require.NoError(t, err)
}

func TestFilter_ErrNonExistingProjectFile(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	h := testHeartbeat()
	h.Entity = tmpFile.Name()

	err = filter.Filter(h, filter.Config{
		IncludeOnlyWithProjectFile: true,
	})

	var errv filter.Err

	assert.True(t, errors.As(err, &errv))
	assert.Equal(t, filter.Err("skipping because of non-existing project file in parent path"), errv)
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
		Language:       heartbeat.String("golang"),
		LineNumber:     heartbeat.Int(42),
		Lines:          heartbeat.Int(100),
		Project:        heartbeat.String("wakatime"),
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}
}
