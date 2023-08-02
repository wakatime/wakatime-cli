package apikey_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/apikey"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/gandarez/go-realpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithReplacing(t *testing.T) {
	first := heartbeat.Heartbeat{
		Entity: "/tmp/main.go",
	}

	second := heartbeat.Heartbeat{
		Entity: "/workdir/main.go",
	}

	config := apikey.Config{
		DefaultAPIKey: "00000000-0000-4000-8000-000000000000",
		MapPatterns: []apikey.MapPattern{
			{
				APIKey: "00000000-0000-4000-8000-000000000001",
				Regex:  regexp.MustCompile(`.workdir.`),
			},
		},
	}

	opt := apikey.WithReplacing(config)
	h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				APIKey: "00000000-0000-4000-8000-000000000000",
				Entity: "/tmp/main.go",
			},
			{
				APIKey: "00000000-0000-4000-8000-000000000001",
				Entity: "/workdir/main.go",
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

func TestApiKey_MatchPattern(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	rp, err := realpath.Realpath(filepath.Join("testdata", "entity.any"))
	require.NoError(t, err)

	patterns := []apikey.MapPattern{
		{
			APIKey: "00000000-0000-4000-8000-000000000000",
			Regex:  regexp.MustCompile(formatRegex(filepath.Join(wd, "path", "to", "otherfolder"))),
		},
		{
			APIKey: "00000000-0000-4000-8000-000000000001",
			Regex:  regexp.MustCompile(formatRegex(filepath.Join(wd, `test([a-zA-Z]+)`))),
		},
	}

	result, ok := apikey.MatchPattern(rp, patterns)

	assert.True(t, ok)
	assert.Equal(t, "00000000-0000-4000-8000-000000000001", result)
}

func TestApiKey_MatchPattern_NoMatch(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	rp, err := realpath.Realpath(filepath.Join("testdata", "entity.any"))
	require.NoError(t, err)

	patterns := []apikey.MapPattern{
		{
			APIKey: "00000000-0000-4000-8000-000000000000",
			Regex:  regexp.MustCompile(formatRegex(filepath.Join(wd, "path", "to", "otherfolder"))),
		},
		{
			APIKey: "00000000-0000-4000-8000-000000000001",
			Regex:  regexp.MustCompile(formatRegex(filepath.Join(wd, "path", "to", "temp"))),
		},
	}

	_, ok := apikey.MatchPattern(rp, patterns)

	assert.False(t, ok)
}

func TestApiKey_MatchPattern_ZeroPatterns(t *testing.T) {
	_, ok := apikey.MatchPattern("", []apikey.MapPattern{})

	assert.False(t, ok)
}

func formatRegex(fp string) string {
	if runtime.GOOS != "windows" {
		return fp
	}

	return strings.ReplaceAll(fp, `\`, `\\`)
}
