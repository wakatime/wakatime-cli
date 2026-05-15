package params_test

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/apikey"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	inipkg "github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/log/setup"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/output"
	paramspkg "github.com/wakatime/wakatime-cli/pkg/params"
	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/regex"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	viperini "github.com/go-viper/encoding/ini"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	iniv1 "gopkg.in/ini.v1"
)

func TestLoadHeartbeatParams_FlagTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("exclude-unknown-project", true)
	v.Set("settings.exclude_unknown_project", false)
	v.Set("include-only-with-project-file", true)
	v.Set("settings.include_only_with_project_file", false)
	v.Set("hide-branch-names", true)
	v.Set("settings.hide_branch_names", false)
	v.Set("hide-dependencies", true)
	v.Set("settings.hide_dependencies", false)
	v.Set("hide-project-names", true)
	v.Set("settings.hide_project_names", false)
	v.Set("hide-file-names", true)
	v.Set("settings.hide_file_names", false)
	v.Set("hide-project-folder", true)
	v.Set("settings.hide_project_folder", false)
	v.Set("guess-language", true)
	v.Set("settings.guess_language", false)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.True(t, params.Filter.ExcludeUnknownProject)
	assert.True(t, params.Filter.IncludeOnlyWithProjectFile)
	assert.Equal(t, []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))}, params.Sanitize.HideBranchNames)
	assert.Equal(t, []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))}, params.Sanitize.HideDependencies)
	assert.Equal(t, []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))}, params.Sanitize.HideProjectNames)
	assert.Equal(t, []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))}, params.Sanitize.HideFileNames)
	assert.True(t, params.Sanitize.HideProjectFolder)
	assert.True(t, params.GuessLanguage)
}

func TestLoadHeartbeatParams_ProjectConfigTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("exclude-unknown-project", true)
	v.Set("settings.exclude_unknown_project", false)
	v.Set("include-only-with-project-file", true)
	v.Set("settings.include_only_with_project_file", false)
	v.Set("hide-branch-names", true)
	v.Set("settings.hide_branch_names", false)
	v.Set("hide-dependencies", true)
	v.Set("settings.hide_dependencies", false)
	v.Set("hide-project-names", true)
	v.Set("settings.hide_project_names", false)
	v.Set("hide-file-names", true)
	v.Set("settings.hide_file_names", false)
	v.Set("hide-project-folder", true)
	v.Set("settings.hide_project_folder", false)
	v.Set("guess-language", true)
	v.Set("settings.guess_language", false)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderProjectConfigPrecedence)
	require.NoError(t, err)

	assert.False(t, params.Filter.ExcludeUnknownProject)
	assert.False(t, params.Filter.IncludeOnlyWithProjectFile)
	assert.Equal(t, []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile("a^"))}, params.Sanitize.HideBranchNames)
	assert.Equal(t, []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile("a^"))}, params.Sanitize.HideDependencies)
	assert.Equal(t, []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile("a^"))}, params.Sanitize.HideProjectNames)
	assert.Equal(t, []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile("a^"))}, params.Sanitize.HideFileNames)
	assert.False(t, params.Sanitize.HideProjectFolder)
	assert.False(t, params.GuessLanguage)
}

func TestLoadHeartbeatParams_AILineChanges(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("ai-line-changes", "789")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, 789, *params.AILineChanges)
}

func TestLoadHeartbeatParams_AILineChangesNegative(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("ai-line-changes", "-789")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, -789, *params.AILineChanges)
}

func TestLoadHeartbeatParams_AlternateProject(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("alternate-project", "web")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "web", params.Project.Alternate)
}

func TestLoadHeartbeatParams_AlternateProject_Unset(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Empty(t, params.Project.Alternate)
}

func TestLoadHeartbeatParams_Category(t *testing.T) {
	ctx := t.Context()

	tests := map[string]heartbeat.Category{
		"advising":       heartbeat.AdvisingCategory,
		"ai coding":      heartbeat.AICodingCategory,
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
		"meeting":        heartbeat.MeetingCategory,
		"notes":          heartbeat.NotesCategory,
		"planning":       heartbeat.PlanningCategory,
		"researching":    heartbeat.ResearchingCategory,
		"running tests":  heartbeat.RunningTestsCategory,
		"supporting":     heartbeat.SupportingCategory,
		"translating":    heartbeat.TranslatingCategory,
		"writing docs":   heartbeat.WritingDocsCategory,
		"writing tests":  heartbeat.WritingTestsCategory,
	}

	for name, category := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("category", name)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, category, params.Category)
		})
	}
}

func TestLoadHeartbeatParams_Category_Default(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, heartbeat.UndefinedCategory, params.Category)
}

func TestLoadHeartbeatParams_Category_Invalid(t *testing.T) {
	v := vipertools.MustNew()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("category", "invalid")

	_, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.Error(t, err)

	assert.Equal(t, "failed to parse category: invalid category \"invalid\"", err.Error())
}

func TestLoadHeartbeatParams_CursorPosition(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("cursorpos", 42)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, 42, *params.CursorPosition)
}

func TestLoadHeartbeatParams_CursorPosition_Zero(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("cursorpos", 0)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Zero(t, *params.CursorPosition)
}

func TestLoadHeartbeatParams_CursorPosition_Unset(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Nil(t, params.CursorPosition)
}

func TestLoadHeartbeatParams_Entity_FlagTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("file", "ignored")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "/path/to/file", params.Entity)
}

func TestLoadHeartbeatParams_Entity_FileFlag(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("file", "~/path/to/file")

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(home, "/path/to/file"), params.Entity)
}

func TestLoadHeartbeatParams_Entity_Unset(t *testing.T) {
	v := vipertools.MustNew()

	_, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.Error(t, err)

	assert.Equal(t, "failed to retrieve entity", err.Error())
}

func TestLoadHeartbeatParams_EntityType(t *testing.T) {
	ctx := t.Context()

	tests := map[string]heartbeat.EntityType{
		"file":   heartbeat.FileType,
		"domain": heartbeat.DomainType,
		"app":    heartbeat.AppType,
	}

	for name, entityType := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("entity-type", name)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, entityType, params.EntityType)
		})
	}
}

func TestLoadHeartbeatParams_EntityType_Default(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, heartbeat.FileType, params.EntityType)
}

func TestLoadHeartbeatParams_EntityType_Invalid(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("entity-type", "invalid")

	_, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.Error(t, err)

	assert.Equal(
		t,
		"failed to parse entity type: invalid entity type \"invalid\"",
		err.Error())
}

