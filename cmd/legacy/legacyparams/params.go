package legacyparams

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/config"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const errMsgTemplate = "invalid url %q. Must be in format" +
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
	OfflineDisabled         bool
	OfflineQueueFile        string
	OfflineSyncMax          int
	StatusBarHideCategories bool
	API                     API
}

// API contains api related parameters.
type API struct {
	BackoffAt        time.Time
	BackoffRetries   int
	DisableSSLVerify bool
	Hostname         string
	Key              string
	Plugin           string
	ProxyURL         string
	SSLCertFilepath  string
	Timeout          time.Duration
	URL              string
}

// String implements fmt.Stringer interface.
func (p API) String() string {
	var backoffAt string
	if !p.BackoffAt.IsZero() {
		backoffAt = p.BackoffAt.Format(config.DateFormat)
	}

	return fmt.Sprintf(
		"api key: '%s', api url: '%s', backoff at: '%s', backoff retries: %d,"+
			" hostname: '%s', plugin: '%s', timeout: %s, disable ssl verify: %t,"+
			" proxy url: '%s', ssl cert filepath: '%s'",
		p.Key[:4]+"...",
		p.URL,
		backoffAt,
		p.BackoffRetries,
		p.Hostname,
		p.Plugin,
		p.Timeout,
		p.DisableSSLVerify,
		p.ProxyURL,
		p.SSLCertFilepath,
	)
}

// Load loads legacy params from viper.Viper instance. Returns ErrAuth
// if failed to retrieve api key.
func Load(v *viper.Viper, apiKeyRequired bool) (Params, error) {
	if v == nil {
		return Params{}, errors.New("viper instance unset")
	}

	apiParams, err := loadAPIParams(v, apiKeyRequired)
	if err != nil {
		return Params{}, fmt.Errorf("failed to load api params: %w", err)
	}

	offlineDisabled := vipertools.FirstNonEmptyBool(v, "disable-offline", "disableoffline")
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

	statusBarHideCategories := vipertools.FirstNonEmptyBool(
		v,
		"today-hide-categories",
		"settings.status_bar_hide_categories",
	)

	return Params{
		OfflineDisabled:         offlineDisabled,
		OfflineQueueFile:        vipertools.GetString(v, "offline-queue-file"),
		OfflineSyncMax:          offlineSyncMax,
		StatusBarHideCategories: statusBarHideCategories,
		API:                     apiParams,
	}, nil
}

func loadAPIParams(v *viper.Viper, apiKeyRequired bool) (API, error) {
	apiKey, ok := vipertools.FirstNonEmptyString(v, "key", "settings.api_key", "settings.apikey")
	if !ok && apiKeyRequired {
		return API{}, api.ErrAuth("failed to load api key")
	}

	if apiKeyRequired && !apiKeyRegex.Match([]byte(apiKey)) {
		return API{}, api.ErrAuth("invalid api key format")
	}

	apiURL := api.BaseURL

	if u, ok := vipertools.FirstNonEmptyString(v, "api-url", "apiurl", "settings.api_url"); ok {
		apiURL = u
	}

	// remove endpoint from api base url to support legacy api_url param
	apiURL = strings.TrimSuffix(apiURL, "/")
	apiURL = strings.TrimSuffix(apiURL, ".bulk")
	apiURL = strings.TrimSuffix(apiURL, "/users/current/heartbeats")
	apiURL = strings.TrimSuffix(apiURL, "/heartbeats")
	apiURL = strings.TrimSuffix(apiURL, "/heartbeat")

	var backoffAt time.Time

	backoffAtStr := vipertools.GetString(v, "internal.backoff_at")
	if backoffAtStr != "" {
		parsed, err := time.Parse(config.DateFormat, backoffAtStr)
		if err != nil {
			log.Warnf("failed to parse backoff_at: %s", err)
		} else {
			backoffAt = parsed
		}
	}

	backoffRetries, _ := vipertools.FirstNonEmptyInt(v, "internal.backoff_retries")

	var (
		hostname string
		err      error
	)

	hostname, ok = vipertools.FirstNonEmptyString(v, "hostname", "settings.hostname")
	if !ok {
		hostname, err = os.Hostname()
		if err != nil {
			return API{}, fmt.Errorf("failed to retrieve hostname from system: %s", err)
		}
	}

	proxyURL, _ := vipertools.FirstNonEmptyString(v, "proxy", "settings.proxy")

	rgx := proxyRegex
	if strings.Contains(proxyURL, `\\`) {
		rgx = ntlmProxyRegex
	}

	if proxyURL != "" && !rgx.MatchString(proxyURL) {
		return API{}, fmt.Errorf(errMsgTemplate, proxyURL)
	}

	var sslCertFilepath string

	sslCertFilepath, ok = vipertools.FirstNonEmptyString(v, "ssl-certs-file", "settings.ssl_certs_file")
	if ok {
		sslCertFilepath, err = homedir.Expand(sslCertFilepath)
		if err != nil {
			if err != nil {
				return API{},
					fmt.Errorf("failed expanding ssl certs file: %s", err)
			}
		}
	}

	var timeout time.Duration

	timeoutSecs, ok := vipertools.FirstNonEmptyInt(v, "timeout", "settings.timeout")
	if ok {
		timeout = time.Duration(timeoutSecs) * time.Second
	}

	return API{
		BackoffAt:        backoffAt,
		BackoffRetries:   backoffRetries,
		DisableSSLVerify: vipertools.FirstNonEmptyBool(v, "no-ssl-verify", "settings.no_ssl_verify"),
		Hostname:         hostname,
		Key:              apiKey,
		Plugin:           vipertools.GetString(v, "plugin"),
		ProxyURL:         proxyURL,
		SSLCertFilepath:  sslCertFilepath,
		Timeout:          timeout,
		URL:              apiURL,
	}, nil
}
