package params_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	paramscmd "github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/apikey"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	inipkg "github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/output"
	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/regex"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/ini.v1"
)

func TestLoadParams_AlternateProject(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("alternate-project", "web")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, "web", params.Project.Alternate)
}

func TestLoadParams_AlternateProject_Unset(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Empty(t, params.Project.Alternate)
}

func TestLoadParams_Category(t *testing.T) {
	tests := map[string]heartbeat.Category{
		"browsing":       heartbeat.BrowsingCategory,
		"building":       heartbeat.BuildingCategory,
		"coding":         heartbeat.CodingCategory,
		"code reviewing": heartbeat.CodeReviewingCategory,
		"communicating":  heartbeat.CommunicatingCategory,
		"debugging":      heartbeat.DebuggingCategory,
		"designing":      heartbeat.DesigningCategory,
		"indexing":       heartbeat.IndexingCategory,
		"learning":       heartbeat.LearningCategory,
		"manual testing": heartbeat.ManualTestingCategory,
		"planning":       heartbeat.PlanningCategory,
		"researching":    heartbeat.ResearchingCategory,
		"running tests":  heartbeat.RunningTestsCategory,
		"translating":    heartbeat.TranslatingCategory,
		"writing docs":   heartbeat.WritingDocsCategory,
		"writing tests":  heartbeat.WritingTestsCategory,
	}

	for name, category := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", "/path/to/file")
			v.Set("category", name)

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			assert.Equal(t, category, params.Category)
		})
	}
}

func TestLoadParams_Category_Default(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, heartbeat.CodingCategory, params.Category)
}

func TestLoadParams_Category_Invalid(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("category", "invalid")

	_, err := paramscmd.LoadHeartbeatParams(v)
	require.Error(t, err)

	assert.Equal(t, "failed to parse category: invalid category \"invalid\"", err.Error())
}

func TestLoadParams_CursorPosition(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("cursorpos", 42)

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, 42, *params.CursorPosition)
}

func TestLoadParams_CursorPosition_Zero(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("cursorpos", 0)

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Zero(t, *params.CursorPosition)
}

func TestLoadParams_CursorPosition_Unset(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Nil(t, params.CursorPosition)
}

func TestLoadParams_Entity_EntityFlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("file", "ignored")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, "/path/to/file", params.Entity)
}

func TestLoadParams_Entity_FileFlag(t *testing.T) {
	v := viper.New()
	v.Set("file", "~/path/to/file")

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(home, "/path/to/file"), params.Entity)
}

func TestLoadParams_Entity_Unset(t *testing.T) {
	v := viper.New()

	_, err := paramscmd.LoadHeartbeatParams(v)
	require.Error(t, err)

	assert.Equal(t, "failed to retrieve entity", err.Error())
}

func TestLoadParams_EntityType(t *testing.T) {
	tests := map[string]heartbeat.EntityType{
		"file":   heartbeat.FileType,
		"domain": heartbeat.DomainType,
		"app":    heartbeat.AppType,
	}

	for name, entityType := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", "/path/to/file")
			v.Set("entity-type", name)

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			assert.Equal(t, entityType, params.EntityType)
		})
	}
}

func TestLoadParams_EntityType_Default(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, heartbeat.FileType, params.EntityType)
}

func TestLoadParams_EntityType_Invalid(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("entity-type", "invalid")

	_, err := paramscmd.LoadHeartbeatParams(v)
	require.Error(t, err)

	assert.Equal(
		t,
		"failed to parse entity type: invalid entity type \"invalid\"",
		err.Error())
}

func TestLoadParams_ExtraHeartbeats(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		r.Close()
		w.Close()
	}()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = r

	data, err := os.ReadFile("testdata/extra_heartbeats.json")
	require.NoError(t, err)

	go func() {
		_, err := w.Write(data)
		require.NoError(t, err)

		w.Close()
	}()

	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("extra-heartbeats", true)

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Len(t, params.ExtraHeartbeats, 2)

	assert.NotNil(t, params.ExtraHeartbeats[0].Language)
	assert.Equal(t, heartbeat.LanguageGo.String(), *params.ExtraHeartbeats[0].Language)
	assert.NotNil(t, params.ExtraHeartbeats[1].Language)
	assert.Equal(t, heartbeat.LanguagePython.String(), *params.ExtraHeartbeats[1].Language)

	assert.Equal(t, []heartbeat.Heartbeat{
		{
			Category:          heartbeat.CodingCategory,
			CursorPosition:    heartbeat.PointerTo(12),
			Entity:            "testdata/main.go",
			EntityType:        heartbeat.FileType,
			IsUnsavedEntity:   true,
			IsWrite:           heartbeat.PointerTo(true),
			LanguageAlternate: "Golang",
			LineNumber:        heartbeat.PointerTo(42),
			Lines:             heartbeat.PointerTo(45),
			ProjectAlternate:  "billing",
			ProjectOverride:   "wakatime-cli",
			Time:              1585598059,
			// tested above
			Language: params.ExtraHeartbeats[0].Language,
		},
		{
			Category:          heartbeat.DebuggingCategory,
			Entity:            "testdata/main.py",
			EntityType:        heartbeat.FileType,
			IsWrite:           nil,
			LanguageAlternate: "Py",
			LineNumber:        nil,
			Lines:             nil,
			ProjectOverride:   "wakatime-cli",
			Time:              1585598060,
			// tested above
			Language: params.ExtraHeartbeats[1].Language,
		},
	}, params.ExtraHeartbeats)
}

func TestLoadParams_ExtraHeartbeats_WithStringValues(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		r.Close()
		w.Close()
	}()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = r

	data, err := os.ReadFile("testdata/extra_heartbeats_with_string_values.json")
	require.NoError(t, err)

	go func() {
		_, err := w.Write(data)
		require.NoError(t, err)

		w.Close()
	}()

	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("extra-heartbeats", true)

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Len(t, params.ExtraHeartbeats, 2)

	assert.NotNil(t, params.ExtraHeartbeats[0].Language)
	assert.Equal(t, heartbeat.LanguageGo.String(), *params.ExtraHeartbeats[0].Language)
	assert.NotNil(t, params.ExtraHeartbeats[1].Language)
	assert.Equal(t, heartbeat.LanguagePython.String(), *params.ExtraHeartbeats[1].Language)

	assert.Equal(t, []heartbeat.Heartbeat{
		{
			Category:        heartbeat.CodingCategory,
			CursorPosition:  heartbeat.PointerTo(12),
			Entity:          "testdata/main.go",
			EntityType:      heartbeat.FileType,
			IsUnsavedEntity: true,
			IsWrite:         heartbeat.PointerTo(true),
			Language:        params.ExtraHeartbeats[0].Language,
			Lines:           heartbeat.PointerTo(45),
			LineNumber:      heartbeat.PointerTo(42),
			Time:            1585598059,
		},
		{
			Category:        heartbeat.CodingCategory,
			CursorPosition:  heartbeat.PointerTo(13),
			Entity:          "testdata/main.go",
			EntityType:      heartbeat.FileType,
			IsUnsavedEntity: true,
			IsWrite:         heartbeat.PointerTo(true),
			Language:        params.ExtraHeartbeats[1].Language,
			LineNumber:      heartbeat.PointerTo(43),
			Lines:           heartbeat.PointerTo(46),
			Time:            1585598060,
		},
	}, params.ExtraHeartbeats)
}