func TestLoadHeartbeatParams_ExtraHeartbeats(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		r.Close()
		w.Close()
	}()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = r

	paramspkg.Once = sync.Once{}

	data, err := os.ReadFile("testdata/extra_heartbeats.json")
	require.NoError(t, err)

	go func() {
		_, err := w.Write(data)
		require.NoError(t, err)

		w.Close()
	}()

	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("extra-heartbeats", true)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Len(t, params.ExtraHeartbeats, 2)

	assert.NotNil(t, params.ExtraHeartbeats[0].Language)
	assert.Equal(t, heartbeat.LanguageGo.String(), *params.ExtraHeartbeats[0].Language)
	assert.NotNil(t, params.ExtraHeartbeats[1].Language)
	assert.Equal(t, heartbeat.LanguagePython.String(), *params.ExtraHeartbeats[1].Language)

	assert.Equal(t, []heartbeat.Heartbeat{
		{
			Category:          heartbeat.CodingCategory.String(),
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
			Category:          heartbeat.DebuggingCategory.String(),
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

func TestLoadHeartbeatParams_ExtraHeartbeats_WithStringValues(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		r.Close()
		w.Close()
	}()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = r

	paramspkg.Once = sync.Once{}

	data, err := os.ReadFile("testdata/extra_heartbeats_with_string_values.json")
	require.NoError(t, err)

	go func() {
		_, err := w.Write(data)
		require.NoError(t, err)

		w.Close()
	}()

	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("extra-heartbeats", true)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Len(t, params.ExtraHeartbeats, 2)

	assert.NotNil(t, params.ExtraHeartbeats[0].Language)
	assert.Equal(t, heartbeat.LanguageGo.String(), *params.ExtraHeartbeats[0].Language)
	assert.NotNil(t, params.ExtraHeartbeats[1].Language)
	assert.Equal(t, heartbeat.LanguagePython.String(), *params.ExtraHeartbeats[1].Language)

	assert.Equal(t, []heartbeat.Heartbeat{
		{
			Category:        heartbeat.CodingCategory.String(),
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
			Category:        heartbeat.DebuggingCategory.String(),
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

func TestLoadHeartbeatParams_ExtraHeartbeats_WithEOF(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		r.Close()
		w.Close()
	}()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = r

	paramspkg.Once = sync.Once{}

	data, err := os.ReadFile("testdata/extra_heartbeats.json")
	require.NoError(t, err)

	go func() {
		// trim trailing newline and make sure we still parse extra heartbeats
		_, err := w.Write(bytes.TrimRight(data, "\n"))
		require.NoError(t, err)

		w.Close()
	}()

	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("extra-heartbeats", true)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Len(t, params.ExtraHeartbeats, 2)

	assert.NotNil(t, params.ExtraHeartbeats[0].Language)
	assert.Equal(t, heartbeat.LanguageGo.String(), *params.ExtraHeartbeats[0].Language)
	assert.NotNil(t, params.ExtraHeartbeats[1].Language)
	assert.Equal(t, heartbeat.LanguagePython.String(), *params.ExtraHeartbeats[1].Language)

	assert.Equal(t, []heartbeat.Heartbeat{
		{
			Category:          heartbeat.CodingCategory.String(),
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
			Category:          heartbeat.DebuggingCategory.String(),
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

func TestLoadHeartbeatParams_ExtraHeartbeats_NoData(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		r.Close()
		w.Close()
	}()

	ctx := t.Context()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = r

	paramspkg.Once = sync.Once{}

	go func() {
		_, err := w.Write([]byte{})
		require.NoError(t, err)

		w.Close()
	}()

	logFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer logFile.Close()

	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("extra-heartbeats", true)
	v.Set("log-file", logFile.Name())
	v.Set("verbose", true)

	logger, err := setup.Logging(ctx, v)
	require.NoError(t, err)

	defer logger.Flush()

	ctx = log.ToContext(ctx, logger)

	params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Empty(t, params.ExtraHeartbeats)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "skipping extra heartbeats, as no data was provided")
	assert.NotContains(t, string(output), "failed to read extra heartbeats: failed parsing")
}

func TestLoadHeartbeat_GuessLanguage_FlagTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("guess-language", false)
	v.Set("settings.guess_language", true)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.False(t, params.GuessLanguage)
}

func TestLoadHeartbeat_GuessLanguage_FromConfig(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("settings.guess_language", true)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.True(t, params.GuessLanguage)
}

func TestLoadHeartbeat_GuessLanguage_Default(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.False(t, params.GuessLanguage)
}

func TestLoadHeartbeatParams_HumanAdditions(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("human-line-changes", "456")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, 456, *params.HumanLineChanges)
}

func TestLoadHeartbeatParams_HumanAdditionsNegative(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("human-line-changes", "-456")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, -456, *params.HumanLineChanges)
}

func TestLoadHeartbeatParams_IsUnsavedEntity(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("is-unsaved-entity", true)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.True(t, params.IsUnsavedEntity)
}

func TestLoadHeartbeatParams_IsWrite(t *testing.T) {
	ctx := t.Context()

	tests := map[string]bool{
		"is write":    true,
		"is no write": false,
	}

	for name, isWrite := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("write", isWrite)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, isWrite, *params.IsWrite)
		})
	}
}

func TestLoadHeartbeatParams_IsWrite_Unset(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Nil(t, params.IsWrite)
}

func TestLoadHeartbeatParams_Language(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("language", "Go")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageGo.String(), *params.Language)
}

func TestLoadHeartbeatParams_LanguageAlternate(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("alternate-language", "Go")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageGo.String(), params.LanguageAlternate)
	assert.Nil(t, params.Language)
}

func TestLoadHeartbeatParams_LineNumber(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("lineno", 42)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, 42, *params.LineNumber)
}

func TestLoadHeartbeatParams_LineNumber_Zero(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("lineno", 0)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Zero(t, *params.LineNumber)
}

func TestLoadHeartbeatParams_LineNumber_Unset(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Nil(t, params.LineNumber)
}

func TestLoadHeartbeatParams_LocalFile(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("local-file", "/path/to/file")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "/path/to/file", params.LocalFile)
}

func TestLoadHeartbeatParams_Project(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("project", "billing")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "billing", params.Project.Override)
}

func TestLoadHeartbeatParams_Project_Unset(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Empty(t, params.Project.Override)
}

func TestLoadHeartbeatParams_ProjectMap(t *testing.T) {
	ctx := t.Context()

	tests := map[string]struct {
		Entity   string
		Regex    regex.Regex
		Project  string
		Expected []project.MapPattern
	}{
		"simple regex": {
			Entity:  "/home/user/projects/foo/file",
			Regex:   regex.NewRegexpWrap(regexp.MustCompile("projects/foo")),
			Project: "My Awesome Project",
			Expected: []project.MapPattern{
				{
					Name:  "My Awesome Project",
					Regex: regex.NewRegexpWrap(regexp.MustCompile("(?i)projects/foo")),
				},
			},
		},
		"regex with group replacement": {
			Entity:  "/home/user/projects/bar123/file",
			Regex:   regex.NewRegexpWrap(regexp.MustCompile(`^/home/user/projects/bar(\\d+)/`)),
			Project: "project{0}",
			Expected: []project.MapPattern{
				{
					Name:  "project{0}",
					Regex: regex.NewRegexpWrap(regexp.MustCompile(`(?i)^/home/user/projects/bar(\\d+)/`)),
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", test.Entity)
			v.Set(fmt.Sprintf("projectmap.%s", test.Regex.String()), test.Project)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params.Project.MapPatterns)
		})
	}
}

func TestLoadHeartbeatParams_Time(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("time", 1590609206.1)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, 1590609206.1, params.Time)
}

func TestLoadHeartbeatParams_Time_Default(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	now := float64(time.Now().UnixNano()) / 1000000000
	assert.GreaterOrEqual(t, now, params.Time)
	assert.GreaterOrEqual(t, params.Time, now-60)
}

func TestLoadHeartbeatParams_Filter_Exclude(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("exclude", []string{".*", "wakatime.*"})
	v.Set("settings.exclude", []string{".+", "wakatime.+"})
	v.Set("settings.ignore", []string{".?", "wakatime.?"})

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	require.Len(t, params.Filter.Exclude, 6)
	assert.Equal(t, "(?i).*", params.Filter.Exclude[0].String())
	assert.Equal(t, "(?i)wakatime.*", params.Filter.Exclude[1].String())
	assert.Equal(t, "(?i).+", params.Filter.Exclude[2].String())
	assert.Equal(t, "(?i)wakatime.+", params.Filter.Exclude[3].String())
	assert.Equal(t, "(?i).?", params.Filter.Exclude[4].String())
	assert.Equal(t, "(?i)wakatime.?", params.Filter.Exclude[5].String())
}

func TestLoadHeartbeatParams_Filter_Exclude_All(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("exclude", []string{"true"})

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	require.Len(t, params.Filter.Exclude, 1)
	assert.Equal(t, ".*", params.Filter.Exclude[0].String())
}

func TestLoadHeartbeatParams_Filter_Exclude_Multiline(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("settings.ignore", "\t.?\n\twakatime.? \t\n")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	require.Len(t, params.Filter.Exclude, 2)
	assert.Equal(t, "(?i).?", params.Filter.Exclude[0].String())
	assert.Equal(t, "(?i)wakatime.?", params.Filter.Exclude[1].String())
}

func TestLoadHeartbeatParams_Filter_Exclude_IgnoresInvalidRegex(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("exclude", []string{".*", "["})

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	require.Len(t, params.Filter.Exclude, 1)
	assert.Equal(t, "(?i).*", params.Filter.Exclude[0].String())
}

func TestLoadHeartbeatParams_Filter_Exclude_PerlRegexPatterns(t *testing.T) {
	tests := map[string]string{
		"negative lookahead": `^/var/(?!www/).*`,
		"positive lookahead": `^/var/(?=www/).*`,
	}

	for name, pattern := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("exclude", []string{pattern})

			params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			require.Len(t, params.Filter.Exclude, 1)
			assert.Equal(t, "(?i)"+pattern, params.Filter.Exclude[0].String())
		})
	}
}

func TestLoadHeartbeatParams_Filter_ExcludeUnknownProject(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("exclude-unknown-project", true)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.True(t, params.Filter.ExcludeUnknownProject)
}

func TestLoadHeartbeatParams_Filter_ExcludeUnknownProject_FromConfig(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("settings.exclude_unknown_project", true)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.True(t, params.Filter.ExcludeUnknownProject)
}

func TestLoadHeartbeatParams_Filter_ExcludeUnknownProject_FlagTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("exclude-unknown-project", false)
	v.Set("settings.exclude_unknown_project", true)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.False(t, params.Filter.ExcludeUnknownProject)
}

func TestLoadHeartbeatParams_Filter_Include(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("include", []string{".*", "wakatime.*"})
	v.Set("settings.include", []string{".+", "wakatime.+"})

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	require.Len(t, params.Filter.Include, 4)
	assert.Equal(t, "(?i).*", params.Filter.Include[0].String())
	assert.Equal(t, "(?i)wakatime.*", params.Filter.Include[1].String())
	assert.Equal(t, "(?i).+", params.Filter.Include[2].String())
	assert.Equal(t, "(?i)wakatime.+", params.Filter.Include[3].String())
}

