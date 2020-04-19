package heartbeat_test

import (
	"regexp"
	"testing"

	"github.com/alanhamlett/wakatime-cli/lib/api"
	"github.com/alanhamlett/wakatime-cli/lib/heartbeat"
	"github.com/alanhamlett/wakatime-cli/lib/heartbeat/subtypes"

	"github.com/stretchr/testify/assert"
)

func TestSanitize(t *testing.T) {
	r := heartbeat.Sanitize(testHeartbeat(), heartbeat.Obfuscate{})

	assert.Equal(t, api.Heartbeat{
		Branch:         String("heartbeat"),
		Category:       subtypes.CodingCategory,
		CursorPosition: Int(12),
		Dependencies:   []string{"dep1", "dep2"},
		Entity:         "/tmp/main.go",
		EntityType:     subtypes.FileType,
		IsWrite:        true,
		Language:       "golang",
		LineNumber:     Int(42),
		Lines:          Int(100),
		Project:        "wakatime",
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}, r)
}

func TestSanitize_EmptyDependencies(t *testing.T) {
	h := testHeartbeat()
	h.Dependencies = []string{}

	r := heartbeat.Sanitize(h, heartbeat.Obfuscate{})

	assert.Equal(t, api.Heartbeat{
		Branch:         String("heartbeat"),
		Category:       subtypes.CodingCategory,
		CursorPosition: Int(12),
		Dependencies:   nil,
		Entity:         "/tmp/main.go",
		EntityType:     subtypes.FileType,
		IsWrite:        true,
		Language:       "golang",
		LineNumber:     Int(42),
		Lines:          Int(100),
		Project:        "wakatime",
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}, r)
}

func TestSanitize_ObfuscateFile(t *testing.T) {
	r := heartbeat.Sanitize(testHeartbeat(), heartbeat.Obfuscate{
		HideFileNames: []*regexp.Regexp{regexp.MustCompile(".*")},
	})

	assert.Equal(t, api.Heartbeat{
		Branch:         nil,
		Category:       subtypes.CodingCategory,
		CursorPosition: nil,
		Dependencies:   nil,
		Entity:         "HIDDEN.go",
		EntityType:     subtypes.FileType,
		IsWrite:        true,
		Language:       "golang",
		LineNumber:     nil,
		Lines:          nil,
		Project:        "wakatime",
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}, r)
}

func TestSanitize_ObfuscateFile_SkipBranchIfNotMatching(t *testing.T) {
	r := heartbeat.Sanitize(testHeartbeat(), heartbeat.Obfuscate{
		HideFileNames:   []*regexp.Regexp{regexp.MustCompile(".*")},
		HideBranchNames: []*regexp.Regexp{regexp.MustCompile("not_matching")},
	})

	assert.Equal(t, api.Heartbeat{
		Branch:         String("heartbeat"),
		Category:       subtypes.CodingCategory,
		CursorPosition: nil,
		Dependencies:   nil,
		Entity:         "HIDDEN.go",
		EntityType:     subtypes.FileType,
		IsWrite:        true,
		Language:       "golang",
		LineNumber:     nil,
		Lines:          nil,
		Project:        "wakatime",
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}, r)
}

func TestSanitize_ObfuscateProject(t *testing.T) {
	r := heartbeat.Sanitize(testHeartbeat(), heartbeat.Obfuscate{
		HideProjectNames: []*regexp.Regexp{regexp.MustCompile(".*")},
	})

	assert.Equal(t, api.Heartbeat{
		Branch:         nil,
		Category:       subtypes.CodingCategory,
		CursorPosition: nil,
		Dependencies:   nil,
		Entity:         "/tmp/main.go",
		EntityType:     subtypes.FileType,
		IsWrite:        true,
		Language:       "golang",
		LineNumber:     nil,
		Lines:          nil,
		Project:        "wakatime",
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}, r)
}

func TestSanitize_ObfuscateProject_SkipBranchIfNotMatching(t *testing.T) {
	r := heartbeat.Sanitize(testHeartbeat(), heartbeat.Obfuscate{
		HideProjectNames: []*regexp.Regexp{regexp.MustCompile(".*")},
		HideBranchNames:  []*regexp.Regexp{regexp.MustCompile("not_matching")},
	})

	assert.Equal(t, api.Heartbeat{
		Branch:         String("heartbeat"),
		Category:       subtypes.CodingCategory,
		CursorPosition: nil,
		Dependencies:   nil,
		Entity:         "/tmp/main.go",
		EntityType:     subtypes.FileType,
		IsWrite:        true,
		Language:       "golang",
		LineNumber:     nil,
		Lines:          nil,
		Project:        "wakatime",
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}, r)
}

func TestSanitize_ObfuscateBranch(t *testing.T) {
	r := heartbeat.Sanitize(testHeartbeat(), heartbeat.Obfuscate{
		HideBranchNames: []*regexp.Regexp{regexp.MustCompile(".*")},
	})

	assert.Equal(t, api.Heartbeat{
		Branch:         nil,
		Category:       subtypes.CodingCategory,
		CursorPosition: Int(12),
		Dependencies:   []string{"dep1", "dep2"},
		Entity:         "/tmp/main.go",
		EntityType:     subtypes.FileType,
		IsWrite:        true,
		Language:       "golang",
		LineNumber:     Int(42),
		Lines:          Int(100),
		Project:        "wakatime",
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}, r)
}

func TestSanitize_EntityTypeNotFile_DoesNothing(t *testing.T) {
	tests := map[string]subtypes.EntityType{
		"domain": subtypes.DomainType,
		"app":    subtypes.AppType,
	}

	for name, entityType := range tests {
		t.Run(name, func(t *testing.T) {
			h := testHeartbeat()
			h.EntityType = entityType

			r := heartbeat.Sanitize(h, heartbeat.Obfuscate{
				HideBranchNames:  []*regexp.Regexp{regexp.MustCompile(".*")},
				HideFileNames:    []*regexp.Regexp{regexp.MustCompile(".*")},
				HideProjectNames: []*regexp.Regexp{regexp.MustCompile(".*")},
			})

			assert.Equal(t, api.Heartbeat{
				Branch:         String("heartbeat"),
				Category:       subtypes.CodingCategory,
				CursorPosition: Int(12),
				Dependencies:   []string{"dep1", "dep2"},
				Entity:         "/tmp/main.go",
				EntityType:     entityType,
				IsWrite:        true,
				Language:       "golang",
				LineNumber:     Int(42),
				Lines:          Int(100),
				Project:        "wakatime",
				Time:           1585598060,
				UserAgent:      "wakatime/13.0.7",
			}, r)
		})
	}
}

func testHeartbeat() heartbeat.Heartbeat {
	return heartbeat.Heartbeat{
		Branch:         "heartbeat",
		Category:       subtypes.CodingCategory,
		CursorPosition: 12,
		Dependencies:   []string{"dep1", "dep2"},
		Entity:         "/tmp/main.go",
		EntityType:     subtypes.FileType,
		IsWrite:        true,
		Language:       "golang",
		LineNumber:     42,
		Lines:          100,
		Project:        "wakatime",
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}
}

// Int returns a pointer to the int value passed in.
func Int(v int) *int {
	return &v
}

// String returns a pointer to the string value passed in.
func String(v string) *string {
	return &v
}
