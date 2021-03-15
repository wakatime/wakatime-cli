package todaygoal

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/certifi/gocertifi"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

var (
	// nolint
	uuid4Regex = regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$")
	// nolint
	proxyRegex = regexp.MustCompile(`^((https?|socks5)://)?([^:@]+(:([^:@])+)?@)?[^:]+(:\d+)?$`)
	// nolint
	ntlmProxyRegex = regexp.MustCompile(`^.*\\.+$`)
)

// Params contains today-goal command parameters.
type Params struct {
	APIKey  string
	APIUrl  string
	Plugin  string
	Timeout time.Duration
	GoalID  string
	Network NetworkParams
}

// NetworkParams contains network related command parameters.
type NetworkParams struct {
	DisableSSLVerify bool
	ProxyURL         string
	SSLCertFilepath  string
}

// Run executes the today-goal command.
func Run(v *viper.Viper) {
	output, err := Goal(v)
	if err != nil {
		var errauth api.ErrAuth
		if errors.As(err, &errauth) {
			jww.CRITICAL.Printf(
				"%s. Find your api key from wakatime.com/settings/api-key",
				errauth,
			)
			os.Exit(exitcode.ErrAuth)
		}

		var errapi api.Err
		if errors.As(err, &errapi) {
			jww.CRITICAL.Println(err)
			os.Exit(exitcode.ErrAPI)
		}

		jww.CRITICAL.Println(err)
		os.Exit(exitcode.ErrDefault)
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
		Secret: params.APIKey,
	})
	if err != nil {
		return "", fmt.Errorf("error setting up auth option on api client: %w", err)
	}

	opts := []api.Option{
		auth,
		api.WithTimeout(params.Timeout),
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

	goal, err := c.Goal(params.GoalID)
	if err != nil {
		return "", fmt.Errorf("failed fetching todays goal from api: %w", err)
	}

	return goal.Total, nil
}

// LoadParams loads today config params from viper.Viper instance. Returns ErrAuth
// if failed to retrieve api key.
func LoadParams(v *viper.Viper) (Params, error) {
	apiKey, ok := vipertools.FirstNonEmptyString(v, "key", "settings.api_key", "settings.apikey")
	if !ok {
		return Params{}, api.ErrAuth("failed to load api key")
	}

	if !uuid4Regex.Match([]byte(apiKey)) {
		return Params{}, api.ErrAuth("api key invalid")
	}

	params := Params{
		APIKey: apiKey,
		Plugin: v.GetString("plugin"),
	}

	apiURL, ok := vipertools.FirstNonEmptyString(v, "api-url", "apiurl", "settings.api_url")
	if ok {
		params.APIUrl = apiURL
	}

	timeoutSecs, ok := vipertools.FirstNonEmptyInt(v, "timeout", "settings.timeout")
	if ok {
		params.Timeout = time.Duration(timeoutSecs) * time.Second
	}

	if !v.IsSet("today-goal") {
		return Params{}, api.ErrAuth("goal id invalid")
	}

	goalID := v.GetString("today-goal")
	if !uuid4Regex.Match([]byte(goalID)) {
		return Params{}, api.ErrAuth("goal id invalid")
	}

	params.GoalID = goalID

	networkParams, err := loadNetworkParams(v)
	if err != nil {
		return Params{}, fmt.Errorf("failed to parse network params: %s", err)
	}

	params.Network = networkParams

	return params, nil
}

func loadNetworkParams(v *viper.Viper) (NetworkParams, error) {
	if v == nil {
		return NetworkParams{}, errors.New("viper instance unset")
	}

	errMsgTemplate := "Invalid url %%q. Must be in format" +
		"'https://user:pass@host:port' or " +
		"'socks5://user:pass@host:port' or " +
		"'domain\\\\user:pass.'"

	proxyURL, _ := vipertools.FirstNonEmptyString(v, "proxy", "settings.proxy")

	rgx := proxyRegex
	if strings.Contains(proxyURL, `\\`) {
		rgx = ntlmProxyRegex
	}

	if proxyURL != "" && !rgx.MatchString(proxyURL) {
		return NetworkParams{}, fmt.Errorf(errMsgTemplate, proxyURL)
	}

	sslCertFilepath, _ := vipertools.FirstNonEmptyString(v, "ssl-certs-file", "settings.ssl_certs_file")

	return NetworkParams{
		DisableSSLVerify: vipertools.FirstNonEmptyBool(v, "no-ssl-verify", "settings.no_ssl_verify"),
		ProxyURL:         proxyURL,
		SSLCertFilepath:  sslCertFilepath,
	}, nil
}
