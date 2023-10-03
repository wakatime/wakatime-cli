package git_test

import (
	"errors"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/git"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCountLinesChanged(t *testing.T) {
	gc := git.New("testdata/main.go")
	gc.GitCmd = func(args ...string) (string, error) {
		assert.Equal(t, args, []string{"-C", "testdata", "diff", "--numstat", "testdata/main.go"})

		return "4       1       testdata/main.go", nil
	}

	added, removed, err := gc.CountLinesChanged()
	require.NoError(t, err)

	require.NotNil(t, added)
	require.NotNil(t, removed)

	assert.Equal(t, 4, *added)
	assert.Equal(t, 1, *removed)
}

func TestCountLinesChanged_Err(t *testing.T) {
	gc := git.New("testdata/main.go")
	gc.GitCmd = func(args ...string) (string, error) {
		assert.Equal(t, args, []string{"-C", "testdata", "diff", "--numstat", "testdata/main.go"})

		return "", errors.New("some error")
	}

	added, removed, err := gc.CountLinesChanged()
	assert.EqualError(t, err, "failed to count lines changed: some error")

	assert.Nil(t, added)
	assert.Nil(t, removed)
}

func TestCountLinesChanged_Staged(t *testing.T) {
	gc := git.New("testdata/main.go")

	var numCalls int

	gc.GitCmd = func(args ...string) (string, error) {
		numCalls++

		switch numCalls {
		case 1:
			assert.Equal(t, args, []string{"-C", "testdata", "diff", "--numstat", "testdata/main.go"})
		case 2:
			assert.Equal(t, args, []string{"-C", "testdata", "diff", "--numstat", "--cached", "testdata/main.go"})

			return "4       1       testdata/main.go", nil
		}

		return "", nil
	}

	added, removed, err := gc.CountLinesChanged()
	assert.NoError(t, err)

	require.NotNil(t, added)
	require.NotNil(t, removed)

	assert.Equal(t, 4, *added)
	assert.Equal(t, 1, *removed)
}

func TestCountLinesChanged_MissingFile(t *testing.T) {
	gc := git.New("/tmp/missing-file")

	added, removed, err := gc.CountLinesChanged()
	assert.NoError(t, err)

	assert.Nil(t, added)
	assert.Nil(t, removed)
}

func TestCountLinesChanged_NoOutput(t *testing.T) {
	gc := git.New("testdata/main.go")

	var numCalls int

	gc.GitCmd = func(args ...string) (string, error) {
		numCalls++

		switch numCalls {
		case 1:
			assert.Equal(t, args, []string{"-C", "testdata", "diff", "--numstat", "testdata/main.go"})
		case 2:
			assert.Equal(t, args, []string{"-C", "testdata", "diff", "--numstat", "--cached", "testdata/main.go"})
		}

		return "", nil
	}

	added, removed, err := gc.CountLinesChanged()
	assert.NoError(t, err)

	assert.Nil(t, added)
	assert.Nil(t, removed)
}

func TestCountLinesChanged_MalformedOutput(t *testing.T) {
	gc := git.New("testdata/main.go")
	gc.GitCmd = func(args ...string) (string, error) {
		assert.Equal(t, args, []string{"-C", "testdata", "diff", "--numstat", "testdata/main.go"})

		return "malformed output", nil
	}

	added, removed, err := gc.CountLinesChanged()
	assert.NoError(t, err)

	assert.Nil(t, added)
	assert.Nil(t, removed)
}
