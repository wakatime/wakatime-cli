package offline

import (
	"context"
	"errors"
	"fmt"

	paramscmd "github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/pkg/deps"
	"github.com/wakatime/wakatime-cli/pkg/filestats"
	"github.com/wakatime/wakatime-cli/pkg/filter"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/language"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/remote"

	"github.com/spf13/viper"
)

// SaveHeartbeats saves heartbeats to the offline db without trying to send to the API.
// Used when we have more heartbeats than `offline.SendLimit`, when we couldn't send
// heartbeats to the API, or the API returned an auth error.
func SaveHeartbeats(ctx context.Context, v *viper.Viper, heartbeats []heartbeat.Heartbeat, queueFilepath string) error {
	params, err := loadParams(ctx, v)
	if err != nil {
		return fmt.Errorf("failed to load command parameters: %w", err)
	}

	logger := log.Extract(ctx)

	setLogFields(ctx, params)
	logger.Debugf("params: %s", params)

	if params.Offline.Disabled {
		return errors.New("saving to offline db disabled")
	}

	if heartbeats == nil {
		// We're not saving surplus extra heartbeats, so save
		// main heartbeat and all extra heartbeats to offline db
		heartbeats = buildHeartbeats(ctx, params)
	}

	handleOpts := initHandleOptions(params)

	handleOpts = append(handleOpts, offline.WithQueue(queueFilepath))

	sender := offline.Noop{}
	handle := heartbeat.NewHandle(sender, handleOpts...)

	_, _ = handle(ctx, heartbeats)

	return nil
}

func loadParams(ctx context.Context, v *viper.Viper) (paramscmd.Params, error) {
	logger := log.Extract(ctx)

	paramAPI, err := paramscmd.LoadAPIParams(ctx, v)
	if err != nil {
		logger.Warnf("failed to load API parameters: %s", err)
	}

	paramHeartbeat, err := paramscmd.LoadHeartbeatParams(ctx, v)
	if err != nil {
		return paramscmd.Params{}, fmt.Errorf("failed to load heartbeat parameters: %s", err)
	}

	return paramscmd.Params{
		API:       paramAPI,
		Heartbeat: paramHeartbeat,
		Offline:   paramscmd.LoadOfflineParams(ctx, v),
	}, nil
}

func buildHeartbeats(ctx context.Context, params paramscmd.Params) []heartbeat.Heartbeat {
	heartbeats := []heartbeat.Heartbeat{}

	userAgent := heartbeat.UserAgent(ctx, params.API.Plugin)

	heartbeats = append(heartbeats, heartbeat.New(
		params.Heartbeat.Project.BranchAlternate,
		params.Heartbeat.Category,
		params.Heartbeat.CursorPosition,
		params.Heartbeat.Entity,
		params.Heartbeat.EntityType,
		params.Heartbeat.IsUnsavedEntity,
		params.Heartbeat.IsWrite,
		params.Heartbeat.Language,
		params.Heartbeat.LanguageAlternate,
		params.Heartbeat.LineAdditions,
		params.Heartbeat.LineDeletions,
		params.Heartbeat.LineNumber,
		params.Heartbeat.LinesInFile,
		params.Heartbeat.LocalFile,
		params.Heartbeat.Project.Alternate,
		params.Heartbeat.Project.ProjectFromGitRemote,
		params.Heartbeat.Project.Override,
		params.Heartbeat.Sanitize.ProjectPathOverride,
		params.Heartbeat.Time,
		userAgent,
	))

	if len(params.Heartbeat.ExtraHeartbeats) > 0 {
		logger := log.Extract(ctx)
		logger.Debugf("include %d extra heartbeat(s) from stdin", len(params.Heartbeat.ExtraHeartbeats))

		for _, h := range params.Heartbeat.ExtraHeartbeats {
			heartbeats = append(heartbeats, heartbeat.New(
				h.BranchAlternate,
				h.Category,
				h.CursorPosition,
				h.Entity,
				h.EntityType,
				h.IsUnsavedEntity,
				h.IsWrite,
				h.Language,
				h.LanguageAlternate,
				h.LineAdditions,
				h.LineDeletions,
				h.LineNumber,
				h.Lines,
				h.LocalFile,
				h.ProjectAlternate,
				h.ProjectFromGitRemote,
				h.ProjectOverride,
				h.ProjectPathOverride,
				h.Time,
				userAgent,
			))
		}
	}

	return heartbeats
}

func initHandleOptions(params paramscmd.Params) []heartbeat.HandleOption {
	return []heartbeat.HandleOption{
		heartbeat.WithFormatting(),
		heartbeat.WithEntityModifier(),
		filter.WithFiltering(filter.Config{
			Exclude:                    params.Heartbeat.Filter.Exclude,
			Include:                    params.Heartbeat.Filter.Include,
			IncludeOnlyWithProjectFile: params.Heartbeat.Filter.IncludeOnlyWithProjectFile,
		}),
		remote.WithDetection(),
		filestats.WithDetection(),
		language.WithDetection(language.Config{
			GuessLanguage: params.Heartbeat.GuessLanguage,
		}),
		deps.WithDetection(deps.Config{
			FilePatterns: params.Heartbeat.Sanitize.HideFileNames,
		}),
		project.WithDetection(project.Config{
			HideProjectNames:     params.Heartbeat.Sanitize.HideProjectNames,
			MapPatterns:          params.Heartbeat.Project.MapPatterns,
			ProjectFromGitRemote: params.Heartbeat.Project.ProjectFromGitRemote,
			Submodule: project.Submodule{
				DisabledPatterns: params.Heartbeat.Project.SubmodulesDisabled,
				MapPatterns:      params.Heartbeat.Project.SubmoduleMapPatterns,
			},
		}),
		project.WithFiltering(project.FilterConfig{
			ExcludeUnknownProject: params.Heartbeat.Filter.ExcludeUnknownProject,
		}),
		heartbeat.WithSanitization(heartbeat.SanitizeConfig{
			BranchPatterns:    params.Heartbeat.Sanitize.HideBranchNames,
			FilePatterns:      params.Heartbeat.Sanitize.HideFileNames,
			HideProjectFolder: params.Heartbeat.Sanitize.HideProjectFolder,
			ProjectPatterns:   params.Heartbeat.Sanitize.HideProjectNames,
		}),
		remote.WithCleanup(),
		filter.WithLengthValidator(),
	}
}

func setLogFields(ctx context.Context, params paramscmd.Params) {
	log.AddField(ctx, "file", params.Heartbeat.Entity)
	log.AddField(ctx, "time", params.Heartbeat.Time)

	if params.API.Plugin != "" {
		log.AddField(ctx, "plugin", params.API.Plugin)
	}

	if params.Heartbeat.LineNumber != nil {
		log.AddField(ctx, "lineno", params.Heartbeat.LineNumber)
	}

	if params.Heartbeat.IsWrite != nil {
		log.AddField(ctx, "is_write", params.Heartbeat.IsWrite)
	}
}
