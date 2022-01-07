package offline

import (
	"errors"
	"fmt"

	"github.com/wakatime/wakatime-cli/cmd/params"
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

// SaveHeartbeats saves heartbeats to the offline db without trying to send
// to the API. Should only be used after a config file parse error.
func SaveHeartbeats(v *viper.Viper, heartbeats []heartbeat.Heartbeat, queueFilepath string) error {
	config := params.Config{}

	if heartbeats == nil {
		config.HeartbeatRequired = true
	}

	params, err := params.Load(v, config)
	if err != nil {
		return fmt.Errorf("failed to load command parameters: %w", err)
	}

	setLogFields(&params)

	log.Debugf("params: %s", params)

	if params.Offline.Disabled {
		return errors.New("abort saving to offline queue due to being disabled")
	}

	if heartbeats == nil {
		heartbeats = buildHeartbeats(params)
	}

	handleOpts := initHandleOptions(params)

	if params.Offline.QueueFile != "" {
		queueFilepath = params.Offline.QueueFile
	}

	offlineHandleOpt, err := offline.WithQueue(queueFilepath)
	if err != nil {
		return fmt.Errorf("failed to initialize offline queue handle option: %w", err)
	}

	handleOpts = append(handleOpts, offlineHandleOpt)

	sender := offline.Sender{}
	handle := heartbeat.NewHandle(&sender, handleOpts...)

	_, _ = handle(heartbeats)

	return nil
}

func buildHeartbeats(params params.Params) []heartbeat.Heartbeat {
	heartbeats := []heartbeat.Heartbeat{}

	userAgent := heartbeat.UserAgentUnknownPlugin()
	if params.API.Plugin != "" {
		userAgent = heartbeat.UserAgent(params.API.Plugin)
	}

	heartbeats = append(heartbeats, heartbeat.New(
		params.Heartbeat.Category,
		params.Heartbeat.CursorPosition,
		params.Heartbeat.Entity,
		params.Heartbeat.EntityType,
		params.Heartbeat.IsWrite,
		params.Heartbeat.Language,
		params.Heartbeat.LanguageAlternate,
		params.Heartbeat.LineNumber,
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
				h.Category,
				h.CursorPosition,
				h.Entity,
				h.EntityType,
				h.IsWrite,
				h.Language,
				h.LanguageAlternate,
				h.LineNumber,
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

func setLogFields(params *params.Params) {
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

func initHandleOptions(params params.Params) []heartbeat.HandleOption {
	return []heartbeat.HandleOption{
		heartbeat.WithFormatting(heartbeat.FormatConfig{
			RemoteAddressPattern: remote.RemoteAddressRegex,
		}),
		filter.WithFiltering(filter.Config{
			Exclude:                    params.Heartbeat.Filter.Exclude,
			ExcludeUnknownProject:      params.Heartbeat.Filter.ExcludeUnknownProject,
			Include:                    params.Heartbeat.Filter.Include,
			IncludeOnlyWithProjectFile: params.Heartbeat.Filter.IncludeOnlyWithProjectFile,
			RemoteAddressPattern:       remote.RemoteAddressRegex,
		}),
		heartbeat.WithEntityModifer(),
		remote.WithDetection(),
		filestats.WithDetection(filestats.Config{
			LinesInFile: params.Heartbeat.LinesInFile,
		}),
		language.WithDetection(),
		deps.WithDetection(deps.Config{
			FilePatterns: params.Heartbeat.Sanitize.HideFileNames,
		}),
		project.WithDetection(project.Config{
			ShouldObfuscateProject: heartbeat.ShouldSanitize(
				params.Heartbeat.Entity, params.Heartbeat.Sanitize.HideProjectNames),
			MapPatterns:       params.Heartbeat.Project.MapPatterns,
			SubmodulePatterns: params.Heartbeat.Project.DisableSubmodule,
		}),
		heartbeat.WithSanitization(heartbeat.SanitizeConfig{
			BranchPatterns:       params.Heartbeat.Sanitize.HideBranchNames,
			FilePatterns:         params.Heartbeat.Sanitize.HideFileNames,
			HideProjectFolder:    params.Heartbeat.Sanitize.HideProjectFolder,
			ProjectPatterns:      params.Heartbeat.Sanitize.HideProjectNames,
			RemoteAddressPattern: remote.RemoteAddressRegex,
		}),
	}
}
