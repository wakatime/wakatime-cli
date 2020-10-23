package heartbeat_test

import (
	"regexp"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithSanitization_ObfuscateFile(t *testing.T) {
	tests := map[string]struct {
		Heartbeat heartbeat.Heartbeat
		Expected  heartbeat.Heartbeat
	}{
		"file": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:         heartbeat.String("heartbeat"),
				Category:       heartbeat.CodingCategory,
				CursorPosition: heartbeat.Int(12),
				Dependencies:   []string{"dep1", "dep2"},
				Entity:         "/tmp/main.go",
				EntityType:     heartbeat.FileType,
				IsWrite:        heartbeat.Bool(true),
				Language:       heartbeat.LanguageGo,
				LineNumber:     heartbeat.Int(42),
				Lines:          heartbeat.Int(100),
				Project:        heartbeat.String("wakatime"),
				Time:           1585598060,
				UserAgent:      "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "HIDDEN.go",
				EntityType: heartbeat.FileType,
				IsWrite:    heartbeat.Bool(true),
				Language:   heartbeat.LanguageGo,
				Project:    heartbeat.String("wakatime"),
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
		"app": {
			Heartbeat: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "Slack",
				EntityType: heartbeat.AppType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "HIDDEN",
				EntityType: heartbeat.AppType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
		"domain": {
			Heartbeat: heartbeat.Heartbeat{
				Category:   heartbeat.BrowsingCategory,
				Entity:     "wakatime.com.br",
				EntityType: heartbeat.DomainType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.BrowsingCategory,
				Entity:     "HIDDEN",
				EntityType: heartbeat.DomainType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
	}

	opt := heartbeat.WithSanitization(heartbeat.SanitizeConfig{
		FilePatterns: []*regexp.Regexp{regexp.MustCompile(".*")},
	})

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
				assert.Equal(t, []heartbeat.Heartbeat{
					test.Expected,
				}, hh)

				return []heartbeat.Result{
					{
						Status: 201,
					},
				}, nil
			})

			result, err := handle([]heartbeat.Heartbeat{test.Heartbeat})
			require.NoError(t, err)

			assert.Equal(t, []heartbeat.Result{
				{
					Status: 201,
				},
			}, result)
		})
	}
}

