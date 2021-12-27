package today

import (
	"errors"
	"fmt"

	apicmd "github.com/wakatime/wakatime-cli/cmd/api"
	paramscmd "github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/summary"

	"github.com/spf13/viper"
)

// Run executes the today command.
func Run(v *viper.Viper) (int, error) {
	output, err := Today(v)
	if err != nil {
		var errauth api.ErrAuth
		if errors.As(err, &errauth) {
			return exitcode.ErrAuth, fmt.Errorf(
				"today fetch failed: invalid api key... find yours at wakatime.com/api-key. %s",
				errauth,
			)
		}

		var errapi api.Err
		if errors.As(err, &errapi) {
			return exitcode.ErrAPI, fmt.Errorf(
				"today fetch failed: api error: %s",
				err,
			)
		}

		var errbadRequest api.ErrBadRequest
		if errors.As(err, &errbadRequest) {
			return exitcode.ErrGeneric, fmt.Errorf(
				"today fetch failed: bad request: %s",
				err,
			)
		}

		return exitcode.ErrGeneric, fmt.Errorf(
			"today fetch failed: %s",
			err,
		)
	}

	log.Debugln("successfully fetched today for status bar")
	fmt.Println(output)

	return exitcode.Success, nil
}

// Today returns a rendered summary of todays coding activity.
func Today(v *viper.Viper) (string, error) {
	params, err := paramscmd.Load(v, paramscmd.Config{APIKeyRequired: true})
	if err != nil {
		return "", fmt.Errorf("failed to load command parameters: %w", err)
	}

	apiClient, err := apicmd.NewClient(params.API)
	if err != nil {
		return "", fmt.Errorf("failed to initialize api client: %w", err)
	}

	s, err := apiClient.Today()
	if err != nil {
		return "", fmt.Errorf("failed fetching today from api: %w", err)
	}

	output, err := summary.RenderToday(s, params.StatusBar.HideCategories)
	if err != nil {
		return "", fmt.Errorf("failed generating today output: %s", err)
	}

	return output, nil
}
