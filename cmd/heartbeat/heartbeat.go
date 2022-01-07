package heartbeat

import (
	"errors"
	"fmt"
	"strings"

	apicmd "github.com/wakatime/wakatime-cli/cmd/api"
	offlinecmd "github.com/wakatime/wakatime-cli/cmd/offline"
	"github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/backoff"
	"github.com/wakatime/wakatime-cli/pkg/deps"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
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

// Run executes the heartbeat command.
func Run(v *viper.Viper) (int, error) {
	queueFilepath, err := offline.QueueFilepath()
	if err != nil {
		return exitcode.ErrGeneric, fmt.Errorf(
			"sending heartbeat(s) failed: failed to load offline queue filepath: %s",
			err,
		)
	}

	err = SendHeartbeats(v, queueFilepath)
	if err != nil {
		var errauth api.ErrAuth
		if errors.As(err, &errauth) {
			return exitcode.ErrAuth, fmt.Errorf(
				"sending heartbeat(s) failed: invalid api key... find yours at wakatime.com/api-key. %w",
				err,
			)
		}

		var errbadRequest api.ErrBadRequest
		if errors.As(err, &errbadRequest) {
			return exitcode.ErrGeneric, fmt.Errorf(
				"sending heartbeat(s) later due to bad request: %w",
				err,
			)
		}

		var errapi api.Err
		if errors.As(err, &errapi) {
			return exitcode.ErrAPI, fmt.Errorf(
				"sending heartbeat(s) later due to api error: %w",
				err,
			)
		}

		return exitcode.ErrGeneric, fmt.Errorf(
			"sending heartbeat(s) failed: %w",
			err,
		)
	}

	log.Debugln("successfully sent heartbeat(s)")

	return exitcode.Success, nil
}

// SendHeartbeats sends a heartbeat to the wakatime api and includes additional
// heartbeats from the offline queue, if available and offline sync is not
// explicitly disabled.
func SendHeartbeats(v *viper.Viper, queueFilepath string) error {
	params, err := params.Load(v, params.Config{APIKeyRequired: true, HeartbeatRequired: true})
	if err != nil {
		return fmt.Errorf("failed to load command parameters: %w", err)
	}

	setLogFields(&params)

	log.Debugf("params: %s", params)

	heartbeats := buildHeartbeats(params)

	// only send at once the maximum amount of `offline.SendLimit`.
	if len(heartbeats) > offline.SendLimit {
		extraHeartbeats := heartbeats[offline.SendLimit:]

		log.Debugf("save %d extra heartbeat(s) to offline queue", len(extraHeartbeats))

		go func() {
			if err := offlinecmd.SaveHeartbeats(v, extraHeartbeats, queueFilepath); err != nil {
				log.Errorf("failed to save extra heartbeats to offline queue: %s", err)
			}
		}()

		heartbeats = heartbeats[:offline.SendLimit]
	}

	handleOpts := initHandleOptions(params)

	if !params.Offline.Disabled {
		if params.Offline.QueueFile != "" {
			queueFilepath = params.Offline.QueueFile
		}

		offlineHandleOpt, err := offline.WithQueue(queueFilepath)
		if err != nil {
			return fmt.Errorf("failed to initialize offline queue handle option: %w", err)
		}

		handleOpts = append(handleOpts, offlineHandleOpt)
	}

	handleOpts = append(handleOpts, backoff.WithBackoff(backoff.Config{
		V:       v,
		At:      params.API.BackoffAt,
		Retries: params.API.BackoffRetries,
	}))

	apiClient, err := apicmd.NewClient(params.API)
	if err != nil {
		return fmt.Errorf("failed to initialize api client: %w", err)
	}

	handle := heartbeat.NewHandle(apiClient, handleOpts...)

	results, err := handle(heartbeats)
	if err != nil {
		return err
	}

	for _, result := range results {
		if len(result.Errors) > 0 {
			log.Warnln(strings.Join(result.Errors, " "))
		}
	}

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
