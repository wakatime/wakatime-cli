package params

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/apikey"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/output"
	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/regex"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"golang.org/x/net/http/httpproxy"
)

const (
	errMsgTemplate = "invalid url %q. Must be in format" +
		"'https://user:pass@host:port' or " +
		"'socks5://user:pass@host:port' or " +
		"'domain\\\\user:pass.'"
	gitpodHostname = "Gitpod"
)

var (
	// nolint
	apiKeyRegex = regexp.MustCompile("^(waka_)?[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$")
	// nolint
	matchAllRegex = regexp.MustCompile(".*")
	// nolint
	matchNoneRegex = regexp.MustCompile("a^")
	// nolint
	ntlmProxyRegex = regexp.MustCompile(`^.*\\.+$`)
	// nolint
	proxyRegex = regexp.MustCompile(`^((https?|socks5)://)?([^:@]+(:([^:@])+)?@)?([^:]+|(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])))(:\d+)?$`)
)

// FlagReadOrder defines the order in which flags are read when loading parameters.
type FlagReadOrder uint8

const (
	// FlagReadOrderFlagPrecedence defines the order where command line flags take precedence.
	FlagReadOrderFlagPrecedence FlagReadOrder = iota
	// FlagReadOrderProjectConfigPrecedence defines the order where project config takes precedence over command line flags.
	FlagReadOrderProjectConfigPrecedence
)

// These variables are used to define the order in which parameters are read from the viper instance.
// It's needed when a project config file is used, so we can read the parameters in the correct order.
// nolint:gochecknoglobals
var (
	apiKeyOrder = map[FlagReadOrder][]string{
		FlagReadOrderFlagPrecedence: {
			"key",
			"settings.api_key",
			"api_key",
			"settings.apikey",
			"apikey",
		},
		FlagReadOrderProjectConfigPrecedence: {
			"settings.api_key",
			"api_key",
			"key",
			"settings.apikey",
			"apikey",
		},
	}
	guessLanguageOrder = map[FlagReadOrder][]string{
		FlagReadOrderFlagPrecedence:          {"guess-language", "settings.guess_language"},
		FlagReadOrderProjectConfigPrecedence: {"settings.guess_language", "guess-language"},
	}
	excludeUnknownProjectOrder = map[FlagReadOrder][]string{
		FlagReadOrderFlagPrecedence:          {"exclude-unknown-project", "settings.exclude_unknown_project"},
		FlagReadOrderProjectConfigPrecedence: {"settings.exclude_unknown_project", "exclude-unknown-project"},
	}
	includeOnlyWithProjectFileOrder = map[FlagReadOrder][]string{
		FlagReadOrderFlagPrecedence:          {"include-only-with-project-file", "settings.include_only_with_project_file"},
		FlagReadOrderProjectConfigPrecedence: {"settings.include_only_with_project_file", "include-only-with-project-file"},
	}
	hideBranchNamesOrder = map[FlagReadOrder][]string{
		FlagReadOrderFlagPrecedence: {
			"hide-branch-names",
			"settings.hide_branch_names",
			"settings.hide_branchnames",
			"settings.hidebranchnames",
		},
		FlagReadOrderProjectConfigPrecedence: {
			"settings.hide_branch_names",
			"settings.hide_branchnames",
			"settings.hidebranchnames",
			"hide-branch-names",
		},
	}
	hideDependenciesOrder = map[FlagReadOrder][]string{
		FlagReadOrderFlagPrecedence:          {"hide-dependencies", "settings.hide_dependencies"},
		FlagReadOrderProjectConfigPrecedence: {"settings.hide_dependencies", "hide-dependencies"},
	}
	hideProjectNamesOrder = map[FlagReadOrder][]string{
		FlagReadOrderFlagPrecedence: {
			"hide-project-names",
			"settings.hide_project_names",
			"settings.hide_projectnames",
			"settings.hideprojectnames"},
		FlagReadOrderProjectConfigPrecedence: {
			"settings.hide_project_names",
			"settings.hide_projectnames",
			"settings.hideprojectnames",
			"hide-project-names",
		},
	}
	hideFileNamesOrder = map[FlagReadOrder][]string{
		FlagReadOrderFlagPrecedence: {
			"hide-file-names",
			"hide-filenames",
			"hidefilenames",
			"settings.hide_file_names",
			"settings.hide_filenames",
			"settings.hidefilenames",
		},
		FlagReadOrderProjectConfigPrecedence: {
			"settings.hide_file_names",
			"settings.hide_filenames",
			"settings.hidefilenames",
			"hide-file-names",
			"hide-filenames",
			"hidefilenames",
		},
	}
	hideProjectFolderOrder = map[FlagReadOrder][]string{
		FlagReadOrderFlagPrecedence:          {"hide-project-folder", "settings.hide_project_folder"},
		FlagReadOrderProjectConfigPrecedence: {"settings.hide_project_folder", "hide-project-folder"},
	}
	disableOfflineOrder = map[FlagReadOrder][]string{
		FlagReadOrderFlagPrecedence:          {"disable-offline", "disableoffline", "settings.offline"},
		FlagReadOrderProjectConfigPrecedence: {"settings.offline", "disable-offline", "disableoffline"},
	}
	heartbeatRateLimitOrder = map[FlagReadOrder][]string{
		FlagReadOrderFlagPrecedence:          {"heartbeat-rate-limit-seconds", "settings.heartbeat_rate_limit_seconds"},
		FlagReadOrderProjectConfigPrecedence: {"settings.heartbeat_rate_limit_seconds", "heartbeat-rate-limit-seconds"},
	}
	todayCategoriesOrder = map[FlagReadOrder][]string{
		FlagReadOrderFlagPrecedence:          {"today-hide-categories", "settings.status_bar_show_categories"},
		FlagReadOrderProjectConfigPrecedence: {"settings.status_bar_show_categories", "today-hide-categories"},
	}
	todayMaxCategoriesOrder = map[FlagReadOrder][]string{
		FlagReadOrderFlagPrecedence:          {"today-max-categories", "settings.status_bar_max_categories"},
		FlagReadOrderProjectConfigPrecedence: {"settings.status_bar_max_categories", "today-max-categories"},
	}
)

