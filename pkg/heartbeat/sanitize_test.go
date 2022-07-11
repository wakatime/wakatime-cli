package heartbeat_test

import (
	"regexp"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/regex"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithSanitization_ObfuscateFile(t *testing.T) {
	opt := heartbeat.WithSanitization(heartbeat.SanitizeConfig{
		FilePatterns: []regex.Regex{regexp.MustCompile(".*")},
	})

	handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Category:   heartbeat.CodingCategory,
				Entity:     "HIDDEN.go",
				EntityType: heartbeat.FileType,
				IsWrite:    heartbeat.PointerTo(true),
				Language:   heartbeat.PointerTo("Go"),
				Project:    heartbeat.PointerTo("wakatime"),
				Time:       1585598060,
				UserAgent:  "wakatime/13.0.7",
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := handle([]heartbeat.Heartbeat{testHeartbeat()})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestSanitize_Obfuscate(t *testing.T) {
	tests := map[string]struct {
		Heartbeat heartbeat.Heartbeat
		Expected  heartbeat.Heartbeat
	}{
		"file": {
			Heartbeat: heartbeat.Heartbeat{
				Branch:         heartbeat.PointerTo("heartbeat"),
				Category:       heartbeat.CodingCategory,
				CursorPosition: heartbeat.PointerTo(12),
				Dependencies:   []string{"dep1", "dep2"},
				Entity:         "/tmp/main.go",
				EntityType:     heartbeat.FileType,
				IsWrite:        heartbeat.PointerTo(true),
				Language:       heartbeat.PointerTo("Go"),
				LineNumber:     heartbeat.PointerTo(42),
				Lines:          heartbeat.PointerTo(100),
				Project:        heartbeat.PointerTo("wakatime"),
				Time:           1585598060,
				UserAgent:      "wakatime/13.0.7",
			},
			Expected: heartbeat.Heartbeat{
				Category:   heartbeat.CodingCategory,
				Entity:     "HIDDEN.go",
				EntityType: heartbeat.FileType,
				IsWrite:    heartbeat.PointerTo(true),
				Language:   heartbeat.PointerTo("Go"),
				Project:    heartbeat.PointerTo("wakatime"),
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
				Entity:     "wakatime.com",
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
				FilePatterns: []regex.Regex{regexp.MustCompile(".*")},
			})

			assert.Equal(t, test.Expected, r)
		})
	}
}

func TestSanitize_ObfuscateFile_SkipBranchIfNotMatching(t *testing.T) {
	r := heartbeat.Sanitize(testHeartbeat(), heartbeat.SanitizeConfig{
		FilePatterns:   []regex.Regex{regexp.MustCompile(".*")},
		BranchPatterns: []regex.Regex{regexp.MustCompile("not_matching")},
	})

	assert.Equal(t, heartbeat.Heartbeat{
		Branch:     heartbeat.PointerTo("heartbeat"),
		Category:   heartbeat.CodingCategory,
		Entity:     "HIDDEN.go",
		EntityType: heartbeat.FileType,
		IsWrite:    heartbeat.PointerTo(true),
		Language:   heartbeat.PointerTo("Go"),
		Project:    heartbeat.PointerTo("wakatime"),
		Time:       1585598060,
		UserAgent:  "wakatime/13.0.7",
	}, r)
}

func TestSanitize_ObfuscateFile_NilFields(t *testing.T) {
	h := testHeartbeat()
	h.Branch = nil

	r := heartbeat.Sanitize(h, heartbeat.SanitizeConfig{
		FilePatterns:   []regex.Regex{regexp.MustCompile(".*")},
		BranchPatterns: []regex.Regex{regexp.MustCompile(".*")},
	})

	assert.Equal(t, heartbeat.Heartbeat{
		Category:   heartbeat.CodingCategory,
		Entity:     "HIDDEN.go",
		EntityType: heartbeat.FileType,
		IsWrite:    heartbeat.PointerTo(true),
		Language:   heartbeat.PointerTo("Go"),
		Project:    heartbeat.PointerTo("wakatime"),
		Time:       1585598060,
		UserAgent:  "wakatime/13.0.7",
	}, r)
}

func TestSanitize_ObfuscateProject(t *testing.T) {
	r := heartbeat.Sanitize(testHeartbeat(), heartbeat.SanitizeConfig{
		ProjectPatterns: []regex.Regex{regexp.MustCompile(".*")},
	})

	assert.Equal(t, heartbeat.Heartbeat{
		Category:   heartbeat.CodingCategory,
		Entity:     "/tmp/main.go",
		EntityType: heartbeat.FileType,
		IsWrite:    heartbeat.PointerTo(true),
		Language:   heartbeat.PointerTo("Go"),
		Project:    heartbeat.PointerTo("wakatime"),
		Time:       1585598060,
		UserAgent:  "wakatime/13.0.7",
	}, r)
}