func TestSanitize_ObfuscateFile(t *testing.T) {
	tests := map[string]struct {
		Heartbeat heartbeat.Heartbeat
		Expected  heartbeat.Heartbeat
	}{
		"file": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:         heartbeat.String("heartbeat"),
				Category:       heartbeat.CodingCategory,
				CursorPosition: heartbeat.Int(12),
				Dependencies:   []string{"dep1", "dep2"},
				Entity:         "/tmp/main.go",
				EntityType:     heartbeat.FileType,
				IsWrite:        heartbeat.Bool(true),
				Language:       heartbeat.LanguageGo,
				LineNumber:     heartbeat.Int(42),
				Lines:          heartbeat.Int(100),
				Project:        heartbeat.String("wakatime"),
				Time:           1585598060,
				UserAgent:      "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "HIDDEN.go",
				EntityType: heartbeat.FileType,
				IsWrite:    heartbeat.Bool(true),
				Language:   heartbeat.LanguageGo,
				Project:    heartbeat.String("wakatime"),
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
		"app": {
			Heartbeat: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "Slack",
				EntityType: heartbeat.AppType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "HIDDEN",
				EntityType: heartbeat.AppType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
		"domain": {
			Heartbeat: heartbeat.Heartbeat{
				Category:   heartbeat.BrowsingCategory,
				Entity:     "wakatime.com.br",
				EntityType: heartbeat.DomainType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.BrowsingCategory,
				Entity:     "HIDDEN",
				EntityType: heartbeat.DomainType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r := heartbeat.Sanitize(test.Heartbeat, heartbeat.SanitizeConfig{
				FilePatterns: []*regexp.Regexp{regexp.MustCompile(".*")},
			})

			assert.Equal(t, test.Expected, r)
		})
	}
}

func TestSanitize_Obfuscate_SkipBranchIfNotMatching(t *testing.T) {
	tests := map[string]struct {
		Heartbeat heartbeat.Heartbeat
		Expected  heartbeat.Heartbeat
	}{
		"file": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:         heartbeat.String("heartbeat"),
				Category:       heartbeat.CodingCategory,
				CursorPosition: heartbeat.Int(12),
				Dependencies:   []string{"dep1", "dep2"},
				Entity:         "/tmp/main.go",
				EntityType:     heartbeat.FileType,
				IsWrite:        heartbeat.Bool(true),
				Language:       heartbeat.LanguageGo,
				LineNumber:     heartbeat.Int(42),
				Lines:          heartbeat.Int(100),
				Project:        heartbeat.String("wakatime"),
				Time:           1585598060,
				UserAgent:      "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Branch:     heartbeat.String("heartbeat"),
				Category:   heartbeat.CodingCategory,
				Entity:     "HIDDEN.go",
				EntityType: heartbeat.FileType,
				IsWrite:    heartbeat.Bool(true),
				Language:   heartbeat.LanguageGo,
				Project:    heartbeat.String("wakatime"),
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
		"app": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:     heartbeat.String("heartbeat"),
				Category:   heartbeat.CodingCategory,
				Entity:     "Slack",
				EntityType: heartbeat.AppType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Branch:     heartbeat.String("heartbeat"),
				Category:   heartbeat.CodingCategory,
				Entity:     "HIDDEN",
				EntityType: heartbeat.AppType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
		"domain": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:     heartbeat.String("heartbeat"),
				Category:   heartbeat.BrowsingCategory,
				Entity:     "wakatime.com.br",
				EntityType: heartbeat.DomainType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Branch:     heartbeat.String("heartbeat"),
				Category:   heartbeat.BrowsingCategory,
				Entity:     "HIDDEN",
				EntityType: heartbeat.DomainType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r := heartbeat.Sanitize(test.Heartbeat, heartbeat.SanitizeConfig{
				FilePatterns:   []*regexp.Regexp{regexp.MustCompile(".*")},
				BranchPatterns: []*regexp.Regexp{regexp.MustCompile("not_matching")},
			})

			assert.Equal(t, test.Expected, r)
		})
	}
}

func TestSanitize_ObfuscateFile_NilFields(t *testing.T) {
	tests := map[string]struct {
		Heartbeat heartbeat.Heartbeat
		Expected  heartbeat.Heartbeat
	}{
		"file": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:         nil,
				Category:       heartbeat.CodingCategory,
				CursorPosition: heartbeat.Int(12),
				Dependencies:   []string{"dep1", "dep2"},
				Entity:         "/tmp/main.go",
				EntityType:     heartbeat.FileType,
				IsWrite:        heartbeat.Bool(true),
				Language:       heartbeat.LanguageGo,
				LineNumber:     heartbeat.Int(42),
				Lines:          heartbeat.Int(100),
				Project:        heartbeat.String("wakatime"),
				Time:           1585598060,
				UserAgent:      "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "HIDDEN.go",
				EntityType: heartbeat.FileType,
				IsWrite:    heartbeat.Bool(true),
				Language:   heartbeat.LanguageGo,
				Project:    heartbeat.String("wakatime"),
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
		"app": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:     nil,
				Category:   heartbeat.CodingCategory,
				Entity:     "Slack",
				EntityType: heartbeat.AppType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "HIDDEN",
				EntityType: heartbeat.AppType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
		"domain": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:     nil,
				Category:   heartbeat.BrowsingCategory,
				Entity:     "wakatime.com.br",
				EntityType: heartbeat.DomainType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.BrowsingCategory,
				Entity:     "HIDDEN",
				EntityType: heartbeat.DomainType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r := heartbeat.Sanitize(test.Heartbeat, heartbeat.SanitizeConfig{
				FilePatterns:   []*regexp.Regexp{regexp.MustCompile(".*")},
				BranchPatterns: []*regexp.Regexp{regexp.MustCompile(".*")},
			})

			assert.Equal(t, test.Expected, r)
		})
	}
}

