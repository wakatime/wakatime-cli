package offline

import (
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

// SaveHeartbeats saves heartbeats to the offline db, when we haven't
// tried sending them to the API. If we tried sending to API already,
// to the API. Used when we have heartbeats unsent to API.
func SaveHeartbeats(v *viper.Viper, heartbeats []heartbeat.Heartbeat, queueFilepath string) error {
	params, err := loadParams(v)
	if err != nil {
		return fmt.Errorf("failed to load command parameters: %w", err)
	}

	setLogFields(params)

	log.Debugf("params: %s", params)

	if params.Offline.Disabled {
		return errors.New("saving to offline db disabled")
	}

	if heartbeats == nil {
		// We're not saving surplus extra heartbeats, so save
		// main heartbeat and all extra heartbeats to offline db
		heartbeats = buildHeartbeats(params)
	}

	handleOpts := initHandleOptions(params)

	if params.Offline.QueueFile != "" {
		queueFilepath = params.Offline.QueueFile
	}

	handleOpts = append(handleOpts, offline.WithQueue(queueFilepath))

	sender := offline.Noop{}
	handle := heartbeat.NewHandle(sender, handleOpts...)

	_, _ = handle(heartbeats)

	return nil
}

func loadParams(v *viper.Viper) (paramscmd.Params, error) {
	paramAPI, err := paramscmd.LoadAPIParams(v)
	if err != nil {
		log.Warnf("failed to load API parameters: %s", err)
	}

	paramOffline, err := paramscmd.LoadOfflineParams(v)
	if err != nil {
		log.Warnf("failed to load offline parameters: %s", err)
	}

	params := paramscmd.Params{
		API:     paramAPI,
		Offline: paramOffline,
	}

	paramHeartbeat, err := paramscmd.LoadHeartbeatParams(v)
	if err != nil {
		return paramscmd.Params{}, fmt.Errorf("failed to load heartbeat parameters: %s", err)
	}

	params.Heartbeat = paramHeartbeat

	return params, nil
}

func buildHeartbeats(params paramscmd.Params) []heartbeat.Heartbeat {
	heartbeats := []heartbeat.Heartbeat{}

	userAgent := heartbeat.UserAgent(params.API.Plugin)

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
		params.Heartbeat.LineNumber,
		params.Heartbeat.LinesInFile,
		params.Heartbeat.LocalFile,
		params.Heartbeat.Project.Alternate,
		params.Heartbeat.Project.Override,
		params.Heartbeat.Sanitize.ProjectPathOverride,
		params.Heartbeat.Time,
		userAgent,
	))

	if len(params.Heartbeat.ExtraHeartbeats) > 0 {
		log.Debugf("include %d extra heartbeat(s) from stdin", len(params.Heartbeat.ExtraHeartbeats))

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
				h.LineNumber,
				h.Lines,
				h.LocalFile,
				h.ProjectAlternate,
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
		heartbeat.WithFormatting(heartbeat.FormatConfig{
			RemoteAddressPattern: remote.RemoteAddressRegex,
		}),
		heartbeat.WithEntityModifer(),
		remote.WithDetection(),
		filter.WithFiltering(filter.Config{
			Exclude:                    params.Heartbeat.Filter.Exclude,
			Include:                    params.Heartbeat.Filter.Include,
			IncludeOnlyWithProjectFile: params.Heartbeat.Filter.IncludeOnlyWithProjectFile,
		}),
		filestats.WithDetection(),
		language.WithDetection(),
		deps.WithDetection(deps.Config{
			FilePatterns: params.Heartbeat.Sanitize.HideFileNames,
		}),
		project.WithDetection(project.Config{
			HideProjectNames:  params.Heartbeat.Sanitize.HideProjectNames,
			MapPatterns:       params.Heartbeat.Project.MapPatterns,
			SubmodulePatterns: params.Heartbeat.Project.DisableSubmodule,
		}),
		project.WithFiltering(project.FilterConfig{
			ExcludeUnknownProject: params.Heartbeat.Filter.ExcludeUnknownProject,
		}),
		heartbeat.WithSanitization(heartbeat.SanitizeConfig{
			BranchPatterns:       params.Heartbeat.Sanitize.HideBranchNames,
			FilePatterns:         params.Heartbeat.Sanitize.HideFileNames,
			HideProjectFolder:    params.Heartbeat.Sanitize.HideProjectFolder,
			ProjectPatterns:      params.Heartbeat.Sanitize.HideProjectNames,
			RemoteAddressPattern: remote.RemoteAddressRegex,
		}),
		remote.WithCleanup(),
		filter.WithLengthValidator(),
	}
}

func setLogFields(params paramscmd.Params) {
	if params.API.Plugin != "" {
		log.WithField("plugin", params.API.Plugin)
	}

	log.WithField("time", params.Heartbeat.Time)

	if params.Heartbeat.LineNumber != nil {
		log.WithField("lineno", params.Heartbeat.LineNumber)
	}

	if params.Heartbeat.IsWrite != nil {
		log.WithField("is_write", params.Heartbeat.IsWrite)
	}

	log.WithField("file", params.Heartbeat.Entity)
}