func TestSanitize_ObfuscateProject_SkipBranchIfNotMatching(t *testing.T) {
	r := heartbeat.Sanitize(testHeartbeat(), heartbeat.SanitizeConfig{
		ProjectPatterns: []regex.Regex{regexp.MustCompile(".*")},
		BranchPatterns:  []regex.Regex{regexp.MustCompile("not_matching")},
	})

	assert.Equal(t, heartbeat.Heartbeat{
		Branch:     heartbeat.PointerTo("heartbeat"),
		Category:   heartbeat.CodingCategory,
		Entity:     "/tmp/main.go",
		EntityType: heartbeat.FileType,
		IsWrite:    heartbeat.PointerTo(true),
		Language:   heartbeat.PointerTo("Go"),
		Project:    heartbeat.PointerTo("wakatime"),
		Time:       1585598060,
		UserAgent:  "wakatime/13.0.7",
	}, r)
}

func TestSanitize_ObfuscateProject_NilFields(t *testing.T) {
	h := testHeartbeat()
	h.Branch = nil

	r := heartbeat.Sanitize(h, heartbeat.SanitizeConfig{
		ProjectPatterns: []regex.Regex{regexp.MustCompile(".*")},
		BranchPatterns:  []regex.Regex{regexp.MustCompile(".*")},
	})

	assert.Equal(t, heartbeat.Heartbeat{
		Category:   heartbeat.CodingCategory,
		Entity:     "/tmp/main.go",
		EntityType: heartbeat.FileType,
		IsWrite:    heartbeat.PointerTo(true),
		Language:   heartbeat.PointerTo("Go"),
		Project:    heartbeat.PointerTo("wakatime"),
		Time:       1585598060,
		UserAgent:  "wakatime/13.0.7",
	}, r)
}

func TestSanitize_ObfuscateBranch(t *testing.T) {
	r := heartbeat.Sanitize(testHeartbeat(), heartbeat.SanitizeConfig{
		BranchPatterns: []regex.Regex{regexp.MustCompile(".*")},
	})

	assert.Equal(t, heartbeat.Heartbeat{
		Category:       heartbeat.CodingCategory,
		CursorPosition: heartbeat.PointerTo(12),
		Dependencies:   []string{"dep1", "dep2"},
		Entity:         "/tmp/main.go",
		EntityType:     heartbeat.FileType,
		IsWrite:        heartbeat.PointerTo(true),
		Language:       heartbeat.PointerTo("Go"),
		LineNumber:     heartbeat.PointerTo(42),
		Lines:          heartbeat.PointerTo(100),
		Project:        heartbeat.PointerTo("wakatime"),
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}, r)
}

func TestSanitize_ObfuscateBranch_NilFields(t *testing.T) {
	h := testHeartbeat()
	h.Branch = nil
	h.Project = nil

	r := heartbeat.Sanitize(h, heartbeat.SanitizeConfig{
		BranchPatterns: []regex.Regex{regexp.MustCompile(".*")},
	})

	assert.Equal(t, heartbeat.Heartbeat{
		Category:       heartbeat.CodingCategory,
		CursorPosition: heartbeat.PointerTo(12),
		Dependencies:   []string{"dep1", "dep2"},
		Entity:         "/tmp/main.go",
		EntityType:     heartbeat.FileType,
		IsWrite:        heartbeat.PointerTo(true),
		Language:       heartbeat.PointerTo("Go"),
		LineNumber:     heartbeat.PointerTo(42),
		Lines:          heartbeat.PointerTo(100),
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}, r)
}

func TestSanitize_EmptyConfigDoNothing(t *testing.T) {
	r := heartbeat.Sanitize(testHeartbeat(), heartbeat.SanitizeConfig{})

	assert.Equal(t, heartbeat.Heartbeat{
		Branch:         heartbeat.PointerTo("heartbeat"),
		Category:       heartbeat.CodingCategory,
		CursorPosition: heartbeat.PointerTo(12),
		Dependencies:   []string{"dep1", "dep2"},
		Entity:         "/tmp/main.go",
		EntityType:     heartbeat.FileType,
		IsWrite:        heartbeat.PointerTo(true),
		Language:       heartbeat.PointerTo("Go"),
		LineNumber:     heartbeat.PointerTo(42),
		Lines:          heartbeat.PointerTo(100),
		Project:        heartbeat.PointerTo("wakatime"),
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}, r)
}

func TestSanitize_EmptyConfigDoNothing_EmptyDependencies(t *testing.T) {
	h := testHeartbeat()
	h.Dependencies = []string{}

	r := heartbeat.Sanitize(h, heartbeat.SanitizeConfig{})

	assert.Equal(t, heartbeat.Heartbeat{
		Branch:         heartbeat.PointerTo("heartbeat"),
		Category:       heartbeat.CodingCategory,
		CursorPosition: heartbeat.PointerTo(12),
		Entity:         "/tmp/main.go",
		EntityType:     heartbeat.FileType,
		IsWrite:        heartbeat.PointerTo(true),
		Language:       heartbeat.PointerTo("Go"),
		LineNumber:     heartbeat.PointerTo(42),
		Lines:          heartbeat.PointerTo(100),
		Project:        heartbeat.PointerTo("wakatime"),
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}, r)
}

