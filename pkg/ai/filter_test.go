package ai_test

import (
	"context"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/regex"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAIFilter_NonAICategoryNeverSkipped(t *testing.T) {
	h := heartbeat.Heartbeat{
		Category: heartbeat.CodingCategory.String(),
		Entity:   "/tmp/main.py",
		Project:  heartbeat.PointerTo("nautilus"),
	}

	err := ai.Filter(t.Context(), h, ai.FilterConfig{
		SyncDisabled:    true,
		ExcludeProjects: []regex.Regex{regex.MustCompile("(?i)^nautilus$")},
	})

	assert.NoError(t, err)
}

func TestAIFilter_SyncDisabledSkipsAICoding(t *testing.T) {
	h := heartbeat.Heartbeat{
		Category: heartbeat.AICodingCategory.String(),
		Entity:   "OpenCode ses_123",
		Project:  heartbeat.PointerTo("some-project"),
	}

	err := ai.Filter(t.Context(), h, ai.FilterConfig{
		SyncDisabled: true,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ai tracking is disabled")
}

func TestAIFilter_ExcludeProjectsSkipsMatchingAICoding(t *testing.T) {
	tests := map[string]struct {
		Project  *string
		Patterns []string
		Skipped  bool
	}{
		"exact match": {
			Project:  heartbeat.PointerTo("nautilus"),
			Patterns: []string{"(?i)^nautilus$"},
			Skipped:  true,
		},
		"case insensitive": {
			Project:  heartbeat.PointerTo("Nautilus"),
			Patterns: []string{"(?i)^nautilus$"},
			Skipped:  true,
		},
		"no match keeps heartbeat": {
			Project:  heartbeat.PointerTo("other-project"),
			Patterns: []string{"(?i)^nautilus$"},
			Skipped:  false,
		},
		"nil project keeps heartbeat": {
			Project:  nil,
			Patterns: []string{"(?i)^nautilus$"},
			Skipped:  false,
		},
		"empty project keeps heartbeat": {
			Project:  heartbeat.PointerTo(""),
			Patterns: []string{"(?i)^nautilus$"},
			Skipped:  false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			h := heartbeat.Heartbeat{
				Category: heartbeat.AICodingCategory.String(),
				Entity:   "OpenCode ses_123",
				Project:  test.Project,
			}

			var patterns []regex.Regex

			for _, p := range test.Patterns {
				patterns = append(patterns, regex.MustCompile(p))
			}

			err := ai.Filter(t.Context(), h, ai.FilterConfig{
				ExcludeProjects: patterns,
			})

			if test.Skipped {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "ai exclude projects")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAIWithFiltering_DropsOnlyAICoding(t *testing.T) {
	aiHeartbeat := heartbeat.Heartbeat{
		Category: heartbeat.AICodingCategory.String(),
		Entity:   "OpenCode ses_123",
		Project:  heartbeat.PointerTo("nautilus"),
	}
	codeHeartbeat := heartbeat.Heartbeat{
		Category:   heartbeat.CodingCategory.String(),
		Entity:     "/tmp/main.py",
		EntityType: heartbeat.FileType,
		Project:    heartbeat.PointerTo("nautilus"),
	}

	opt := ai.WithFiltering(ai.FilterConfig{
		ExcludeProjects: []regex.Regex{regex.MustCompile("(?i)^nautilus$")},
	})
	h := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, []heartbeat.Heartbeat{codeHeartbeat}, hh)

		return []heartbeat.Result{{Status: 201}}, nil
	})

	result, err := h(t.Context(), []heartbeat.Heartbeat{aiHeartbeat, codeHeartbeat})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{{Status: 201}}, result)
}
