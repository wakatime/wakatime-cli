package params_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	paramscmd "github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	inipkg "github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/regex"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/ini.v1"
)

func TestLoadParams_AlternateProject(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("alternate-project", "web")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, "web", params.Heartbeat.Project.Alternate)
}

func TestLoadParams_AlternateProject_Unset(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Empty(t, params.Heartbeat.Project.Alternate)
}

func TestLoadParams_Category(t *testing.T) {
	tests := map[string]heartbeat.Category{
		"coding":         heartbeat.CodingCategory,
		"browsing":       heartbeat.BrowsingCategory,
		"building":       heartbeat.BuildingCategory,
		"code reviewing": heartbeat.CodeReviewingCategory,
		"debugging":      heartbeat.DebuggingCategory,
		"designing":      heartbeat.DesigningCategory,
		"indexing":       heartbeat.IndexingCategory,
		"manual testing": heartbeat.ManualTestingCategory,
		"running tests":  heartbeat.RunningTestsCategory,
		"writing tests":  heartbeat.WritingTestsCategory,
	}

	for name, category := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("category", name)

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			assert.Equal(t, category, params.Heartbeat.Category)
		})
	}
}

func TestLoadParams_Category_Default(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.Equal(t, heartbeat.CodingCategory, params.Heartbeat.Category)
}

func TestLoadParams_Category_Invalid(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("category", "invalid")

	_, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.Error(t, err)

	assert.Equal(t, "failed to load heartbeat params: failed to parse category: invalid category \"invalid\"", err.Error())
}

func TestLoadParams_CursorPosition(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("cursorpos", 42)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, 42, *params.Heartbeat.CursorPosition)
}

func TestLoadParams_CursorPosition_Zero(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("cursorpos", 0)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, 0, *params.Heartbeat.CursorPosition)
}

func TestLoadParams_CursorPosition_Unset(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Nil(t, params.Heartbeat.CursorPosition)
}

func TestLoadParams_Entity_EntityFlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("file", "ignored")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, "/path/to/file", params.Heartbeat.Entity)
}

func TestLoadParams_Entity_FileFlag(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("file", "~/path/to/file")

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(home, "/path/to/file"), params.Heartbeat.Entity)
}

func TestLoadParams_Entity_Unset(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	_, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.Error(t, err)

	assert.Equal(t, "failed to load heartbeat params: failed to retrieve entity", err.Error())
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
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("entity-type", name)

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			assert.Equal(t, entityType, params.Heartbeat.EntityType)
		})
	}
}

func TestLoadParams_EntityType_Default(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, heartbeat.FileType, params.Heartbeat.EntityType)
}

