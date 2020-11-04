package heartbeat

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/regex"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	jww "github.com/spf13/jwalterweatherman"
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
	APIKey          string
	APIUrl          string
	Category        heartbeat.Category
	CursorPosition  *int
	Entity          string
	EntityType      heartbeat.EntityType
	ExtraHeartbeats []heartbeat.Heartbeat
	Hostname        string
	IsWrite         *bool
	LineNumber      *int
	LocalFile       string
	OfflineDisabled bool
	OfflineSyncMax  int
	Plugin          string
	Time            float64
	Timeout         time.Duration
	Filter          FilterParams
	Language        LanguageParams
	Network         NetworkParams
	Project         ProjectParams
	Sanitize        SanitizeParams
}

// FilterParams contains heartbeat filtering related command parameters.
type FilterParams struct {
	Exclude                    []regex.Regex
	ExcludeUnknownProject      bool
	Include                    []regex.Regex
	IncludeOnlyWithProjectFile bool
}

// LanguageParams contains language detection related command parameters.
type LanguageParams struct {
	Alternate string
	Override  string
}

// NetworkParams contains network related command parameters.
type NetworkParams struct {
	DisableSSLVerify bool
	ProxyURL         string
	SSLCertFilepath  string
}

// ProjectParams params for project name sanitization.
type ProjectParams struct {
	Alternate        string
	DisableSubmodule []regex.Regex
	MapPatterns      []project.MapPattern
	Override         string
}