func TestSanitize_ObfuscateProjectFolder(t *testing.T) {
	h := testHeartbeat()
	h.Entity = "/path/to/project/main.go"
	h.ProjectPath = "/path/to"

	r := heartbeat.Sanitize(h, heartbeat.SanitizeConfig{
		HideProjectFolder: true,
	})

	assert.Equal(t, heartbeat.Heartbeat{
		Branch:         heartbeat.PointerTo("heartbeat"),
		Category:       heartbeat.CodingCategory,
		CursorPosition: heartbeat.PointerTo(12),
		Dependencies:   []string{"dep1", "dep2"},
		Entity:         "project/main.go",
		EntityType:     heartbeat.FileType,
		IsWrite:        heartbeat.PointerTo(true),
		Language:       heartbeat.PointerTo("Go"),
		LineNumber:     heartbeat.PointerTo(42),
		Lines:          heartbeat.PointerTo(100),
		Project:        heartbeat.PointerTo("wakatime"),
		ProjectPath:    "/path/to/",
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}, r)
}

func TestSanitize_ObfuscateProjectFolder_Override(t *testing.T) {
	h := testHeartbeat()
	h.Entity = "/path/to/project/main.go"
	h.ProjectPath = "/original/folder"
	h.ProjectPathOverride = "/path/to"

	r := heartbeat.Sanitize(h, heartbeat.SanitizeConfig{
		HideProjectFolder: true,
	})

	assert.Equal(t, heartbeat.Heartbeat{
		Branch:              heartbeat.PointerTo("heartbeat"),
		Category:            heartbeat.CodingCategory,
		CursorPosition:      heartbeat.PointerTo(12),
		Dependencies:        []string{"dep1", "dep2"},
		Entity:              "project/main.go",
		EntityType:          heartbeat.FileType,
		IsWrite:             heartbeat.PointerTo(true),
		Language:            heartbeat.PointerTo("Go"),
		LineNumber:          heartbeat.PointerTo(42),
		Lines:               heartbeat.PointerTo(100),
		Project:             heartbeat.PointerTo("wakatime"),
		ProjectPath:         "/original/folder/",
		ProjectPathOverride: "/path/to/",
		Time:                1585598060,
		UserAgent:           "wakatime/13.0.7",
	}, r)
}

func TestSanitize_ObfuscateCredentials_RemoteFile(t *testing.T) {
	h := testHeartbeat()
	h.Entity = "ssh://wakatime:1234@192.168.1.1/path/to/remote/main.go"

	r := heartbeat.Sanitize(h, heartbeat.SanitizeConfig{})

	assert.Equal(t, heartbeat.Heartbeat{
		Branch:         heartbeat.PointerTo("heartbeat"),
		Category:       heartbeat.CodingCategory,
		CursorPosition: heartbeat.PointerTo(12),
		Dependencies:   []string{"dep1", "dep2"},
		Entity:         "ssh://192.168.1.1/path/to/remote/main.go",
		EntityType:     heartbeat.FileType,
		IsWrite:        heartbeat.PointerTo(true),
		Language:       heartbeat.PointerTo("Go"),
		LineNumber:     heartbeat.PointerTo(42),
		Lines:          heartbeat.PointerTo(100),
		Project:        heartbeat.PointerTo("wakatime"),
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}, r)
}

func TestShouldSanitize(t *testing.T) {
	tests := map[string]struct {
		Subject  string
		Regex    []regex.Regex
		Expected bool
	}{
		"match_single": {
			Subject: "fix.123",
			Regex: []regex.Regex{
				regexp.MustCompile("fix.*"),
			},
			Expected: true,
		},
		"match_multiple": {
			Subject: "fix.456",
			Regex: []regex.Regex{
				regexp.MustCompile("bar.*"),
				regexp.MustCompile("fix.*"),
			},
			Expected: true,
		},
		"not_match": {
			Subject: "foo",
			Regex: []regex.Regex{
				regexp.MustCompile("bar.*"),
				regexp.MustCompile("fix.*"),
			},
			Expected: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			shouldSanitize := heartbeat.ShouldSanitize(test.Subject, test.Regex)

			assert.Equal(t, test.Expected, shouldSanitize)
		})
	}
}

func testHeartbeat() heartbeat.Heartbeat {
	return heartbeat.Heartbeat{
		Branch:         heartbeat.PointerTo("heartbeat"),
		Category:       heartbeat.CodingCategory,
		CursorPosition: heartbeat.PointerTo(12),
		Dependencies:   []string{"dep1", "dep2"},
		Entity:         "/tmp/main.go",
		EntityType:     heartbeat.FileType,
		IsWrite:        heartbeat.PointerTo(true),
		Language:       heartbeat.PointerTo("Go"),
		LineNumber:     heartbeat.PointerTo(42),
		Lines:          heartbeat.PointerTo(100),
		Project:        heartbeat.PointerTo("wakatime"),
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}
}