func TestLoadParams_EntityType_Invalid(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("entity-type", "invalid")

	_, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.Error(t, err)

	assert.Equal(
		t,
		"failed to load heartbeat params: failed to parse entity type: invalid entity type \"invalid\"",
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
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("extra-heartbeats", true)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Len(t, params.Heartbeat.ExtraHeartbeats, 2)

	assert.NotNil(t, params.Heartbeat.ExtraHeartbeats[0].Language)
	assert.Equal(t, heartbeat.LanguageGo.String(), *params.Heartbeat.ExtraHeartbeats[0].Language)
	assert.NotNil(t, params.Heartbeat.ExtraHeartbeats[1].Language)
	assert.Equal(t, heartbeat.LanguagePython.String(), *params.Heartbeat.ExtraHeartbeats[1].Language)

	assert.Equal(t, []heartbeat.Heartbeat{
		{
			Category:          heartbeat.CodingCategory,
			CursorPosition:    heartbeat.Int(12),
			Entity:            "testdata/main.go",
			EntityType:        heartbeat.FileType,
			IsWrite:           heartbeat.Bool(true),
			LanguageAlternate: "Golang",
			LineNumber:        heartbeat.Int(42),
			Lines:             heartbeat.Int(45),
			ProjectAlternate:  "billing",
			ProjectOverride:   "wakatime-cli",
			Time:              1585598059,
			// tested above
			Language: params.Heartbeat.ExtraHeartbeats[0].Language,
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
			Language: params.Heartbeat.ExtraHeartbeats[1].Language,
		},
	}, params.Heartbeat.ExtraHeartbeats)
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
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("extra-heartbeats", true)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Len(t, params.Heartbeat.ExtraHeartbeats, 2)

	assert.NotNil(t, params.Heartbeat.ExtraHeartbeats[0].Language)
	assert.Equal(t, heartbeat.LanguageGo.String(), *params.Heartbeat.ExtraHeartbeats[0].Language)
	assert.NotNil(t, params.Heartbeat.ExtraHeartbeats[1].Language)
	assert.Equal(t, heartbeat.LanguagePython.String(), *params.Heartbeat.ExtraHeartbeats[1].Language)

	assert.Equal(t, []heartbeat.Heartbeat{
		{
			Category:       heartbeat.CodingCategory,
			CursorPosition: heartbeat.Int(12),
			Entity:         "testdata/main.go",
			EntityType:     heartbeat.FileType,
			IsWrite:        heartbeat.Bool(true),
			Language:       params.Heartbeat.ExtraHeartbeats[0].Language,
			Lines:          heartbeat.Int(45),
			LineNumber:     heartbeat.Int(42),
			Time:           1585598059,
		},
		{
			Category:       heartbeat.CodingCategory,
			CursorPosition: heartbeat.Int(13),
			Entity:         "testdata/main.go",
			EntityType:     heartbeat.FileType,
			IsWrite:        heartbeat.Bool(true),
			Language:       params.Heartbeat.ExtraHeartbeats[1].Language,
			LineNumber:     heartbeat.Int(43),
			Lines:          heartbeat.Int(46),
			Time:           1585598060,
		},
	}, params.Heartbeat.ExtraHeartbeats)
}

func TestLoadParams_IsWrite(t *testing.T) {
	tests := map[string]bool{
		"is write":    true,
		"is no write": false,
	}

	for name, isWrite := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("write", isWrite)

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			assert.Equal(t, isWrite, *params.Heartbeat.IsWrite)
		})
	}
}

func TestLoadParams_IsWrite_Unset(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Nil(t, params.Heartbeat.IsWrite)
}

func TestLoadParams_Language(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("language", "Go")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.NotNil(t, params.Heartbeat.Language)
	assert.Equal(t, heartbeat.LanguageGo.String(), *params.Heartbeat.Language)
}

func TestLoadParams_LanguageAlternate(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("alternate-language", "Go")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageGo.String(), params.Heartbeat.LanguageAlternate)
	assert.Nil(t, params.Heartbeat.Language)
}

func TestLoadParams_LineNumber(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("lineno", 42)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, 42, *params.Heartbeat.LineNumber)
}

func TestLoadParams_LineNumber_Zero(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("lineno", 0)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, 0, *params.Heartbeat.LineNumber)
}

func TestLoadParams_LineNumber_Unset(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Nil(t, params.Heartbeat.LineNumber)
}

func TestLoadParams_LocalFile(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("local-file", "/path/to/file")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, "/path/to/file", params.Heartbeat.LocalFile)
}

func TestLoadParams_Plugin(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("plugin", "plugin/10.0.0")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, "plugin/10.0.0", params.API.Plugin)
}

func TestLoadParams_Plugin_Unset(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Empty(t, params.API.Plugin)
}

func TestLoadParams_Project(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("project", "billing")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, "billing", params.Heartbeat.Project.Override)
}

func TestLoadParams_Project_Unset(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Empty(t, params.Heartbeat.Project.Override)
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
					Regex: regexp.MustCompile("projects/foo"),
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
					Regex: regexp.MustCompile(`^/home/user/projects/bar(\\d+)/`),
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", test.Entity)
			v.Set(fmt.Sprintf("projectmap.%s", test.Regex.String()), test.Project)

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params.Heartbeat.Project.MapPatterns)
		})
	}
}