// SanitizeParams params for heartbeat sanitization.
type SanitizeParams struct {
	HideBranchNames  []regex.Regex
	HideFileNames    []regex.Regex
	HideProjectNames []regex.Regex
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

	var cursorPosition *int
	if pos := v.GetInt("cursorpos"); v.IsSet("cursorpos") {
		cursorPosition = heartbeat.Int(pos)
	}

	entity, ok := vipertools.FirstNonEmptyString(v, "entity", "file")
	if !ok && !v.IsSet("sync-offline-activity") {
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

	var (
		extraHeartbeats []heartbeat.Heartbeat
		err             error
	)

	if v.GetBool("extra-heartbeats") {
		extraHeartbeats, err = readExtraHeartbeats()
		if err != nil {
			jww.ERROR.Printf("failed to read extra heartbeats: %s", err)
		}
	}

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

	var lineNumber *int
	if num := v.GetInt("lineno"); v.IsSet("lineno") {
		lineNumber = heartbeat.Int(num)
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
	case v.GetString("sync-offline-activity") == "none":
		break
	default:
		offlineSyncMax, err = strconv.Atoi(v.GetString("sync-offline-activity"))
		if err != nil {
			return Params{}, errors.New("argument --sync-offline-activity must be \"none\" or a positive integer number: %s")
		}
	}

	if offlineSyncMax < 0 {
		return Params{}, errors.New("argument --sync-offline-activity must be \"none\" or a positive integer number")
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

	languageParams, err := loadLanguageParams(v)
	if err != nil {
		return Params{}, fmt.Errorf("failed to parse language params: %s", err)
	}

	networkParams, err := loadNetworkParams(v)
	if err != nil {
		return Params{}, fmt.Errorf("failed to parse network params: %s", err)
	}

	projectParams, err := loadProjectParams(v)
	if err != nil {
		return Params{}, fmt.Errorf("failed to parse project params: %s", err)
	}

	sanitizeParams, err := loadSanitizeParams(v)
	if err != nil {
		return Params{}, fmt.Errorf("failed to load sanitize params: %s", err)
	}

	return Params{
		APIKey:          apiKey,
		APIUrl:          apiURL,
		Category:        category,
		CursorPosition:  cursorPosition,
		Entity:          entity,
		ExtraHeartbeats: extraHeartbeats,
		EntityType:      entityType,
		Hostname:        hostname,
		IsWrite:         isWrite,
		LineNumber:      lineNumber,
		LocalFile:       v.GetString("local-file"),
		OfflineDisabled: offlineDisabled,
		OfflineSyncMax:  offlineSyncMax,
		Plugin:          v.GetString("plugin"),
		Time:            timeSecs,
		Timeout:         timeout,
		Filter:          loadFilterParams(v),
		Language:        languageParams,
		Network:         networkParams,
		Project:         projectParams,
		Sanitize:        sanitizeParams,
	}, nil
}

func readExtraHeartbeats() ([]heartbeat.Heartbeat, error) {
	var heartbeats []heartbeat.Heartbeat

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read data from stdin: %s", err)
	}

	errFirst := json.Unmarshal(data, &heartbeats)
	if errFirst == nil {
		return heartbeats, nil
	}

	// try again accepting string properties for int values
	heartbeats, errSecond := parseExtraHeartbeatWithStringValues(data)
	if errSecond != nil {
		return nil, fmt.Errorf(
			"failed to json decode: %s: failed to json decode again, accepting string values: %s",
			errFirst,
			errSecond,
		)
	}

	return heartbeats, nil
}

func parseExtraHeartbeatWithStringValues(data []byte) ([]heartbeat.Heartbeat, error) {
	var incoming []struct {
		Category       heartbeat.Category   `json:"category"`
		CursorPosition *string              `json:"cursorpos"`
		Entity         string               `json:"entity"`
		EntityType     heartbeat.EntityType `json:"type"`
		IsWrite        *bool                `json:"is_write"`
		LineNumber     *string              `json:"lineno"`
		Time           float64              `json:"time"`
		UserAgent      string               `json:"user_agent"`
	}

	err := json.Unmarshal(data, &incoming)
	if err != nil {
		return nil, fmt.Errorf("failed to json decode from data %q: %s", string(data), err)
	}

	var heartbeats []heartbeat.Heartbeat

	for _, h := range incoming {
		var cursorPosition *int

		if h.CursorPosition != nil {
			parsed, err := strconv.Atoi(*h.CursorPosition)
			if err != nil {
				return nil, fmt.Errorf("failed to convert cursor position to int: %s", err)
			}

			cursorPosition = &parsed
		}

		var lineNumber *int

		if h.LineNumber != nil {
			parsed, err := strconv.Atoi(*h.LineNumber)
			if err != nil {
				return nil, fmt.Errorf("failed to convert line number to int: %s", err)
			}

			lineNumber = &parsed
		}

		heartbeats = append(heartbeats, heartbeat.Heartbeat{
			Category:       h.Category,
			CursorPosition: cursorPosition,
			Entity:         h.Entity,
			EntityType:     h.EntityType,
			IsWrite:        h.IsWrite,
			LineNumber:     lineNumber,
			Time:           h.Time,
			UserAgent:      h.UserAgent,
		})
	}

	return heartbeats, nil
}

func loadFilterParams(v *viper.Viper) FilterParams {
	exclude := v.GetStringSlice("exclude")
	exclude = append(exclude, v.GetStringSlice("settings.exclude")...)
	exclude = append(exclude, v.GetStringSlice("settings.ignore")...)

	var excludePatterns []regex.Regex

	for _, s := range exclude {
		compiled, err := regex.Compile(s)
		if err != nil {
			jww.WARN.Printf("failed to compile exclude regex pattern %q", s)
			continue
		}

		excludePatterns = append(excludePatterns, compiled)
	}

	include := v.GetStringSlice("include")
	include = append(include, v.GetStringSlice("settings.include")...)

	var includePatterns []regex.Regex

	for _, s := range include {
		compiled, err := regex.Compile(s)
		if err != nil {
			jww.WARN.Printf("failed to compile include regex pattern %q", s)
			continue
		}

		includePatterns = append(includePatterns, compiled)
	}

	return FilterParams{
		Exclude: excludePatterns,
		ExcludeUnknownProject: vipertools.FirstNonEmptyBool(
			v,
			"exclude-unknown-project",
			"settings.exclude_unknown_project",
		),
		Include: includePatterns,
		IncludeOnlyWithProjectFile: vipertools.FirstNonEmptyBool(
			v,
			"include-only-with-project-file",
			"settings.include_only_with_project_file",
		),
	}
}

func loadLanguageParams(v *viper.Viper) (LanguageParams, error) {
	if v == nil {
		return LanguageParams{}, errors.New("viper instance unset")
	}

	return LanguageParams{
		Alternate: v.GetString("alternate-language"),
		Override:  v.GetString("language"),
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

func loadProjectParams(v *viper.Viper) (ProjectParams, error) {
	if v == nil {
		return ProjectParams{}, errors.New("viper instance unset")
	}

	disableSubmodule, err := parseBoolOrRegexList(v.GetString("git.submodules_disabled"))
	if err != nil {
		return ProjectParams{}, fmt.Errorf(
			"failed to parse regex submodules disabled param: %s",
			err,
		)
	}

	var mapPatterns []project.MapPattern

	projectMap := v.GetStringMapString("projectmap")

	for k, s := range projectMap {
		compiled, err := regexp.Compile(k)
		if err != nil {
			jww.WARN.Printf("failed to compile projectmap regex pattern %q", k)
			continue
		}

		mapPatterns = append(mapPatterns, project.MapPattern{
			Name:  s,
			Regex: compiled,
		})
	}

	return ProjectParams{
		Alternate:        v.GetString("alternate-project"),
		DisableSubmodule: disableSubmodule,
		MapPatterns:      mapPatterns,
		Override:         v.GetString("project"),
	}, nil
}

func parseBoolOrRegexList(s string) ([]regex.Regex, error) {
	var patterns []regex.Regex

	switch {
	case s == "":
		break
	case strings.ToLower(s) == "false":
		break
	case strings.ToLower(s) == "true":
		patterns = []regex.Regex{matchAllRegex}
	default:
		splitted := strings.Split(s, "\n")
		for _, s := range splitted {
			compiled, err := regex.Compile(s)
			if err != nil {
				return nil, fmt.Errorf("failed to compile regex %q: %s", s, err)
			}

			patterns = append(patterns, compiled)
		}
	}

	return patterns, nil
}
