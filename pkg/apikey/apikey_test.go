package apikey_test

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/apikey"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/regex"

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
		DefaultAPIURL: "https://api.wakatime.com/api/v1",
		MapPatterns: []apikey.MapPattern{
			{
				APIKey: "00000000-0000-4000-8000-000000000001",
				Regex:  regex.NewRegexpWrap(regexp.MustCompile(`.workdir.`)),
			},
		},
	}

	opt := apikey.WithReplacing(config)
	h := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				APIKey: "00000000-0000-4000-8000-000000000000",
				APIURL: "https://api.wakatime.com/api/v1",
				Entity: "/tmp/main.go",
			},
			{
				APIKey: "00000000-0000-4000-8000-000000000001",
				APIURL: "https://api.wakatime.com/api/v1",
				Entity: "/workdir/main.go",
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := h(t.Context(), []heartbeat.Heartbeat{first, second})
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
			Regex:  regex.NewRegexpWrap(regexp.MustCompile(formatRegex(filepath.Join(wd, "path", "to", "otherfolder")))),
		},
		{
			APIKey: "00000000-0000-4000-8000-000000000001",
			Regex:  regex.NewRegexpWrap(regexp.MustCompile(formatRegex(filepath.Join(wd, `test([a-zA-Z]+)`)))),
		},
	}

	result, ok := apikey.MatchPattern(t.Context(), rp, patterns)

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
			Regex:  regex.NewRegexpWrap(regexp.MustCompile(formatRegex(filepath.Join(wd, "path", "to", "otherfolder")))),
		},
		{
			APIKey: "00000000-0000-4000-8000-000000000001",
			Regex:  regex.NewRegexpWrap(regexp.MustCompile(formatRegex(filepath.Join(wd, "path", "to", "temp")))),
		},
	}

	_, ok := apikey.MatchPattern(t.Context(), rp, patterns)

	assert.False(t, ok)
}

func TestApiKey_MatchPattern_ZeroPatterns(t *testing.T) {
	_, ok := apikey.MatchPattern(t.Context(), "", []apikey.MapPattern{})

	assert.False(t, ok)
}

func TestWithReplacing_URLPatterns(t *testing.T) {
	first := heartbeat.Heartbeat{
		Entity: "/tmp/main.go",
	}

	second := heartbeat.Heartbeat{
		Entity: "/workdir/main.go",
	}

	third := heartbeat.Heartbeat{
		Entity: "/custom/main.go",
	}

	config := apikey.Config{
		DefaultAPIKey: "00000000-0000-4000-8000-000000000000",
		DefaultAPIURL: "https://api.wakatime.com/api/v1",
		MapPatterns: []apikey.MapPattern{
			{
				APIKey: "00000000-0000-4000-8000-000000000001",
				Regex:  regex.NewRegexpWrap(regexp.MustCompile(`.workdir.`)),
			},
		},
		URLPatterns: []apikey.URLPattern{
			{
				APIURL: "https://custom.example.com/api/v1",
				APIKey: "00000000-0000-4000-8000-000000000002",
				Regex:  regex.NewRegexpWrap(regexp.MustCompile(`.custom.`)),
			},
		},
	}

	opt := apikey.WithReplacing(config)
	h := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.ElementsMatch(t, []heartbeat.Heartbeat{
			{
				APIKey: "00000000-0000-4000-8000-000000000000",
				APIURL: "https://api.wakatime.com/api/v1",
				Entity: "/tmp/main.go",
			},
			{
				APIKey: "00000000-0000-4000-8000-000000000001",
				APIURL: "https://api.wakatime.com/api/v1",
				Entity: "/workdir/main.go",
			},
			{
				APIKey: "00000000-0000-4000-8000-000000000000",
				APIURL: "https://api.wakatime.com/api/v1",
				Entity: "/custom/main.go",
			},
			{
				APIKey: "00000000-0000-4000-8000-000000000002",
				APIURL: "https://custom.example.com/api/v1",
				Entity: "/custom/main.go",
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := h(t.Context(), []heartbeat.Heartbeat{first, second, third})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestWithReplacing_Deduplication(t *testing.T) {
	// Test that when the same api_url+api_key appears in both default and api_urls,
	// it only sends once (deduplication by hash)
	h := heartbeat.Heartbeat{
		Entity: "/work/main.go",
	}

	config := apikey.Config{
		DefaultAPIKey: "00000000-0000-4000-8000-000000000000",
		DefaultAPIURL: "https://api.wakatime.com/api/v1",
		URLPatterns: []apikey.URLPattern{
			{
				// Same URL and key as default - should be deduplicated
				APIURL: "https://api.wakatime.com/api/v1",
				APIKey: "00000000-0000-4000-8000-000000000000",
				Regex:  regex.NewRegexpWrap(regexp.MustCompile(`.work.`)),
			},
		},
	}

	opt := apikey.WithReplacing(config)
	handler := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		// Should only have 1 heartbeat due to deduplication
		assert.Len(t, hh, 1)
		assert.Equal(t, "00000000-0000-4000-8000-000000000000", hh[0].APIKey)
		assert.Equal(t, "https://api.wakatime.com/api/v1", hh[0].APIURL)

		return []heartbeat.Result{{Status: 201}}, nil
	})

	_, err := handler(t.Context(), []heartbeat.Heartbeat{h})
	require.NoError(t, err)
}

func TestWithReplacing_MultipleURLPatternsMatch(t *testing.T) {
	h := heartbeat.Heartbeat{
		Entity: "/work/project/main.go",
	}

	config := apikey.Config{
		DefaultAPIKey: "00000000-0000-4000-8000-000000000000",
		DefaultAPIURL: "https://api.wakatime.com/api/v1",
		URLPatterns: []apikey.URLPattern{
			{
				APIURL: "https://work.example.com/api/v1",
				APIKey: "00000000-0000-4000-8000-000000000001",
				Regex:  regex.NewRegexpWrap(regexp.MustCompile(`.work.`)),
			},
			{
				APIURL: "https://project.example.com/api/v1",
				APIKey: "00000000-0000-4000-8000-000000000002",
				Regex:  regex.NewRegexpWrap(regexp.MustCompile(`.project.`)),
			},
		},
	}

	opt := apikey.WithReplacing(config)
	handler := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		// Should have 3 heartbeats:
		// 1. Default API
		// 2. work.example.com (matches .work.)
		// 3. project.example.com (matches .project.)
		assert.Len(t, hh, 3)

		// Verify all expected destinations are present
		destinations := make(map[string]bool)
		for _, h := range hh {
			destinations[h.APIURL+"|"+h.APIKey] = true
		}

		assert.True(t, destinations["https://api.wakatime.com/api/v1|00000000-0000-4000-8000-000000000000"])
		assert.True(t, destinations["https://work.example.com/api/v1|00000000-0000-4000-8000-000000000001"])
		assert.True(t, destinations["https://project.example.com/api/v1|00000000-0000-4000-8000-000000000002"])

		return []heartbeat.Result{{Status: 201}}, nil
	})

	_, err := handler(t.Context(), []heartbeat.Heartbeat{h})
	require.NoError(t, err)
}

func formatRegex(fp string) string {
	if runtime.GOOS != "windows" {
		return fp
	}

	return strings.ReplaceAll(fp, `\`, `\\`)
}
