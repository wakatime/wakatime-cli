package regex_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/regex"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	tests := map[string]string{
		"standard":           `.*`,
		"negative lookahead": `^/var/(?!www/).*`,
		"positive lookahead": `^/var/(?=www/).*`,
	}

	for name, pattern := range tests {
		t.Run(name, func(t *testing.T) {
			r, err := regex.Compile(pattern)
			require.NoError(t, err)

			assert.Equal(t, pattern, r.String())
		})
	}
}

func TestMustCompile(t *testing.T) {
	tests := map[string]string{
		"standard":           `.*`,
		"negative lookahead": `^/var/(?!www/).*`,
		"positive lookahead": `^/var/(?=www/).*`,
	}

	for name, pattern := range tests {
		t.Run(name, func(t *testing.T) {
			r := regex.MustCompile(pattern)
			assert.Equal(t, pattern, r.String())
		})
	}
}

func TestMustCompilePanicsOnInvalidPattern(t *testing.T) {
	assert.Panics(t, func() {
		regex.MustCompile(`(?`)
	})
}

func TestRegexpWrapMethods(t *testing.T) {
	wrapped := regex.NewRegexpWrap(regexp.MustCompile(`^hello (.+)$`))

	assert.True(t, wrapped.MatchString(context.Background(), "hello world"))
	assert.False(t, wrapped.MatchString(context.Background(), "goodbye world"))
	assert.Equal(t, []string{"hello world", "world"}, wrapped.FindStringSubmatch(context.Background(), "hello world"))
	assert.Equal(t, `^hello (.+)$`, wrapped.String())
}

func TestRegexp2WrapMethods(t *testing.T) {
	wrapped, err := regex.Compile(`^hello (.+)(?=!)`)
	require.NoError(t, err)

	assert.True(t, wrapped.MatchString(context.Background(), "hello world!"))
	assert.False(t, wrapped.MatchString(context.Background(), "hello world"))
	assert.Equal(t, []string{"hello world", "world"}, wrapped.FindStringSubmatch(context.Background(), "hello world!"))
	assert.Nil(t, wrapped.FindStringSubmatch(context.Background(), "hello world"))
}
