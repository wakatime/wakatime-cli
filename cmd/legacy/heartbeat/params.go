package heartbeat

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/spf13/viper"
)

var (
	// nolint
	apiKeyRegex = regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$")
	// nolint
	matchAllRegex = regexp.MustCompile(".*")
	// nolint
	proxyRegex = regexp.MustCompile(`^((https?|socks5)://)?([^:@]+(:([^:@])+)?@)?[^:]+(:\d+)?$`)
)

// Params contains heartbeat command parameters.
type Params struct {
	APIKey     string
	APIUrl     string
	Category   heartbeat.Category
	Entity     string
	EntityType heartbeat.EntityType
	Hostname   string
	IsWrite    *bool
	Plugin     string
	Time       float64
	Timeout    time.Duration
	Network    NetworkParams
	Sanitize   SanitizeParams
}

// NetworkParams contains network related command parameters.
type NetworkParams struct {
	DisableSSLVerify bool
	ProxyURL         string
	SSLCertFilepath  string
}

// SanitizeParams params for heartbeat sanitization.
type SanitizeParams struct {
	HideBranchNames  []*regexp.Regexp
	HideFileNames    []*regexp.Regexp
	HideProjectNames []*regexp.Regexp
}

// LoadParams loads heartbeat config params from viper.Viper instance. Returns ErrAuth
// if failed to retrieve api key.
func LoadParams(v *viper.Viper) (Params, error) {
	apiKey, ok := vipertools.FirstNonEmptyString(v, "key", "settings.api_key", "settings.apikey")
	if !ok {
		return Params{}, api.ErrAuth("failed to load api key")
	}

	if !apiKeyRegex.Match([]byte(apiKey)) {
		return Params{}, api.ErrAuth("invalid api key format")
	}

	apiURL := api.BaseURL
	if url, ok := vipertools.FirstNonEmptyString(v, "api-url", "apiurl", "settings.api_url"); ok {
		apiURL = url
	}

	var category heartbeat.Category

	if categoryStr := v.GetString("category"); categoryStr != "" {
		parsed, err := heartbeat.ParseCategory(categoryStr)
		if err != nil {
			return Params{}, fmt.Errorf("failed to parse category: %s", err)
		}

		category = parsed
	}

	entity, ok := vipertools.FirstNonEmptyString(v, "entity", "file")
	if !ok {
		return Params{}, errors.New("failed to retrieve entity")
	}

	var entityType heartbeat.EntityType

	if entityTypeStr := v.GetString("entity-type"); entityTypeStr != "" {
		parsed, err := heartbeat.ParseEntityType(entityTypeStr)
		if err != nil {
			return Params{}, fmt.Errorf("failed to parse entity type: %s", err)
		}

		entityType = parsed
	}

	var err error

	hostname, ok := vipertools.FirstNonEmptyString(v, "hostname", "settings.hostname")
	if !ok {
		hostname, err = os.Hostname()
		if err != nil {
			return Params{}, fmt.Errorf("failed to retrieve hostname from system: %s", err)
		}
	}

	var isWrite *bool
	if b := v.GetBool("write"); v.IsSet("write") {
		isWrite = heartbeat.Bool(b)
	}

	timeSecs := v.GetFloat64("time")
	if timeSecs == 0 {
		timeSecs = float64(time.Now().UnixNano()) / 1000000000
	}

	var timeout time.Duration

	timeoutSecs, ok := vipertools.FirstNonEmptyInt(v, "timeout", "settings.timeout")
	if ok {
		timeout = time.Duration(timeoutSecs) * time.Second
	}

	networkParams, err := loadNetworkParams(v)
	if err != nil {
		return Params{}, fmt.Errorf("failed to parse network params: %s", err)
	}

	sanitizeParams, err := loadSanitizeParams(v)
	if err != nil {
		return Params{}, fmt.Errorf("failed to load sanitize params: %s", err)
	}

	return Params{
		APIKey:     apiKey,
		APIUrl:     apiURL,
		Category:   category,
		Entity:     entity,
		EntityType: entityType,
		Hostname:   hostname,
		IsWrite:    isWrite,
		Plugin:     v.GetString("plugin"),
		Time:       timeSecs,
		Timeout:    timeout,
		Network:    networkParams,
		Sanitize:   sanitizeParams,
	}, nil
}