type (
	// Params contains params.
	Params struct {
		AI        AIParams
		API       API
		Heartbeat Heartbeat
		Offline   Offline
		StatusBar StatusBar
	}

	// API contains api related parameters.
	API struct {
		BackoffAt        time.Time
		BackoffRetries   int
		DisableSSLVerify bool
		Hostname         string
		Key              string
		KeyPatterns      []apikey.MapPattern
		Plugin           string
		ProxyURL         string
		SSLCertFilepath  string
		Timeout          time.Duration
		URL              string
		URLPatterns      []apikey.URLPattern
	}

	// ExtraHeartbeat contains extra heartbeat.
	ExtraHeartbeat struct {
		AILineChanges     any     `json:"ai_line_changes"`
		BranchAlternate   string  `json:"alternate_branch"`
		Category          string  `json:"category"`
		CursorPosition    any     `json:"cursorpos"`
		Entity            string  `json:"entity"`
		EntityType        string  `json:"entity_type"`
		HumanLineChanges  any     `json:"human_line_changes"`
		IsUnsavedEntity   any     `json:"is_unsaved_entity"`
		IsWrite           any     `json:"is_write"`
		Language          *string `json:"language"`
		LanguageAlternate string  `json:"alternate_language"`
		LineNumber        any     `json:"lineno"`
		Lines             any     `json:"lines"`
		Project           string  `json:"project"`
		ProjectAlternate  string  `json:"alternate_project"`
		Time              any     `json:"time"`
		Timestamp         any     `json:"timestamp"`
		Type              string  `json:"type"`
	}

	// Heartbeat contains heartbeat command parameters.
	Heartbeat struct {
		AIParams          AIParams
		AILineChanges     *int
		Category          heartbeat.Category
		CursorPosition    *int
		Entity            string
		EntityType        heartbeat.EntityType
		ExtraHeartbeats   []heartbeat.Heartbeat
		GuessLanguage     bool
		HumanLineChanges  *int
		IsUnsavedEntity   bool
		IsWrite           *bool
		Language          *string
		LanguageAlternate string
		LineNumber        *int
		LinesInFile       *int
		LocalFile         string
		Time              float64
		Filter            FilterParams
		Project           ProjectParams
		Sanitize          SanitizeParams
	}

	// FilterParams contains heartbeat filtering related command parameters.
	FilterParams struct {
		Exclude                    []regex.Regex
		ExcludeUnknownProject      bool
		Include                    []regex.Regex
		IncludeOnlyWithProjectFile bool
	}

	// AIParams contains AI sync command parameters.
	AIParams struct {
		CodexSource  string
		SyncDisabled bool
	}

	// Offline contains offline related parameters.
	Offline struct {
		Disabled   bool
		LastSentAt time.Time
		PrintMax   int
		RateLimit  time.Duration
		SyncMax    int
	}

	// ProjectParams params for project name sanitization.
	ProjectParams struct {
		Alternate            string
		BranchAlternate      string
		MapPatterns          []project.MapPattern
		Override             string
		ProjectFromGitRemote bool
		SubmodulesDisabled   []regex.Regex
		SubmoduleMapPatterns []project.MapPattern
	}

	// SanitizeParams params for heartbeat sanitization.
	SanitizeParams struct {
		HideBranchNames     []regex.Regex
		HideDependencies    []regex.Regex
		HideFileNames       []regex.Regex
		HideProjectFolder   bool
		HideProjectNames    []regex.Regex
		ProjectPathOverride string
	}

	// StatusBar contains status bar related parameters.
	StatusBar struct {
		HideCategories bool
		HideMinutes    bool
		MaxCategories  int
		Output         output.Output
	}
)

// LoadAPIParams loads API params from viper.Viper instance. Returns ErrAuth
// if failed to retrieve api key.
func LoadAPIParams(ctx context.Context, v *viper.Viper, order FlagReadOrder) (API, error) {
	apiKey, err := loadAPIKey(ctx, v, order)
	if err != nil {
		return API{}, err
	}

	apiURL, err := loadAPIURL(v)
	if err != nil {
		return API{}, err
	}

	apiKeyPatterns, err := loadAPIKeyPatterns(ctx, v, apiKey)
	if err != nil {
		return API{}, err
	}

	apiURLPatterns, err := loadAPIURLPatterns(ctx, v, apiURL, apiKey)
	if err != nil {
		return API{}, err
	}

	backoffAt, backoffRetries := loadBackoffParams(ctx, v)

	hostname := loadHostname(ctx, v)

	proxyURL, err := loadProxyURL(ctx, v, apiURL)
	if err != nil {
		return API{}, err
	}

	sslCertFilepath := vipertools.FirstNonEmptyString(v, "ssl-certs-file", "settings.ssl_certs_file")
	if sslCertFilepath != "" {
		sslCertFilepath, err = homedir.Expand(sslCertFilepath)
		if err != nil {
			return API{}, api.ErrAuth{Err: fmt.Errorf("failed expanding ssl certs file: %s", err)}
		}
	}

	timeout := api.DefaultTimeoutSecs

	if timeoutSecs, ok := vipertools.FirstNonEmptyInt(v, "timeout", "settings.timeout"); ok {
		timeout = timeoutSecs
	}

	return API{
		BackoffAt:        backoffAt,
		BackoffRetries:   backoffRetries,
		DisableSSLVerify: vipertools.FirstNonEmptyBool(v, "no-ssl-verify", "settings.no_ssl_verify"),
		Hostname:         hostname,
		Key:              apiKey,
		KeyPatterns:      apiKeyPatterns,
		Plugin:           vipertools.GetString(v, "plugin"),
		ProxyURL:         proxyURL,
		SSLCertFilepath:  sslCertFilepath,
		Timeout:          time.Duration(timeout) * time.Second,
		URL:              apiURL.String(),
		URLPatterns:      apiURLPatterns,
	}, nil
}

