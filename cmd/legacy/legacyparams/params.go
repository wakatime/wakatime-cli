package legacyparams

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/spf13/viper"
)

var (
	// nolint
	apiKeyRegex = regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$")
	// nolint
	proxyRegex = regexp.MustCompile(`^((https?|socks5)://)?([^:@]+(:([^:@])+)?@)?[^:]+(:\d+)?$`)
	// nolint
	ntlmProxyRegex = regexp.MustCompile(`^.*\\.+$`)
)

// APIParams contains api related command parameters.
type APIParams struct {
	Key     string
	URL     string
	Plugin  string
	Timeout time.Duration
}

// NetworkParams contains network related command parameters.
type NetworkParams struct {
	DisableSSLVerify bool
	ProxyURL         string
	SSLCertFilepath  string
}

func (p APIParams) String() string {
	return fmt.Sprintf(
		"api key: '%s', api url: '%s', plugin: '%s', timeout: %s",
		p.Key[:4]+"...",
		p.URL,
		p.Plugin,
		p.Timeout,
	)
}

func (p NetworkParams) String() string {
	return fmt.Sprintf(
		"disable ssl verify: %t, proxy url: '%s', ssl cert filepath: '%s'",
		p.DisableSSLVerify,
		p.ProxyURL,
		p.SSLCertFilepath,
	)
}

// LoadParams loads legacy params from viper.Viper instance. Returns ErrAuth
// if failed to retrieve api key.
func LoadParams(v *viper.Viper) (APIParams, NetworkParams, error) {
	if v == nil {
		return APIParams{}, NetworkParams{}, errors.New("viper instance unset")
	}

	apiKey, ok := vipertools.FirstNonEmptyString(v, "key", "settings.api_key", "settings.apikey")
	if !ok {
		return APIParams{}, NetworkParams{}, api.ErrAuth("failed to load api key")
	}

	if !apiKeyRegex.Match([]byte(apiKey)) {
		return APIParams{}, NetworkParams{}, api.ErrAuth("invalid api key format")
	}

	apiParams := APIParams{
		Key:    apiKey,
		Plugin: vipertools.GetString(v, "plugin"),
	}

	apiParams.URL = api.BaseURL

	apiURL, ok := vipertools.FirstNonEmptyString(v, "api-url", "apiurl", "settings.api_url")
	if ok {
		apiParams.URL = apiURL
	}

	timeoutSecs, ok := vipertools.FirstNonEmptyInt(v, "timeout", "settings.timeout")
	if ok {
		apiParams.Timeout = time.Duration(timeoutSecs) * time.Second
	}

	networkParams, err := loadNetworkParams(v)
	if err != nil {
		return APIParams{}, NetworkParams{}, fmt.Errorf("failed to parse network params: %s", err)
	}

	return apiParams, networkParams, nil
}

func loadNetworkParams(v *viper.Viper) (NetworkParams, error) {
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