func TestLoadHeartbeatParams_Filter_Include_All(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("include", []string{"true"})

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	require.Len(t, params.Filter.Include, 1)
	assert.Equal(t, ".*", params.Filter.Include[0].String())
}

func TestLoadHeartbeatParams_Filter_Include_IgnoresInvalidRegex(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("include", []string{".*", "["})

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	require.Len(t, params.Filter.Include, 1)
	assert.Equal(t, "(?i).*", params.Filter.Include[0].String())
}

func TestLoadHeartbeatParams_Filter_Include_PerlRegexPatterns(t *testing.T) {
	ctx := t.Context()

	tests := map[string]string{
		"negative lookahead": `^/var/(?!www/).*`,
		"positive lookahead": `^/var/(?=www/).*`,
	}

	for name, pattern := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("include", []string{pattern})

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			require.Len(t, params.Filter.Include, 1)
			assert.Equal(t, "(?i)"+pattern, params.Filter.Include[0].String())
		})
	}
}

func TestLoadHeartbeatParams_Filter_IncludeOnlyWithProjectFile(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("include-only-with-project-file", true)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.True(t, params.Filter.IncludeOnlyWithProjectFile)
}

func TestLoadHeartbeatParams_Filter_IncludeOnlyWithProjectFile_FromConfig(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("settings.include_only_with_project_file", true)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.True(t, params.Filter.IncludeOnlyWithProjectFile)
}

func TestLoadHeartbeatParams_SanitizeParams_HideBranchNames_True(t *testing.T) {
	ctx := t.Context()

	tests := map[string]any{
		"lowercase":       "true",
		"uppercase":       "TRUE",
		"first uppercase": "True",
		"boolean":         true,
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("hide-branch-names", viperValue)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, paramspkg.SanitizeParams{
				HideBranchNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
			}, params.Sanitize)
		})
	}
}

func TestLoadHeartbeatParams_SanitizeParams_HideBranchNames_False(t *testing.T) {
	ctx := t.Context()

	tests := map[string]string{
		"lowercase":       "false",
		"uppercase":       "FALSE",
		"first uppercase": "False",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("hide-branch-names", viperValue)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, paramspkg.SanitizeParams{
				HideBranchNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile("a^"))},
			}, params.Sanitize)
		})
	}
}

func TestLoadHeartbeatParams_SanitizeParams_HideBranchNames_List(t *testing.T) {
	ctx := t.Context()

	tests := map[string]struct {
		ViperValue string
		Expected   []regex.Regex
	}{
		"regex": {
			ViperValue: "fix.*",
			Expected: []regex.Regex{
				regex.NewRegexpWrap(regexp.MustCompile("(?i)fix.*")),
			},
		},
		"regex list": {
			ViperValue: ".*secret.*\nfix.*",
			Expected: []regex.Regex{
				regex.NewRegexpWrap(regexp.MustCompile("(?i).*secret.*")),
				regex.NewRegexpWrap(regexp.MustCompile("(?i)fix.*")),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("hide-branch-names", test.ViperValue)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, paramspkg.SanitizeParams{
				HideBranchNames: test.Expected,
			}, params.Sanitize)
		})
	}
}

func TestLoadHeartbeatParams_SanitizeParams_HideBranchNames_FlagTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("hide-branch-names", true)
	v.Set("settings.hide_branch_names", "ignored")
	v.Set("settings.hide_branchnames", "ignored")
	v.Set("settings.hidebranchnames", "ignored")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideBranchNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideBranchNames_ConfigTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_branch_names", true)
	v.Set("settings.hide_branchnames", "ignored")
	v.Set("settings.hidebranchnames", "ignored")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideBranchNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideBranchNames_ConfigDeprecatedOneTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_branchnames", true)
	v.Set("settings.hidebranchnames", "ignored")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideBranchNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideBranchNames_ConfigDeprecatedTwo(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hidebranchnames", true)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideBranchNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideBranchNames_InvalidRegex(t *testing.T) {
	logFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer logFile.Close()

	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("hide-branch-names", ".*secret.*\n[0-9+")
	v.Set("log-file", logFile.Name())

	logger, err := setup.Logging(ctx, v)
	require.NoError(t, err)

	defer logger.Flush()

	ctx = log.ToContext(ctx, logger)

	_, err = paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "failed to compile regex pattern \\\"(?i)[0-9+\\\", it will be ignored")
}

func TestLoadHeartbeatParams_SanitizeParams_HideDependencies_Flag(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_dependencies", true)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideDependencies: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideDependencies_True(t *testing.T) {
	ctx := t.Context()

	tests := map[string]string{
		"lowercase":       "true",
		"uppercase":       "TRUE",
		"first uppercase": "True",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("hide-dependencies", viperValue)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, paramspkg.SanitizeParams{
				HideDependencies: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
			}, params.Sanitize)
		})
	}
}

func TestLoadHeartbeatParams_SanitizeParams_HideDependencies_False(t *testing.T) {
	ctx := t.Context()

	tests := map[string]string{
		"lowercase":       "false",
		"uppercase":       "FALSE",
		"first uppercase": "False",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("hide-dependencies", viperValue)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, paramspkg.SanitizeParams{
				HideDependencies: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile("a^"))},
			}, params.Sanitize)
		})
	}
}

func TestLoadHeartbeatParams_SanitizeParams_HideDependencies_List(t *testing.T) {
	ctx := t.Context()

	tests := map[string]struct {
		ViperValue string
		Expected   []regex.Regex
	}{
		"regex": {
			ViperValue: "fix.*",
			Expected: []regex.Regex{
				regex.NewRegexpWrap(regexp.MustCompile("(?i)fix.*")),
			},
		},
		"regex list": {
			ViperValue: ".*secret.*\nfix.*",
			Expected: []regex.Regex{
				regex.NewRegexpWrap(regexp.MustCompile("(?i).*secret.*")),
				regex.NewRegexpWrap(regexp.MustCompile("(?i)fix.*")),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("hide-dependencies", test.ViperValue)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, paramspkg.SanitizeParams{
				HideDependencies: test.Expected,
			}, params.Sanitize)
		})
	}
}

func TestLoadHeartbeatParams_SanitizeParams_HideDependencies_FlagTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("hide-dependencies", true)
	v.Set("settings.hide_dependencies", "ignored")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideDependencies: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideDependencies_FromConfig(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_dependencies", true)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideDependencies: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideDependencies_InvalidRegex(t *testing.T) {
	logFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer logFile.Close()

	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("hide-dependencies", ".*secret.*\n[0-9+")
	v.Set("log-file", logFile.Name())

	logger, err := setup.Logging(ctx, v)
	require.NoError(t, err)

	defer logger.Flush()

	ctx = log.ToContext(ctx, logger)

	_, err = paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "failed to compile regex pattern \\\"(?i)[0-9+\\\", it will be ignored")
}

func TestLoadHeartbeatParams_SanitizeParams_HideProjectNames_True(t *testing.T) {
	ctx := t.Context()

	tests := map[string]string{
		"lowercase":       "true",
		"uppercase":       "TRUE",
		"first uppercase": "True",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("hide-project-names", viperValue)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, paramspkg.SanitizeParams{
				HideProjectNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
			}, params.Sanitize)
		})
	}
}

func TestLoadHeartbeatParams_SanitizeParams_HideProjectNames_False(t *testing.T) {
	ctx := t.Context()

	tests := map[string]string{
		"lowercase":       "false",
		"uppercase":       "FALSE",
		"first uppercase": "False",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("hide-project-names", viperValue)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, paramspkg.SanitizeParams{
				HideProjectNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile("a^"))},
			}, params.Sanitize)
		})
	}
}

func TestLoadHeartbeatParams_SanitizeParams_HideProjecthNames_List(t *testing.T) {
	ctx := t.Context()

	tests := map[string]struct {
		ViperValue string
		Expected   []regex.Regex
	}{
		"regex": {
			ViperValue: "fix.*",
			Expected: []regex.Regex{
				regex.NewRegexpWrap(regexp.MustCompile("(?i)fix.*")),
			},
		},
		"regex list": {
			ViperValue: ".*secret.*\nfix.*",
			Expected: []regex.Regex{
				regex.NewRegexpWrap(regexp.MustCompile("(?i).*secret.*")),
				regex.NewRegexpWrap(regexp.MustCompile("(?i)fix.*")),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("hide-project-names", test.ViperValue)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, paramspkg.SanitizeParams{
				HideProjectNames: test.Expected,
			}, params.Sanitize)
		})
	}
}