func TestSanitize_ObfuscateProject(t *testing.T) {
	tests := map[string]struct {
		Heartbeat heartbeat.Heartbeat
		Expected  heartbeat.Heartbeat
	}{
		"file": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:         heartbeat.String("heartbeat"),
				Category:       heartbeat.CodingCategory,
				CursorPosition: heartbeat.Int(12),
				Dependencies:   []string{"dep1", "dep2"},
				Entity:         "/tmp/main.go",
				EntityType:     heartbeat.FileType,
				IsWrite:        heartbeat.Bool(true),
				Language:       heartbeat.LanguageGo,
				LineNumber:     heartbeat.Int(42),
				Lines:          heartbeat.Int(100),
				Project:        heartbeat.String("wakatime"),
				Time:           1585598060,
				UserAgent:      "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "/tmp/main.go",
				EntityType: heartbeat.FileType,
				IsWrite:    heartbeat.Bool(true),
				Language:   heartbeat.LanguageGo,
				Project:    heartbeat.String("wakatime"),
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
		"app": {
			Heartbeat: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "Slack",
				EntityType: heartbeat.AppType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "Slack",
				EntityType: heartbeat.AppType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
		"domain": {
			Heartbeat: heartbeat.Heartbeat{
				Category:   heartbeat.BrowsingCategory,
				Entity:     "wakatime.com.br",
				EntityType: heartbeat.DomainType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.BrowsingCategory,
				Entity:     "wakatime.com.br",
				EntityType: heartbeat.DomainType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r := heartbeat.Sanitize(test.Heartbeat, heartbeat.SanitizeConfig{
				ProjectPatterns: []*regexp.Regexp{regexp.MustCompile(".*")},
			})

			assert.Equal(t, test.Expected, r)
		})
	}
}

func TestSanitize_ObfuscateProject_SkipBranchIfNotMatching(t *testing.T) {
	tests := map[string]struct {
		Heartbeat heartbeat.Heartbeat
		Expected  heartbeat.Heartbeat
	}{
		"file": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:         heartbeat.String("heartbeat"),
				Category:       heartbeat.CodingCategory,
				CursorPosition: heartbeat.Int(12),
				Dependencies:   []string{"dep1", "dep2"},
				Entity:         "/tmp/main.go",
				EntityType:     heartbeat.FileType,
				IsWrite:        heartbeat.Bool(true),
				Language:       heartbeat.LanguageGo,
				LineNumber:     heartbeat.Int(42),
				Lines:          heartbeat.Int(100),
				Project:        heartbeat.String("wakatime"),
				Time:           1585598060,
				UserAgent:      "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Branch:     heartbeat.String("heartbeat"),
				Category:   heartbeat.CodingCategory,
				Entity:     "/tmp/main.go",
				EntityType: heartbeat.FileType,
				IsWrite:    heartbeat.Bool(true),
				Language:   heartbeat.LanguageGo,
				Project:    heartbeat.String("wakatime"),
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
		"app": {
			Heartbeat: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "Slack",
				EntityType: heartbeat.AppType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "Slack",
				EntityType: heartbeat.AppType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
		"domain": {
			Heartbeat: heartbeat.Heartbeat{
				Category:   heartbeat.BrowsingCategory,
				Entity:     "wakatime.com.br",
				EntityType: heartbeat.DomainType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.BrowsingCategory,
				Entity:     "wakatime.com.br",
				EntityType: heartbeat.DomainType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r := heartbeat.Sanitize(test.Heartbeat, heartbeat.SanitizeConfig{
				ProjectPatterns: []*regexp.Regexp{regexp.MustCompile(".*")},
				BranchPatterns:  []*regexp.Regexp{regexp.MustCompile("not_matching")},
			})

			assert.Equal(t, test.Expected, r)
		})
	}
}

func TestSanitize_ObfuscateProject_NilFields(t *testing.T) {
	tests := map[string]struct {
		Heartbeat heartbeat.Heartbeat
		Expected  heartbeat.Heartbeat
	}{
		"file": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:         nil,
				Category:       heartbeat.CodingCategory,
				CursorPosition: heartbeat.Int(12),
				Dependencies:   []string{"dep1", "dep2"},
				Entity:         "/tmp/main.go",
				EntityType:     heartbeat.FileType,
				IsWrite:        heartbeat.Bool(true),
				Language:       heartbeat.LanguageGo,
				LineNumber:     heartbeat.Int(42),
				Lines:          heartbeat.Int(100),
				Project:        heartbeat.String("wakatime"),
				Time:           1585598060,
				UserAgent:      "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "/tmp/main.go",
				EntityType: heartbeat.FileType,
				IsWrite:    heartbeat.Bool(true),
				Language:   heartbeat.LanguageGo,
				Project:    heartbeat.String("wakatime"),
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
		"app": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:     nil,
				Category:   heartbeat.CodingCategory,
				Entity:     "Slack",
				EntityType: heartbeat.AppType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "Slack",
				EntityType: heartbeat.AppType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
		"domain": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:     nil,
				Category:   heartbeat.BrowsingCategory,
				Entity:     "wakatime.com.br",
				EntityType: heartbeat.DomainType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.BrowsingCategory,
				Entity:     "wakatime.com.br",
				EntityType: heartbeat.DomainType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r := heartbeat.Sanitize(test.Heartbeat, heartbeat.SanitizeConfig{
				ProjectPatterns: []*regexp.Regexp{regexp.MustCompile(".*")},
				BranchPatterns:  []*regexp.Regexp{regexp.MustCompile(".*")},
			})

			assert.Equal(t, test.Expected, r)
		})
	}
}