func TestLoadParams_Timeout_FlagTakesPreceedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("timeout", 5)
	v.Set("settings.timeout", 10)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, 5*time.Second, params.API.Timeout)
}

func TestLoadParams_Timeout_FromConfig(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.timeout", 10)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, 10*time.Second, params.API.Timeout)
}

func TestLoadParams_Time(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("time", 1590609206.1)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, 1590609206.1, params.Heartbeat.Time)
}

func TestLoadParams_Time_Default(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	now := float64(time.Now().UnixNano()) / 1000000000
	assert.GreaterOrEqual(t, now, params.Heartbeat.Time)
	assert.GreaterOrEqual(t, params.Heartbeat.Time, now-60)
}

func TestLoadParams_Filter_Exclude(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("exclude", []string{".*", "wakatime.*"})
	v.Set("settings.exclude", []string{".+", "wakatime.+"})
	v.Set("settings.ignore", []string{".?", "wakatime.?"})

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	require.Len(t, params.Heartbeat.Filter.Exclude, 6)
	assert.Equal(t, ".*", params.Heartbeat.Filter.Exclude[0].String())
	assert.Equal(t, "wakatime.*", params.Heartbeat.Filter.Exclude[1].String())
	assert.Equal(t, ".+", params.Heartbeat.Filter.Exclude[2].String())
	assert.Equal(t, "wakatime.+", params.Heartbeat.Filter.Exclude[3].String())
	assert.Equal(t, ".?", params.Heartbeat.Filter.Exclude[4].String())
	assert.Equal(t, "wakatime.?", params.Heartbeat.Filter.Exclude[5].String())
}

func TestLoadParams_Filter_Exclude_Multiline(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.ignore", "\t.?\n\twakatime.? \t\n")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	require.Len(t, params.Heartbeat.Filter.Exclude, 2)
	assert.Equal(t, ".?", params.Heartbeat.Filter.Exclude[0].String())
	assert.Equal(t, "wakatime.?", params.Heartbeat.Filter.Exclude[1].String())
}

func TestLoadParams_Filter_Exclude_IgnoresInvalidRegex(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("exclude", []string{".*", "["})

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	require.Len(t, params.Heartbeat.Filter.Exclude, 1)
	assert.Equal(t, ".*", params.Heartbeat.Filter.Exclude[0].String())
}

func TestLoadParams_Filter_Exclude_PerlRegexPatterns(t *testing.T) {
	tests := map[string]string{
		"negative lookahead": `^/var/(?!www/).*`,
		"positive lookahead": `^/var/(?=www/).*`,
	}

	for name, pattern := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("exclude", []string{pattern})

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			require.Len(t, params.Heartbeat.Filter.Exclude, 1)
			assert.Equal(t, pattern, params.Heartbeat.Filter.Exclude[0].String())
		})
	}
}

func TestLoadParams_Filter_ExcludeUnknownProject(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("exclude-unknown-project", true)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, true, params.Heartbeat.Filter.ExcludeUnknownProject)
}

func TestLoadParams_Filter_ExcludeUnknownProject_FromConfig(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("exclude-unknown-project", false)
	v.Set("settings.exclude_unknown_project", true)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, true, params.Heartbeat.Filter.ExcludeUnknownProject)
}

func TestLoadParams_Filter_Include(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("include", []string{".*", "wakatime.*"})
	v.Set("settings.include", []string{".+", "wakatime.+"})

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	require.Len(t, params.Heartbeat.Filter.Include, 4)
	assert.Equal(t, ".*", params.Heartbeat.Filter.Include[0].String())
	assert.Equal(t, "wakatime.*", params.Heartbeat.Filter.Include[1].String())
	assert.Equal(t, ".+", params.Heartbeat.Filter.Include[2].String())
	assert.Equal(t, "wakatime.+", params.Heartbeat.Filter.Include[3].String())
}