func TestLoadHeartbeatParams_SanitizeParams_HideProjectNames_FlagTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("hide-project-names", true)
	v.Set("settings.hide_project_names", "ignored")
	v.Set("settings.hide_projectnames", "ignored")
	v.Set("settings.hideprojectnames", "ignored")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideProjectNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideProjectNames_ConfigTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_project_names", true)
	v.Set("settings.hide_projectnames", "ignored")
	v.Set("settings.hideprojectnames", "ignored")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideProjectNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideProjectNames_ConfigDeprecatedOneTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_projectnames", true)
	v.Set("settings.hideprojectnames", "ignored")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideProjectNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideProjectNames_ConfigDeprecatedTwo(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hideprojectnames", "true")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideProjectNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideProjectNames_InvalidRegex(t *testing.T) {
	logFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer logFile.Close()

	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("hide-project-names", ".*secret.*\n[0-9+")
	v.Set("log-file", logFile.Name())

	logger, err := setup.Logging(ctx, v)
	require.NoError(t, err)

	defer logger.Flush()

	ctx = log.ToContext(ctx, logger)

	_, err = paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "failed to compile regex pattern \\\"(?i)[0-9+\\\", it will be ignored")
}

func TestLoadHeartbeatParams_SanitizeParams_HideFileNames_True(t *testing.T) {
	ctx := t.Context()

	tests := map[string]string{
		"lowercase":       "true",
		"uppercase":       "TRUE",
		"first uppercase": "True",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("hide-file-names", viperValue)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, paramspkg.SanitizeParams{
				HideFileNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
			}, params.Sanitize)
		})
	}
}

func TestLoadHeartbeatParams_SanitizeParams_HideFileNames_False(t *testing.T) {
	ctx := t.Context()

	tests := map[string]string{
		"lowercase":       "false",
		"uppercase":       "FALSE",
		"first uppercase": "False",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("hide-file-names", viperValue)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, paramspkg.SanitizeParams{
				HideFileNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile("a^"))},
			}, params.Sanitize)
		})
	}
}

func TestLoadHeartbeatParams_SanitizeParams_HideFileNames_List(t *testing.T) {
	ctx := t.Context()

	tests := map[string]struct {
		ViperValue string
		Expected   []regex.Regex
	}{
		"regex": {
			ViperValue: "fix.*",
			Expected: []regex.Regex{
				regex.NewRegexpWrap(regexp.MustCompile("(?i)fix.*")),
			},
		},
		"regex list": {
			ViperValue: ".*secret.*\nfix.*",
			Expected: []regex.Regex{
				regex.NewRegexpWrap(regexp.MustCompile("(?i).*secret.*")),
				regex.NewRegexpWrap(regexp.MustCompile("(?i)fix.*")),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("hide-file-names", test.ViperValue)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, paramspkg.SanitizeParams{
				HideFileNames: test.Expected,
			}, params.Sanitize)
		})
	}
}

func TestLoadheartbeatParams_SanitizeParams_HideFileNames_FlagTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("hide-file-names", true)
	v.Set("hide-filenames", "ignored")
	v.Set("hidefilenames", "ignored")
	v.Set("settings.hide_file_names", "ignored")
	v.Set("settings.hide_filenames", "ignored")
	v.Set("settings.hidefilenames", "ignored")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideFileNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideFileNames_FlagDeprecatedOneTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("hide-filenames", true)
	v.Set("hidefilenames", "ignored")
	v.Set("settings.hide_file_names", "ignored")
	v.Set("settings.hide_filenames", "ignored")
	v.Set("settings.hidefilenames", "ignored")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideFileNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideFileNames_FlagDeprecatedTwoTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("hidefilenames", true)
	v.Set("settings.hide_file_names", "ignored")
	v.Set("settings.hide_filenames", "ignored")
	v.Set("settings.hidefilenames", "ignored")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideFileNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideFileNames_ConfigTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_file_names", true)
	v.Set("settings.hide_filenames", "ignored")
	v.Set("settings.hidefilenames", "ignored")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideFileNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideFileNames_ConfigDeprecatedOneTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_filenames", true)
	v.Set("settings.hidefilenames", "ignored")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideFileNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideFileNames_ConfigDeprecatedTwo(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hidefilenames", "true")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideFileNames: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideFileNames_InvalidRegex(t *testing.T) {
	logFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer logFile.Close()

	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("hide-file-names", ".*secret.*\n[0-9+")
	v.Set("log-file", logFile.Name())

	logger, err := setup.Logging(ctx, v)
	require.NoError(t, err)

	defer logger.Flush()

	ctx = log.ToContext(ctx, logger)

	_, err = paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "failed to compile regex pattern \\\"(?i)[0-9+\\\", it will be ignored")
}

func TestLoadHeartbeatParams_SanitizeParams_HideProjectFolder(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("hide-project-folder", true)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideProjectFolder: true,
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_HideProjectFolder_ConfigTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("settings.hide_project_folder", true)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		HideProjectFolder: true,
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SanitizeParams_OverrideProjectPath(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("project-folder", "/custom-path")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, paramspkg.SanitizeParams{
		ProjectPathOverride: "/custom-path",
	}, params.Sanitize)
}

func TestLoadHeartbeatParams_SubmodulesDisabled_True(t *testing.T) {
	tests := map[string]string{
		"lowercase":       "true",
		"uppercase":       "TRUE",
		"first uppercase": "True",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("git.submodules_disabled", viperValue)

			params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))}, params.Project.SubmodulesDisabled)
		})
	}
}

func TestLoadHeartbeatParams_SubmodulesDisabled_False(t *testing.T) {
	ctx := t.Context()

	tests := map[string]string{
		"lowercase":       "false",
		"uppercase":       "FALSE",
		"first uppercase": "False",
	}

	for name, viperValue := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", "/path/to/file")
			v.Set("git.submodules_disabled", viperValue)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, params.Project.SubmodulesDisabled, []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile("a^"))})
		})
	}
}

func TestLoadHeartbeatsParams_SubmodulesDisabled_List(t *testing.T) {
	ctx := t.Context()

	tests := map[string]struct {
		ViperValue string
		Expected   []regex.Regex
	}{
		"regex": {
			ViperValue: "fix.*",
			Expected: []regex.Regex{
				regex.NewRegexpWrap(regexp.MustCompile("(?i)fix.*")),
			},
		},
		"regex_list": {
			ViperValue: "\n.*secret.*\nfix.*",
			Expected: []regex.Regex{
				regex.NewRegexpWrap(regexp.MustCompile("(?i).*secret.*")),
				regex.NewRegexpWrap(regexp.MustCompile("(?i)fix.*")),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			multilineOption := iniv1.LoadOptions{AllowPythonMultilineValues: true}
			iniCodec := viperini.Codec{LoadOptions: multilineOption}

			codecRegistry := viper.NewCodecRegistry()
			err := codecRegistry.RegisterCodec("ini", iniCodec)
			require.NoError(t, err)

			v := viper.NewWithOptions(viper.WithCodecRegistry(codecRegistry))
			v.Set("entity", "/path/to/file")
			v.Set("git.submodules_disabled", test.ViperValue)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params.Project.SubmodulesDisabled)
		})
	}
}

func TestLoadHeartbeatsParams_SubmoduleProjectMap(t *testing.T) {
	ctx := t.Context()

	tests := map[string]struct {
		Entity   string
		Regex    regex.Regex
		Project  string
		Expected []project.MapPattern
	}{
		"simple regex": {
			Entity:  "/home/user/projects/foo/file",
			Regex:   regex.NewRegexpWrap(regexp.MustCompile("projects/foo")),
			Project: "My Awesome Project",
			Expected: []project.MapPattern{
				{
					Name:  "My Awesome Project",
					Regex: regex.NewRegexpWrap(regexp.MustCompile("(?i)projects/foo")),
				},
			},
		},
		"regex with group replacement": {
			Entity:  "/home/user/projects/bar123/file",
			Regex:   regex.NewRegexpWrap(regexp.MustCompile(`^/home/user/projects/bar(\\d+)/`)),
			Project: "project{0}",
			Expected: []project.MapPattern{
				{
					Name:  "project{0}",
					Regex: regex.NewRegexpWrap(regexp.MustCompile(`(?i)^/home/user/projects/bar(\\d+)/`)),
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("entity", test.Entity)
			v.Set(fmt.Sprintf("git_submodule_projectmap.%s", test.Regex.String()), test.Project)

			params, err := paramspkg.LoadHeartbeatParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params.Project.SubmoduleMapPatterns)
		})
	}
}

