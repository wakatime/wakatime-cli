package heartbeat

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/filestats"
	"github.com/wakatime/wakatime-cli/pkg/filter"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/offline"

	_ "github.com/mattn/go-sqlite3" // not used directly
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

// Run executes the heartbeat command.
func Run(v *viper.Viper) {
	err := SendHeartbeats(v)
	if err != nil {
		var errauth api.ErrAuth
		if errors.As(err, &errauth) {
			jww.CRITICAL.Printf(
				"failed to send heartbeat: %s. Find your api key from wakatime.com/settings/api-key",
				errauth,
			)
			os.Exit(exitcode.ErrAuth)
		}

		var errapi api.Err
		if errors.As(err, &errapi) {
			jww.CRITICAL.Printf("failed to send heartbeat(s): %s", err)
			os.Exit(exitcode.ErrAPI)
		}

		jww.CRITICAL.Printf("failed to send heartbeat(s): %s", err)
		os.Exit(exitcode.ErrDefault)
	}

	jww.DEBUG.Println("successfully handled heartbeat(s)")
	os.Exit(exitcode.Success)
}

// SendHeartbeats sends a heartbeat to the wakatime api and includes additional
// heartbeats from the offline queue, if available and offline sync is not
// explicitly disabled.
func SendHeartbeats(v *viper.Viper) error {
	params, err := LoadParams(v)
	if err != nil {
		return fmt.Errorf("failed to load command parameters: %w", err)
	}

	withAuth, err := api.WithAuth(api.BasicAuth{
		Secret: params.APIKey,
	})
	if err != nil {
		return fmt.Errorf("failed to set up auth option on api client: %w", err)
	}

	clientOpts := []api.Option{
		withAuth,
		api.WithTimeout(params.Timeout),
	}

	if params.Network.DisableSSLVerify {
		clientOpts = append(clientOpts, api.WithDisableSSLVerify())
	}

	if params.Network.ProxyURL != "" {
		withProxy, err := api.WithProxy(params.Network.ProxyURL)
		if err != nil {
			return fmt.Errorf("failed to set up proxy option on api client: %w", err)
		}

		clientOpts = append(clientOpts, withProxy)
	}

	if params.Network.SSLCertFilepath != "" {
		withSSLCert, err := api.WithSSLCert(params.Network.SSLCertFilepath)
		if err != nil {
			return fmt.Errorf("failed to set up ssl cert option on api client: %w", err)
		}

		clientOpts = append(clientOpts, withSSLCert)
	}

	var userAgent string
	if params.Plugin != "" {
		userAgent = heartbeat.UserAgent(params.Plugin)
		clientOpts = append(clientOpts, api.WithUserAgent(params.Plugin))
	} else {
		userAgent = heartbeat.UserAgentUnknownPlugin()
		clientOpts = append(clientOpts, api.WithUserAgentUnknownPlugin())
	}

	c := api.NewClient(params.APIUrl, http.DefaultClient, clientOpts...)

	h := heartbeat.Heartbeat{
		Entity:         params.Entity,
		EntityType:     params.EntityType,
		Category:       params.Category,
		CursorPosition: params.CursorPosition,
		IsWrite:        params.IsWrite,
		LineNumber:     params.LineNumber,
		Time:           params.Time,
		UserAgent:      userAgent,
	}

	handleOpts := []heartbeat.HandleOption{
		filter.WithFiltering(filter.Config{
			Exclude:                    params.Filter.Exclude,
			ExcludeUnknownProject:      params.Filter.ExcludeUnknownProject,
			Include:                    params.Filter.Include,
			IncludeOnlyWithProjectFile: params.Filter.IncludeOnlyWithProjectFile,
		}),
		filestats.WithDetection(),
		heartbeat.WithSanitization(heartbeat.SanitizeConfig{
			BranchPatterns:  params.Sanitize.HideBranchNames,
			FilePatterns:    params.Sanitize.HideFileNames,
			ProjectPatterns: params.Sanitize.HideProjectNames,
		}),
	}

	if !params.OfflineDisabled {
		filepath, err := offline.QueueFilepath()
		if err != nil {
			return fmt.Errorf("failed to load offline queue filepath: %w", err)
		}

		offlineHandleOpt, err := offline.WithQueue(filepath, params.OfflineSyncMax)
		if err != nil {
			return fmt.Errorf("failed to initialize offline queue handle option: %w", err)
		}

		handleOpts = append(handleOpts, offlineHandleOpt)
	}

	handle := heartbeat.NewHandle(c, handleOpts...)

	_, err = handle([]heartbeat.Heartbeat{h})
	if err != nil {
		return fmt.Errorf("failed to send heartbeats via api client: %w", err)
	}

	return nil
}