func TestLoadParams_ExtraHeartbeats_WithEOF(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		r.Close()
		w.Close()
	}()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = r

	data, err := os.ReadFile("testdata/extra_heartbeats.json")
	require.NoError(t, err)

	go func() {
		// trim trailing newline and make sure we still parse extra heartbeats
		_, err := w.Write(bytes.TrimRight(data, "\n"))
		require.NoError(t, err)

		w.Close()
	}()

	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("extra-heartbeats", true)

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Len(t, params.ExtraHeartbeats, 2)

	assert.NotNil(t, params.ExtraHeartbeats[0].Language)
	assert.Equal(t, heartbeat.LanguageGo.String(), *params.ExtraHeartbeats[0].Language)
	assert.NotNil(t, params.ExtraHeartbeats[1].Language)
	assert.Equal(t, heartbeat.LanguagePython.String(), *params.ExtraHeartbeats[1].Language)

	assert.Equal(t, []heartbeat.Heartbeat{
		{
			Category:          heartbeat.CodingCategory,
			CursorPosition:    heartbeat.PointerTo(12),
			Entity:            "testdata/main.go",
			EntityType:        heartbeat.FileType,
			IsUnsavedEntity:   true,
			IsWrite:           heartbeat.PointerTo(true),
			LanguageAlternate: "Golang",
			LineNumber:        heartbeat.PointerTo(42),
			Lines:             heartbeat.PointerTo(45),
			ProjectAlternate:  "billing",
			ProjectOverride:   "wakatime-cli",
			Time:              1585598059,
			// tested above
			Language: params.ExtraHeartbeats[0].Language,
		},
		{
			Category:          heartbeat.DebuggingCategory,
			Entity:            "testdata/main.py",
			EntityType:        heartbeat.FileType,
			IsWrite:           nil,
			LanguageAlternate: "Py",
			LineNumber:        nil,
			Lines:             nil,
			ProjectOverride:   "wakatime-cli",
			Time:              1585598060,
			// tested above
			Language: params.ExtraHeartbeats[1].Language,
		},
	}, params.ExtraHeartbeats)
}

func TestLoadParams_Filter_IsUnsavedEntity(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("is-unsaved-entity", true)

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.True(t, params.IsUnsavedEntity)
}

func TestLoadParams_IsWrite(t *testing.T) {
	tests := map[string]bool{
		"is write":    true,
		"is no write": false,
	}

	for name, isWrite := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", "/path/to/file")
			v.Set("write", isWrite)

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			assert.Equal(t, isWrite, *params.IsWrite)
		})
	}
}

func TestLoadParams_IsWrite_Unset(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Nil(t, params.IsWrite)
}

func TestLoadParams_Language(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("language", "Go")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageGo.String(), *params.Language)
}

func TestLoadParams_LanguageAlternate(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("alternate-language", "Go")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageGo.String(), params.LanguageAlternate)
	assert.Nil(t, params.Language)
}

func TestLoadParams_LineNumber(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("lineno", 42)

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, 42, *params.LineNumber)
}

func TestLoadParams_LineNumber_Zero(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("lineno", 0)

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Zero(t, *params.LineNumber)
}

func TestLoadParams_LineNumber_Unset(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Nil(t, params.LineNumber)
}

func TestLoadParams_LocalFile(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("local-file", "/path/to/file")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, "/path/to/file", params.LocalFile)
}

func TestLoadParams_Project(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("project", "billing")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, "billing", params.Project.Override)
}

func TestLoadParams_Project_Unset(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Empty(t, params.Project.Override)
}

func TestLoadParams_ProjectMap(t *testing.T) {
	tests := map[string]struct {
		Entity   string
		Regex    regex.Regex
		Project  string
		Expected []project.MapPattern
	}{
		"simple regex": {
			Entity:  "/home/user/projects/foo/file",
			Regex:   regexp.MustCompile("projects/foo"),
			Project: "My Awesome Project",
			Expected: []project.MapPattern{
				{
					Name:  "My Awesome Project",
					Regex: regexp.MustCompile("(?i)projects/foo"),
				},
			},
		},
		"regex with group replacement": {
			Entity:  "/home/user/projects/bar123/file",
			Regex:   regexp.MustCompile(`^/home/user/projects/bar(\\d+)/`),
			Project: "project{0}",
			Expected: []project.MapPattern{
				{
					Name:  "project{0}",
					Regex: regexp.MustCompile(`(?i)^/home/user/projects/bar(\\d+)/`),
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", test.Entity)
			v.Set(fmt.Sprintf("projectmap.%s", test.Regex.String()), test.Project)

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params.Project.MapPatterns)
		})
	}
}

func TestLoadParams_ProjectApiKey(t *testing.T) {
	tests := map[string]struct {
		Entity   string
		Regex    regex.Regex
		APIKey   string
		Expected []apikey.MapPattern
	}{
		"simple regex": {
			Regex:  regexp.MustCompile("projects/foo"),
			APIKey: "00000000-0000-4000-8000-000000000001",
			Expected: []apikey.MapPattern{
				{
					APIKey: "00000000-0000-4000-8000-000000000001",
					Regex:  regexp.MustCompile(`(?i)projects/foo`),
				},
			},
		},
		"complex regex": {
			Regex:  regexp.MustCompile(`^/home/user/projects/bar(\\d+)/`),
			APIKey: "00000000-0000-4000-8000-000000000002",
			Expected: []apikey.MapPattern{
				{
					APIKey: "00000000-0000-4000-8000-000000000002",
					Regex:  regexp.MustCompile(`(?i)^/home/user/projects/bar(\\d+)/`),
				},
			},
		},
		"case insensitive": {
			Regex:  regexp.MustCompile("projects/foo"),
			APIKey: "00000000-0000-4000-8000-000000000001",
			Expected: []apikey.MapPattern{
				{
					APIKey: "00000000-0000-4000-8000-000000000001",
					Regex:  regexp.MustCompile(`(?i)projects/foo`),
				},
			},
		},
		"api key equal to default": {
			Regex:    regexp.MustCompile(`/some/path`),
			APIKey:   "00000000-0000-4000-8000-000000000000",
			Expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set(fmt.Sprintf("project_api_key.%s", test.Regex.String()), test.APIKey)

			params, err := paramscmd.LoadAPIParams(v)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params.KeyPatterns)
		})
	}
}

