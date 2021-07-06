package heartbeat

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/wakatime/wakatime-cli/cmd/legacy/legacyapi"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/deps"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/filestats"
	"github.com/wakatime/wakatime-cli/pkg/filter"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/language"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/spf13/viper"
)

// Run executes the heartbeat command.
func Run(v *viper.Viper) (int, error) {
	queueFilepath, err := offline.QueueFilepath()
	if err != nil {
		return exitcode.ErrDefault, fmt.Errorf(
			"failed to load offline queue filepath: %s",
			err,
		)
	}

	err = SendHeartbeats(v, queueFilepath)
	if err != nil {
		var errauth api.ErrAuth
		if errors.As(err, &errauth) {
			return exitcode.ErrAuth, fmt.Errorf(
				"failed to send heartbeat: %s. Find your api key from wakatime.com/settings/api-key",
				errauth,
			)
		}

		var errapi api.Err
		if errors.As(err, &errapi) {
			return exitcode.ErrAPI, fmt.Errorf(
				"failed to send heartbeat(s) due to api error: %s",
				err,
			)
		}

		return exitcode.ErrDefault, fmt.Errorf(
			"failed to send heartbeat(s): %s",
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
	params, err := LoadParams(v)
	if err != nil {
		return fmt.Errorf("failed to load command parameters: %w", err)
	}

	if params.EntityType == heartbeat.FileType && !isFile(params.Entity) {
		log.Warnf("file '%s' does not exist. ignoring this heartbeat", params.Entity)
		return nil
	}

	setLogFields(&params)

	log.Debugf("heartbeat params: %s", params)

	userAgent := heartbeat.UserAgentUnknownPlugin()
	if params.API.Plugin != "" {
		userAgent = heartbeat.UserAgent(params.API.Plugin)
	}

	heartbeats := []heartbeat.Heartbeat{
		heartbeat.New(
			params.Category,
			params.CursorPosition,
			params.Entity,
			params.EntityType,
			params.IsWrite,
			params.Language,
			params.LanguageAlternate,
			params.LineNumber,
			params.LocalFile,
			params.Project.Alternate,
			params.Project.Override,
			params.Time,
			userAgent,
		),
	}

	if len(params.ExtraHeartbeats) > 0 {
		log.Debugf("include %d extra heartbeat(s) from stdin", len(params.ExtraHeartbeats))

		for _, h := range params.ExtraHeartbeats {
			if h.EntityType == heartbeat.FileType && !isFile(h.Entity) {
				log.Warnf("file '%s' does not exist. ignoring this extra heartbeat", h.Entity)
				return nil
			}

			heartbeats = append(heartbeats, heartbeat.New(
				h.Category,
				h.CursorPosition,
				h.Entity,
				h.EntityType,
				h.IsWrite,
				h.Language,
				"",
				h.LineNumber,
				h.LocalFile,
				h.ProjectAlternate,
				h.ProjectOverride,
				h.Time,
				userAgent,
			))
		}
	}

	handleOpts := []heartbeat.HandleOption{
		filter.WithFiltering(filter.Config{
			Exclude:                    params.Filter.Exclude,
			ExcludeUnknownProject:      params.Filter.ExcludeUnknownProject,
			Include:                    params.Filter.Include,
			IncludeOnlyWithProjectFile: params.Filter.IncludeOnlyWithProjectFile,
		}),
		filestats.WithDetection(filestats.Config{
			LinesInFile: params.LinesInFile,
		}),
		language.WithDetection(),
		deps.WithDetection(deps.Config{
			FilePatterns: params.Sanitize.HideFileNames,
		}),
		project.WithDetection(project.Config{
			ShouldObfuscateProject: heartbeat.ShouldSanitize(params.Entity, params.Sanitize.HideProjectNames),
			MapPatterns:            params.Project.MapPatterns,
			SubmodulePatterns:      params.Project.DisableSubmodule,
		}),
		heartbeat.WithSanitization(heartbeat.SanitizeConfig{
			BranchPatterns:  params.Sanitize.HideBranchNames,
			FilePatterns:    params.Sanitize.HideFileNames,
			ProjectPatterns: params.Sanitize.HideProjectNames,
		}),
	}

	if !params.OfflineDisabled {
		if params.OfflineQueueFile != "" {
			queueFilepath = params.OfflineQueueFile
		}

		offlineHandleOpt, err := offline.WithQueue(queueFilepath, params.OfflineSyncMax)
		if err != nil {
			return fmt.Errorf("failed to initialize offline queue handle option: %w", err)
		}

		handleOpts = append(handleOpts, offlineHandleOpt)
	}

	apiClient, err := legacyapi.NewClient(params.API)
	if err != nil {
		return fmt.Errorf("failed to initialize api client: %w", err)
	}

	handle := heartbeat.NewHandle(apiClient, handleOpts...)

	results, err := handle(heartbeats)
	if err != nil {
		return fmt.Errorf("failed to send heartbeats via api client: %w", err)
	}

	for _, result := range results {
		if len(result.Errors) > 0 {
			log.Warnln(strings.Join(result.Errors, " "))
		}
	}

	return nil
}

func setLogFields(params *Params) {
	if params.API.Plugin != "" {
		log.WithField("plugin", params.API.Plugin)
	}

	log.WithField("time", params.Time)

	if params.LineNumber != nil {
		log.WithField("lineno", params.LineNumber)
	}

	if params.IsWrite != nil {
		log.WithField("is_write", params.IsWrite)
	}

	log.WithField("file", params.Entity)
}

// isFile checks if the passed in filepath is a valid file.
func isFile(filepath string) bool {
	info, err := os.Stat(filepath)
	return !(os.IsNotExist(err) || info.IsDir())
}