func TestLoadAPIParams_ProjectApiKey(t *testing.T) {
	ctx := t.Context()

	tests := map[string]struct {
		Entity   string
		Regex    regex.Regex
		APIKey   string
		Expected []apikey.MapPattern
	}{
		"simple regex": {
			Regex:  regex.NewRegexpWrap(regexp.MustCompile("projects/foo")),
			APIKey: "00000000-0000-4000-8000-000000000001",
			Expected: []apikey.MapPattern{
				{
					APIKey: "00000000-0000-4000-8000-000000000001",
					Regex:  regex.NewRegexpWrap(regexp.MustCompile(`(?i)projects/foo`)),
				},
			},
		},
		"complex regex": {
			Regex:  regex.NewRegexpWrap(regexp.MustCompile(`^/home/user/projects/bar(\\d+)/`)),
			APIKey: "00000000-0000-4000-8000-000000000002",
			Expected: []apikey.MapPattern{
				{
					APIKey: "00000000-0000-4000-8000-000000000002",
					Regex:  regex.NewRegexpWrap(regexp.MustCompile(`(?i)^/home/user/projects/bar(\\d+)/`)),
				},
			},
		},
		"case insensitive": {
			Regex:  regex.NewRegexpWrap(regexp.MustCompile("projects/foo")),
			APIKey: "00000000-0000-4000-8000-000000000001",
			Expected: []apikey.MapPattern{
				{
					APIKey: "00000000-0000-4000-8000-000000000001",
					Regex:  regex.NewRegexpWrap(regexp.MustCompile(`(?i)projects/foo`)),
				},
			},
		},
		"api key equal to default": {
			Regex:    regex.NewRegexpWrap(regexp.MustCompile(`/some/path`)),
			APIKey:   "00000000-0000-4000-8000-000000000000",
			Expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set(fmt.Sprintf("project_api_key.%s", test.Regex.String()), test.APIKey)

			params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params.KeyPatterns)
		})
	}
}

func TestLoadAPIParams_ProjectApiKey_ParseConfig(t *testing.T) {
	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("config", "testdata/.wakatime.cfg")
	v.Set("entity", "testdata/heartbeat_go.json")

	configFile, err := inipkg.FilePath(ctx, v)
	require.NoError(t, err)

	err = inipkg.ReadInConfig(v, configFile)
	require.NoError(t, err)

	params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	expected := []apikey.MapPattern{
		{
			APIKey: "00000000-0000-4000-8000-000000000001",
			Regex:  regex.NewRegexpWrap(regexp.MustCompile("(?i)/some/path")),
		},
	}

	assert.Equal(t, expected, params.KeyPatterns)
}

func TestLoadAPIParams_APIKeyPrefixSupported(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "waka_00000000-0000-4000-8000-000000000000")

	_, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)
}

func TestLoadAPIParams_Plugin(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("plugin", "plugin/10.0.0")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "plugin/10.0.0", params.Plugin)
}

func TestLoadAPIParams_Plugin_Unset(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Empty(t, params.Plugin)
}

func TestLoadAPIParams_Timeout_FlagTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("timeout", 5)
	v.Set("settings.timeout", 10)

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, 5*time.Second, params.Timeout)
}

func TestLoadAPIParams_Timeout_ConfigTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.timeout", 10)

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, 10*time.Second, params.Timeout)
}

func TestLoadAPIParams_Timeout_FromConfig(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.timeout", 10)

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, 10*time.Second, params.Timeout)
}

func TestLoadAPIParams_Timeout_Zero(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("timeout", 0)

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Zero(t, params.Timeout)
}

func TestLoadAPIParams_Timeout_Default(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.SetDefault("timeout", api.DefaultTimeoutSecs)

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, time.Duration(api.DefaultTimeoutSecs)*time.Second, params.Timeout)
}

func TestLoadAPIParams_Timeout_NegativeNumber(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("timeout", 0)

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Zero(t, params.Timeout)
}

func TestLoadAPIParams_Timeout_NonIntegerValue(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("timeout", "invalid")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, time.Duration(api.DefaultTimeoutSecs)*time.Second, params.Timeout)
}

func TestLoadAPIParams_APIKey(t *testing.T) {
	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-4000-8000-000000000000", params.Key)
}

func TestLoadAPIParams_APIKey_FlagTakesPrecedence(t *testing.T) {
	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.api_key", "10000000-0000-4000-8000-000000000000")

	params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-4000-8000-000000000000", params.Key)
}

func TestLoadAPIParams_APIKey_FromConfig(t *testing.T) {
	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("settings.api_key", "10000000-0000-4000-8000-000000000000")

	params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "10000000-0000-4000-8000-000000000000", params.Key)
}

func TestLoadAPIParams_APIKey_FromTopLevelConfig(t *testing.T) {
	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("api_key", "30000000-0000-4000-8000-000000000000")

	params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "30000000-0000-4000-8000-000000000000", params.Key)
}

func TestLoadAPIParams_APIKey_ConfigDeprecatedTakesPrecedence(t *testing.T) {
	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("settings.apikey", "20000000-0000-4000-8000-000000000000")

	params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "20000000-0000-4000-8000-000000000000", params.Key)
}

func TestLoadAPIParams_APIKey_TopLevelDeprecatedTakesPrecedence(t *testing.T) {
	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("apikey", "40000000-0000-4000-8000-000000000000")

	params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "40000000-0000-4000-8000-000000000000", params.Key)
}

func TestLoadAPIParams_APIKeyUnset(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "")

	_, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.Error(t, err)

	var errauth api.ErrAuth

	assert.ErrorAs(t, err, &errauth)
	assert.EqualError(t, errauth, "api key not found or empty")
}

func TestLoadAPIParams_APIKeyInvalid(t *testing.T) {
	ctx := t.Context()

	tests := map[string]string{
		"invalid format 1": "not-uuid",
		"invalid format 2": "00000000-0000-0000-8000-000000000000",
		"invalid format 3": "00000000-0000-4000-0000-000000000000",
	}

	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("key", value)

			_, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.Error(t, err)

			var errauth api.ErrAuth

			assert.ErrorAs(t, err, &errauth)
			assert.EqualError(t, errauth, "invalid api key format")
		})
	}
}

func TestLoadAPIParams_APIKey_ConfigFileTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("config", "testdata/.wakatime.cfg")
	v.Set("entity", "testdata/heartbeat_go.json")

	configFile, err := inipkg.FilePath(t.Context(), v)
	require.NoError(t, err)

	err = inipkg.ReadInConfig(v, configFile)
	require.NoError(t, err)

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-4000-8000-000000000000", params.Key)
}

func TestLoadAPIParams_APIKey_FromVault(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("config", "testdata/.wakatime-vault.cfg")
	v.Set("entity", "testdata/heartbeat_go.json")

	configFile, err := inipkg.FilePath(t.Context(), v)
	require.NoError(t, err)

	err = inipkg.ReadInConfig(v, configFile)
	require.NoError(t, err)

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-4000-8000-000000000000", params.Key)
}

func TestLoadParams_APIKey_FromVault_Err_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping because OS is not darwin.")
	}

	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("config", "testdata/.wakatime-vault-error.cfg")
	v.Set("entity", "testdata/heartbeat_go.json")

	configFile, err := inipkg.FilePath(ctx, v)
	require.NoError(t, err)

	err = inipkg.ReadInConfig(v, configFile)
	require.NoError(t, err)

	_, err = paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)

	assert.EqualError(t, err, "failed to read api key from vault: exit status 1")
}

func TestLoadAPIParams_APIKeyFromEnv(t *testing.T) {
	v := vipertools.MustNew()

	t.Setenv("WAKATIME_API_KEY", "00000000-0000-4000-8000-000000000000")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-4000-8000-000000000000", params.Key)
}

func TestLoadAPIParams_APIKeyFromEnv_Invalid(t *testing.T) {
	v := vipertools.MustNew()

	t.Setenv("WAKATIME_API_KEY", "00000000-0000-4000-0000-000000000000")

	_, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.Error(t, err)

	var errauth api.ErrAuth

	assert.ErrorAs(t, err, &errauth)
	assert.EqualError(t, errauth, "invalid api key format")
}

func TestLoadAPIParams_APIKeyFromEnv_ConfigTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("settings.api_key", "00000000-0000-4000-8000-000000000000")

	t.Setenv("WAKATIME_API_KEY", "10000000-0000-4000-8000-000000000000")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-4000-8000-000000000000", params.Key)
}

func TestLoadAPIParams_APIUrl_Sanitize(t *testing.T) {
	ctx := t.Context()

	tests := map[string]struct {
		URL      string
		Expected string
	}{
		"api url with legacy heartbeats endpoint": {
			URL:      "http://localhost:8080/api/v1/heartbeats.bulk",
			Expected: "http://localhost:8080/api/v1",
		},
		"api url with users heartbeats endpoint": {
			URL:      "http://localhost:8080/users/current/heartbeats",
			Expected: "http://localhost:8080",
		},
		"api url with trailing slash": {
			URL:      "http://localhost:8080/api/",
			Expected: "http://localhost:8080/api",
		},
		"api url with wakapi style endpoint": {
			URL:      "http://localhost:8080/api/heartbeat",
			Expected: "http://localhost:8080/api",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("api-url", test.URL)

			params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params.URL)
		})
	}
}

func TestLoadAPIParams_Url(t *testing.T) {
	ctx := t.Context()

	v := vipertools.MustNew()

	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", "http://localhost:8080")

	params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:8080", params.URL)
}