func TestLoadParams_ProjectApiKey_ParseConfig(t *testing.T) {
	v := viper.New()
	v.Set("config", "testdata/.wakatime.cfg")
	v.Set("entity", "testdata/heartbeat_go.json")

	configFile, err := inipkg.FilePath(v)
	require.NoError(t, err)

	err = inipkg.ReadInConfig(v, configFile)
	require.NoError(t, err)

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	expected := []apikey.MapPattern{
		{
			APIKey: "00000000-0000-4000-8000-000000000001",
			Regex:  regex.MustCompile("(?i)/some/path"),
		},
	}

	assert.Equal(t, expected, params.KeyPatterns)
}

func TestLoadParams_Time(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("time", 1590609206.1)

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, 1590609206.1, params.Time)
}

func TestLoadParams_Time_Default(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	now := float64(time.Now().UnixNano()) / 1000000000
	assert.GreaterOrEqual(t, now, params.Time)
	assert.GreaterOrEqual(t, params.Time, now-60)
}

func TestLoadParams_Filter_Exclude(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("exclude", []string{".*", "wakatime.*"})
	v.Set("settings.exclude", []string{".+", "wakatime.+"})
	v.Set("settings.ignore", []string{".?", "wakatime.?"})

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	require.Len(t, params.Filter.Exclude, 6)
	assert.Equal(t, "(?i).*", params.Filter.Exclude[0].String())
	assert.Equal(t, "(?i)wakatime.*", params.Filter.Exclude[1].String())
	assert.Equal(t, "(?i).+", params.Filter.Exclude[2].String())
	assert.Equal(t, "(?i)wakatime.+", params.Filter.Exclude[3].String())
	assert.Equal(t, "(?i).?", params.Filter.Exclude[4].String())
	assert.Equal(t, "(?i)wakatime.?", params.Filter.Exclude[5].String())
}

func TestLoadParams_Filter_Exclude_Multiline(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("settings.ignore", "\t.?\n\twakatime.? \t\n")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	require.Len(t, params.Filter.Exclude, 2)
	assert.Equal(t, "(?i).?", params.Filter.Exclude[0].String())
	assert.Equal(t, "(?i)wakatime.?", params.Filter.Exclude[1].String())
}

func TestLoadParams_Filter_Exclude_IgnoresInvalidRegex(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("exclude", []string{".*", "["})

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	require.Len(t, params.Filter.Exclude, 1)
	assert.Equal(t, "(?i).*", params.Filter.Exclude[0].String())
}

func TestLoadParams_Filter_Exclude_PerlRegexPatterns(t *testing.T) {
	tests := map[string]string{
		"negative lookahead": `^/var/(?!www/).*`,
		"positive lookahead": `^/var/(?=www/).*`,
	}

	for name, pattern := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", "/path/to/file")
			v.Set("exclude", []string{pattern})

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			require.Len(t, params.Filter.Exclude, 1)
			assert.Equal(t, "(?i)"+pattern, params.Filter.Exclude[0].String())
		})
	}
}

func TestLoadParams_Filter_ExcludeUnknownProject(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("exclude-unknown-project", true)

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.True(t, params.Filter.ExcludeUnknownProject)
}

func TestLoadParams_Filter_ExcludeUnknownProject_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("exclude-unknown-project", false)
	v.Set("settings.exclude_unknown_project", true)

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.True(t, params.Filter.ExcludeUnknownProject)
}

func TestLoadParams_Filter_Include(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("include", []string{".*", "wakatime.*"})
	v.Set("settings.include", []string{".+", "wakatime.+"})

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	require.Len(t, params.Filter.Include, 4)
	assert.Equal(t, "(?i).*", params.Filter.Include[0].String())
	assert.Equal(t, "(?i)wakatime.*", params.Filter.Include[1].String())
	assert.Equal(t, "(?i).+", params.Filter.Include[2].String())
	assert.Equal(t, "(?i)wakatime.+", params.Filter.Include[3].String())
}

func TestLoadParams_Filter_Include_IgnoresInvalidRegex(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("include", []string{".*", "["})

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	require.Len(t, params.Filter.Include, 1)
	assert.Equal(t, "(?i).*", params.Filter.Include[0].String())
}

func TestLoadParams_Filter_Include_PerlRegexPatterns(t *testing.T) {
	tests := map[string]string{
		"negative lookahead": `^/var/(?!www/).*`,
		"positive lookahead": `^/var/(?=www/).*`,
	}

	for name, pattern := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", "/path/to/file")
			v.Set("include", []string{pattern})

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			require.Len(t, params.Filter.Include, 1)
			assert.Equal(t, "(?i)"+pattern, params.Filter.Include[0].String())
		})
	}
}

func TestLoadParams_Filter_IncludeOnlyWithProjectFile(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("include-only-with-project-file", true)

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.True(t, params.Filter.IncludeOnlyWithProjectFile)
}

func TestLoadParams_Filter_IncludeOnlyWithProjectFile_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("include-only-with-project-file", false)
	v.Set("settings.include_only_with_project_file", true)

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.True(t, params.Filter.IncludeOnlyWithProjectFile)
}