func loadAPIKeyPatterns(ctx context.Context, v *viper.Viper, defaultAPIKey string) ([]apikey.MapPattern, error) {
	logger := log.Extract(ctx)

	var patterns []apikey.MapPattern

	apiKeyMap := vipertools.GetStringMapString(v, "project_api_key")

	for k, s := range apiKeyMap {
		// make all regex case insensitive
		if !strings.HasPrefix(k, "(?i)") {
			k = "(?i)" + k
		}

		compiled, err := regex.Compile(k)
		if err != nil {
			logger.Warnf("failed to compile project_api_key regex pattern %q", k)
			continue
		}

		if !apiKeyRegex.MatchString(s) {
			return nil, api.ErrAuth{Err: fmt.Errorf("invalid api key format for %q", k)}
		}

		if s == defaultAPIKey {
			continue
		}

		patterns = append(patterns, apikey.MapPattern{
			APIKey: s,
			Regex:  compiled,
		})
	}

	return patterns, nil
}

func loadAPIURLPatterns(
	ctx context.Context,
	v *viper.Viper,
	defaultURL *url.URL,
	defaultKey string,
) ([]apikey.URLPattern, error) {
	logger := log.Extract(ctx)

	var patterns []apikey.URLPattern

	apiURLMap := vipertools.GetStringMapString(v, "api_urls")

	for k, s := range apiURLMap {
		// make all regex case insensitive
		if !strings.HasPrefix(k, "(?i)") {
			k = "(?i)" + k
		}

		compiled, err := regex.Compile(k)
		if err != nil {
			logger.Warnf("failed to compile api_urls regex pattern %q", k)
			continue
		}

		// split value by | to get api_url and api_key
		parts := strings.SplitN(s, "|", 2)
		if len(parts) > 2 {
			logger.Warnf("invalid api_urls format for %q, expected 'api_url|api_key'", k)
			continue
		}

		apiURLValue := defaultURL.String()
		apiKeyValue := defaultKey

		if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
			apiURLValue = strings.TrimSpace(parts[0])
		}

		if len(parts) == 2 {
			apiKeyValue = strings.TrimSpace(parts[1])
		}

		if apiURLValue == "" {
			logger.Warnf("empty api_url for %q", k)
			continue
		}

		if !apiKeyRegex.MatchString(apiKeyValue) {
			return nil, api.ErrAuth{Err: fmt.Errorf("invalid api key format in api_urls for %q", k)}
		}

		apiURLValue, err = normalizeURL(apiURLValue)
		if err != nil {
			logger.Warnf("invalid api_url format for %q: %s", k, err)
			continue
		}

		patterns = append(patterns, apikey.URLPattern{
			APIURL: apiURLValue,
			APIKey: apiKeyValue,
			Regex:  compiled,
		})
	}

	return patterns, nil
}

func loadAPIURL(v *viper.Viper) (*url.URL, error) {
	apiURLStr := api.BaseURL

	if u := vipertools.FirstNonEmptyString(v, "api-url", "apiurl", "api_url", "settings.api_url"); u != "" {
		apiURLStr = u
	}

	apiURLStr, err := normalizeURL(apiURLStr)
	if err != nil {
		return nil, api.ErrAuth{Err: fmt.Errorf("invalid api url: %s", err)}
	}

	apiURL, err := url.Parse(apiURLStr)
	if err != nil {
		return nil, api.ErrAuth{Err: fmt.Errorf("failed to parse api url: %s", err)}
	}

	return apiURL, nil
}

func loadBackoffParams(ctx context.Context, v *viper.Viper) (backoffAt time.Time, backoffRetries int) {
	logger := log.Extract(ctx)

	backoffAtStr := vipertools.GetString(v, "internal.backoff_at")
	if backoffAtStr != "" {
		parsed, err := vipertools.SafeTimeParse(ini.DateFormat, backoffAtStr)
		// nolint:gocritic
		if err != nil {
			logger.Warnf("failed to parse backoff_at: %s", err)
		} else if parsed.After(time.Now()) {
			backoffAt = time.Now()
		} else {
			backoffAt = parsed
		}
	}

	backoffRetriesStr := vipertools.GetString(v, "internal.backoff_retries")
	if backoffRetriesStr != "" {
		parsed, err := strconv.Atoi(backoffRetriesStr)
		if err != nil {
			logger.Warnf("failed to parse backoff_retries: %s", err)
		} else {
			backoffRetries = parsed
		}
	}

	return backoffAt, backoffRetries
}

func loadHostname(ctx context.Context, v *viper.Viper) string {
	logger := log.Extract(ctx)

	hostname := vipertools.FirstNonEmptyString(v, "hostname", "settings.hostname")
	gitpod := os.Getenv("GITPOD_WORKSPACE_ID")

	if hostname == "" && gitpod != "" {
		hostname = gitpodHostname
	}

	if hostname == "" {
		var err error

		hostname, err = os.Hostname()
		if err != nil {
			logger.Warnf("failed to retrieve hostname from system: %s", err)
		}
	}

	return hostname
}

func loadProxyURL(ctx context.Context, v *viper.Viper, apiURL *url.URL) (string, error) {
	logger := log.Extract(ctx)

	proxyURL := vipertools.FirstNonEmptyString(v, "proxy", "settings.proxy")

	rgx := proxyRegex
	if strings.Contains(proxyURL, `\\`) {
		rgx = ntlmProxyRegex
	}

	if proxyURL != "" && !rgx.MatchString(proxyURL) {
		return "", api.ErrAuth{Err: fmt.Errorf(errMsgTemplate, proxyURL)}
	}

	proxyEnv := httpproxy.FromEnvironment()

	proxyEnvURL, err := proxyEnv.ProxyFunc()(apiURL)
	if err != nil {
		logger.Warnf("failed to get proxy url from environment for api url: %s", err)
	}

	// try use proxy from environment if no custom proxy is set
	if proxyURL == "" && proxyEnvURL != nil {
		proxyURL = proxyEnvURL.String()
	}

	return proxyURL, nil
}

