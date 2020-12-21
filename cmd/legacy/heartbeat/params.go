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
	"github.com/wakatime/wakatime-cli/pkg/language"
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
	// nolint
	ntlmProxyRegex = regexp.MustCompile(`^.*\\.+$`)
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
	Language        *heartbeat.Language
	LineNumber      *int
	LinesInFile     *int
	LocalFile       string
	OfflineDisabled bool
	OfflineSyncMax  int
	Plugin          string
	Time            float64
	Timeout         time.Duration
	Filter          FilterParams
	Network         NetworkParams
	Project         ProjectParams
	Sanitize        SanitizeParams
}

func (p Params) String() string {
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
		language = p.Language.String()
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
		"api key: %q, api url: %q, category: %q, cursor position: %q, entity: %q,"+
			" entity type: %q, num extra heartbeats: %d, hostname: %q, is write: %t,"+
			" language: %q, line number: %q, lines in file: %q, offline disabled: %t, "+
			" offline sync max: %d, plugin: %q, time: %.5f, timeout: %s, filter params: (%s), "+
			"network params: (%s), project params: (%s), sanitize params: (%s)",
		p.APIKey[:4]+"...",
		p.APIUrl,
		p.Category,
		cursorPosition,
		p.Entity,
		p.EntityType,
		len(p.ExtraHeartbeats),
		p.Hostname,
		isWrite,
		language,
		lineNumber,
		linesInFile,
		p.OfflineDisabled,
		p.OfflineSyncMax,
		p.Plugin,
		p.Time,
		p.Timeout,
		p.Filter,
		p.Network,
		p.Project,
		p.Sanitize,
	)
}

// FilterParams contains heartbeat filtering related command parameters.
type FilterParams struct {
	Exclude                    []regex.Regex
	ExcludeUnknownProject      bool
	Include                    []regex.Regex
	IncludeOnlyWithProjectFile bool
}

func (p FilterParams) String() string {
	return fmt.Sprintf(
		"exclude: %q, exclude unknown project: %t, include: %q, include only with project file: %t",
		p.Exclude,
		p.ExcludeUnknownProject,
		p.Include,
		p.IncludeOnlyWithProjectFile,
	)
}

// NetworkParams contains network related command parameters.
type NetworkParams struct {
	DisableSSLVerify bool
	ProxyURL         string
	SSLCertFilepath  string
}

func (p NetworkParams) String() string {
	return fmt.Sprintf(
		"disable ssl verify: %t, proxy url: %q, ssl cert filepath: %q",
		p.DisableSSLVerify,
		p.ProxyURL,
		p.SSLCertFilepath,
	)
}

// ProjectParams params for project name sanitization.
type ProjectParams struct {
	Alternate        string
	DisableSubmodule []regex.Regex
	MapPatterns      []project.MapPattern
	Override         string
}

func (p ProjectParams) String() string {
	return fmt.Sprintf(
		"alternate: %q, disable submodule: %q, map patterns: %q, override: %q",
		p.Alternate,
		p.DisableSubmodule,
		p.MapPatterns,
		p.Override,
	)
}

// SanitizeParams params for heartbeat sanitization.
type SanitizeParams struct {
	HideBranchNames  []regex.Regex
	HideFileNames    []regex.Regex
	HideProjectNames []regex.Regex
}

func (p SanitizeParams) String() string {
	return fmt.Sprintf(
		"hide branch names: %q, hide file names: %q, hide project names: %q",
		p.HideBranchNames,
		p.HideFileNames,
		p.HideProjectNames,
	)
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

	var linesInFile *int
	if num := v.GetInt("lines-in-file"); v.IsSet("lines-in-file") {
		linesInFile = heartbeat.Int(num)
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

	plugin := v.GetString("plugin")

	timeSecs := v.GetFloat64("time")
	if timeSecs == 0 {
		timeSecs = float64(time.Now().UnixNano()) / 1000000000
	}

	var timeout time.Duration

	timeoutSecs, ok := vipertools.FirstNonEmptyInt(v, "timeout", "settings.timeout")
	if ok {
		timeout = time.Duration(timeoutSecs) * time.Second
	}

	language, err := loadLanguage(v, plugin)
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
		Language:        language,
		LineNumber:      lineNumber,
		LinesInFile:     linesInFile,
		LocalFile:       v.GetString("local-file"),
		OfflineDisabled: offlineDisabled,
		OfflineSyncMax:  offlineSyncMax,
		Plugin:          plugin,
		Time:            timeSecs,
		Timeout:         timeout,
		Filter:          loadFilterParams(v),
		Network:         networkParams,
		Project:         projectParams,
		Sanitize:        sanitizeParams,
	}, nil
}

func readExtraHeartbeats() ([]heartbeat.Heartbeat, error) {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read data from stdin: %s", err)
	}

	heartbeats, errFirst := parseExtraHeartbeat(data)
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

