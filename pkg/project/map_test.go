package project_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/gandarez/go-realpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMap_Detect(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	rp, err := realpath.Realpath(filepath.Join("testdata", "entity.any"))
	require.NoError(t, err)

	m := project.Map{
		Filepath: rp,
		Patterns: []project.MapPattern{
			{
				Name:  "my-project-1",
				Regex: regexp.MustCompile(formatRegex(filepath.Join(wd, "testdata"))),
			},
		},
	}

	result, detected, err := m.Detect()
	require.NoError(t, err)

	assert.True(t, detected)

	assert.Contains(t, result.Folder, "testdata")
	assert.Equal(t, "my-project-1", result.Project)
}

func TestMap_Detect_RegexReplace(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	rp, err := realpath.Realpath(filepath.Join("testdata", "entity.any"))
	require.NoError(t, err)

	m := project.Map{
		Filepath: rp,
		Patterns: []project.MapPattern{
			{
				Name:  "my-project-1",
				Regex: regexp.MustCompile(formatRegex(filepath.Join(wd, "path", "to", "otherfolder"))),
			},
			{
				Name:  "my-project-2-{0}",
				Regex: regexp.MustCompile(formatRegex(filepath.Join(wd, `test([a-zA-Z]+)`))),
			},
		},
	}

	result, detected, err := m.Detect()
	require.NoError(t, err)

	assert.True(t, detected)

	assert.Contains(t, result.Folder, "testdata")
	assert.Equal(t, "my-project-2-data", result.Project)
}

func TestMap_Detect_NoMatch(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	m := project.Map{
		Filepath: "testdata/entity.any",
		Patterns: []project.MapPattern{
			{
				Name:  "my_project_1",
				Regex: regexp.MustCompile(formatRegex(filepath.Join(wd, "path", "to", "otherfolder"))),
			},
			{
				Name:  "my_project_2",
				Regex: regexp.MustCompile(formatRegex(filepath.Join(wd, "path", "to", "temp"))),
			},
		},
	}

	result, detected, err := m.Detect()
	require.NoError(t, err)

	assert.False(t, detected)

	assert.Empty(t, result.Folder)
	assert.Empty(t, result.Project)
}

func TestMap_Detect_ZeroPatterns(t *testing.T) {
	m := project.Map{
		Patterns: []project.MapPattern{},
	}

	_, detected, err := m.Detect()
	require.NoError(t, err)

	assert.False(t, detected)
}

func TestMap_ID(t *testing.T) {
	m := project.Map{}

	assert.Equal(t, project.MapDetector, m.ID())
}
