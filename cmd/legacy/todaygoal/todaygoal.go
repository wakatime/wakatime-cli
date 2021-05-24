package todaygoal

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/wakatime/wakatime-cli/cmd/legacy/legacyapi"
	"github.com/wakatime/wakatime-cli/cmd/legacy/legacyparams"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/spf13/viper"
)

var uuid4Regex = regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$") // nolint

// Params contains today-goal command parameters.
type Params struct {
	GoalID string
	API    legacyparams.API
}

// Run executes the today-goal command.
func Run(v *viper.Viper) {
	output, err := Goal(v)
	if err != nil {
		var errauth api.ErrAuth
		if errors.As(err, &errauth) {
			log.Errorf(
				"%s. Find your api key from wakatime.com/settings/api-key",
				errauth,
			)
			os.Exit(exitcode.ErrAuth)
		}

		var errapi api.Err
		if errors.As(err, &errapi) {
			log.Errorln(err)
			os.Exit(exitcode.ErrAPI)
		}

		log.Fatalln(err)
	}

	fmt.Println(output)
	os.Exit(exitcode.Success)
}

// Goal returns total time of given goal id for todays coding activity.
func Goal(v *viper.Viper) (string, error) {
	params, err := LoadParams(v)
	if err != nil {
		return "", fmt.Errorf("failed to load command parameters: %w", err)
	}

	apiClient, err := legacyapi.NewClient(params.API)
	if err != nil {
		return "", fmt.Errorf("failed to initialize api client: %w", err)
	}

	goal, err := apiClient.Goal(params.GoalID)
	if err != nil {
		return "", fmt.Errorf("failed fetching todays goal from api: %w", err)
	}

	return goal.Total, nil
}

// LoadParams loads todaygoal config params from viper.Viper instance. Returns ErrAuth
// if failed to retrieve api key.
func LoadParams(v *viper.Viper) (Params, error) {
	params, err := legacyparams.Load(v)
	if err != nil {
		return Params{}, fmt.Errorf("failed to load params: %w", err)
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
		API:    params.API,
	}, nil
}