func loadNetworkParams(v *viper.Viper) (NetworkParams, error) {
	if v == nil {
		return NetworkParams{}, errors.New("viper instance unset")
	}

	errMsgTemplate := "Invalid url %%q. Must be in format" +
		"'https://user:pass@host:port' or " +
		"'socks5://user:pass@host:port' or " +
		"'domain\\user:pass.'"

	proxyURL, _ := vipertools.FirstNonEmptyString(v, "proxy", "settings.proxy")
	if proxyURL != "" && !proxyRegex.MatchString(proxyURL) {
		return NetworkParams{}, fmt.Errorf(errMsgTemplate, proxyURL)
	}

	sslCertFilepath, _ := vipertools.FirstNonEmptyString(v, "ssl-certs-file", "settings.ssl_certs_file")

	return NetworkParams{
		DisableSSLVerify: vipertools.FirstNonEmptyBool(v, "no-ssl-verify", "settings.no_ssl_verify"),
		ProxyURL:         proxyURL,
		SSLCertFilepath:  sslCertFilepath,
	}, nil
}

func loadSanitizeParams(v *viper.Viper) (SanitizeParams, error) {
	var params SanitizeParams

	// hide branch names
	hideBranchNamesStr, _ := vipertools.FirstNonEmptyString(
		v,
		"hide-branch-names",
		"settings.hide_branch_names",
		"settings.hide_branchnames",
		"settings.hidebranchnames",
	)

	hideBranchNamesPatterns, err := parseBoolOrRegexList(hideBranchNamesStr)
	if err != nil {
		return SanitizeParams{}, fmt.Errorf(
			"failed to parse regex hide branch names param %q: %s",
			hideBranchNamesStr,
			err,
		)
	}

	params.HideBranchNames = hideBranchNamesPatterns

	// hide project names
	hideProjectNamesStr, _ := vipertools.FirstNonEmptyString(
		v,
		"hide-project-names",
		"settings.hide_project_names",
		"settings.hide_projectnames",
		"settings.hideprojectnames",
	)

	hideProjectNamesPatterns, err := parseBoolOrRegexList(hideProjectNamesStr)
	if err != nil {
		return SanitizeParams{}, fmt.Errorf(
			"failed to parse regex hide project names param %q: %s",
			hideProjectNamesStr,
			err,
		)
	}

	params.HideProjectNames = hideProjectNamesPatterns

	// hide file names
	hideFileNamesStr, _ := vipertools.FirstNonEmptyString(
		v,
		"hide-file-names",
		"hide-filenames",
		"hidefilenames",
		"settings.hide_file_names",
		"settings.hide_filenames",
		"settings.hidefilenames",
	)

	hideFileNamesPatterns, err := parseBoolOrRegexList(hideFileNamesStr)
	if err != nil {
		return SanitizeParams{}, fmt.Errorf(
			"failed to parse regex hide file names param %q: %s",
			hideFileNamesStr,
			err,
		)
	}

	params.HideFileNames = hideFileNamesPatterns

	return params, nil
}

func parseBoolOrRegexList(s string) ([]*regexp.Regexp, error) {
	var patterns []*regexp.Regexp

	switch {
	case s == "":
		break
	case strings.ToLower(s) == "false":
		break
	case strings.ToLower(s) == "true":
		patterns = []*regexp.Regexp{matchAllRegex}
	default:
		splitted := strings.Split(s, "\n")
		for _, s := range splitted {
			compiled, err := regexp.Compile(s)
			if err != nil {
				return nil, fmt.Errorf("failed to compile regex %q: %s", s, err)
			}

			patterns = append(patterns, compiled)
		}
	}

	return patterns, nil
}
