package deps

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type panicParser struct{}

func (panicParser) Parse(context.Context, string) ([]string, error) {
	panic("lexer failed")
}

type dependenciesParser struct{}

func (dependenciesParser) Parse(context.Context, string) ([]string, error) {
	return []string{"requests", "requests"}, nil
}

type errorParser struct{}

func (errorParser) Parse(context.Context, string) ([]string, error) {
	return nil, errors.New("failed")
}

func TestParseDependencies_Panic(t *testing.T) {
	dependencies, err := parseDependencies(t.Context(), panicParser{}, "testdata/python.py")

	require.Error(t, err)
	assert.Nil(t, dependencies)
	assert.Contains(t, err.Error(), "panicked: lexer failed")
}

func TestParseDependencies_Success(t *testing.T) {
	dependencies, err := parseDependencies(t.Context(), dependenciesParser{}, "testdata/python.py")

	require.NoError(t, err)
	assert.Equal(t, []string{"requests", "requests"}, dependencies)
}

func TestParseDependencies_Error(t *testing.T) {
	dependencies, err := parseDependencies(t.Context(), errorParser{}, "testdata/python.py")

	require.EqualError(t, err, "failed")
	assert.Nil(t, dependencies)
}