func TestLoadParams_SanitizeParams_HideBranchNames_True(t *testing.T) {
	tests := map[string]string{
		"lowercase":       "true",
		"uppercase":       "TRUE",
		"first uppercase": "True",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", "/path/to/file")
			v.Set("hide-branch-names", viperValue)

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{
				HideBranchNames: []regex.Regex{regex.MustCompile(".*")},
			}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideBranchNames_False(t *testing.T) {
	tests := map[string]string{
		"lowercase":       "false",
		"uppercase":       "FALSE",
		"first uppercase": "False",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", "/path/to/file")
			v.Set("hide-branch-names", viperValue)

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{
				HideBranchNames: []regex.Regex{regex.MustCompile("a^")},
			}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideBranchNames_List(t *testing.T) {
	tests := map[string]struct {
		ViperValue string
		Expected   []regex.Regex
	}{
		"regex": {
			ViperValue: "fix.*",
			Expected: []regex.Regex{
				regexp.MustCompile("fix.*"),
			},
		},
		"regex list": {
			ViperValue: ".*secret.*\nfix.*",
			Expected: []regex.Regex{
				regexp.MustCompile(".*secret.*"),
				regexp.MustCompile("fix.*"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", "/path/to/file")
			v.Set("hide-branch-names", test.ViperValue)

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{
				HideBranchNames: test.Expected,
			}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideBranchNames_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("hide-branch-names", true)
	v.Set("settings.hide_branch_names", "ignored")
	v.Set("settings.hide_branchnames", "ignored")
	v.Set("settings.hidebranchnames", "ignored")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideBranchNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideBranchNames_ConfigTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_branch_names", "true")
	v.Set("settings.hide_branchnames", "ignored")
	v.Set("settings.hidebranchnames", "ignored")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideBranchNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideBranchNames_ConfigDeprecatedOneTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_branchnames", "true")
	v.Set("settings.hidebranchnames", "ignored")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideBranchNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideBranchNames_ConfigDeprecatedTwo(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hidebranchnames", "true")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideBranchNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideBranchNames_InvalidRegex(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("hide-branch-names", ".*secret.*\n[0-9+")

	_, err := paramscmd.LoadHeartbeatParams(v)
	require.Error(t, err)

	assert.True(t, strings.HasPrefix(
		err.Error(),
		"failed to load sanitize params:"+
			" failed to parse regex hide branch names param \".*secret.*\\n[0-9+\":"+
			" failed to compile regex \"[0-9+\":",
	))
}

func TestLoadParams_SanitizeParams_HideProjectNames_True(t *testing.T) {
	tests := map[string]string{
		"lowercase":       "true",
		"uppercase":       "TRUE",
		"first uppercase": "True",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", "/path/to/file")
			v.Set("hide-project-names", viperValue)

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{
				HideProjectNames: []regex.Regex{regexp.MustCompile(".*")},
			}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideProjectNames_False(t *testing.T) {
	tests := map[string]string{
		"lowercase":       "false",
		"uppercase":       "FALSE",
		"first uppercase": "False",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", "/path/to/file")
			v.Set("hide-project-names", viperValue)

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{
				HideProjectNames: []regex.Regex{regexp.MustCompile("a^")},
			}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideProjecthNames_List(t *testing.T) {
	tests := map[string]struct {
		ViperValue string
		Expected   []regex.Regex
	}{
		"regex": {
			ViperValue: "fix.*",
			Expected: []regex.Regex{
				regexp.MustCompile("fix.*"),
			},
		},
		"regex list": {
			ViperValue: ".*secret.*\nfix.*",
			Expected: []regex.Regex{
				regexp.MustCompile(".*secret.*"),
				regexp.MustCompile("fix.*"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", "/path/to/file")
			v.Set("hide-project-names", test.ViperValue)

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{
				HideProjectNames: test.Expected,
			}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideProjectNames_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("hide-project-names", "true")
	v.Set("settings.hide_project_names", "ignored")
	v.Set("settings.hide_projectnames", "ignored")
	v.Set("settings.hideprojectnames", "ignored")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideProjectNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideProjectNames_ConfigTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_project_names", "true")
	v.Set("settings.hide_projectnames", "ignored")
	v.Set("settings.hideprojectnames", "ignored")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideProjectNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideProjectNames_ConfigDeprecatedOneTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_projectnames", "true")
	v.Set("settings.hideprojectnames", "ignored")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideProjectNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideProjectNames_ConfigDeprecatedTwo(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hideprojectnames", "true")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideProjectNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideProjectNames_InvalidRegex(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("hide-project-names", ".*secret.*\n[0-9+")

	_, err := paramscmd.LoadHeartbeatParams(v)
	require.Error(t, err)

	assert.True(t, strings.HasPrefix(
		err.Error(),
		"failed to load sanitize params:"+
			" failed to parse regex hide project names param \".*secret.*\\n[0-9+\":"+
			" failed to compile regex \"[0-9+\":",
	))
}

func TestLoadParams_SanitizeParams_HideFileNames_True(t *testing.T) {
	tests := map[string]string{
		"lowercase":       "true",
		"uppercase":       "TRUE",
		"first uppercase": "True",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", "/path/to/file")
			v.Set("hide-file-names", viperValue)

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{
				HideFileNames: []regex.Regex{regexp.MustCompile(".*")},
			}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideFileNames_False(t *testing.T) {
	tests := map[string]string{
		"lowercase":       "false",
		"uppercase":       "FALSE",
		"first uppercase": "False",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", "/path/to/file")
			v.Set("hide-file-names", viperValue)

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{
				HideFileNames: []regex.Regex{regexp.MustCompile("a^")},
			}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideFileNames_List(t *testing.T) {
	tests := map[string]struct {
		ViperValue string
		Expected   []regex.Regex
	}{
		"regex": {
			ViperValue: "fix.*",
			Expected: []regex.Regex{
				regexp.MustCompile("fix.*"),
			},
		},
		"regex list": {
			ViperValue: ".*secret.*\nfix.*",
			Expected: []regex.Regex{
				regexp.MustCompile(".*secret.*"),
				regexp.MustCompile("fix.*"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", "/path/to/file")
			v.Set("hide-file-names", test.ViperValue)

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{
				HideFileNames: test.Expected,
			}, params.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideFileNames_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("hide-file-names", "true")
	v.Set("hide-filenames", "ignored")
	v.Set("hidefilenames", "ignored")
	v.Set("settings.hide_file_names", "ignored")
	v.Set("settings.hide_filenames", "ignored")
	v.Set("settings.hidefilenames", "ignored")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideFileNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_FlagDeprecatedOneTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("hide-filenames", "true")
	v.Set("hidefilenames", "ignored")
	v.Set("settings.hide_file_names", "ignored")
	v.Set("settings.hide_filenames", "ignored")
	v.Set("settings.hidefilenames", "ignored")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideFileNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_FlagDeprecatedTwoTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("hidefilenames", "true")
	v.Set("settings.hide_file_names", "ignored")
	v.Set("settings.hide_filenames", "ignored")
	v.Set("settings.hidefilenames", "ignored")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideFileNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_ConfigTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_file_names", "true")
	v.Set("settings.hide_filenames", "ignored")
	v.Set("settings.hidefilenames", "ignored")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideFileNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_ConfigDeprecatedOneTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_filenames", "true")
	v.Set("settings.hidefilenames", "ignored")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideFileNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_ConfigDeprecatedTwo(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hidefilenames", "true")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideFileNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_InvalidRegex(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("hide-file-names", ".*secret.*\n[0-9+")

	_, err := paramscmd.LoadHeartbeatParams(v)
	require.Error(t, err)

	assert.True(t, strings.HasPrefix(
		err.Error(),
		"failed to load sanitize params:"+
			" failed to parse regex hide file names param \".*secret.*\\n[0-9+\":"+
			" failed to compile regex \"[0-9+\":",
	))
}

func TestLoadParams_SanitizeParams_HideProjectFolder(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("hide-project-folder", true)

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideProjectFolder: true,
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_HideProjectFolder_ConfigTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_project_folder", true)

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideProjectFolder: true,
	}, params.Sanitize)
}

func TestLoadParams_SanitizeParams_OverrideProjectPath(t *testing.T) {
	v := viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("project-folder", "/custom-path")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		ProjectPathOverride: "/custom-path",
	}, params.Sanitize)
}

func TestLoadParams_SubmodulesDisabled_True(t *testing.T) {
	tests := map[string]string{
		"lowercase":       "true",
		"uppercase":       "TRUE",
		"first uppercase": "True",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", "/path/to/file")
			v.Set("git.submodules_disabled", viperValue)

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			assert.Equal(t, []regex.Regex{regexp.MustCompile(".*")}, params.Project.SubmodulesDisabled)
		})
	}
}

func TestLoadParams_SubmodulesDisabled_False(t *testing.T) {
	tests := map[string]string{
		"lowercase":       "false",
		"uppercase":       "FALSE",
		"first uppercase": "False",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", "/path/to/file")
			v.Set("git.submodules_disabled", viperValue)

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			assert.Equal(t, params.Project.SubmodulesDisabled, []regex.Regex{regexp.MustCompile("a^")})
		})
	}
}

func TestLoadParams_SubmodulesDisabled_List(t *testing.T) {
	tests := map[string]struct {
		ViperValue string
		Expected   []regex.Regex
	}{
		"regex": {
			ViperValue: "fix.*",
			Expected: []regex.Regex{
				regexp.MustCompile("fix.*"),
			},
		},
		"regex_list": {
			ViperValue: "\n.*secret.*\nfix.*",
			Expected: []regex.Regex{
				regexp.MustCompile(".*secret.*"),
				regexp.MustCompile("fix.*"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			multilineOption := viper.IniLoadOptions(ini.LoadOptions{AllowPythonMultilineValues: true})
			v := viper.NewWithOptions(multilineOption)
			v.Set("entity", "/path/to/file")
			v.Set("git.submodules_disabled", test.ViperValue)

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params.Project.SubmodulesDisabled)
		})
	}
}

func TestLoadParams_SubmoduleProjectMap(t *testing.T) {
	tests := map[string]struct {
		Entity   string
		Regex    regex.Regex
		Project  string
		Expected []project.MapPattern
	}{
		"simple regex": {
			Entity:  "/home/user/projects/foo/file",
			Regex:   regexp.MustCompile("projects/foo"),
			Project: "My Awesome Project",
			Expected: []project.MapPattern{
				{
					Name:  "My Awesome Project",
					Regex: regexp.MustCompile("(?i)projects/foo"),
				},
			},
		},
		"regex with group replacement": {
			Entity:  "/home/user/projects/bar123/file",
			Regex:   regexp.MustCompile(`^/home/user/projects/bar(\\d+)/`),
			Project: "project{0}",
			Expected: []project.MapPattern{
				{
					Name:  "project{0}",
					Regex: regexp.MustCompile(`(?i)^/home/user/projects/bar(\\d+)/`),
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("entity", test.Entity)
			v.Set(fmt.Sprintf("git_submodule_projectmap.%s", test.Regex.String()), test.Project)

			params, err := paramscmd.LoadHeartbeatParams(v)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params.Project.SubmoduleMapPatterns)
		})
	}
}

func TestLoadParams_Plugin(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("plugin", "plugin/10.0.0")

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, "plugin/10.0.0", params.Plugin)
}

func TestLoadParams_Plugin_Unset(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Empty(t, params.Plugin)
}

func TestLoadParams_Timeout_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("timeout", 5)
	v.Set("settings.timeout", 10)

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, 5*time.Second, params.Timeout)
}

func TestLoadParams_Timeout_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.timeout", 10)

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, 10*time.Second, params.Timeout)
}

func TestLoad_OfflineDisabled_ConfigTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("disable-offline", false)
	v.Set("disableoffline", false)
	v.Set("settings.offline", false)

	params := paramscmd.LoadOfflineParams(v)

	assert.True(t, params.Disabled)
}

func TestLoad_OfflineDisabled_FlagDeprecatedTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("disable-offline", false)
	v.Set("disableoffline", true)

	params := paramscmd.LoadOfflineParams(v)

	assert.True(t, params.Disabled)
}

func TestLoad_OfflineDisabled_FromFlag(t *testing.T) {
	v := viper.New()
	v.Set("disable-offline", true)

	params := paramscmd.LoadOfflineParams(v)

	assert.True(t, params.Disabled)
}

func TestLoad_OfflineQueueFile(t *testing.T) {
	v := viper.New()
	v.Set("offline-queue-file", "/path/to/file")

	params := paramscmd.LoadOfflineParams(v)

	assert.Equal(t, "/path/to/file", params.QueueFile)
}

func TestLoad_OfflineSyncMax(t *testing.T) {
	v := viper.New()
	v.Set("sync-offline-activity", 42)

	params := paramscmd.LoadOfflineParams(v)

	assert.Equal(t, 42, params.SyncMax)
}

func TestLoad_OfflineSyncMax_Zero(t *testing.T) {
	v := viper.New()
	v.Set("sync-offline-activity", "0")

	params := paramscmd.LoadOfflineParams(v)

	assert.Zero(t, params.SyncMax)
}

func TestLoad_OfflineSyncMax_Default(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)

	params := paramscmd.LoadOfflineParams(v)

	assert.Equal(t, 1000, params.SyncMax)
}

func TestLoad_OfflineSyncMax_NegativeNumber(t *testing.T) {
	v := viper.New()
	v.Set("sync-offline-activity", -1)

	params := paramscmd.LoadOfflineParams(v)

	assert.Equal(t, 0, params.SyncMax)
}

func TestLoad_OfflineSyncMax_NonIntegerValue(t *testing.T) {
	v := viper.New()
	v.Set("sync-offline-activity", "invalid")

	params := paramscmd.LoadOfflineParams(v)

	assert.Equal(t, 0, params.SyncMax)
}

func TestLoad_API_APIKey(t *testing.T) {
	tests := map[string]struct {
		ViperAPIKey          string
		ViperAPIKeyConfig    string
		ViperAPIKeyConfigOld string
		Expected             paramscmd.API
	}{
		"api key flag takes precedence": {
			ViperAPIKey:          "00000000-0000-4000-8000-000000000000",
			ViperAPIKeyConfig:    "10000000-0000-4000-8000-000000000000",
			ViperAPIKeyConfigOld: "20000000-0000-4000-8000-000000000000",
			Expected: paramscmd.API{
				Key:      "00000000-0000-4000-8000-000000000000",
				URL:      "https://api.wakatime.com/api/v1",
				Hostname: "my-computer",
			},
		},
		"api from config takes precedence": {
			ViperAPIKeyConfig:    "00000000-0000-4000-8000-000000000000",
			ViperAPIKeyConfigOld: "10000000-0000-4000-8000-000000000000",
			Expected: paramscmd.API{
				Key:      "00000000-0000-4000-8000-000000000000",
				URL:      "https://api.wakatime.com/api/v1",
				Hostname: "my-computer",
			},
		},
		"api key from config deprecated": {
			ViperAPIKeyConfigOld: "00000000-0000-4000-8000-000000000000",
			Expected: paramscmd.API{
				Key:      "00000000-0000-4000-8000-000000000000",
				URL:      "https://api.wakatime.com/api/v1",
				Hostname: "my-computer",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("hostname", "my-computer")
			v.Set("key", test.ViperAPIKey)
			v.Set("settings.api_key", test.ViperAPIKeyConfig)
			v.Set("settings.apikey", test.ViperAPIKeyConfigOld)

			params, err := paramscmd.LoadAPIParams(v)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params)
		})
	}
}

func TestLoad_API_APIKeyUnset(t *testing.T) {
	v := viper.New()
	v.Set("key", "")

	_, err := paramscmd.LoadAPIParams(v)
	require.Error(t, err)

	var errauth api.ErrAuth

	assert.ErrorAs(t, err, &errauth)

	assert.EqualError(t, errauth, "api key not found or empty")
}

func TestLoad_API_APIKeyInvalid(t *testing.T) {
	tests := map[string]string{
		"invalid format 1": "not-uuid",
		"invalid format 2": "00000000-0000-0000-8000-000000000000",
		"invalid format 3": "00000000-0000-4000-0000-000000000000",
	}

	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", value)

			_, err := paramscmd.LoadAPIParams(v)
			require.Error(t, err)

			var errauth api.ErrAuth
			assert.ErrorAs(t, err, &errauth)

			assert.EqualError(t, errauth, "invalid api key format")
		})
	}
}

func TestLoadParams_ApiKey_SettingTakePrecedence(t *testing.T) {
	v := viper.New()
	v.Set("config", "testdata/.wakatime.cfg")
	v.Set("entity", "testdata/heartbeat_go.json")

	configFile, err := inipkg.FilePath(v)
	require.NoError(t, err)

	err = inipkg.ReadInConfig(v, configFile)
	require.NoError(t, err)

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-4000-8000-000000000000", params.Key)
}

func TestLoadParams_ApiKey_FromVault(t *testing.T) {
	v := viper.New()
	v.Set("config", "testdata/.wakatime-vault.cfg")
	v.Set("entity", "testdata/heartbeat_go.json")

	configFile, err := inipkg.FilePath(v)
	require.NoError(t, err)

	err = inipkg.ReadInConfig(v, configFile)
	require.NoError(t, err)

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-4000-8000-000000000000", params.Key)
}

func TestLoadParams_ApiKey_FromVault_Err_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping because OS is not darwin.")
	}

	v := viper.New()
	v.Set("config", "testdata/.wakatime-vault-error.cfg")
	v.Set("entity", "testdata/heartbeat_go.json")

	configFile, err := inipkg.FilePath(v)
	require.NoError(t, err)

	err = inipkg.ReadInConfig(v, configFile)
	require.NoError(t, err)

	_, err = paramscmd.LoadAPIParams(v)

	assert.EqualError(t, err, "failed to read api key from vault: exit status 1")
}

func TestLoad_API_APIUrl(t *testing.T) {
	tests := map[string]struct {
		ViperAPIUrl       string
		ViperAPIUrlConfig string
		ViperAPIUrlOld    string
		Expected          paramscmd.API
	}{
		"api url flag takes precedence": {
			ViperAPIUrl:       "http://localhost:8080",
			ViperAPIUrlConfig: "http://localhost:8081",
			ViperAPIUrlOld:    "http://localhost:8082",
			Expected: paramscmd.API{
				Key:      "00000000-0000-4000-8000-000000000000",
				URL:      "http://localhost:8080",
				Hostname: "my-computer",
			},
		},
		"api url deprecated flag takes precedence": {
			ViperAPIUrlConfig: "http://localhost:8081",
			ViperAPIUrlOld:    "http://localhost:8082",
			Expected: paramscmd.API{
				Key:      "00000000-0000-4000-8000-000000000000",
				URL:      "http://localhost:8082",
				Hostname: "my-computer",
			},
		},
		"api url from config": {
			ViperAPIUrlConfig: "http://localhost:8081",
			Expected: paramscmd.API{
				Key:      "00000000-0000-4000-8000-000000000000",
				URL:      "http://localhost:8081",
				Hostname: "my-computer",
			},
		},
		"api url with legacy heartbeats endpoint": {
			ViperAPIUrl: "http://localhost:8080/api/v1/heartbeats.bulk",
			Expected: paramscmd.API{
				Key:      "00000000-0000-4000-8000-000000000000",
				URL:      "http://localhost:8080/api/v1",
				Hostname: "my-computer",
			},
		},
		"api url with trailing slash": {
			ViperAPIUrl: "http://localhost:8080/api/",
			Expected: paramscmd.API{
				Key:      "00000000-0000-4000-8000-000000000000",
				URL:      "http://localhost:8080/api",
				Hostname: "my-computer",
			},
		},
		"api url with wakapi style endpoint": {
			ViperAPIUrl: "http://localhost:8080/api/heartbeat",
			Expected: paramscmd.API{
				Key:      "00000000-0000-4000-8000-000000000000",
				URL:      "http://localhost:8080/api",
				Hostname: "my-computer",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("hostname", "my-computer")
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("api-url", test.ViperAPIUrl)
			v.Set("apiurl", test.ViperAPIUrlOld)
			v.Set("settings.api_url", test.ViperAPIUrlConfig)

			params, err := paramscmd.LoadAPIParams(v)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params)
		})
	}
}

func TestLoad_APIUrl_Default(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, api.BaseURL, params.URL)
}

func TestLoad_APIUrl_InvalidFormat(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", "http://in valid")

	_, err := paramscmd.LoadAPIParams(v)

	var errauth api.ErrAuth

	assert.ErrorAs(t, err, &errauth)

	assert.EqualError(t, errauth, `invalid api url: parse "http://in valid": invalid character " " in host name`)
}

func TestLoad_API_BackoffAt(t *testing.T) {
	v := viper.New()
	v.Set("hostname", "my-computer")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("internal.backoff_at", "2021-08-30T18:50:42-03:00")
	v.Set("internal.backoff_retries", "3")

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	backoffAt, err := time.Parse(inipkg.DateFormat, "2021-08-30T18:50:42-03:00")
	require.NoError(t, err)

	assert.Equal(t, paramscmd.API{
		BackoffAt:      backoffAt,
		BackoffRetries: 3,
		Key:            "00000000-0000-4000-8000-000000000000",
		URL:            "https://api.wakatime.com/api/v1",
		Hostname:       "my-computer",
	}, params)
}

func TestLoad_API_BackoffAtErr(t *testing.T) {
	v := viper.New()
	v.Set("hostname", "my-computer")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("internal.backoff_at", "2021-08-30")
	v.Set("internal.backoff_retries", "2")

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.API{
		BackoffAt:      time.Time{},
		BackoffRetries: 2,
		Key:            "00000000-0000-4000-8000-000000000000",
		URL:            "https://api.wakatime.com/api/v1",
		Hostname:       "my-computer",
	}, params)
}

func TestLoad_API_Plugin(t *testing.T) {
	v := viper.New()
	v.Set("hostname", "my-computer")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("plugin", "plugin/10.0.0")

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, paramscmd.API{
		Key:      "00000000-0000-4000-8000-000000000000",
		URL:      "https://api.wakatime.com/api/v1",
		Plugin:   "plugin/10.0.0",
		Hostname: "my-computer",
	}, params)
}

func TestLoad_API_Timeout_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("timeout", 5)
	v.Set("settings.timeout", 10)

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, 5*time.Second, params.Timeout)
}

func TestLoad_API_Timeout_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.timeout", 10)

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, 10*time.Second, params.Timeout)
}