func TestLoadParams_Filter_Include_IgnoresInvalidRegex(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("include", []string{".*", "["})

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	require.Len(t, params.Heartbeat.Filter.Include, 1)
	assert.Equal(t, ".*", params.Heartbeat.Filter.Include[0].String())
}

func TestLoadParams_Filter_Include_PerlRegexPatterns(t *testing.T) {
	tests := map[string]string{
		"negative lookahead": `^/var/(?!www/).*`,
		"positive lookahead": `^/var/(?=www/).*`,
	}

	for name, pattern := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("include", []string{pattern})

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			require.Len(t, params.Heartbeat.Filter.Include, 1)
			assert.Equal(t, pattern, params.Heartbeat.Filter.Include[0].String())
		})
	}
}

func TestLoadParams_Filter_IncludeOnlyWithProjectFile(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("include-only-with-project-file", true)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, true, params.Heartbeat.Filter.IncludeOnlyWithProjectFile)
}

func TestLoadParams_Filter_IncludeOnlyWithProjectFile_FromConfig(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("include-only-with-project-file", false)
	v.Set("settings.include_only_with_project_file", true)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, true, params.Heartbeat.Filter.IncludeOnlyWithProjectFile)
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
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-branch-names", viperValue)

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{
				HideBranchNames: []regex.Regex{regex.MustCompile(".*")},
			}, params.Heartbeat.Sanitize)
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
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-branch-names", viperValue)

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{}, params.Heartbeat.Sanitize)
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
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-branch-names", test.ViperValue)

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{
				HideBranchNames: test.Expected,
			}, params.Heartbeat.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideBranchNames_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hide-branch-names", "true")
	v.Set("settings.hide_branch_names", "ignored")
	v.Set("settings.hide_branchnames", "ignored")
	v.Set("settings.hidebranchnames", "ignored")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideBranchNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Heartbeat.Sanitize)
}

func TestLoadParams_SanitizeParams_HideBranchNames_ConfigTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_branch_names", "true")
	v.Set("settings.hide_branchnames", "ignored")
	v.Set("settings.hidebranchnames", "ignored")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideBranchNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Heartbeat.Sanitize)
}

func TestLoadParams_SanitizeParams_HideBranchNames_ConfigDeprecatedOneTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_branchnames", "true")
	v.Set("settings.hidebranchnames", "ignored")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideBranchNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Heartbeat.Sanitize)
}

func TestLoadParams_SanitizeParams_HideBranchNames_ConfigDeprecatedTwo(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hidebranchnames", "true")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideBranchNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Heartbeat.Sanitize)
}

func TestLoadParams_SanitizeParams_HideBranchNames_InvalidRegex(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hide-branch-names", ".*secret.*\n[0-9+")

	_, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.Error(t, err)

	assert.True(t, strings.HasPrefix(
		err.Error(),
		"failed to load heartbeat params: failed to load sanitize params:"+
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
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-project-names", viperValue)

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{
				HideProjectNames: []regex.Regex{regexp.MustCompile(".*")},
			}, params.Heartbeat.Sanitize)
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
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-project-names", viperValue)

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{}, params.Heartbeat.Sanitize)
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
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-project-names", test.ViperValue)

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{
				HideProjectNames: test.Expected,
			}, params.Heartbeat.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideProjectNames_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hide-project-names", "true")
	v.Set("settings.hide_project_names", "ignored")
	v.Set("settings.hide_projectnames", "ignored")
	v.Set("settings.hideprojectnames", "ignored")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideProjectNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Heartbeat.Sanitize)
}

func TestLoadParams_SanitizeParams_HideProjectNames_ConfigTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_project_names", "true")
	v.Set("settings.hide_projectnames", "ignored")
	v.Set("settings.hideprojectnames", "ignored")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideProjectNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Heartbeat.Sanitize)
}

