package today

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/summary"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

// nolint
var apiKeyRegex = regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$")

// Params contains today command parameters.
type Params struct {
	APIKey  string
	APIUrl  string
	Plugin  string
	Timeout time.Duration
}

// Run executes the today command.
func Run(v *viper.Viper) {
	params, err := LoadParams(v)
	if err != nil {
		var errauth api.ErrAuth
		if errors.As(err, &errauth) {
			jww.CRITICAL.Printf(
				"error loading api key: %s. Find your api key from wakatime.com/settings/api-key",
				err,
			)
			os.Exit(exitcode.ErrAuth)
		}

		jww.CRITICAL.Printf("failed loading params: %s", err)
		os.Exit(exitcode.ErrDefault)
	}

	auth, err := api.WithAuth(api.BasicAuth{
		Secret: params.APIKey,
	})
	if err != nil {
		jww.CRITICAL.Printf(
			"error setting up auth option on api client: %s. Find your api key from wakatime.com/settings/api-key",
			err,
		)
		os.Exit(exitcode.ErrAuth)
	}

	opts := []api.Option{
		auth,
		api.WithTimeout(params.Timeout),
	}

	if params.Plugin != "" {
		opts = append(opts, api.WithUserAgent(params.Plugin))
	} else {
		opts = append(opts, api.WithUserAgentUnknownPlugin())
	}

	url := api.BaseURL
	if params.APIUrl != "" {
		url = params.APIUrl
	}

	c := api.NewClient(url, http.DefaultClient, opts...)

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), now.Location())

	summaries, err := c.Summaries(todayStart, todayEnd)
	if err != nil {
		var errapi api.Err
		if errors.As(err, &errapi) {
			jww.CRITICAL.Printf("failed fetching summaries from api: %s", err)
			os.Exit(exitcode.ErrAPI)
		}

		var errauth api.ErrAuth
		if errors.As(err, &errauth) {
			jww.CRITICAL.Printf("authentication error: %s. Find your api key from wakatime.com/settings/api-key", err)
			os.Exit(exitcode.ErrAuth)
		}

		jww.CRITICAL.Printf("failed fetching summaries from api: %s", err)
		os.Exit(exitcode.ErrDefault)
	}

	output, err := summary.RenderToday(summaries)
	if err != nil {
		jww.CRITICAL.Printf("failed generating today summary output: %s", err)
		os.Exit(exitcode.ErrDefault)
	}

	fmt.Println(output)
	os.Exit(exitcode.Success)
}

// LoadParams loads needed data from the configuration file. Returns ErrAuth on problems with api key.
func LoadParams(v *viper.Viper) (Params, error) {
	apiKey := firstNonEmptyString(v, "key", "settings.api_key", "settings.apikey")
	if apiKey == "" {
		return Params{}, api.ErrAuth("api key unset")
	}

	if !apiKeyRegex.Match([]byte(apiKey)) {
		return Params{}, api.ErrAuth("api key invalid")
	}

	var (
		apiURL      = firstNonEmptyString(v, "api-url", "apiurl", "settings.api_url")
		timeoutSecs = firstNonEmptyInt(v, "timeout", "settings.timeout")
	)

	return Params{
		APIKey:  apiKey,
		APIUrl:  apiURL,
		Plugin:  v.GetString("plugin"),
		Timeout: time.Duration(timeoutSecs) * time.Second,
	}, nil
}

func firstNonEmptyInt(v *viper.Viper, keys ...string) int {
	for _, key := range keys {
		if value := v.GetInt(key); value != 0 {
			return value
		}
	}

	return 0
}
func firstNonEmptyString(v *viper.Viper, keys ...string) string {
	for _, key := range keys {
		if value := v.GetString(key); value != "" {
			return value
		}
	}

	return ""
}
