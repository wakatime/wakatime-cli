package todaygoal

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/cmd/legacy/legacyparams"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/certifi/gocertifi"
	"github.com/spf13/viper"
)

var uuid4Regex = regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$") // nolint

// Params contains today-goal command parameters.
type Params struct {
	GoalID  string
	API     legacyparams.APIParams
	Network legacyparams.NetworkParams
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

	auth, err := api.WithAuth(api.BasicAuth{
		Secret: params.API.Key,
	})
	if err != nil {
		return "", fmt.Errorf("error setting up auth option on api client: %w", err)
	}

	opts := []api.Option{
		auth,
		api.WithTimeout(params.API.Timeout),
	}

	if params.Network.DisableSSLVerify {
		opts = append(opts, api.WithDisableSSLVerify())
	}

	if !params.Network.DisableSSLVerify && params.Network.SSLCertFilepath != "" {
		withSSLCert, err := api.WithSSLCertFile(params.Network.SSLCertFilepath)
		if err != nil {
			return "", fmt.Errorf("failed to set up ssl cert file option on api client: %s", err)
		}

		opts = append(opts, withSSLCert)
	} else if !params.Network.DisableSSLVerify {
		certPool, err := gocertifi.CACerts()
		if err != nil {
			return "", fmt.Errorf("failed to build certifi cert pool: %s", err)
		}

		withSSLCert, err := api.WithSSLCertPool(certPool)
		if err != nil {
			return "", fmt.Errorf("failed to set up ssl cert pool option on api client: %s", err)
		}

		opts = append(opts, withSSLCert)
	}

	if params.Network.ProxyURL != "" {
		withProxy, err := api.WithProxy(params.Network.ProxyURL)
		if err != nil {
			return "", fmt.Errorf("failed to set up proxy option on api client: %w", err)
		}

		opts = append(opts, withProxy)

		if strings.Contains(params.Network.ProxyURL, `\\`) {
			withNTLMRetry, err := api.WithNTLMRequestRetry(params.Network.ProxyURL)
			if err != nil {
				return "", fmt.Errorf("failed to set up ntlm request retry option on api client: %w", err)
			}

			opts = append(opts, withNTLMRetry)
		}
	}

	if params.API.Plugin != "" {
		opts = append(opts, api.WithUserAgent(params.API.Plugin))
	} else {
		opts = append(opts, api.WithUserAgentUnknownPlugin())
	}

	url := api.BaseURL
	if params.API.URL != "" {
		url = params.API.URL
	}

	c := api.NewClient(url, http.DefaultClient, opts...)

	goal, err := c.Goal(params.GoalID)
	if err != nil {
		return "", fmt.Errorf("failed fetching todays goal from api: %w", err)
	}

	return goal.Total, nil
}

// LoadParams loads todaygoal config params from viper.Viper instance. Returns ErrAuth
// if failed to retrieve api key.
func LoadParams(v *viper.Viper) (Params, error) {
	params := Params{}

	apiParams, networkParams, err := legacyparams.LoadParams(v)
	if err != nil {
		return Params{}, err
	}

	params.API = apiParams
	params.Network = networkParams

	if !v.IsSet("today-goal") {
		return Params{}, fmt.Errorf("goal id unset")
	}

	goalID := vipertools.GetString(v, "today-goal")
	if !uuid4Regex.Match([]byte(goalID)) {
		return Params{}, fmt.Errorf("goal id invalid")
	}

	params.GoalID = goalID

	return params, nil
}
