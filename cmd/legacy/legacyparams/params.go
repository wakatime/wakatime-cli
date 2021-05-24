package legacyparams

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/spf13/viper"
)

const errMsgTemplate = "Invalid url %q. Must be in format" +
	"'https://user:pass@host:port' or " +
	"'socks5://user:pass@host:port' or " +
	"'domain\\\\user:pass.'"

var (
	// nolint
	apiKeyRegex = regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$")
	// nolint
	proxyRegex = regexp.MustCompile(`^((https?|socks5)://)?([^:@]+(:([^:@])+)?@)?[^:]+(:\d+)?$`)
	// nolint
	ntlmProxyRegex = regexp.MustCompile(`^.*\\.+$`)
)

// Params contains legacy params.
type Params struct {
	OfflineDisabled bool
	OfflineSyncMax  int
	API             API
}

// API contains api related parameters.
type API struct {
	DisableSSLVerify bool
	Key              string
	Plugin           string
	ProxyURL         string
	SSLCertFilepath  string
	Timeout          time.Duration
	URL              string
}

// String implements fmt.Stringer interface.
func (p API) String() string {
	return fmt.Sprintf(
		"api key: '%s', api url: '%s', plugin: '%s', timeout: %s, disable ssl verify: %t,"+
			"proxy url: '%s', ssl cert filepath: '%s'",
		p.Key[:4]+"...",
		p.URL,
		p.Plugin,
		p.Timeout,
		p.DisableSSLVerify,
		p.ProxyURL,
		p.SSLCertFilepath,
	)
}

// Load loads legacy params from viper.Viper instance. Returns ErrAuth
// if failed to retrieve api key.
func Load(v *viper.Viper) (Params, error) {
	if v == nil {
		return Params{}, errors.New("viper instance unset")
	}

	apiParams, err := loadAPIParams(v)
	if err != nil {
		return Params{}, fmt.Errorf("failed to load api params: %w", err)
	}

	offlineDisabled := vipertools.FirstNonEmptyBool(v, "disableoffline", "disable-offline")
	if b := v.GetBool("settings.offline"); v.IsSet("settings.offline") {
		offlineDisabled = !b
	}

	var offlineSyncMax int

	switch {
	case !v.IsSet("sync-offline-activity"):
		// use default
		offlineSyncMax = v.GetInt("sync-offline-activity")
	case vipertools.GetString(v, "sync-offline-activity") == "none":
		break
	default:
		offlineSyncMax, err = strconv.Atoi(vipertools.GetString(v, "sync-offline-activity"))
		if err != nil {
			return Params{}, errors.New("argument --sync-offline-activity must be \"none\" or a positive integer number: %s")
		}
	}

	if offlineSyncMax < 0 {
		return Params{}, errors.New("argument --sync-offline-activity must be \"none\" or a positive integer number")
	}

	return Params{
		OfflineDisabled: offlineDisabled,
		OfflineSyncMax:  offlineSyncMax,
		API:             apiParams,
	}, nil
}

func loadAPIParams(v *viper.Viper) (API, error) {
	apiKey, ok := vipertools.FirstNonEmptyString(v, "key", "settings.api_key", "settings.apikey")
	if !ok {
		return API{}, api.ErrAuth("failed to load api key")
	}

	if !apiKeyRegex.Match([]byte(apiKey)) {
		return API{}, api.ErrAuth("invalid api key format")
	}

	apiURL := api.BaseURL

	if u, ok := vipertools.FirstNonEmptyString(v, "api-url", "apiurl", "settings.api_url"); ok {
		apiURL = u
	}

	proxyURL, _ := vipertools.FirstNonEmptyString(v, "proxy", "settings.proxy")

	rgx := proxyRegex
	if strings.Contains(proxyURL, `\\`) {
		rgx = ntlmProxyRegex
	}

	if proxyURL != "" && !rgx.MatchString(proxyURL) {
		return API{}, fmt.Errorf(errMsgTemplate, proxyURL)
	}

	sslCertFilepath, _ := vipertools.FirstNonEmptyString(v, "ssl-certs-file", "settings.ssl_certs_file")

	var timeout time.Duration

	timeoutSecs, ok := vipertools.FirstNonEmptyInt(v, "timeout", "settings.timeout")
	if ok {
		timeout = time.Duration(timeoutSecs) * time.Second
	}

	return API{
		DisableSSLVerify: vipertools.FirstNonEmptyBool(v, "no-ssl-verify", "settings.no_ssl_verify"),
		Key:              apiKey,
		Plugin:           vipertools.GetString(v, "plugin"),
		ProxyURL:         proxyURL,
		SSLCertFilepath:  sslCertFilepath,
		Timeout:          timeout,
		URL:              apiURL,
	}, nil
}