// loadAPIKey loads a valid default WakaTime API Key or returns an error.
func loadAPIKey(ctx context.Context, v *viper.Viper, order FlagReadOrder) (string, error) {
	apiKey := vipertools.FirstNonEmptyString(v, apiKeyOrder[order]...)
	if apiKey != "" {
		if !apiKeyRegex.MatchString(apiKey) {
			return "", api.ErrAuth{Err: errors.New("invalid api key format")}
		}

		return apiKey, nil
	}

	apiKey, err := readAPIKeyFromCommand(vipertools.GetString(v, "settings.api_key_vault_cmd"))
	if err != nil {
		return "", api.ErrAuth{Err: fmt.Errorf("failed to read api key from vault: %s", err)}
	}

	logger := log.Extract(ctx)

	if apiKey != "" {
		if !apiKeyRegex.MatchString(apiKey) {
			return "", api.ErrAuth{Err: errors.New("invalid api key format")}
		}

		logger.Debugln("loaded api key from vault")

		return apiKey, nil
	}

	apiKey = os.Getenv("WAKATIME_API_KEY")
	if apiKey != "" {
		if !apiKeyRegex.MatchString(apiKey) {
			return "", api.ErrAuth{Err: errors.New("invalid api key format")}
		}

		logger.Debugln("loaded api key from env var")

		return apiKey, nil
	}

	if apiKey == "" {
		return "", api.ErrAuth{Err: errors.New("api key not found or empty")}
	}

	return apiKey, nil
}

// LoadHeartbeatParams loads heartbeats params from viper.Viper instance.
func LoadHeartbeatParams(ctx context.Context, v *viper.Viper, order FlagReadOrder) (Heartbeat, error) {
	var aiLineChanges *int
	if num := v.GetInt("ai-line-changes"); v.IsSet("ai-line-changes") {
		aiLineChanges = heartbeat.PointerTo(num)
	}

	var category heartbeat.Category

	if categoryStr := vipertools.GetString(v, "category"); categoryStr != "" {
		parsed, err := heartbeat.ParseCategory(categoryStr)
		if err != nil {
			return Heartbeat{}, fmt.Errorf("failed to parse category: %s", err)
		}

		category = parsed
	}

	var cursorPosition *int
	if pos := v.GetInt("cursorpos"); v.IsSet("cursorpos") {
		cursorPosition = heartbeat.PointerTo(pos)
	}

	entity := vipertools.FirstNonEmptyString(v, "entity", "file")
	if entity == "" {
		return Heartbeat{}, errors.New("failed to retrieve entity")
	}

	entity, err := homedir.Expand(entity)
	if err != nil {
		return Heartbeat{}, fmt.Errorf("failed expanding entity: %s", err)
	}

	var entityType heartbeat.EntityType

	if entityTypeStr := vipertools.GetString(v, "entity-type"); entityTypeStr != "" {
		parsed, err := heartbeat.ParseEntityType(entityTypeStr)
		if err != nil {
			return Heartbeat{}, fmt.Errorf("failed to parse entity type: %s", err)
		}

		entityType = parsed
	}

	var extraHeartbeats []heartbeat.Heartbeat

	if v.GetBool("extra-heartbeats") {
		extraHeartbeats = readExtraHeartbeats(ctx)
	}

	var isWrite *bool
	if b := v.GetBool("write"); v.IsSet("write") {
		isWrite = heartbeat.PointerTo(b)
	}

	var humanLineChanges *int
	if num := v.GetInt("human-line-changes"); v.IsSet("human-line-changes") {
		humanLineChanges = heartbeat.PointerTo(num)
	}

	var lineNumber *int
	if num := v.GetInt("lineno"); v.IsSet("lineno") {
		lineNumber = heartbeat.PointerTo(num)
	}

	var linesInFile *int
	if num := v.GetInt("lines-in-file"); v.IsSet("lines-in-file") {
		linesInFile = heartbeat.PointerTo(num)
	}

	timeSecs := v.GetFloat64("time")
	if timeSecs == 0 {
		timeSecs = float64(time.Now().UnixNano()) / 1000000000
	}

	processingParams, err := LoadHeartbeatProcessingParams(ctx, v, order)
	if err != nil {
		return Heartbeat{}, err
	}

	var language *string
	if l := vipertools.GetString(v, "language"); l != "" {
		language = &l
	}

	return Heartbeat{
		AIParams:          processingParams.AIParams,
		AILineChanges:     aiLineChanges,
		Category:          category,
		CursorPosition:    cursorPosition,
		Entity:            entity,
		ExtraHeartbeats:   extraHeartbeats,
		EntityType:        entityType,
		GuessLanguage:     processingParams.GuessLanguage,
		HumanLineChanges:  humanLineChanges,
		IsUnsavedEntity:   v.GetBool("is-unsaved-entity"),
		IsWrite:           isWrite,
		Language:          language,
		LanguageAlternate: vipertools.GetString(v, "alternate-language"),
		LineNumber:        lineNumber,
		LinesInFile:       linesInFile,
		LocalFile:         vipertools.GetString(v, "local-file"),
		Time:              timeSecs,
		Filter:            processingParams.Filter,
		Project:           processingParams.Project,
		Sanitize:          processingParams.Sanitize,
	}, nil
}

// LoadHeartbeatProcessingParams loads the heartbeat params needed by the preprocessing pipeline
// without requiring a command-line heartbeat entity.
func LoadHeartbeatProcessingParams(ctx context.Context, v *viper.Viper, order FlagReadOrder) (Heartbeat, error) {
	filterParams, err := loadFilterParams(ctx, v, order)
	if err != nil {
		return Heartbeat{}, fmt.Errorf("failed to load filter params: %s", err)
	}

	projectParams, err := LoadProjectParams(ctx, v)
	if err != nil {
		return Heartbeat{}, fmt.Errorf("failed to parse project params: %s", err)
	}

	sanitizeParams, err := loadSanitizeParams(ctx, v, order)
	if err != nil {
		return Heartbeat{}, fmt.Errorf("failed to load sanitize params: %s", err)
	}

	aiParams, err := LoadAIParams(ctx, v, order)
	if err != nil {
		return Heartbeat{}, fmt.Errorf("failed to load ai params: %s", err)
	}

	return Heartbeat{
		AIParams:      aiParams,
		GuessLanguage: vipertools.FirstNonEmptyBool(v, guessLanguageOrder[order]...),
		Filter:        filterParams,
		Project:       projectParams,
		Sanitize:      sanitizeParams,
	}, nil
}