func TestLoad_API_DisableSSLVerify_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("no-ssl-verify", true)
	v.Set("settings.no_ssl_verify", false)

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.True(t, params.DisableSSLVerify)
}

func TestLoad_API_DisableSSLVerify_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.no_ssl_verify", true)

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.True(t, params.DisableSSLVerify)
}

func TestLoad_API_DisableSSLVerify_Default(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.False(t, params.DisableSSLVerify)
}

func TestLoad_API_ProxyURL(t *testing.T) {
	tests := map[string]string{
		"https":  "https://john:secret@example.org:8888",
		"http":   "http://john:secret@example.org:8888",
		"ipv6":   "socks5://john:secret@2001:0db8:85a3:0000:0000:8a2e:0370:7334:8888",
		"ntlm":   `domain\\john:123456`,
		"socks5": "socks5://john:secret@example.org:8888",
	}

	for name, proxyURL := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("proxy", proxyURL)

			params, err := paramscmd.LoadAPIParams(v)
			require.NoError(t, err)

			assert.Equal(t, proxyURL, params.ProxyURL)
		})
	}
}

func TestLoad_API_ProxyURL_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("proxy", "https://john:secret@example.org:8888")
	v.Set("settings.proxy", "ignored")

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, "https://john:secret@example.org:8888", params.ProxyURL)
}