func TestLoadParams_SanitizeParams_HideProjectNames_ConfigDeprecatedOneTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_projectnames", "true")
	v.Set("settings.hideprojectnames", "ignored")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideProjectNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Heartbeat.Sanitize)
}

func TestLoadParams_SanitizeParams_HideProjectNames_ConfigDeprecatedTwo(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hideprojectnames", "true")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideProjectNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Heartbeat.Sanitize)
}

func TestLoadParams_SanitizeParams_HideProjectNames_InvalidRegex(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hide-project-names", ".*secret.*\n[0-9+")

	_, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.Error(t, err)

	assert.True(t, strings.HasPrefix(
		err.Error(),
		"failed to load heartbeat params: failed to load sanitize params:"+
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
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-file-names", viperValue)

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{
				HideFileNames: []regex.Regex{regexp.MustCompile(".*")},
			}, params.Heartbeat.Sanitize)
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
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-file-names", viperValue)

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{}, params.Heartbeat.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideFilehNames_List(t *testing.T) {
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
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("hide-file-names", test.ViperValue)

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			assert.Equal(t, paramscmd.SanitizeParams{
				HideFileNames: test.Expected,
			}, params.Heartbeat.Sanitize)
		})
	}
}

func TestLoadParams_SanitizeParams_HideFileNames_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hide-file-names", "true")
	v.Set("hide-filenames", "ignored")
	v.Set("hidefilenames", "ignored")
	v.Set("settings.hide_file_names", "ignored")
	v.Set("settings.hide_filenames", "ignored")
	v.Set("settings.hidefilenames", "ignored")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideFileNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Heartbeat.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_FlagDeprecatedOneTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hide-filenames", "true")
	v.Set("hidefilenames", "ignored")
	v.Set("settings.hide_file_names", "ignored")
	v.Set("settings.hide_filenames", "ignored")
	v.Set("settings.hidefilenames", "ignored")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideFileNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Heartbeat.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_FlagDeprecatedTwoTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hidefilenames", "true")
	v.Set("settings.hide_file_names", "ignored")
	v.Set("settings.hide_filenames", "ignored")
	v.Set("settings.hidefilenames", "ignored")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideFileNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Heartbeat.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_ConfigTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_file_names", "true")
	v.Set("settings.hide_filenames", "ignored")
	v.Set("settings.hidefilenames", "ignored")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideFileNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Heartbeat.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_ConfigDeprecatedOneTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_filenames", "true")
	v.Set("settings.hidefilenames", "ignored")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideFileNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Heartbeat.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_ConfigDeprecatedTwo(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hidefilenames", "true")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.NoError(t, err)

	assert.Equal(t, paramscmd.SanitizeParams{
		HideFileNames: []regex.Regex{regexp.MustCompile(".*")},
	}, params.Heartbeat.Sanitize)
}

func TestLoadParams_SanitizeParams_HideFileNames_InvalidRegex(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hide-file-names", ".*secret.*\n[0-9+")

	_, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
	require.Error(t, err)

	assert.True(t, strings.HasPrefix(
		err.Error(),
		"failed to load heartbeat params: failed to load sanitize params:"+
			" failed to parse regex hide file names param \".*secret.*\\n[0-9+\":"+
			" failed to compile regex \"[0-9+\":",
	))
}

func TestLoadParams_DisableSubmodule_True(t *testing.T) {
	tests := map[string]string{
		"lowercase":       "true",
		"uppercase":       "TRUE",
		"first uppercase": "True",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("git.submodules_disabled", viperValue)

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			assert.Equal(t, []regex.Regex{regexp.MustCompile(".*")}, params.Heartbeat.Project.DisableSubmodule)
		})
	}
}

func TestLoadParams_DisableSubmodule_False(t *testing.T) {
	tests := map[string]string{
		"lowercase":       "false",
		"uppercase":       "FALSE",
		"first uppercase": "False",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("git.submodules_disabled", viperValue)

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			assert.Equal(t, []regex.Regex(nil), params.Heartbeat.Project.DisableSubmodule)
		})
	}
}