func TestSanitize_ObfuscateBranch(t *testing.T) {
	tests := map[string]struct {
		Heartbeat heartbeat.Heartbeat
		Expected  heartbeat.Heartbeat
	}{
		"file": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:         heartbeat.String("wakatime"),
				Category:       heartbeat.CodingCategory,
				CursorPosition: heartbeat.Int(12),
				Dependencies:   []string{"dep1", "dep2"},
				Entity:         "/tmp/main.go",
				EntityType:     heartbeat.FileType,
				IsWrite:        heartbeat.Bool(true),
				Language:       heartbeat.LanguageGo,
				LineNumber:     heartbeat.Int(42),
				Lines:          heartbeat.Int(100),
				Project:        heartbeat.String("wakatime"),
				Time:           1585598060,
				UserAgent:      "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:       heartbeat.CodingCategory,
				CursorPosition: heartbeat.Int(12),
				Dependencies:   []string{"dep1", "dep2"},
				Entity:         "/tmp/main.go",
				EntityType:     heartbeat.FileType,
				IsWrite:        heartbeat.Bool(true),
				Language:       heartbeat.LanguageGo,
				LineNumber:     heartbeat.Int(42),
				Lines:          heartbeat.Int(100),
				Project:        heartbeat.String("wakatime"),
				Time:           1585598060,
				UserAgent:      "wakatime/13.0.7",
			},
		},
		"app": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:     heartbeat.String("wakatime"),
				Category:   heartbeat.CodingCategory,
				Entity:     "Slack",
				EntityType: heartbeat.AppType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "Slack",
				EntityType: heartbeat.AppType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
		"domain": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:     heartbeat.String("wakatime"),
				Category:   heartbeat.BrowsingCategory,
				Entity:     "wakatime.com.br",
				EntityType: heartbeat.DomainType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.BrowsingCategory,
				Entity:     "wakatime.com.br",
				EntityType: heartbeat.DomainType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r := heartbeat.Sanitize(test.Heartbeat, heartbeat.SanitizeConfig{
				BranchPatterns: []*regexp.Regexp{regexp.MustCompile(".*")},
			})

			assert.Equal(t, test.Expected, r)
		})
	}
}