func TestLoad_API_ProxyURL_UserDefinedTakesPrecedenceOverEnvironment(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("proxy", "https://john:secret@example.org:8888")

	err := os.Setenv("HTTPS_PROXY", "https://papa:secret@company.org:9000")
	require.NoError(t, err)

	defer os.Unsetenv("HTTPS_PROXY")

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, "https://john:secret@example.org:8888", params.ProxyURL)
}

func TestLoad_API_ProxyURL_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.proxy", "https://john:secret@example.org:8888")

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, "https://john:secret@example.org:8888", params.ProxyURL)
}

func TestLoad_API_ProxyURL_FromEnvironment(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	err := os.Setenv("HTTPS_PROXY", "https://john:secret@example.org:8888")
	require.NoError(t, err)

	defer os.Unsetenv("HTTPS_PROXY")

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, "https://john:secret@example.org:8888", params.ProxyURL)
}

func TestLoad_API_ProxyURL_NoProxyFromEnvironment(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	err := os.Setenv("NO_PROXY", "https://some.org,https://api.wakatime.com")
	require.NoError(t, err)

	defer os.Unsetenv("NO_PROXY")

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Empty(t, params.ProxyURL)
}

func TestLoad_API_ProxyURL_InvalidFormat(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("proxy", "ftp://john:secret@example.org:8888")

	_, err := paramscmd.LoadAPIParams(v)

	var errauth api.ErrAuth

	assert.ErrorAs(t, err, &errauth)

	assert.EqualError(
		t,
		err,
		"invalid url \"ftp://john:secret@example.org:8888\". Must be in format'https://user:pass@host:port' or"+
			" 'socks5://user:pass@host:port' or 'domain\\\\user:pass.'")
}

