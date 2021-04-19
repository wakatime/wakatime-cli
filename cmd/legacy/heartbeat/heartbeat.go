package heartbeat

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/deps"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/filestats"
	"github.com/wakatime/wakatime-cli/pkg/filter"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/language"
	log "github.com/wakatime/wakatime-cli/pkg/logfile2"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/certifi/gocertifi"
	"github.com/spf13/viper"
)

// Run executes the heartbeat command.
func Run(v *viper.Viper) {
	queueFilepath, err := offline.QueueFilepath()
	if err != nil {
		log.LogEntry.Fatalf("failed to load offline queue filepath: %s", err)
	}

	err = SendHeartbeats(v, queueFilepath)
	if err != nil {
		var errauth api.ErrAuth
		if errors.As(err, &errauth) {
			log.LogEntry.Errorf(
				"failed to send heartbeat: %s. Find your api key from wakatime.com/settings/api-key",
				errauth,
			)
			os.Exit(exitcode.ErrAuth)
		}

		var errapi api.Err
		if errors.As(err, &errapi) {
			log.LogEntry.Errorf("failed to send heartbeat(s): %s", err)
			os.Exit(exitcode.ErrAPI)
		}

		log.LogEntry.Fatalf("failed to send heartbeat(s): %s", err)
	}

	log.LogEntry.Debugln("successfully handled heartbeat(s)")
	os.Exit(exitcode.Success)
}

// SendHeartbeats sends a heartbeat to the wakatime api and includes additional
// heartbeats from the offline queue, if available and offline sync is not
// explicitly disabled.
func SendHeartbeats(v *viper.Viper, queueFilepath string) error {
	params, err := LoadParams(v)
	if err != nil {
		return fmt.Errorf("failed to load command parameters: %w", err)
	}

	setLogEntryWithParams(&params)

	log.LogEntry.Debugf("heartbeat params: %s", params)

	withAuth, err := api.WithAuth(api.BasicAuth{
		Secret: params.API.Key,
	})
	if err != nil {
		return fmt.Errorf("failed to set up auth option on api client: %w", err)
	}

	clientOpts := []api.Option{
		withAuth,
		api.WithTimeout(params.API.Timeout),
	}

	if params.Network.DisableSSLVerify {
		clientOpts = append(clientOpts, api.WithDisableSSLVerify())
	}

	if !params.Network.DisableSSLVerify && params.Network.SSLCertFilepath != "" {
		withSSLCert, err := api.WithSSLCertFile(params.Network.SSLCertFilepath)
		if err != nil {
			return fmt.Errorf("failed to set up ssl cert file option on api client: %s", err)
		}

		clientOpts = append(clientOpts, withSSLCert)
	} else if !params.Network.DisableSSLVerify {
		certPool, err := gocertifi.CACerts()
		if err != nil {
			return fmt.Errorf("failed to build certifi cert pool: %s", err)
		}

		withSSLCert, err := api.WithSSLCertPool(certPool)
		if err != nil {
			return fmt.Errorf("failed to set up ssl cert pool option on api client: %s", err)
		}

		clientOpts = append(clientOpts, withSSLCert)
	}

	if params.Network.ProxyURL != "" {
		withProxy, err := api.WithProxy(params.Network.ProxyURL)
		if err != nil {
			return fmt.Errorf("failed to set up proxy option on api client: %w", err)
		}

		clientOpts = append(clientOpts, withProxy)

		if strings.Contains(params.Network.ProxyURL, `\\`) {
			withNTLMRetry, err := api.WithNTLMRequestRetry(params.Network.ProxyURL)
			if err != nil {
				return fmt.Errorf("failed to set up ntlm request retry option on api client: %w", err)
			}

			clientOpts = append(clientOpts, withNTLMRetry)
		}
	}

	var userAgent string
	if params.API.Plugin != "" {
		userAgent = heartbeat.UserAgent(params.API.Plugin)
		clientOpts = append(clientOpts, api.WithUserAgent(params.API.Plugin))
	} else {
		userAgent = heartbeat.UserAgentUnknownPlugin()
		clientOpts = append(clientOpts, api.WithUserAgentUnknownPlugin())
	}

	c := api.NewClient(params.API.URL, http.DefaultClient, clientOpts...)

	heartbeats := []heartbeat.Heartbeat{
		heartbeat.New(
			params.Category,
			params.CursorPosition,
			params.Entity,
			params.EntityType,
			params.IsWrite,
			params.Language,
			params.LineNumber,
			params.LocalFile,
			params.Project.Alternate,
			params.Project.Override,
			params.Time,
			userAgent,
		),
	}

	if len(params.ExtraHeartbeats) > 0 {
		log.LogEntry.Debugf("include %d extra heartbeat(s) from stdin", len(params.ExtraHeartbeats))

		for _, h := range params.ExtraHeartbeats {
			heartbeats = append(heartbeats, heartbeat.New(
				h.Category,
				h.CursorPosition,
				h.Entity,
				h.EntityType,
				h.IsWrite,
				h.Language,
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
		offlineHandleOpt, err := offline.WithQueue(queueFilepath, params.OfflineSyncMax)
		if err != nil {
			return fmt.Errorf("failed to initialize offline queue handle option: %w", err)
		}

		handleOpts = append(handleOpts, offlineHandleOpt)
	}

	handle := heartbeat.NewHandle(c, handleOpts...)

	_, err = handle(heartbeats)
	if err != nil {
		return fmt.Errorf("failed to send heartbeats via api client: %w", err)
	}

	return nil
}

func setLogEntryWithParams(params *Params) {
	if params.API.Plugin != "" {
		log.LogEntry.Data["plugin"] = params.API.Plugin
	}

	log.LogEntry.Data["time"] = params.Time

	if params.LineNumber != nil {
		log.LogEntry.Data["lineno"] = params.LineNumber
	}

	if params.IsWrite != nil {
		log.LogEntry.Data["is_write"] = params.IsWrite
	}

	log.LogEntry.Data["file"] = params.Entity
}