func TestLoadParams_DisableSubmodule_List(t *testing.T) {
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
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("git.submodules_disabled", test.ViperValue)

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true, HeartbeatRequired: true})
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params.Heartbeat.Project.DisableSubmodule)
		})
	}
}

func TestLoad_OfflineDisabled_ConfigTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("disable-offline", false)
	v.Set("disableoffline", false)
	v.Set("settings.offline", false)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.True(t, params.Offline.Disabled)
}

func TestLoad_OfflineDisabled_FlagDeprecatedTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("disable-offline", false)
	v.Set("disableoffline", true)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.True(t, params.Offline.Disabled)
}

func TestLoad_OfflineDisabled_FromFlag(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("disable-offline", true)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.True(t, params.Offline.Disabled)
}

func TestLoad_OfflineQueueFile(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("offline-queue-file", "/path/to/file")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.Equal(t, "/path/to/file", params.Offline.QueueFile)
}

func TestLoad_OfflineSyncMax(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("sync-offline-activity", 42)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.Equal(t, 42, params.Offline.SyncMax)
}

func TestLoad_OfflineSyncMax_NoEntity(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("sync-offline-activity", 42)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.Equal(t, 42, params.Offline.SyncMax)
}

func TestLoad_OfflineSyncMax_None(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("sync-offline-activity", "none")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.Equal(t, 0, params.Offline.SyncMax)
}

func TestLoad_OfflineSyncMax_Default(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.Equal(t, 1000, params.Offline.SyncMax)
}

func TestLoad_OfflineSyncMax_NegativeNumber(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("sync-offline-activity", -1)

	_, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.Error(t, err)

	assert.Contains(t, err.Error(), "--sync-offline-activity")
}

func TestLoad_OfflineSyncMax_NonIntegerValue(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("sync-offline-activity", "invalid")

	_, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.Error(t, err)

	assert.Contains(t, err.Error(), "--sync-offline-activity")
}

func TestLoad_API_APIKey(t *testing.T) {
	tests := map[string]struct {
		ViperAPIKey          string
		ViperAPIKeyConfig    string
		ViperAPIKeyConfigOld string
		Expected             paramscmd.Params
	}{
		"api key flag takes preceedence": {
			ViperAPIKey:          "00000000-0000-4000-8000-000000000000",
			ViperAPIKeyConfig:    "10000000-0000-4000-8000-000000000000",
			ViperAPIKeyConfigOld: "20000000-0000-4000-8000-000000000000",
			Expected: paramscmd.Params{
				API: paramscmd.API{

					Key:      "00000000-0000-4000-8000-000000000000",
					URL:      "https://api.wakatime.com/api/v1",
					Hostname: "my-computer",
				},
			},
		},
		"api from config takes preceedence": {
			ViperAPIKeyConfig:    "00000000-0000-4000-8000-000000000000",
			ViperAPIKeyConfigOld: "10000000-0000-4000-8000-000000000000",
			Expected: paramscmd.Params{
				API: paramscmd.API{

					Key:      "00000000-0000-4000-8000-000000000000",
					URL:      "https://api.wakatime.com/api/v1",
					Hostname: "my-computer",
				},
			},
		},
		"api key from config deprecated": {
			ViperAPIKeyConfigOld: "00000000-0000-4000-8000-000000000000",
			Expected: paramscmd.Params{
				API: paramscmd.API{

					Key:      "00000000-0000-4000-8000-000000000000",
					URL:      "https://api.wakatime.com/api/v1",
					Hostname: "my-computer",
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("sync-offline-activity", "none")
			v.Set("key", test.ViperAPIKey)
			v.Set("settings.api_key", test.ViperAPIKeyConfig)
			v.Set("settings.apikey", test.ViperAPIKeyConfigOld)
			v.Set("hostname", "my-computer")

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params)
		})
	}
}