func TestLoad_API_SSLCertFilepath_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("ssl-certs-file", "~/path/to/cert.pem")

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(home, "/path/to/cert.pem"), params.SSLCertFilepath)
}

func TestLoad_API_SSLCertFilepath_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.ssl_certs_file", "/path/to/cert.pem")

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, "/path/to/cert.pem", params.SSLCertFilepath)
}

func TestLoadParams_Hostname_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("hostname", "my-machine")
	v.Set("settings.hostname", "ignored")

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, "my-machine", params.Hostname)
}

func TestLoadParams_Hostname_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.hostname", "my-machine")

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	assert.Equal(t, "my-machine", params.Hostname)
}

func TestLoadParams_Hostname_DefaultFromSystem(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	params, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)

	expected, err := os.Hostname()
	require.NoError(t, err)

	assert.Equal(t, expected, params.Hostname)
}

func TestLoadParams_StatusBar_HideCategories_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("today-hide-categories", true)
	v.Set("settings.status_bar_hide_categories", "ignored")

	params, err := paramscmd.LoadStatusBarParams(v)
	require.NoError(t, err)

	assert.True(t, params.HideCategories)
}

func TestLoadParams_StatusBar_HideCategories_ConfigTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("settings.status_bar_hide_categories", true)

	params, err := paramscmd.LoadStatusBarParams(v)
	require.NoError(t, err)

	assert.True(t, params.HideCategories)
}

func TestLoadParams_StatusBar_Output(t *testing.T) {
	tests := map[string]output.Output{
		"text": output.TextOutput,
		"json": output.JSONOutput,
	}

	for name, out := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("output", name)

			params, err := paramscmd.LoadStatusBarParams(v)
			require.NoError(t, err)

			assert.Equal(t, out, params.Output)
		})
	}
}

