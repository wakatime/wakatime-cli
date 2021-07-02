package today

import (
	"errors"
	"fmt"
	"time"

	"github.com/wakatime/wakatime-cli/cmd/legacy/legacyapi"
	"github.com/wakatime/wakatime-cli/cmd/legacy/legacyparams"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/summary"

	"github.com/spf13/viper"
)

// Run executes the today command.
func Run(v *viper.Viper) (int, error) {
	output, err := Summary(v)
	if err != nil {
		var errauth api.ErrAuth
		if errors.As(err, &errauth) {
			return exitcode.ErrAuth, fmt.Errorf(
				"failed to fetch today summary: %s. Find your api key from wakatime.com/settings/api-key",
				errauth,
			)
		}

		var errapi api.Err
		if errors.As(err, &errapi) {
			return exitcode.ErrAPI, fmt.Errorf(
				"failed to fetch today summary due to api error: %s",
				err,
			)
		}

		return exitcode.ErrDefault, fmt.Errorf(
			"failed to fetch today summary: %s",
			err,
		)
	}

	log.Debugln("successfully fetched today summary")
	fmt.Println(output)

	return exitcode.Success, nil
}

// Summary returns a rendered summary of todays coding activity.
func Summary(v *viper.Viper) (string, error) {
	params, err := legacyparams.Load(v)
	if err != nil {
		return "", fmt.Errorf("failed to load command parameters: %w", err)
	}

	apiClient, err := legacyapi.NewClient(params.API)
	if err != nil {
		return "", fmt.Errorf("failed to initialize api client: %w", err)
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), now.Location())

	summaries, err := apiClient.Summaries(todayStart, todayEnd)
	if err != nil {
		return "", fmt.Errorf("failed fetching summaries from api: %w", err)
	}

	output, err := summary.RenderToday(summaries)
	if err != nil {
		return "", fmt.Errorf("failed generating today summary output: %s", err)
	}

	return output, nil
}