func TestLoad_API_APIKeyInvalid(t *testing.T) {
	tests := map[string]string{
		"unset":            "",
		"invalid format 1": "not-uuid",
		"invalid format 2": "00000000-0000-0000-8000-000000000000",
		"invalid format 3": "00000000-0000-4000-0000-000000000000",
	}

	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", value)

			_, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
			require.Error(t, err)

			var errauth api.ErrAuth
			require.True(t, errors.As(err, &errauth))
		})
	}
}

func TestLoad_API_APIUrl(t *testing.T) {
	tests := map[string]struct {
		ViperAPIUrl       string
		ViperAPIUrlConfig string
		ViperAPIUrlOld    string
		Expected          paramscmd.Params
	}{
		"api url flag takes preceedence": {
			ViperAPIUrl:       "http://localhost:8080",
			ViperAPIUrlConfig: "http://localhost:8081",
			ViperAPIUrlOld:    "http://localhost:8082",
			Expected: paramscmd.Params{
				API: paramscmd.API{

					Key:      "00000000-0000-4000-8000-000000000000",
					URL:      "http://localhost:8080",
					Hostname: "my-computer",
				},
			},
		},
		"api url deprecated flag takes preceedence": {
			ViperAPIUrlConfig: "http://localhost:8081",
			ViperAPIUrlOld:    "http://localhost:8082",
			Expected: paramscmd.Params{
				API: paramscmd.API{

					Key:      "00000000-0000-4000-8000-000000000000",
					URL:      "http://localhost:8082",
					Hostname: "my-computer",
				},
			},
		},
		"api url from config": {
			ViperAPIUrlConfig: "http://localhost:8081",
			Expected: paramscmd.Params{
				API: paramscmd.API{

					Key:      "00000000-0000-4000-8000-000000000000",
					URL:      "http://localhost:8081",
					Hostname: "my-computer",
				},
			},
		},
		"api url with legacy heartbeats endpoint": {
			ViperAPIUrl: "http://localhost:8080/api/v1/heartbeats.bulk",
			Expected: paramscmd.Params{
				API: paramscmd.API{

					Key:      "00000000-0000-4000-8000-000000000000",
					URL:      "http://localhost:8080/api/v1",
					Hostname: "my-computer",
				},
			},
		},
		"api url with trailing slash": {
			ViperAPIUrl: "http://localhost:8080/api/",
			Expected: paramscmd.Params{
				API: paramscmd.API{

					Key:      "00000000-0000-4000-8000-000000000000",
					URL:      "http://localhost:8080/api",
					Hostname: "my-computer",
				},
			},
		},
		"api url with wakapi style endpoint": {
			ViperAPIUrl: "http://localhost:8080/api/heartbeat",
			Expected: paramscmd.Params{
				API: paramscmd.API{

					Key:      "00000000-0000-4000-8000-000000000000",
					URL:      "http://localhost:8080/api",
					Hostname: "my-computer",
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("sync-offline-activity", "none")
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("api-url", test.ViperAPIUrl)
			v.Set("apiurl", test.ViperAPIUrlOld)
			v.Set("settings.api_url", test.ViperAPIUrlConfig)
			v.Set("hostname", "my-computer")

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params)
		})
	}
}

func TestLoad_APIUrl_Default(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.Equal(t, api.BaseURL, params.API.URL)
}

func TestLoad_API_BackoffAt(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("hostname", "my-computer")
	v.Set("internal.backoff_at", "2021-08-30T18:50:42-03:00")
	v.Set("internal.backoff_retries", "3")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	backoffAt, err := time.Parse(inipkg.DateFormat, "2021-08-30T18:50:42-03:00")
	require.NoError(t, err)

	assert.Equal(t, paramscmd.API{
		BackoffAt:      backoffAt,
		BackoffRetries: 3,
		Key:            "00000000-0000-4000-8000-000000000000",
		URL:            "https://api.wakatime.com/api/v1",
		Hostname:       "my-computer",
	}, params.API)
}

