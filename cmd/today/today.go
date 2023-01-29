package today

import (
	"fmt"

	cmdapi "github.com/wakatime/wakatime-cli/cmd/api"
	"github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/summary"
	"github.com/wakatime/wakatime-cli/pkg/wakaerror"

	"github.com/spf13/viper"
)

// Run executes the today command.
func Run(v *viper.Viper) (int, error) {
	output, err := Today(v)
	if err != nil {
		if errwaka, ok := err.(wakaerror.Error); ok {
			return errwaka.ExitCode(), fmt.Errorf("today fetch failed: %s", errwaka.Message())
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
	paramAPI, err := params.LoadAPIParams(v)
	if err != nil {
		return "", fmt.Errorf("failed to load API parameters: %w", err)
	}

	paramStatusBar, err := params.LoadStatusBarParams(v)
	if err != nil {
		return "", fmt.Errorf("failed to load status bar parameters: %w", err)
	}

	apiClient, err := cmdapi.NewClient(paramAPI)
	if err != nil {
		return "", fmt.Errorf("failed to initialize api client: %w", err)
	}

	s, err := apiClient.Today()
	if err != nil {
		return "", fmt.Errorf("failed fetching today from api: %w", err)
	}

	output, err := summary.RenderToday(s, paramStatusBar.HideCategories, paramStatusBar.Output)
	if err != nil {
		return "", fmt.Errorf("failed generating today output: %s", err)
	}

	return output, nil
}