func loadFilterParams(ctx context.Context, v *viper.Viper, order FlagReadOrder) (FilterParams, error) {
	exclude := v.GetStringSlice("exclude")
	exclude = append(exclude, v.GetStringSlice("settings.exclude")...)
	exclude = append(exclude, v.GetStringSlice("settings.ignore")...)

	var excludePatterns []regex.Regex

	for _, s := range exclude {
		patterns, err := parseBoolOrRegexList(ctx, s)
		if err != nil {
			return FilterParams{}, fmt.Errorf(
				"failed to parse regex exclude param %q: %s",
				s,
				err,
			)
		}

		excludePatterns = append(excludePatterns, patterns...)
	}

	include := v.GetStringSlice("include")
	include = append(include, v.GetStringSlice("settings.include")...)

	var includePatterns []regex.Regex

	for _, s := range include {
		patterns, err := parseBoolOrRegexList(ctx, s)
		if err != nil {
			return FilterParams{}, fmt.Errorf(
				"failed to parse regex include param %q: %s",
				s,
				err,
			)
		}

		includePatterns = append(includePatterns, patterns...)
	}

	return FilterParams{
		Exclude:                    excludePatterns,
		ExcludeUnknownProject:      vipertools.FirstNonEmptyBool(v, excludeUnknownProjectOrder[order]...),
		Include:                    includePatterns,
		IncludeOnlyWithProjectFile: vipertools.FirstNonEmptyBool(v, includeOnlyWithProjectFileOrder[order]...),
	}, nil
}

func loadSanitizeParams(ctx context.Context, v *viper.Viper, order FlagReadOrder) (SanitizeParams, error) {
	// hide branch names
	hideBranchNamesStr := vipertools.FirstNonEmptyString(v, hideBranchNamesOrder[order]...)

	hideBranchNamesPatterns, err := parseBoolOrRegexList(ctx, hideBranchNamesStr)
	if err != nil {
		return SanitizeParams{}, fmt.Errorf(
			"failed to parse regex hide branch names param %q: %s",
			hideBranchNamesStr,
			err,
		)
	}

	// hide dependencies
	hideDependenciesStr := vipertools.FirstNonEmptyString(v, hideDependenciesOrder[order]...)

	hideDependenciesPatterns, err := parseBoolOrRegexList(ctx, hideDependenciesStr)
	if err != nil {
		return SanitizeParams{}, fmt.Errorf(
			"failed to parse regex hide dependencies param %q: %s",
			hideDependenciesStr,
			err,
		)
	}

	// hide project names
	hideProjectNamesStr := vipertools.FirstNonEmptyString(v, hideProjectNamesOrder[order]...)

	hideProjectNamesPatterns, err := parseBoolOrRegexList(ctx, hideProjectNamesStr)
	if err != nil {
		return SanitizeParams{}, fmt.Errorf(
			"failed to parse regex hide project names param %q: %s",
			hideProjectNamesStr,
			err,
		)
	}

	// hide file names
	hideFileNamesStr := vipertools.FirstNonEmptyString(v, hideFileNamesOrder[order]...)

	hideFileNamesPatterns, err := parseBoolOrRegexList(ctx, hideFileNamesStr)
	if err != nil {
		return SanitizeParams{}, fmt.Errorf(
			"failed to parse regex hide file names param %q: %s",
			hideFileNamesStr,
			err,
		)
	}

	return SanitizeParams{
		HideBranchNames:     hideBranchNamesPatterns,
		HideDependencies:    hideDependenciesPatterns,
		HideFileNames:       hideFileNamesPatterns,
		HideProjectFolder:   vipertools.FirstNonEmptyBool(v, hideProjectFolderOrder[order]...),
		HideProjectNames:    hideProjectNamesPatterns,
		ProjectPathOverride: vipertools.GetString(v, "project-folder"),
	}, nil
}

// LoadProjectParams loads project detection params from viper.Viper instance.
func LoadProjectParams(ctx context.Context, v *viper.Viper) (ProjectParams, error) {
	submodulesDisabled, err := parseBoolOrRegexList(ctx, vipertools.GetString(v, "git.submodules_disabled"))
	if err != nil {
		return ProjectParams{}, fmt.Errorf(
			"failed to parse regex submodules disabled param: %s",
			err,
		)
	}

	return ProjectParams{
		Alternate:            vipertools.GetString(v, "alternate-project"),
		BranchAlternate:      vipertools.GetString(v, "alternate-branch"),
		MapPatterns:          loadProjectMapPatterns(ctx, v, "projectmap"),
		Override:             vipertools.GetString(v, "project"),
		ProjectFromGitRemote: v.GetBool("git.project_from_git_remote"),
		SubmodulesDisabled:   submodulesDisabled,
		SubmoduleMapPatterns: loadProjectMapPatterns(ctx, v, "git_submodule_projectmap"),
	}, nil
}

func loadProjectMapPatterns(ctx context.Context, v *viper.Viper, prefix string) []project.MapPattern {
	logger := log.Extract(ctx)

	var mapPatterns []project.MapPattern

	values := vipertools.GetStringMapString(v, prefix)

	for k, s := range values {
		// make all regex case insensitive
		if !strings.HasPrefix(k, "(?i)") {
			k = "(?i)" + k
		}

		compiled, err := regex.Compile(k)
		if err != nil {
			logger.Warnf("failed to compile projectmap regex pattern %q", k)
			continue
		}

		mapPatterns = append(mapPatterns, project.MapPattern{
			Name:  s,
			Regex: compiled,
		})
	}

	return mapPatterns
}

// LoadAIParams loads ai sync params from viper.Viper instance.
func LoadAIParams(_ context.Context, v *viper.Viper, _ FlagReadOrder) (AIParams, error) {
	codexSource := vipertools.GetString(v, "codex-source")
	if codexSource == "" {
		codexSource = os.Getenv("WAKATIME_CODEX_SOURCE")
	}

	return AIParams{
		CodexSource: codexSource,
		SyncDisabled: v.GetBool("sync-ai-disable") ||
			v.GetBool("sync-ai-disabled") ||
			v.GetBool("settings.sync_ai_disabled"),
	}, nil
}