func TestSanitize_ObfuscateBranch_NilFields(t *testing.T) {
	tests := map[string]struct {
		Heartbeat heartbeat.Heartbeat
		Expected  heartbeat.Heartbeat
	}{
		"file": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:         nil,
				Category:       heartbeat.CodingCategory,
				CursorPosition: heartbeat.Int(12),
				Dependencies:   []string{"dep1", "dep2"},
				Entity:         "/tmp/main.go",
				EntityType:     heartbeat.FileType,
				IsWrite:        heartbeat.Bool(true),
				Language:       heartbeat.LanguageGo,
				LineNumber:     heartbeat.Int(42),
				Lines:          heartbeat.Int(100),
				Project:        nil,
				Time:           1585598060,
				UserAgent:      "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:       heartbeat.CodingCategory,
				CursorPosition: heartbeat.Int(12),
				Dependencies:   []string{"dep1", "dep2"},
				Entity:         "/tmp/main.go",
				EntityType:     heartbeat.FileType,
				IsWrite:        heartbeat.Bool(true),
				Language:       heartbeat.LanguageGo,
				LineNumber:     heartbeat.Int(42),
				Lines:          heartbeat.Int(100),
				Time:           1585598060,
				UserAgent:      "wakatime/13.0.7",
			},
		},
		"app": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:     nil,
				Category:   heartbeat.CodingCategory,
				Entity:     "Slack",
				EntityType: heartbeat.AppType,
				Project:    nil,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "Slack",
				EntityType: heartbeat.AppType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
		"domain": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:     nil,
				Category:   heartbeat.BrowsingCategory,
				Entity:     "wakatime.com.br",
				EntityType: heartbeat.DomainType,
				Project:    nil,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.BrowsingCategory,
				Entity:     "wakatime.com.br",
				EntityType: heartbeat.DomainType,
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r := heartbeat.Sanitize(test.Heartbeat, heartbeat.SanitizeConfig{
				BranchPatterns: []*regexp.Regexp{regexp.MustCompile(".*")},
			})

			assert.Equal(t, test.Expected, r)
		})
	}
}

func TestSanitize_EmptyConfigDoNothing(t *testing.T) {
	r := heartbeat.Sanitize(testHeartbeat(), heartbeat.SanitizeConfig{})

	assert.Equal(t, heartbeat.Heartbeat{
		Branch:         heartbeat.String("heartbeat"),
		Category:       heartbeat.CodingCategory,
		CursorPosition: heartbeat.Int(12),
		Dependencies:   []string{"dep1", "dep2"},
		Entity:         "/tmp/main.go",
		EntityType:     heartbeat.FileType,
		IsWrite:        heartbeat.Bool(true),
		Language:       heartbeat.LanguageGo,
		LineNumber:     heartbeat.Int(42),
		Lines:          heartbeat.Int(100),
		Project:        heartbeat.String("wakatime"),
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}, r)
}

func TestSanitize_EmptyConfigDoNothing_EmptyDependencies(t *testing.T) {
	h := testHeartbeat()
	h.Dependencies = []string{}

	r := heartbeat.Sanitize(h, heartbeat.SanitizeConfig{})

	assert.Equal(t, heartbeat.Heartbeat{
		Branch:         heartbeat.String("heartbeat"),
		Category:       heartbeat.CodingCategory,
		CursorPosition: heartbeat.Int(12),
		Entity:         "/tmp/main.go",
		EntityType:     heartbeat.FileType,
		IsWrite:        heartbeat.Bool(true),
		Language:       heartbeat.LanguageGo,
		LineNumber:     heartbeat.Int(42),
		Lines:          heartbeat.Int(100),
		Project:        heartbeat.String("wakatime"),
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}, r)
}

func TestShouldSanitize(t *testing.T) {
	tests := map[string]struct {
		Subject  string
		Regex    []*regexp.Regexp
		Expected bool
	}{
		"match_single": {
			Subject: "fix.123",
			Regex: []*regexp.Regexp{
				regexp.MustCompile("fix.*"),
			},
			Expected: true,
		},
		"match_multiple": {
			Subject: "fix.456",
			Regex: []*regexp.Regexp{
				regexp.MustCompile("bar.*"),
				regexp.MustCompile("fix.*"),
			},
			Expected: true,
		},
		"not_match": {
			Subject: "foo",
			Regex: []*regexp.Regexp{
				regexp.MustCompile("bar.*"),
				regexp.MustCompile("fix.*"),
			},
			Expected: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			souldSanitize := heartbeat.ShouldSanitize(test.Subject, test.Regex)

			assert.Equal(t, test.Expected, souldSanitize)
		})
	}
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
		Language:       heartbeat.LanguageGo,
		LineNumber:     heartbeat.Int(42),
		Lines:          heartbeat.Int(100),
		Project:        heartbeat.String("wakatime"),
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}
}
