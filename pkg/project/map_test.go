package project_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMap_Detect(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	m := project.Map{
		Filepath: "testdata/entity.any",
		Patterns: []project.MapPattern{
			{
				Name: "my-project-1",
				Regex: func() *regexp.Regexp {
					r, err := regexp.Compile(filepath.Join(wd, "testdata"))
					require.NoError(t, err)
					return r
				}(),
			},
		},
	}

	result, detected, err := m.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Equal(t, "my-project-1", result.Project)
}

func TestMap_Detect_RegexReplace(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	m := project.Map{
		Filepath: "testdata/entity.any",
		Patterns: []project.MapPattern{
			{
				Name: "my-project-1",
				Regex: func() *regexp.Regexp {
					r, err := regexp.Compile(filepath.Join(wd, "path/to/otherfolder"))
					require.NoError(t, err)
					return r
				}(),
			},
			{
				Name: "my-project-2-{0}",
				Regex: func() *regexp.Regexp {
					r, err := regexp.Compile(filepath.Join(wd, "test(\\w+)"))
					require.NoError(t, err)
					return r
				}(),
			},
		},
	}

	result, detected, err := m.Detect()
	require.NoError(t, err)

	assert.True(t, detected)
	assert.Equal(t, "my-project-2-data", result.Project)
}

func TestMap_Detect_NoMatch(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	m := project.Map{
		Filepath: "testdata/entity.any",
		Patterns: []project.MapPattern{
			{
				Name: "my_project_1",
				Regex: func() *regexp.Regexp {
					r, err := regexp.Compile(filepath.Join(wd, "path/to/otherfolder"))
					require.NoError(t, err)
					return r
				}(),
			},
			{
				Name: "my_project_2",
				Regex: func() *regexp.Regexp {
					r, err := regexp.Compile(filepath.Join(wd, "path/to/temp"))
					require.NoError(t, err)
					return r
				}(),
			},
		},
	}

	result, detected, err := m.Detect()
	require.NoError(t, err)

	assert.False(t, detected)
	assert.Equal(t, "", result.Project)
}

func TestMap_Detect_ZeroPatterns(t *testing.T) {
	m := project.Map{
		Patterns: []project.MapPattern{},
	}

	_, detected, err := m.Detect()
	require.NoError(t, err)

	assert.False(t, detected)
}

func TestMap_String(t *testing.T) {
	m := project.Map{}

	assert.Equal(t, "project-map-detector", m.String())
}