// LoadOfflineParams loads offline params from viper.Viper instance.
func LoadOfflineParams(ctx context.Context, v *viper.Viper, order FlagReadOrder) Offline {
	disabled := vipertools.FirstNonEmptyBool(v, disableOfflineOrder[order]...)
	// if disable-offline or disableoffline is not set, negate settings.offline if set
	if v.IsSet("settings.offline") && !v.IsSet("disable-offline") && !v.IsSet("disableoffline") {
		disabled = !v.GetBool("settings.offline")
	}

	logger := log.Extract(ctx)

	rateLimit := offline.RateLimitDefaultSeconds

	if rateLimitSecs, ok := vipertools.FirstNonEmptyInt(v, heartbeatRateLimitOrder[order]...); ok {
		rateLimit = rateLimitSecs

		if rateLimit < 0 {
			logger.Warnf(
				"argument --heartbeat-rate-limit-seconds must be zero or a positive integer number, got %d",
				rateLimit,
			)

			rateLimit = 0
		}
	}

	syncMax := math.MaxInt32

	if syncOfflineActivity := v.GetInt("sync-offline-activity"); v.IsSet("sync-offline-activity") {
		syncMax = syncOfflineActivity

		if syncMax < 0 {
			logger.Warnf("argument --sync-offline-activity must be zero or a positive integer number, got %d", syncMax)
			syncMax = math.MaxInt32
		}

		if syncMax == 0 {
			syncMax = math.MaxInt32
		}
	}

	var lastSentAt time.Time

	lastSentAtStr := vipertools.GetString(v, "internal.heartbeats_last_sent_at")
	if lastSentAtStr != "" {
		parsed, err := vipertools.SafeTimeParse(ini.DateFormat, lastSentAtStr)
		// nolint:gocritic
		if err != nil {
			logger.Warnf("failed to parse heartbeats_last_sent_at: %s", err)
		} else if parsed.After(time.Now()) {
			lastSentAt = time.Now()
		} else {
			lastSentAt = parsed
		}
	}

	return Offline{
		Disabled:   disabled,
		LastSentAt: lastSentAt,
		PrintMax:   v.GetInt("print-offline-heartbeats"),
		RateLimit:  time.Duration(rateLimit) * time.Second,
		SyncMax:    syncMax,
	}
}

// LoadStatusBarParams loads status bar params from viper.Viper instance.
func LoadStatusBarParams(v *viper.Viper, order FlagReadOrder) (StatusBar, error) {
	hideCategories := true

	for _, key := range todayCategoriesOrder[order] {
		valStr := vipertools.GetString(v, key)
		if valStr == "" {
			continue
		}

		val, err := strconv.ParseBool(valStr)
		if err != nil {
			if key == "settings.status_bar_show_categories" {
				return StatusBar{}, fmt.Errorf("failed to parse status_bar_show_categories: %s", err)
			}

			return StatusBar{}, fmt.Errorf("failed to parse today-hide-categories: %s", err)
		}

		if key == "settings.status_bar_show_categories" {
			hideCategories = !val
		} else {
			hideCategories = val
		}

		break
	}

	hideMinutes := false

	if hideMinutesStr := vipertools.GetString(v, "settings.status_bar_hide_minutes"); hideMinutesStr != "" {
		val, err := strconv.ParseBool(hideMinutesStr)
		if err != nil {
			return StatusBar{}, fmt.Errorf("failed to parse status_bar_hide_minutes: %s", err)
		}

		hideMinutes = val
	}

	maxCategories := 2

	if maxCategoriesStr := vipertools.FirstNonEmptyString(v, todayMaxCategoriesOrder[order]...); maxCategoriesStr != "" {
		val, err := strconv.Atoi(maxCategoriesStr)
		if err != nil {
			return StatusBar{}, fmt.Errorf("failed to parse today-max-categories: %s", err)
		}

		if val < 0 {
			return StatusBar{}, fmt.Errorf("today-max-categories must be a positive number, got %d", val)
		}

		maxCategories = val
	}

	var out output.Output

	if outputStr := vipertools.GetString(v, "output"); outputStr != "" {
		parsed, err := output.Parse(outputStr)
		if err != nil {
			return StatusBar{}, fmt.Errorf("failed to parse output: %s", err)
		}

		out = parsed
	}

	return StatusBar{
		HideCategories: hideCategories,
		HideMinutes:    hideMinutes,
		MaxCategories:  maxCategories,
		Output:         out,
	}, nil
}