func TestLoadParams_StatusBar_Output_Default(t *testing.T) {
	v := viper.New()

	params, err := paramscmd.LoadStatusBarParams(v)
	require.NoError(t, err)

	assert.Equal(t, output.TextOutput, params.Output)
}

func TestLoadParams_StatusBar_Output_Invalid(t *testing.T) {
	v := viper.New()
	v.Set("output", "invalid")

	_, err := paramscmd.LoadStatusBarParams(v)
	require.Error(t, err)

	assert.Equal(t, "failed to parse output: invalid output \"invalid\"", err.Error())
}

func TestAPI_String(t *testing.T) {
	backoffat, err := time.Parse(inipkg.DateFormat, "2021-08-30T18:50:42-03:00")
	require.NoError(t, err)

	api := paramscmd.API{
		BackoffAt:        backoffat,
		BackoffRetries:   5,
		DisableSSLVerify: true,
		Hostname:         "my-machine",
		Key:              "00000000-0000-4000-8000-000000000000",
		KeyPatterns: []apikey.MapPattern{
			{
				APIKey: "00000000-0000-4000-8000-000000000001",
				Regex:  regex.MustCompile("^/api/v1/"),
			},
		},
		Plugin:          "my-plugin",
		ProxyURL:        "https://example.org:23",
		SSLCertFilepath: "/path/to/cert.pem",
		Timeout:         time.Second * 10,
		URL:             "https://example.org:23",
	}

	assert.Equal(
		t,
		"api key: '<hidden>0000', api url: 'https://example.org:23', backoff at: '2021-08-30T18:50:42-03:00',"+
			" backoff retries: 5, hostname: 'my-machine', key patterns: '[{<hidden>0001 ^/api/v1/}]', plugin: 'my-plugin',"+
			" proxy url: 'https://example.org:23', timeout: 10s, disable ssl verify: true,"+
			" ssl cert filepath: '/path/to/cert.pem'",
		api.String(),
	)
}

func TestFilterParams_String(t *testing.T) {
	filterparams := paramscmd.FilterParams{
		Exclude:                    []regex.Regex{regex.MustCompile("^/exclude")},
		ExcludeUnknownProject:      true,
		Include:                    []regex.Regex{regex.MustCompile("^/include")},
		IncludeOnlyWithProjectFile: true,
	}

	assert.Equal(
		t,
		"exclude: '[^/exclude]', exclude unknown project: true, include: '[^/include]',"+
			" include only with project file: true",
		filterparams.String(),
	)
}

func TestHeartbeat_String(t *testing.T) {
	heartbeat := paramscmd.Heartbeat{
		Category:        heartbeat.CodingCategory,
		CursorPosition:  heartbeat.PointerTo(15),
		Entity:          "path/to/entity.go",
		EntityType:      heartbeat.FileType,
		ExtraHeartbeats: make([]heartbeat.Heartbeat, 3),
		IsUnsavedEntity: true,
		IsWrite:         heartbeat.PointerTo(true),
		Language:        heartbeat.PointerTo("Golang"),
		LineNumber:      heartbeat.PointerTo(4),
		LinesInFile:     heartbeat.PointerTo(56),
		Time:            1585598059,
	}

	assert.Equal(
		t,
		"category: 'coding', cursor position: '15', entity: 'path/to/entity.go', entity type: 'file',"+
			" num extra heartbeats: 3, is unsaved entity: true, is write: true, language: 'Golang', line number: '4',"+
			" lines in file: '56', time: 1585598059.00000, filter params: (exclude: '[]', exclude unknown project: false,"+
			" include: '[]', include only with project file: false), project params: (alternate: '', branch alternate: '',"+
			" map patterns: '[]', override: '', git submodules disabled: '[]', git submodule project map: '[]'),"+
			" sanitize params: (hide branch names: '[]', hide project folder: false, hide file names: '[]',"+
			" hide project names: '[]', project path override: '')",
		heartbeat.String(),
	)
}

func TestOffline_String(t *testing.T) {
	offline := paramscmd.Offline{
		Disabled:  true,
		PrintMax:  6,
		QueueFile: "/path/to/queue.file",
		SyncMax:   12,
	}

	assert.Equal(
		t,
		"disabled: true, print max: 6, queue file: '/path/to/queue.file', num sync max: 12",
		offline.String(),
	)
}

func TestProjectParams_String(t *testing.T) {
	projectparams := paramscmd.ProjectParams{
		Alternate:            "alternate",
		BranchAlternate:      "branch-alternate",
		MapPatterns:          []project.MapPattern{{Name: "project-1", Regex: regex.MustCompile("^/regex")}},
		Override:             "override",
		SubmodulesDisabled:   []regex.Regex{regexp.MustCompile(".*")},
		SubmoduleMapPatterns: []project.MapPattern{{Name: "awesome-project", Regex: regex.MustCompile("^/regex")}},
	}

	assert.Equal(
		t,
		"alternate: 'alternate', branch alternate: 'branch-alternate',"+
			" map patterns: '[{project-1 ^/regex}]', override: 'override',"+
			" git submodules disabled: '[.*]', git submodule project map: '[{awesome-project ^/regex}]'",
		projectparams.String(),
	)
}

func TestLoadParams_ProjectFromGitRemote(t *testing.T) {
	v := viper.New()
	v.Set("git.project_from_git_remote", true)
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.LoadHeartbeatParams(v)
	require.NoError(t, err)

	assert.True(t, params.Project.ProjectFromGitRemote)
}

func TestSanitizeParams_String(t *testing.T) {
	sanitizeparams := paramscmd.SanitizeParams{
		HideBranchNames:     []regex.Regex{regex.MustCompile("^/hide")},
		HideProjectFolder:   true,
		HideFileNames:       []regex.Regex{regex.MustCompile("^/hide")},
		HideProjectNames:    []regex.Regex{regex.MustCompile("^/hide")},
		ProjectPathOverride: "path/to/project",
	}

	assert.Equal(
		t,
		"hide branch names: '[^/hide]', hide project folder: true, hide file names: '[^/hide]',"+
			" hide project names: '[^/hide]', project path override: 'path/to/project'",
		sanitizeparams.String(),
	)
}

func TestStatusBar_String(t *testing.T) {
	statusbar := paramscmd.StatusBar{
		HideCategories: true,
		Output:         output.JSONOutput,
	}

	assert.Equal(
		t,
		"hide categories: true, output: 'json'",
		statusbar.String(),
	)
}

func TestLoadParams_APIKeyPrefixSupported(t *testing.T) {
	v := viper.New()
	v.Set("key", "waka_00000000-0000-4000-8000-000000000000")

	_, err := paramscmd.LoadAPIParams(v)
	require.NoError(t, err)
}
