package handler

import (
	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/apikey"
	"github.com/wakatime/wakatime-cli/pkg/deps"
	"github.com/wakatime/wakatime-cli/pkg/fileexperts"
	"github.com/wakatime/wakatime-cli/pkg/filestats"
	"github.com/wakatime/wakatime-cli/pkg/filter"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/language"
	"github.com/wakatime/wakatime-cli/pkg/params"
	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/remote"
)

// Preprocessor is a function used to preprocess heartbeat parameters.
// It takes params.Params as input and returns a heartbeat.HandleOption.
type Preprocessor func(params params.Params) heartbeat.HandleOption

// WithAIParsing returns a Preprocessor that appends ai heartbeats.
func WithAIParsing() Preprocessor {
	return func(params params.Params) heartbeat.HandleOption {
		return ai.WithAISync(ai.Config{
			SyncAfterTime: params.AI.SyncAfterTime,
			SyncDisabled:  params.AI.SyncDisabled,
		})
	}
}

// WithFormatting returns a Preprocessor that applies heartbeat formatting.
func WithFormatting() Preprocessor {
	return func(_ params.Params) heartbeat.HandleOption {
		return heartbeat.WithFormatting()
	}
}

// WithCategoryDetection returns a Preprocessor that applies heartbeat categorization.
func WithCategoryDetection() Preprocessor {
	return func(_ params.Params) heartbeat.HandleOption {
		return heartbeat.WithCategory()
	}
}

// WithEntityModifier returns a Preprocessor that applies entity modification to heartbeats.
func WithEntityModifier() Preprocessor {
	return func(_ params.Params) heartbeat.HandleOption {
		return heartbeat.WithEntityModifier()
	}
}

// WithHeartbeatFiltering returns a Preprocessor that applies filtering to heartbeats.
func WithHeartbeatFiltering() Preprocessor {
	return func(params params.Params) heartbeat.HandleOption {
		return filter.WithFiltering(filter.Config{
			Exclude:                    params.Heartbeat.Filter.Exclude,
			Include:                    params.Heartbeat.Filter.Include,
			IncludeOnlyWithProjectFile: params.Heartbeat.Filter.IncludeOnlyWithProjectFile,
		})
	}
}

// WithRemoteDetection returns a Preprocessor that applies remote detection to heartbeats.
func WithRemoteDetection() Preprocessor {
	return func(_ params.Params) heartbeat.HandleOption {
		return remote.WithDetection()
	}
}

// WithAPIKeyReplacing returns a Preprocessor that replaces API keys and URLs in heartbeats.
func WithAPIKeyReplacing() Preprocessor {
	return func(params params.Params) heartbeat.HandleOption {
		return apikey.WithReplacing(apikey.Config{
			DefaultAPIKey: params.API.Key,
			DefaultAPIURL: params.API.URL,
			MapPatterns:   params.API.KeyPatterns,
			URLPatterns:   params.API.URLPatterns,
		})
	}
}

// WithFileStatsDetection returns a Preprocessor that applies file statistics detection to heartbeats.
func WithFileStatsDetection() Preprocessor {
	return func(_ params.Params) heartbeat.HandleOption {
		return filestats.WithDetection()
	}
}

// WithLanguageDetection returns a Preprocessor that applies language detection to heartbeats.
func WithLanguageDetection() Preprocessor {
	return func(params params.Params) heartbeat.HandleOption {
		return language.WithDetection(language.Config{
			GuessLanguage: params.Heartbeat.GuessLanguage,
		})
	}
}

// WithDependencyDetection returns a Preprocessor that applies dependency detection to heartbeats.
func WithDependencyDetection() Preprocessor {
	return func(params params.Params) heartbeat.HandleOption {
		return deps.WithDetection(deps.Config{
			FilePatterns: params.Heartbeat.Sanitize.HideFileNames,
		})
	}
}

// WithProjectDetection returns a Preprocessor that applies project detection to heartbeats.
func WithProjectDetection() Preprocessor {
	return func(params params.Params) heartbeat.HandleOption {
		return project.WithDetection(project.Config{
			HideProjectNames:     params.Heartbeat.Sanitize.HideProjectNames,
			MapPatterns:          params.Heartbeat.Project.MapPatterns,
			ProjectFromGitRemote: params.Heartbeat.Project.ProjectFromGitRemote,
			Submodule: project.Submodule{
				DisabledPatterns: params.Heartbeat.Project.SubmodulesDisabled,
				MapPatterns:      params.Heartbeat.Project.SubmoduleMapPatterns,
			},
		})
	}
}

// WithProjectFiltering returns a Preprocessor that applies project filtering to heartbeats.
func WithProjectFiltering() Preprocessor {
	return func(params params.Params) heartbeat.HandleOption {
		return project.WithFiltering(project.FilterConfig{
			ExcludeUnknownProject: params.Heartbeat.Filter.ExcludeUnknownProject,
		})
	}
}

// WithHeartbeatSanitization returns a Preprocessor that applies sanitization to heartbeats.
func WithHeartbeatSanitization() Preprocessor {
	return func(params params.Params) heartbeat.HandleOption {
		return heartbeat.WithSanitization(heartbeat.SanitizeConfig{
			HideBranchPatterns:     params.Heartbeat.Sanitize.HideBranchNames,
			HideDependencyPatterns: params.Heartbeat.Sanitize.HideDependencies,
			HideFilePatterns:       params.Heartbeat.Sanitize.HideFileNames,
			HideProjectFolder:      params.Heartbeat.Sanitize.HideProjectFolder,
			HideProjectPatterns:    params.Heartbeat.Sanitize.HideProjectNames,
		})
	}
}

// WithFileExpertsValidation returns a Preprocessor that applies validation for file experts.
func WithFileExpertsValidation() Preprocessor {
	return func(_ params.Params) heartbeat.HandleOption {
		return fileexperts.WithValidation()
	}
}

// WithRemoteCleanup returns a Preprocessor that applies cleanup for remote heartbeats.
func WithRemoteCleanup() Preprocessor {
	return func(_ params.Params) heartbeat.HandleOption {
		return remote.WithCleanup()
	}
}