func readAPIKeyFromCommand(cmdStr string) (string, error) {
	if cmdStr == "" {
		return "", nil
	}

	cmdStr = strings.TrimSpace(cmdStr)
	if cmdStr == "" {
		return "", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmdName := "sh"
	cmdArgs := []string{"-c", cmdStr}

	if runtime.GOOS == "windows" {
		cmdName = "cmd"
		cmdArgs = []string{"/C", cmdStr}
	}

	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...) // nolint:gosec
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

var extraHeartbeatsCache []heartbeat.Heartbeat // nolint:gochecknoglobals

// Once prevents reading from stdin twice.
var Once sync.Once // nolint:gochecknoglobals

func readExtraHeartbeats(ctx context.Context) []heartbeat.Heartbeat {
	Once.Do(func() {
		logger := log.Extract(ctx)

		in := bufio.NewReader(os.Stdin)

		input, err := in.ReadString('\n')
		if err != nil && err != io.EOF {
			logger.Debugf("failed to read data from stdin: %s", err)
		}

		heartbeats, err := parseExtraHeartbeats(ctx, input)
		if err != nil {
			logger.Errorf("failed parsing: %s", err)
		}

		extraHeartbeatsCache = heartbeats
	})

	return extraHeartbeatsCache
}

func parseExtraHeartbeats(ctx context.Context, data string) ([]heartbeat.Heartbeat, error) {
	logger := log.Extract(ctx)

	if data == "" {
		logger.Debugln("skipping extra heartbeats, as no data was provided")

		return nil, nil
	}

	var extraHeartbeats []ExtraHeartbeat

	err := json.Unmarshal([]byte(data), &extraHeartbeats)
	if err != nil {
		return nil, fmt.Errorf("failed to json decode from data %q: %s", data, err)
	}

	var heartbeats []heartbeat.Heartbeat

	for _, h := range extraHeartbeats {
		parsed, err := parseExtraHeartbeat(h)
		if err != nil {
			return nil, err
		}

		heartbeats = append(heartbeats, *parsed)
	}

	return heartbeats, nil
}

func parseExtraHeartbeat(h ExtraHeartbeat) (*heartbeat.Heartbeat, error) {
	var (
		err      error
		category string
	)

	cat, err := heartbeat.ParseCategory(h.Category)
	if err != nil {
		return nil, fmt.Errorf("failed to parse category: %s", err)
	}

	category = cat.String()

	h.Entity, err = homedir.Expand(h.Entity)
	if err != nil {
		return nil, fmt.Errorf("failed expanding entity: %s", err)
	}

	var entityType heartbeat.EntityType

	// Both type or entity_type are acceptable here. Type takes precedence.
	entityTypeStr := firstNonEmptyString(h.Type, h.EntityType)
	if entityTypeStr != "" {
		entityType, err = heartbeat.ParseEntityType(entityTypeStr)
		if err != nil {
			return nil, err
		}
	}

	cursorPosition, err := parseIntegerNumber(h.CursorPosition)
	if err != nil {
		return nil, fmt.Errorf("failed to convert cursorpos to int: %s", err)
	}

	var isWrite *bool

	switch isWriteVal := h.IsWrite.(type) {
	case bool:
		isWrite = heartbeat.PointerTo(isWriteVal)
	case string:
		val, err := strconv.ParseBool(isWriteVal)
		if err != nil {
			return nil, fmt.Errorf("failed to convert is write to bool: %s", err)
		}

		isWrite = heartbeat.PointerTo(val)
	}

	lineNumber, err := parseIntegerNumber(h.LineNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to convert lineno to int: %s", err)
	}

	lines, err := parseIntegerNumber(h.Lines)
	if err != nil {
		return nil, fmt.Errorf("failed to convert lines to int: %s", err)
	}

	var time float64

	switch timeVal := h.Time.(type) {
	case float64:
		time = timeVal
	case string:
		val, err := strconv.ParseFloat(timeVal, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to convert time to float64: %s", err)
		}

		time = val
	}

	var timestamp float64

	switch timestampVal := h.Timestamp.(type) {
	case float64:
		timestamp = timestampVal
	case string:
		val, err := strconv.ParseFloat(timestampVal, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to convert timestamp to float64: %s", err)
		}

		timestamp = val
	}

	var timestampParsed float64

	switch {
	case h.Time != nil && h.Time != 0:
		timestampParsed = time
	case h.Timestamp != nil && h.Timestamp != 0:
		timestampParsed = timestamp
	default:
		return nil, fmt.Errorf("skipping extra heartbeat, as no valid timestamp was defined")
	}

	var isUnsavedEntity bool

	switch isUnsavedEntityVal := h.IsUnsavedEntity.(type) {
	case bool:
		isUnsavedEntity = isUnsavedEntityVal
	case string:
		val, err := strconv.ParseBool(isUnsavedEntityVal)
		if err != nil {
			return nil, fmt.Errorf("failed to convert is_unsaved_entity to bool: %s", err)
		}

		isUnsavedEntity = val
	}

	aiLineChanges, err := parseIntegerNumber(h.AILineChanges)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ai_line_changes to int: %s", err)
	}

	humanLineChanges, err := parseIntegerNumber(h.HumanLineChanges)
	if err != nil {
		return nil, fmt.Errorf("failed to convert human_line_changes to int: %s", err)
	}

	return &heartbeat.Heartbeat{
		AILineChanges:     aiLineChanges,
		BranchAlternate:   h.BranchAlternate,
		Category:          category,
		CursorPosition:    cursorPosition,
		Entity:            h.Entity,
		EntityType:        entityType,
		HumanLineChanges:  humanLineChanges,
		IsUnsavedEntity:   isUnsavedEntity,
		IsWrite:           isWrite,
		Language:          h.Language,
		LanguageAlternate: h.LanguageAlternate,
		LineNumber:        lineNumber,
		Lines:             lines,
		ProjectAlternate:  h.ProjectAlternate,
		ProjectOverride:   h.Project,
		Time:              timestampParsed,
	}, nil
}

// String implements fmt.Stringer interface.
func (p API) String() string {
	var backoffAt string
	if !p.BackoffAt.IsZero() {
		backoffAt = p.BackoffAt.Format(ini.DateFormat)
	}

	apiKey := p.Key
	if len(apiKey) > 4 {
		// only show last 4 chars of api key in logs
		apiKey = fmt.Sprintf("<hidden>%s", apiKey[len(apiKey)-4:])
	}

	keyPatterns := []apikey.MapPattern{}

	for _, k := range p.KeyPatterns {
		if len(k.APIKey) > 4 {
			// only show last 4 chars of api key in logs
			k.APIKey = fmt.Sprintf("<hidden>%s", k.APIKey[len(k.APIKey)-4:])
		}

		keyPatterns = append(keyPatterns, apikey.MapPattern{
			Regex:  k.Regex,
			APIKey: k.APIKey,
		})
	}

	return fmt.Sprintf(
		"api key: '%s', api url: '%s', backoff at: '%s', backoff retries: %d,"+
			" hostname: '%s', key patterns: '%s', plugin: '%s', proxy url: '%s',"+
			" timeout: %s, disable ssl verify: %t, ssl cert filepath: '%s'",
		apiKey,
		p.URL,
		backoffAt,
		p.BackoffRetries,
		p.Hostname,
		keyPatterns,
		p.Plugin,
		p.ProxyURL,
		p.Timeout,
		p.DisableSSLVerify,
		p.SSLCertFilepath,
	)
}

// String implements fmt.Stringer interface.
func (p FilterParams) String() string {
	return fmt.Sprintf(
		"exclude: '%s', exclude unknown project: %t, include: '%s', include only with project file: %t",
		p.Exclude,
		p.ExcludeUnknownProject,
		p.Include,
		p.IncludeOnlyWithProjectFile,
	)
}

