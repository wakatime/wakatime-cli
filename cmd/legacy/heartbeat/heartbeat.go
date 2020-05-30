package heartbeat

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

// Run executes the heartbeat command.
func Run(v *viper.Viper) {
	err := SendHeartbeat(v)
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
			jww.CRITICAL.Printf("failed to send heartbeat: %s", err)
			os.Exit(exitcode.ErrAPI)
		}

		jww.CRITICAL.Printf("failed to send heartbeat: %s", err)
		os.Exit(exitcode.ErrDefault)
	}

	os.Exit(exitcode.Success)
}

// SendHeartbeat sends a heartbeat to the wakatime api.
func SendHeartbeat(v *viper.Viper) error {
	params, err := LoadParams(v)
	if err != nil {
		return fmt.Errorf("failed to load command parameters: %w", err)
	}

	auth, err := api.WithAuth(api.BasicAuth{
		Secret: params.APIKey,
	})
	if err != nil {
		return fmt.Errorf("failed to set up auth option on api client: %w", err)
	}

	clientOpts := []api.Option{
		auth,
		api.WithTimeout(params.Timeout),
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
		Entity:     params.Entity,
		EntityType: params.EntityType,
		Category:   params.Category,
		Time:       params.Time,
		UserAgent:  userAgent,
		IsWrite:    params.IsWrite,
	}

	handleOpts := []heartbeat.HandleOption{}
	handle := heartbeat.NewHandle(c, handleOpts...)

	_, err = handle([]heartbeat.Heartbeat{h})
	if err != nil {
		return fmt.Errorf("failed to send heartbeats via api client: %w", err)
	}

	return nil
}