func TestLoadAPIParams_Url_FlagTakesPrecedence(t *testing.T) {
	ctx := t.Context()

	v := vipertools.MustNew()

	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", "http://localhost:8080")
	v.Set("settings.api_url", "http://localhost:8081")

	params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:8080", params.URL)
}

func TestLoadAPIParams_Url_FlagDeprecatedTakesPrecedence(t *testing.T) {
	ctx := t.Context()

	v := vipertools.MustNew()

	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("apiurl", "http://localhost:8080")
	v.Set("settings.api_url", "http://localhost:8081")

	params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:8080", params.URL)
}

func TestLoadAPIParams_Url_FromConfig(t *testing.T) {
	ctx := t.Context()

	v := vipertools.MustNew()

	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.api_url", "http://localhost:8081")

	params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:8081", params.URL)
}

func TestLoadAPIParams_Url_Default(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, api.BaseURL, params.URL)
}

func TestLoadAPIParams_Url_InvalidFormat(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", "http://in valid")

	_, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	var errauth api.ErrAuth

	require.ErrorAs(t, err, &errauth)
	assert.EqualError(t, errauth, `invalid api url: parse "http://in valid": invalid character " " in host name`)
}

func TestLoadAPIParams_BackoffAt(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("hostname", "my-computer")
	v.Set("timeout", 0)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("internal.backoff_at", "2021-08-30T18:50:42-03:00")
	v.Set("internal.backoff_retries", "3")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	backoffAt, err := time.Parse(inipkg.DateFormat, "2021-08-30T18:50:42-03:00")
	require.NoError(t, err)

	assert.Equal(t, paramspkg.API{
		BackoffAt:      backoffAt,
		BackoffRetries: 3,
		Key:            "00000000-0000-4000-8000-000000000000",
		URL:            "https://api.wakatime.com/api/v1",
		Hostname:       "my-computer",
	}, params)
}

func TestLoadAPIParams_BackoffAtErr(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("hostname", "my-computer")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("timeout", 0)
	v.Set("internal.backoff_at", "2021-08-30")
	v.Set("internal.backoff_retries", "2")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, 2, params.BackoffRetries)
	assert.Empty(t, params.BackoffAt)
}

func TestLoadAPIParams_BackoffAtFuture(t *testing.T) {
	v := vipertools.MustNew()
	backoff := time.Now().Add(time.Duration(2) * time.Hour)

	v.Set("hostname", "my-computer")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("internal.backoff_at", backoff.Format(inipkg.DateFormat))
	v.Set("internal.backoff_retries", "3")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, 3, params.BackoffRetries)
	assert.LessOrEqual(t, params.BackoffAt, time.Now())
}

func TestLoadAPIParams_DisableSSLVerify_FlagTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("no-ssl-verify", false)
	v.Set("settings.no_ssl_verify", true)

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.False(t, params.DisableSSLVerify)
}

func TestLoadAPIParams_DisableSSLVerify_FromConfig(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.no_ssl_verify", true)

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.True(t, params.DisableSSLVerify)
}

func TestLoadAPIParams_DisableSSLVerify_Default(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.False(t, params.DisableSSLVerify)
}

func TestLoadAPIParams_ProxyURL(t *testing.T) {
	ctx := t.Context()

	tests := map[string]string{
		"https":  "https://john:secret@example.org:8888",
		"http":   "http://john:secret@example.org:8888",
		"ipv6":   "socks5://john:secret@2001:0db8:85a3:0000:0000:8a2e:0370:7334:8888",
		"ntlm":   `domain\\john:123456`,
		"socks5": "socks5://john:secret@example.org:8888",
	}

	for name, proxyURL := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("proxy", proxyURL)

			params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, proxyURL, params.ProxyURL)
		})
	}
}

func TestLoadAPIParams_ProxyURL_FlagTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("proxy", "https://john:secret@example.org:8888")
	v.Set("settings.proxy", "ignored")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "https://john:secret@example.org:8888", params.ProxyURL)
}

func TestLoadAPIParams_ProxyURL_FlagTakesPrecedenceOverEnvironment(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("proxy", "https://john:secret@example.org:8888")

	t.Setenv("HTTPS_PROXY", "https://papa:secret@company.org:9000")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "https://john:secret@example.org:8888", params.ProxyURL)
}

func TestLoadAPIParams_ProxyURL_FromConfig(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.proxy", "https://john:secret@example.org:8888")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "https://john:secret@example.org:8888", params.ProxyURL)
}

func TestLoadAPIParams_ProxyURL_FromEnvironment(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	t.Setenv("HTTPS_PROXY", "https://john:secret@example.org:8888")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "https://john:secret@example.org:8888", params.ProxyURL)
}

func TestLoadAPIParams_ProxyURL_NoProxyFromEnvironment(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	t.Setenv("NO_PROXY", "https://some.org,https://api.wakatime.com")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Empty(t, params.ProxyURL)
}

func TestLoadAPIParams_ProxyURL_InvalidFormat(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("proxy", "ftp://john:secret@example.org:8888")

	_, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	var errauth api.ErrAuth

	assert.ErrorAs(t, err, &errauth)
	assert.EqualError(
		t,
		err,
		"invalid url \"ftp://john:secret@example.org:8888\". Must be in format'https://user:pass@host:port' or"+
			" 'socks5://user:pass@host:port' or 'domain\\\\user:pass.'")
}

func TestLoadAPIParams_SSLCertFilepath_FlagTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("ssl-certs-file", "~/path/to/cert.pem")

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(home, "/path/to/cert.pem"), params.SSLCertFilepath)
}

func TestLoadAPIParams_SSLCertFilepath_FromConfig(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.ssl_certs_file", "/path/to/cert.pem")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "/path/to/cert.pem", params.SSLCertFilepath)
}

func TestLoadAPIParams_Hostname_FlagTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("hostname", "my-machine")
	v.Set("settings.hostname", "ignored")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "my-machine", params.Hostname)
}

func TestLoadAPIParams_Hostname_FromConfig(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.hostname", "my-machine")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "my-machine", params.Hostname)
}

func TestLoadAPIParams_Hostname_ConfigTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.hostname", "my-machine")

	t.Setenv("GITPOD_WORKSPACE_ID", "gitpod")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "my-machine", params.Hostname)
}

func TestLoadAPIParams_Hostname_FromGitpodEnv(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	t.Setenv("GITPOD_WORKSPACE_ID", "gitpod")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, "Gitpod", params.Hostname)
}

func TestLoadAPIParams_Hostname_DefaultFromSystem(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	params, err := paramspkg.LoadAPIParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	expected, err := os.Hostname()
	require.NoError(t, err)

	assert.Equal(t, expected, params.Hostname)
}

func TestLoadOfflineParams_Disabled_ConfigTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("settings.offline", false)

	params := paramspkg.LoadOfflineParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	assert.True(t, params.Disabled)
}

func TestLoadOfflineParams_Disabled_FlagDeprecatedTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("disable-offline", false)
	v.Set("disableoffline", true)

	params := paramspkg.LoadOfflineParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	assert.False(t, params.Disabled)
}

func TestLoadOfflineParams_Disabled_FromFlag(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("disable-offline", true)

	params := paramspkg.LoadOfflineParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	assert.True(t, params.Disabled)
}

func TestLoadOfflineParams_RateLimit_FlagTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("heartbeat-rate-limit-seconds", 5)
	v.Set("settings.heartbeat_rate_limit_seconds", 10)

	params := paramspkg.LoadOfflineParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	assert.Equal(t, time.Duration(5)*time.Second, params.RateLimit)
}

func TestLoadOfflineParams_RateLimit_FromConfig(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("settings.heartbeat_rate_limit_seconds", 10)

	params := paramspkg.LoadOfflineParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	assert.Equal(t, time.Duration(10)*time.Second, params.RateLimit)
}

func TestLoadOfflineParams_RateLimit_Zero(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("heartbeat-rate-limit-seconds", 0)

	params := paramspkg.LoadOfflineParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	assert.Zero(t, params.RateLimit)
}

func TestLoadOfflineParams_RateLimit_Default(t *testing.T) {
	v := vipertools.MustNew()
	v.SetDefault("heartbeat-rate-limit-seconds", offline.RateLimitDefaultSeconds)

	params := paramspkg.LoadOfflineParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	assert.Equal(t, time.Duration(offline.RateLimitDefaultSeconds)*time.Second, params.RateLimit)
}

func TestLoadOfflineParams_RateLimit_NegativeNumber(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("heartbeat-rate-limit-seconds", -1)

	params := paramspkg.LoadOfflineParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	assert.Zero(t, params.RateLimit)
}

func TestLoadOfflineParams_RateLimit_NonIntegerValue(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("heartbeat-rate-limit-seconds", "invalid")

	params := paramspkg.LoadOfflineParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	assert.Equal(t, time.Duration(offline.RateLimitDefaultSeconds)*time.Second, params.RateLimit)
}