// String implements fmt.Stringer interface.
func (p Heartbeat) String() string {
	var aiLineChanges string
	if p.AILineChanges != nil {
		aiLineChanges = strconv.Itoa(*p.AILineChanges)
	}

	var cursorPosition string
	if p.CursorPosition != nil {
		cursorPosition = strconv.Itoa(*p.CursorPosition)
	}

	var isWrite bool
	if p.IsWrite != nil {
		isWrite = *p.IsWrite
	}

	var language string
	if p.Language != nil {
		language = *p.Language
	}

	var humanLineChanges string
	if p.HumanLineChanges != nil {
		humanLineChanges = strconv.Itoa(*p.HumanLineChanges)
	}

	var lineNumber string
	if p.LineNumber != nil {
		lineNumber = strconv.Itoa(*p.LineNumber)
	}

	var linesInFile string
	if p.LinesInFile != nil {
		linesInFile = strconv.Itoa(*p.LinesInFile)
	}

	return fmt.Sprintf(
		"ai line changes: '%s', category: '%s', cursor position: '%s', entity: '%s',"+
			" entity type: '%s', num extra heartbeats: %d, guess language: %t, human line changes: '%s',"+
			" is unsaved entity: %t, is write: %t, language: '%s',"+
			" line number: '%s', lines in file: '%s', time: %.5f, filter params: (%s),"+
			" project params: (%s), sanitize params: (%s)",
		aiLineChanges,
		p.Category,
		cursorPosition,
		p.Entity,
		p.EntityType,
		len(p.ExtraHeartbeats),
		p.GuessLanguage,
		humanLineChanges,
		p.IsUnsavedEntity,
		isWrite,
		language,
		lineNumber,
		linesInFile,
		p.Time,
		p.Filter,
		p.Project,
		p.Sanitize,
	)
}

// String implements fmt.Stringer interface.
func (p Offline) String() string {
	var lastSentAt string
	if !p.LastSentAt.IsZero() {
		lastSentAt = p.LastSentAt.Format(ini.DateFormat)
	}

	return fmt.Sprintf(
		"disabled: %t, last sent at: '%s', print max: %d, rate limit: %s, num sync max: %d",
		p.Disabled,
		lastSentAt,
		p.PrintMax,
		p.RateLimit,
		p.SyncMax,
	)
}

// String implements fmt.Stringer interface.
func (p AIParams) String() string {
	return fmt.Sprintf(
		"disabled: %t",
		p.SyncDisabled,
	)
}

// String implements fmt.Stringer interface.
func (p Params) String() string {
	return fmt.Sprintf(
		"api params: (%s), heartbeat params: (%s), ai params: (%s), offline params: (%s), status bar params: (%s)",
		p.API,
		p.Heartbeat,
		p.AI,
		p.Offline,
		p.StatusBar,
	)
}

// String implements fmt.Stringer interface.
func (p ProjectParams) String() string {
	return fmt.Sprintf(
		"alternate: '%s', branch alternate: '%s', map patterns: '%s', override: '%s',"+
			" git submodules disabled: '%s', git submodule project map: '%s'",
		p.Alternate,
		p.BranchAlternate,
		p.MapPatterns,
		p.Override,
		p.SubmodulesDisabled,
		p.SubmoduleMapPatterns,
	)
}

// String implements fmt.Stringer interface.
func (p SanitizeParams) String() string {
	return fmt.Sprintf(
		"hide branch names: '%s', hide project folder: %t, hide file names: '%s',"+
			" hide project names: '%s', hide dependencies: '%s', project path override: '%s'",
		p.HideBranchNames,
		p.HideProjectFolder,
		p.HideFileNames,
		p.HideProjectNames,
		p.HideDependencies,
		p.ProjectPathOverride,
	)
}

// String implements fmt.Stringer interface.
func (p StatusBar) String() string {
	return fmt.Sprintf(
		"hide categories: %t, max categories: %d, output: '%s'",
		p.HideCategories,
		p.MaxCategories,
		p.Output,
	)
}

// String implements fmt.Stringer interface.
func (order FlagReadOrder) String() string {
	switch order {
	case FlagReadOrderFlagPrecedence:
		return "flag-precedence"
	case FlagReadOrderProjectConfigPrecedence:
		return "project-config-precedence"
	default:
		return "unknown"
	}
}

func normalizeURL(apiURL string) (string, error) {
	// validate and normalize the URL
	apiURL = strings.TrimSuffix(apiURL, "/")
	apiURL = strings.TrimSuffix(apiURL, ".bulk")
	apiURL = strings.TrimSuffix(apiURL, "/users/current/heartbeats")
	apiURL = strings.TrimSuffix(apiURL, "/heartbeats")
	apiURL = strings.TrimSuffix(apiURL, "/heartbeat")

	_, err := url.Parse(apiURL)
	if err != nil {
		return "", err
	}

	return apiURL, nil
}

func parseBoolOrRegexList(ctx context.Context, s string) ([]regex.Regex, error) {
	var patterns []regex.Regex

	logger := log.Extract(ctx)

	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.Trim(s, "\n\t ")

	switch {
	case s == "":
	case strings.ToLower(s) == "false":
		patterns = []regex.Regex{regex.NewRegexpWrap(matchNoneRegex)}
	case strings.ToLower(s) == "true":
		patterns = []regex.Regex{regex.NewRegexpWrap(matchAllRegex)}
	default:
		splitted := strings.Split(s, "\n")
		for _, s := range splitted {
			s = strings.Trim(s, "\n\t ")
			if s == "" {
				continue
			}

			// make all regex case insensitive
			if !strings.HasPrefix(s, "(?i)") {
				s = "(?i)" + s
			}

			compiled, err := regex.Compile(s)
			if err != nil {
				logger.Warnf("failed to compile regex pattern %q, it will be ignored", s)
				continue
			}

			patterns = append(patterns, compiled)
		}
	}

	return patterns, nil
}

func parseIntegerNumber(inp any) (*int, error) {
	var result *int

	switch val := inp.(type) {
	case float64:
		result = heartbeat.PointerTo(int(val))
	case string:
		v, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}

		result = &v
	}

	return result, nil
}

// firstNonEmptyString accepts multiple values and return the first non empty string value.
func firstNonEmptyString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}

	return ""
}
