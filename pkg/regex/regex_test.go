package regex_test

import (
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