func TestLoadOfflineParams_LastSentAt(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("internal.heartbeats_last_sent_at", "2021-08-30T18:50:42-03:00")

	params := paramspkg.LoadOfflineParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	lastSentAt, err := time.Parse(inipkg.DateFormat, "2021-08-30T18:50:42-03:00")
	require.NoError(t, err)

	assert.Equal(t, lastSentAt, params.LastSentAt)
}

func TestLoadOfflineParams_LastSentAt_Err(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("internal.heartbeats_last_sent_at", "2021-08-30")

	params := paramspkg.LoadOfflineParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	assert.Zero(t, params.LastSentAt)
}

func TestLoadOfflineParams_LastSentAtFuture(t *testing.T) {
	v := vipertools.MustNew()
	lastSentAt := time.Now().Add(2 * time.Hour)
	v.Set("internal.heartbeats_last_sent_at", lastSentAt.Format(inipkg.DateFormat))

	params := paramspkg.LoadOfflineParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	assert.LessOrEqual(t, params.LastSentAt, time.Now())
}

func TestLoadOfflineParams_SyncMax(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("sync-offline-activity", 42)

	params := paramspkg.LoadOfflineParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	assert.Equal(t, 42, params.SyncMax)
}

func TestLoadOfflineParams_SyncMax_Zero(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("sync-offline-activity", "0")

	params := paramspkg.LoadOfflineParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	assert.Equal(t, math.MaxInt32, params.SyncMax)
}

func TestLoadOfflineParams_SyncMax_Default(t *testing.T) {
	v := vipertools.MustNew()
	v.SetDefault("sync-offline-activity", 1000)

	params := paramspkg.LoadOfflineParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	assert.Equal(t, 1000, params.SyncMax)
}

func TestLoadOfflineParams_SyncMax_NegativeNumber(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("sync-offline-activity", -1)

	params := paramspkg.LoadOfflineParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	assert.Equal(t, math.MaxInt32, params.SyncMax)
}

func TestLoadOfflineParams_SyncMax_NonIntegerValue(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("sync-offline-activity", "invalid")

	params := paramspkg.LoadOfflineParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)

	assert.Equal(t, math.MaxInt32, params.SyncMax)
}

func TestLoadStatusBarParams_HideCategories_Default(t *testing.T) {
	v := vipertools.MustNew()

	params, err := paramspkg.LoadStatusBarParams(v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.True(t, params.HideCategories)
}

func TestLoadStatusBarParams_HideCategories_FlagTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("today-hide-categories", false)
	v.Set("settings.status_bar_show_categories", false)

	params, err := paramspkg.LoadStatusBarParams(v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.False(t, params.HideCategories)
}

func TestLoadStatusBarParams_HideCategories_ConfigTakesPrecedence(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("today-hide-categories", false)
	v.Set("settings.status_bar_show_categories", false)

	params, err := paramspkg.LoadStatusBarParams(v, paramspkg.FlagReadOrderProjectConfigPrecedence)
	require.NoError(t, err)

	assert.True(t, params.HideCategories)
}

func TestLoadStatusBarParams_HideCategories_ShowCategoriesEnabled(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("settings.status_bar_show_categories", true)

	params, err := paramspkg.LoadStatusBarParams(v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.False(t, params.HideCategories)
}

func TestLoadStatusBarParams_HideCategories_ShowCategoriesInvalid(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("settings.status_bar_show_categories", "invalid")

	_, err := paramspkg.LoadStatusBarParams(v, paramspkg.FlagReadOrderFlagPrecedence)
	require.Error(t, err)

	assert.Equal(
		t,
		"failed to parse status_bar_show_categories: strconv.ParseBool: parsing \"invalid\": invalid syntax",
		err.Error(),
	)
}

func TestLoadStatusBarParams_HideMinutes(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("settings.status_bar_hide_minutes", true)

	params, err := paramspkg.LoadStatusBarParams(v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.True(t, params.HideMinutes)
}

func TestLoadStatusBarParams_HideMinutesDefault(t *testing.T) {
	v := vipertools.MustNew()

	params, err := paramspkg.LoadStatusBarParams(v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.False(t, params.HideMinutes)
}

func TestLoadStatusBarParams_HideMinutesInvalid(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("settings.status_bar_hide_minutes", "invalid")

	_, err := paramspkg.LoadStatusBarParams(v, paramspkg.FlagReadOrderFlagPrecedence)
	require.Error(t, err)

	assert.Equal(
		t,
		"failed to parse status_bar_hide_minutes: strconv.ParseBool: parsing \"invalid\": invalid syntax",
		err.Error(),
	)
}

func TestLoadStatusBarParams_Output(t *testing.T) {
	tests := map[string]output.Output{
		"text": output.TextOutput,
		"json": output.JSONOutput,
	}

	for name, out := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("output", name)

			params, err := paramspkg.LoadStatusBarParams(v, paramspkg.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)

			assert.Equal(t, out, params.Output)
		})
	}
}

func TestLoadStatusBarParams_Output_Default(t *testing.T) {
	v := vipertools.MustNew()

	params, err := paramspkg.LoadStatusBarParams(v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, output.TextOutput, params.Output)
}

func TestLoadStatusBarParams_Output_Invalid(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("output", "invalid")

	_, err := paramspkg.LoadStatusBarParams(v, paramspkg.FlagReadOrderFlagPrecedence)
	require.Error(t, err)

	assert.Equal(t, "failed to parse output: invalid output \"invalid\"", err.Error())
}

func TestLoadStatusBarParams_MaxCategories(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("today-max-categories", 0)

	params, err := paramspkg.LoadStatusBarParams(v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, output.TextOutput, params.Output)
}

func TestLoadStatusBarParams_MaxCategoriesNegative(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("today-max-categories", -1)

	_, err := paramspkg.LoadStatusBarParams(v, paramspkg.FlagReadOrderFlagPrecedence)
	require.Error(t, err)

	assert.Equal(t, "today-max-categories must be a positive number, got -1", err.Error())
}

func TestLoadStatusBarParams_MaxCategoriesSetting(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("settings.status_bar_max_categories", 1)

	params, err := paramspkg.LoadStatusBarParams(v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Equal(t, output.TextOutput, params.Output)
}

func TestLoadStatusBarParams_MaxCategoriesNegativeSetting(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("settings.status_bar_max_categories", -1)

	_, err := paramspkg.LoadStatusBarParams(v, paramspkg.FlagReadOrderFlagPrecedence)
	require.Error(t, err)

	assert.Equal(t, "today-max-categories must be a positive number, got -1", err.Error())
}

func TestLoadHeartbeatParams_ExtraHeartbeats_StdinReadOnlyOnce(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		r.Close()
		w.Close()
	}()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = r

	paramspkg.Once = sync.Once{}

	data, err := os.ReadFile("testdata/extra_heartbeats.json")
	require.NoError(t, err)

	_, err = w.Write(data)
	require.NoError(t, err)

	w.Close()

	v := vipertools.MustNew()
	v.Set("entity", "/path/to/file")
	v.Set("extra-heartbeats", true)

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Len(t, params.ExtraHeartbeats, 2)
	assert.Equal(t, "Golang", params.ExtraHeartbeats[0].LanguageAlternate)

	r.Close()
	w.Close()

	// change stdin and make sure loading params uses old stdin
	r, w, err = os.Pipe()
	require.NoError(t, err)

	data, err = os.ReadFile("testdata/extra_heartbeats_with_string_values.json")
	require.NoError(t, err)

	_, err = w.Write(data)
	require.NoError(t, err)

	w.Close()

	os.Stdin = r

	v = viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("extra-heartbeats", true)

	params, err = paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Len(t, params.ExtraHeartbeats, 2)
	assert.Equal(t, "Golang", params.ExtraHeartbeats[0].LanguageAlternate)

	v = viper.New()
	v.Set("entity", "/path/to/file")
	v.Set("extra-heartbeats", true)

	paramspkg.Once = sync.Once{}

	params, err = paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.Len(t, params.ExtraHeartbeats, 2)
	assert.Empty(t, params.ExtraHeartbeats[0].LanguageAlternate)
}

func TestAPI_String(t *testing.T) {
	backoffat, err := time.Parse(inipkg.DateFormat, "2021-08-30T18:50:42-03:00")
	require.NoError(t, err)

	api := paramspkg.API{
		BackoffAt:        backoffat,
		BackoffRetries:   5,
		DisableSSLVerify: true,
		Hostname:         "my-machine",
		Key:              "00000000-0000-4000-8000-000000000000",
		KeyPatterns: []apikey.MapPattern{
			{
				APIKey: "00000000-0000-4000-8000-000000000001",
				Regex:  regex.NewRegexpWrap(regexp.MustCompile("^/api/v1/")),
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
	filterparams := paramspkg.FilterParams{
		Exclude:                    []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile("^/exclude"))},
		ExcludeUnknownProject:      true,
		Include:                    []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile("^/include"))},
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
	heartbeat := paramspkg.Heartbeat{
		AILineChanges:    heartbeat.PointerTo(123),
		Category:         heartbeat.CodingCategory,
		CursorPosition:   heartbeat.PointerTo(15),
		Entity:           "path/to/entity.go",
		EntityType:       heartbeat.FileType,
		ExtraHeartbeats:  make([]heartbeat.Heartbeat, 3),
		GuessLanguage:    true,
		HumanLineChanges: heartbeat.PointerTo(456),
		IsUnsavedEntity:  true,
		IsWrite:          heartbeat.PointerTo(true),
		Language:         heartbeat.PointerTo("Golang"),
		LineNumber:       heartbeat.PointerTo(4),
		LinesInFile:      heartbeat.PointerTo(56),
		Time:             1585598059,
	}

	assert.Equal(
		t,
		"ai line changes: '123', category: 'coding', cursor position: '15',"+
			" entity: 'path/to/entity.go', entity type: 'file', num extra heartbeats: 3,"+
			" guess language: true, human line changes: '456',"+
			" is unsaved entity: true, is write: true, language: 'Golang', line number: '4',"+
			" lines in file: '56', time: 1585598059.00000, filter params: (exclude: '[]',"+
			" exclude unknown project: false, include: '[]', include only with"+
			" project file: false), project params: (alternate: '', branch alternate: '', map patterns:"+
			" '[]', override: '', git submodules disabled: '[]', git submodule project map: '[]'), sanitize"+
			" params: (hide branch names: '[]', hide project folder: false, hide file names: '[]',"+
			" hide project names: '[]', hide dependencies: '[]', project path override: '')",
		heartbeat.String(),
	)
}

func TestOffline_String(t *testing.T) {
	lastSentAt, err := time.Parse(inipkg.DateFormat, "2021-08-30T18:50:42-03:00")
	require.NoError(t, err)

	offline := paramspkg.Offline{
		Disabled:   true,
		LastSentAt: lastSentAt,
		PrintMax:   6,
		RateLimit:  time.Duration(15) * time.Second,
		SyncMax:    12,
	}

	assert.Equal(
		t,
		"disabled: true, last sent at: '2021-08-30T18:50:42-03:00', print max: 6,"+
			" rate limit: 15s, num sync max: 12",
		offline.String(),
	)
}

func TestParamsString_IncludesAIParams(t *testing.T) {
	params := paramspkg.AIParams{
		SyncDisabled: true,
	}

	assert.Equal(
		t,
		"disabled: true",
		params.String(),
	)
}

func TestProjectParams_String(t *testing.T) {
	projectparams := paramspkg.ProjectParams{
		Alternate:       "alternate",
		BranchAlternate: "branch-alternate",
		MapPatterns: []project.MapPattern{{
			Name:  "project-1",
			Regex: regex.NewRegexpWrap(regexp.MustCompile("^/regex")),
		}},
		Override:           "override",
		SubmodulesDisabled: []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile(".*"))},
		SubmoduleMapPatterns: []project.MapPattern{{
			Name:  "awesome-project",
			Regex: regex.NewRegexpWrap(regexp.MustCompile("^/regex")),
		}},
	}

	assert.Equal(
		t,
		"alternate: 'alternate', branch alternate: 'branch-alternate',"+
			" map patterns: '[{project-1 ^/regex}]', override: 'override',"+
			" git submodules disabled: '[.*]', git submodule project map: '[{awesome-project ^/regex}]'",
		projectparams.String(),
	)
}

