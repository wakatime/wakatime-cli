package todaygoal

import (
	"fmt"
	"regexp"

	cmdapi "github.com/wakatime/wakatime-cli/cmd/api"
	"github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/goal"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/output"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"
	"github.com/wakatime/wakatime-cli/pkg/wakaerror"

	"github.com/spf13/viper"
)

var uuid4Regex = regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$") // nolint

// Params contains today-goal command parameters.
type Params struct {
	GoalID string
	Output output.Output
	API    params.API
}

// Run executes the today-goal command.
func Run(v *viper.Viper) (int, error) {
	output, err := Goal(v)
	if err != nil {
		if errwaka, ok := err.(wakaerror.Error); ok {
			return errwaka.ExitCode(), fmt.Errorf("today goal fetch failed: %s", errwaka.Message())
		}

		return exitcode.ErrGeneric, fmt.Errorf(
			"today goal fetch failed: %s",
			err,
		)
	}

	log.Debugln("successfully fetched today goal")
	fmt.Println(output)

	return exitcode.Success, nil
}

// Goal returns total time of given goal id for todays coding activity.
func Goal(v *viper.Viper) (string, error) {
	params, err := LoadParams(v)
	if err != nil {
		return "", fmt.Errorf("failed to load command parameters: %w", err)
	}

	apiClient, err := cmdapi.NewClient(params.API)
	if err != nil {
		return "", fmt.Errorf("failed to initialize api client: %w", err)
	}

	g, err := apiClient.Goal(params.GoalID)
	if err != nil {
		return "", fmt.Errorf("failed fetching todays goal from api: %w", err)
	}

	output, err := goal.RenderToday(g, params.Output)
	if err != nil {
		return "", fmt.Errorf("failed generating today output: %s", err)
	}

	return output, nil
}

// LoadParams loads todaygoal config params from viper.Viper instance. Returns ErrAuth
// if failed to retrieve api key.
func LoadParams(v *viper.Viper) (Params, error) {
	paramAPI, err := params.LoadAPIParams(v)
	if err != nil {
		return Params{}, fmt.Errorf("failed to load API parameters: %w", err)
	}

	paramStatusBar, err := params.LoadStatusBarParams(v)
	if err != nil {
		return Params{}, fmt.Errorf("failed to load status bar parameters: %w", err)
	}

	if !v.IsSet("today-goal") {
		return Params{}, fmt.Errorf("goal id unset")
	}

	goalID := vipertools.GetString(v, "today-goal")
	if !uuid4Regex.Match([]byte(goalID)) {
		return Params{}, fmt.Errorf("goal id invalid")
	}

	return Params{
		GoalID: goalID,
		Output: paramStatusBar.Output,
		API:    paramAPI,
	}, nil
}