func TestLoad_API_BackoffAtErr(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("hostname", "my-computer")
	v.Set("internal.backoff_at", "2021-08-30")
	v.Set("internal.backoff_retries", "2")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.Equal(t, paramscmd.API{
		BackoffAt:      time.Time{},
		BackoffRetries: 2,
		Key:            "00000000-0000-4000-8000-000000000000",
		URL:            "https://api.wakatime.com/api/v1",
		Hostname:       "my-computer",
	}, params.API)
}

func TestLoad_API_Plugin(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("plugin", "plugin/10.0.0")
	v.Set("hostname", "my-computer")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.Equal(t, paramscmd.API{
		Key:      "00000000-0000-4000-8000-000000000000",
		URL:      "https://api.wakatime.com/api/v1",
		Plugin:   "plugin/10.0.0",
		Hostname: "my-computer",
	}, params.API)
}

func TestLoad_API_Timeout_FlagTakesPreceedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("timeout", 5)
	v.Set("settings.timeout", 10)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.Equal(t, 5*time.Second, params.API.Timeout)
}

func TestLoad_API_Timeout_FromConfig(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.timeout", 10)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.Equal(t, 10*time.Second, params.API.Timeout)
}

func TestLoad_API_DisableSSLVerify_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("no-ssl-verify", true)
	v.Set("settings.no_ssl_verify", false)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.True(t, params.API.DisableSSLVerify)
}

func TestLoad_API_DisableSSLVerify_FromConfig(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.no_ssl_verify", true)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.True(t, params.API.DisableSSLVerify)
}

func TestLoad_API_DisableSSLVerify_Default(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.False(t, params.API.DisableSSLVerify)
}

func TestLoad_API_ProxyURL(t *testing.T) {
	tests := map[string]string{
		"https":  "https://john:secret@example.org:8888",
		"http":   "http://john:secret@example.org:8888",
		"socks5": "socks5://john:secret@example.org:8888",
		"ntlm":   `domain\\john:123456`,
	}

	for name, proxyURL := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.SetDefault("sync-offline-activity", 1000)
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("proxy", proxyURL)

			params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
			require.NoError(t, err)

			assert.Equal(t, proxyURL, params.API.ProxyURL)
		})
	}
}

func TestLoad_API_ProxyURL_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("proxy", "https://john:secret@example.org:8888")
	v.Set("settings.proxy", "ignored")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.Equal(t, "https://john:secret@example.org:8888", params.API.ProxyURL)
}

func TestLoad_API_ProxyURL_FromConfig(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.proxy", "https://john:secret@example.org:8888")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.Equal(t, "https://john:secret@example.org:8888", params.API.ProxyURL)
}

func TestLoad_API_ProxyURL_InvalidFormat(t *testing.T) {
	proxyURL := "ftp://john:secret@example.org:8888"

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("proxy", proxyURL)

	_, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.Error(t, err)
}

func TestLoad_API_SSLCertFilepath_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("ssl-certs-file", "~/path/to/cert.pem")

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(home, "/path/to/cert.pem"), params.API.SSLCertFilepath)
}

func TestLoad_API_SSLCertFilepath_FromConfig(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.ssl_certs_file", "/path/to/cert.pem")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.Equal(t, "/path/to/cert.pem", params.API.SSLCertFilepath)
}

func TestLoadParams_Hostname_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("hostname", "my-machine")
	v.Set("settings.hostname", "ignored")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.Equal(t, "my-machine", params.API.Hostname)
}

func TestLoadParams_Hostname_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.hostname", "my-machine")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	assert.Equal(t, "my-machine", params.API.Hostname)
}

func TestLoadParams_Hostname_DefaultFromSystem(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")

	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	require.NoError(t, err)

	expected, err := os.Hostname()
	require.NoError(t, err)

	assert.Equal(t, expected, params.API.Hostname)
}