func TestLoadHeartbeatParams_ProjectFromGitRemote(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("git.project_from_git_remote", true)
	v.Set("entity", "/path/to/file")

	params, err := paramspkg.LoadHeartbeatParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	assert.True(t, params.Project.ProjectFromGitRemote)
}

func TestSanitizeParams_String(t *testing.T) {
	sanitizeparams := paramspkg.SanitizeParams{
		HideBranchNames:     []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile("^/hide"))},
		HideDependencies:    []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile("^/hide"))},
		HideProjectFolder:   true,
		HideFileNames:       []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile("^/hide"))},
		HideProjectNames:    []regex.Regex{regex.NewRegexpWrap(regexp.MustCompile("^/hide"))},
		ProjectPathOverride: "path/to/project",
	}

	assert.Equal(
		t,
		"hide branch names: '[^/hide]', hide project folder: true, hide file names: '[^/hide]',"+
			" hide project names: '[^/hide]', hide dependencies: '[^/hide]', project path override: 'path/to/project'",
		sanitizeparams.String(),
	)
}

func TestStatusBar_String(t *testing.T) {
	statusbar := paramspkg.StatusBar{
		HideCategories: true,
		Output:         output.JSONOutput,
	}

	assert.Equal(
		t,
		"hide categories: true, max categories: 0, output: 'json'",
		statusbar.String(),
	)
}

func TestFlagReadOrder_String(t *testing.T) {
	tests := map[string]paramspkg.FlagReadOrder{
		"flag-precedence":           paramspkg.FlagReadOrderFlagPrecedence,
		"project-config-precedence": paramspkg.FlagReadOrderProjectConfigPrecedence,
	}

	for value, order := range tests {
		t.Run(value, func(t *testing.T) {
			s := order.String()
			assert.Equal(t, value, s)
		})
	}
}

func TestLoadAPIParams_APIURLs(t *testing.T) {
	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api_urls./work/projects", "https://work.example.com/api/v1|00000000-0000-4000-8000-000000000001")

	params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	expected := []apikey.URLPattern{
		{
			APIURL: "https://work.example.com/api/v1",
			APIKey: "00000000-0000-4000-8000-000000000001",
			Regex:  regex.NewRegexpWrap(regexp.MustCompile("(?i)/work/projects")),
		},
	}

	assert.Equal(t, expected, params.URLPatterns)
}

func TestLoadAPIParams_APIURLs_ParseConfig(t *testing.T) {
	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("config", "testdata/.wakatime.cfg")
	v.Set("entity", "testdata/heartbeat_go.json")

	configFile, err := inipkg.FilePath(ctx, v)
	require.NoError(t, err)

	err = inipkg.ReadInConfig(v, configFile)
	require.NoError(t, err)

	params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	expected := []apikey.URLPattern{
		{
			APIURL: "https://work.example.com/api/v1",
			APIKey: "00000000-0000-4000-8000-000000000003",
			Regex:  regex.NewRegexpWrap(regexp.MustCompile("(?i)/work/projects")),
		},
	}

	assert.Equal(t, expected, params.URLPatterns)
}

func TestLoadAPIParams_APIURLs_MissingAPIKey(t *testing.T) {
	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api_url", "https://default.com/api/v1")
	v.Set("api_urls./work/projects", "https://work.example.com/api/v1") // missing api key

	params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	// missing api key should default to api_key
	expected := []apikey.URLPattern{
		{
			APIURL: "https://work.example.com/api/v1",
			APIKey: "00000000-0000-4000-8000-000000000000",
			Regex:  regex.NewRegexpWrap(regexp.MustCompile("(?i)/work/projects")),
		},
	}

	assert.Equal(t, expected, params.URLPatterns)
}

func TestLoadAPIParams_APIURLs_InvalidAPIKey(t *testing.T) {
	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api_urls./work/projects", "https://work.example.com/api/v1|invalid-api-key")

	_, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.Error(t, err)

	var errAuth api.ErrAuth

	assert.ErrorAs(t, err, &errAuth)
}

func TestLoadAPIParams_APIURLs_Empty(t *testing.T) {
	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api_url", "https://default.com/api/v1")
	v.Set("api_urls./work/projects", "")

	params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	// empty string should default to api_url and api_key
	expected := []apikey.URLPattern{
		{
			APIURL: "https://default.com/api/v1",
			APIKey: "00000000-0000-4000-8000-000000000000",
			Regex:  regex.NewRegexpWrap(regexp.MustCompile("(?i)/work/projects")),
		},
	}

	assert.Equal(t, expected, params.URLPatterns)
}

func TestLoadAPIParams_APIURLs_NormalizesURL(t *testing.T) {
	ctx := t.Context()

	v := vipertools.MustNew()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set(
		"api_urls./work/projects",
		"https://work.example.com/api/v1/users/current/heartbeats.bulk|00000000-0000-4000-8000-000000000001",
	)

	params, err := paramspkg.LoadAPIParams(ctx, v, paramspkg.FlagReadOrderFlagPrecedence)
	require.NoError(t, err)

	expected := []apikey.URLPattern{
		{
			APIURL: "https://work.example.com/api/v1",
			APIKey: "00000000-0000-4000-8000-000000000001",
			Regex:  regex.NewRegexpWrap(regexp.MustCompile("(?i)/work/projects")),
		},
	}

	assert.Equal(t, expected, params.URLPatterns)
}