func parseExtraHeartbeat(data []byte) ([]heartbeat.Heartbeat, error) {
	var incoming []struct {
		Category       heartbeat.Category  `json:"category"`
		CursorPosition *int                `json:"cursorpos"`
		Entity         string              `json:"entity"`
		EntityType     string              `json:"entity_type"`
		Type           string              `json:"type"`
		IsWrite        *bool               `json:"is_write"`
		Language       *heartbeat.Language `json:"language"`
		LineNumber     *int                `json:"lineno"`
		Lines          *int                `json:"lines"`
		Time           float64             `json:"time"`
		Timestamp      float64             `json:"timestamp"`
		UserAgent      string              `json:"user_agent"`
	}

	err := json.Unmarshal(data, &incoming)
	if err != nil {
		return nil, fmt.Errorf("failed to json decode from data %q: %s", string(data), err)
	}

	var heartbeats []heartbeat.Heartbeat

	for _, h := range incoming {
		var entityType heartbeat.EntityType

		// Both type or entity_type are acceptable here. Type takes precedence.
		entityTypeStr := firstNonEmptyString(h.Type, h.EntityType)
		if entityTypeStr != "" {
			entityType, err = heartbeat.ParseEntityType(entityTypeStr)
			if err != nil {
				return nil, err
			}
		}

		var timestamp float64

		switch {
		case h.Time != 0:
			timestamp = h.Time
		case h.Timestamp != 0:
			timestamp = h.Timestamp
		default:
			return nil, fmt.Errorf("skipping extra heartbeat, as no valid timestamp was defined")
		}

		heartbeats = append(heartbeats, heartbeat.Heartbeat{
			Category:       h.Category,
			CursorPosition: h.CursorPosition,
			Entity:         h.Entity,
			EntityType:     entityType,
			IsWrite:        h.IsWrite,
			Language:       h.Language,
			LineNumber:     h.LineNumber,
			Lines:          h.Lines,
			Time:           timestamp,
			UserAgent:      h.UserAgent,
		})
	}

	return heartbeats, nil
}

func parseExtraHeartbeatWithStringValues(data []byte) ([]heartbeat.Heartbeat, error) {
	var incoming []struct {
		Category       heartbeat.Category  `json:"category"`
		CursorPosition *string             `json:"cursorpos"`
		Entity         string              `json:"entity"`
		EntityType     string              `json:"entity_type"`
		Type           string              `json:"type"`
		IsWrite        *bool               `json:"is_write"`
		Language       *heartbeat.Language `json:"language"`
		LineNumber     *string             `json:"lineno"`
		Lines          *string             `json:"lines"`
		Time           float64             `json:"time"`
		Timestamp      float64             `json:"timestamp"`
		UserAgent      string              `json:"user_agent"`
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

		var entityType heartbeat.EntityType

		// Both type or entity_type are acceptable here. Type takes precedence.
		entityTypeStr := firstNonEmptyString(h.Type, h.EntityType)
		if entityTypeStr != "" {
			entityType, err = heartbeat.ParseEntityType(entityTypeStr)
			if err != nil {
				return nil, err
			}
		}

		var lineNumber *int

		if h.LineNumber != nil {
			parsed, err := strconv.Atoi(*h.LineNumber)
			if err != nil {
				return nil, fmt.Errorf("failed to convert line number to int: %s", err)
			}

			lineNumber = &parsed
		}

		var lines *int

		if h.Lines != nil {
			parsed, err := strconv.Atoi(*h.Lines)
			if err != nil {
				return nil, fmt.Errorf("failed to convert lines to int: %s", err)
			}

			lines = &parsed
		}

		var timestamp float64

		switch {
		case h.Time != 0:
			timestamp = h.Time
		case h.Timestamp != 0:
			timestamp = h.Timestamp
		default:
			return nil, fmt.Errorf("skipping extra heartbeat, as no valid timestamp was defined")
		}

		heartbeats = append(heartbeats, heartbeat.Heartbeat{
			Category:       h.Category,
			CursorPosition: cursorPosition,
			Entity:         h.Entity,
			EntityType:     entityType,
			IsWrite:        h.IsWrite,
			Language:       h.Language,
			LineNumber:     lineNumber,
			Lines:          lines,
			Time:           timestamp,
			UserAgent:      h.UserAgent,
		})
	}

	return heartbeats, nil
}

func loadLanguage(v *viper.Viper, plugin string) (*heartbeat.Language, error) {
	if v == nil {
		return nil, errors.New("viper instance unset")
	}

	lang, _ := vipertools.FirstNonEmptyString(v, "language", "alternate-language")

	if lang == "" {
		return nil, nil
	}

	parsed, ok := language.Parse(lang, plugin)
	if !ok {
		jww.WARN.Printf("failed to parse language from string %q. plugin: %q", lang, plugin)
		return nil, nil
	}

	return &parsed, nil
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

// firstNonEmptyString accepts multiple values and return the first non empty string value.
func firstNonEmptyString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}

	return ""
}
