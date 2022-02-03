package params

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/regex"
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
	matchAllRegex = regexp.MustCompile(".*")
	// nolint
	pluginRegex = regexp.MustCompile(`(?i)([a-z\/0-9.]+\s)?(?P<editor>[a-z-]+)\-wakatime\/[0-9.]+`)
	// nolint
	proxyRegex = regexp.MustCompile(`^((https?|socks5)://)?([^:@]+(:([^:@])+)?@)?[^:]+(:\d+)?$`)
	// nolint
	ntlmProxyRegex = regexp.MustCompile(`^.*\\.+$`)
)

type (
	// Config contains configuration settings.
	Config struct {
		APIKeyRequired    bool
		HeartbeatRequired bool
		ForSavingOffline  bool
	}

	// Params contains params.
	Params struct {
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
		Plugin           string
		ProxyURL         string
		SSLCertFilepath  string
		Timeout          time.Duration
		URL              string
	}

	// Heartbeat contains heartbeat command parameters.
	Heartbeat struct {
		Category          heartbeat.Category
		CursorPosition    *int
		Entity            string
		EntityType        heartbeat.EntityType
		ExtraHeartbeats   []heartbeat.Heartbeat
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

	// Offline contains offline related parameters.
	Offline struct {
		Disabled  bool
		QueueFile string
		SyncMax   int
	}

	// ProjectParams params for project name sanitization.
	ProjectParams struct {
		Alternate        string
		DisableSubmodule []regex.Regex
		MapPatterns      []project.MapPattern
		Override         string
	}

	// SanitizeParams params for heartbeat sanitization.
	SanitizeParams struct {
		HideBranchNames     []regex.Regex
		HideFileNames       []regex.Regex
		HideProjectFolder   bool
		HideProjectNames    []regex.Regex
		ProjectPathOverride string
	}

	// StatusBar contains status bar related parameters.
	StatusBar struct {
		HideCategories bool
	}
)

// Load loads params from viper.Viper instance. Returns ErrAuth
// if failed to retrieve api key.
func Load(v *viper.Viper, config Config) (Params, error) {
	if v == nil {
		return Params{}, errors.New("viper instance unset")
	}

	heartbeatParams, err := loadHeartbeatParams(v, config.HeartbeatRequired)
	if err != nil {
		return Params{}, fmt.Errorf("failed to load heartbeat params: %w", err)
	}

	apiParams, err := loadAPIParams(v, config.APIKeyRequired)
	if err != nil {
		if config.ForSavingOffline {
			log.Warnf("failed to load api params: %w", err)
		} else {
			return Params{}, fmt.Errorf("failed to load api params: %w", err)
		}
	}

	offlineParams, err := loadOfflineParams(v)
	if err != nil {
		if config.ForSavingOffline {
			log.Warnf("failed to load offline params: %w", err)
		} else {
			return Params{}, fmt.Errorf("failed to load offline params: %w", err)
		}
	}

	statusBarParams := loadStausBarParams(v)

	return Params{
		API:       apiParams,
		Heartbeat: heartbeatParams,
		Offline:   offlineParams,
		StatusBar: statusBarParams,
	}, nil
}

func loadAPIParams(v *viper.Viper, apiKeyRequired bool) (API, error) {
	apiKey, ok := vipertools.FirstNonEmptyString(v, "key", "settings.api_key", "settings.apikey")
	if !ok && apiKeyRequired {
		return API{}, api.ErrAuth("failed to load api key")
	}

	if !apiKeyRegex.Match([]byte(apiKey)) && apiKeyRequired {
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
		parsed, err := time.Parse(ini.DateFormat, backoffAtStr)
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

func loadHeartbeatParams(v *viper.Viper, required bool) (Heartbeat, error) {
	if !required {
		return Heartbeat{}, nil
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
		cursorPosition = heartbeat.Int(pos)
	}

	var entity string

	entity, ok := vipertools.FirstNonEmptyString(v, "entity", "file")
	if !ok {
		return Heartbeat{}, errors.New("failed to retrieve entity")
	}

	entityExpanded, err := homedir.Expand(entity)
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
		extraHeartbeats, err = readExtraHeartbeats()
		if err != nil {
			log.Errorf("failed to read extra heartbeats: %s", err)
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

	timeSecs := v.GetFloat64("time")
	if timeSecs == 0 {
		timeSecs = float64(time.Now().UnixNano()) / 1000000000
	}

	projectParams, err := loadProjectParams(v)
	if err != nil {
		return Heartbeat{}, fmt.Errorf("failed to parse project params: %s", err)
	}

	sanitizeParams, err := loadSanitizeParams(v)
	if err != nil {
		return Heartbeat{}, fmt.Errorf("failed to load sanitize params: %s", err)
	}

	var language *string
	if l := vipertools.GetString(v, "language"); l != "" {
		language = &l
	}

	return Heartbeat{
		Category:          category,
		CursorPosition:    cursorPosition,
		Entity:            entityExpanded,
		ExtraHeartbeats:   extraHeartbeats,
		EntityType:        entityType,
		IsWrite:           isWrite,
		Language:          language,
		LanguageAlternate: vipertools.GetString(v, "alternate-language"),
		LineNumber:        lineNumber,
		LinesInFile:       linesInFile,
		LocalFile:         vipertools.GetString(v, "local-file"),
		Time:              timeSecs,
		Filter:            loadFilterParams(v),
		Project:           projectParams,
		Sanitize:          sanitizeParams,
	}, nil
}

func loadFilterParams(v *viper.Viper) FilterParams {
	exclude := v.GetStringSlice("exclude")
	exclude = append(exclude, v.GetStringSlice("settings.exclude")...)
	exclude = append(exclude, v.GetStringSlice("settings.ignore")...)

	var excludePatterns []regex.Regex

	for _, s := range exclude {
		compiled, err := regex.Compile(s)
		if err != nil {
			log.Warnf("failed to compile exclude regex pattern %q", s)
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
			log.Warnf("failed to compile include regex pattern %q", s)
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

func loadSanitizeParams(v *viper.Viper) (SanitizeParams, error) {
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

	return SanitizeParams{
		HideBranchNames:     hideBranchNamesPatterns,
		HideFileNames:       hideFileNamesPatterns,
		HideProjectFolder:   vipertools.FirstNonEmptyBool(v, "hide-project-folder", "settings.hide_project_folder"),
		HideProjectNames:    hideProjectNamesPatterns,
		ProjectPathOverride: vipertools.GetString(v, "project-folder"),
	}, nil
}

func loadProjectParams(v *viper.Viper) (ProjectParams, error) {
	disableSubmodule, err := parseBoolOrRegexList(vipertools.GetString(v, "git.submodules_disabled"))
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
			log.Warnf("failed to compile projectmap regex pattern %q", k)
			continue
		}

		mapPatterns = append(mapPatterns, project.MapPattern{
			Name:  s,
			Regex: compiled,
		})
	}

	return ProjectParams{
		Alternate:        vipertools.GetString(v, "alternate-project"),
		DisableSubmodule: disableSubmodule,
		MapPatterns:      mapPatterns,
		Override:         vipertools.GetString(v, "project"),
	}, nil
}

func loadOfflineParams(v *viper.Viper) (Offline, error) {
	disabled := vipertools.FirstNonEmptyBool(v, "disable-offline", "disableoffline")
	if b := v.GetBool("settings.offline"); v.IsSet("settings.offline") {
		disabled = !b
	}

	var (
		syncMax int
		err     error
	)

	switch {
	case !v.IsSet("sync-offline-activity"):
		// use default
		syncMax = v.GetInt("sync-offline-activity")
	case vipertools.GetString(v, "sync-offline-activity") == "none":
		break
	default:
		syncMax, err = strconv.Atoi(vipertools.GetString(v, "sync-offline-activity"))
		if err != nil {
			return Offline{}, errors.New("argument --sync-offline-activity must be \"none\" or a positive integer number: %s")
		}
	}

	if syncMax < 0 {
		return Offline{}, errors.New("argument --sync-offline-activity must be \"none\" or a positive integer number")
	}

	return Offline{
		Disabled:  disabled,
		QueueFile: vipertools.GetString(v, "offline-queue-file"),
		SyncMax:   syncMax,
	}, nil
}

func loadStausBarParams(v *viper.Viper) StatusBar {
	hideCategories := vipertools.FirstNonEmptyBool(
		v,
		"today-hide-categories",
		"settings.status_bar_hide_categories",
	)

	return StatusBar{
		HideCategories: hideCategories,
	}
}

func readExtraHeartbeats() ([]heartbeat.Heartbeat, error) {
	in := bufio.NewReader(os.Stdin)

	input, err := in.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read data from stdin: %s", err)
	}

	heartbeats, err := parseExtraHeartbeat(input)
	if err != nil {
		return nil, fmt.Errorf("failed to json decode: %s", err)
	}

	return heartbeats, nil
}

func parseExtraHeartbeat(data string) ([]heartbeat.Heartbeat, error) {
	var incoming []struct {
		Category          heartbeat.Category `json:"category"`
		CursorPosition    interface{}        `json:"cursorpos"`
		Entity            string             `json:"entity"`
		EntityType        string             `json:"entity_type"`
		Type              string             `json:"type"`
		IsWrite           interface{}        `json:"is_write"`
		Language          *string            `json:"language"`
		LanguageAlternate string             `json:"alternate_language"`
		LineNumber        interface{}        `json:"lineno"`
		Lines             interface{}        `json:"lines"`
		Project           string             `json:"project"`
		ProjectAlternate  string             `json:"alternate_project"`
		Time              interface{}        `json:"time"`
		Timestamp         interface{}        `json:"timestamp"`
	}

	err := json.Unmarshal([]byte(data), &incoming)
	if err != nil {
		return nil, fmt.Errorf("failed to json decode from data %q: %s", string(data), err)
	}

	var heartbeats []heartbeat.Heartbeat

	for _, h := range incoming {
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

		var cursorPosition *int

		switch cursorPositionVal := h.CursorPosition.(type) {
		case float64:
			cursorPosition = heartbeat.Int(int(cursorPositionVal))
		case string:
			val, err := strconv.Atoi(cursorPositionVal)
			if err != nil {
				return nil, fmt.Errorf("failed to convert cursor position to int: %s", err)
			}

			cursorPosition = heartbeat.Int(val)
		}

		var isWrite *bool

		switch isWriteVal := h.IsWrite.(type) {
		case bool:
			isWrite = heartbeat.Bool(isWriteVal)
		case string:
			val, err := strconv.ParseBool(isWriteVal)
			if err != nil {
				return nil, fmt.Errorf("failed to convert is write to bool: %s", err)
			}

			isWrite = heartbeat.Bool(val)
		}

		var lineNumber *int

		switch lineNumberVal := h.LineNumber.(type) {
		case float64:
			lineNumber = heartbeat.Int(int(lineNumberVal))
		case string:
			val, err := strconv.Atoi(lineNumberVal)
			if err != nil {
				return nil, fmt.Errorf("failed to convert line number to int: %s", err)
			}

			lineNumber = heartbeat.Int(val)
		}

		var lines *int

		switch linesVal := h.Lines.(type) {
		case float64:
			lines = heartbeat.Int(int(linesVal))
		case string:
			val, err := strconv.Atoi(linesVal)
			if err != nil {
				return nil, fmt.Errorf("failed to convert lines to int: %s", err)
			}

			lines = heartbeat.Int(val)
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

		heartbeats = append(heartbeats, heartbeat.Heartbeat{
			Category:          h.Category,
			CursorPosition:    cursorPosition,
			Entity:            h.Entity,
			EntityType:        entityType,
			IsWrite:           isWrite,
			Language:          h.Language,
			LanguageAlternate: h.LanguageAlternate,
			LineNumber:        lineNumber,
			Lines:             lines,
			ProjectAlternate:  h.ProjectAlternate,
			ProjectOverride:   h.Project,
			Time:              timestampParsed,
		})
	}

	return heartbeats, nil
}

func parseEditorFromPlugin(plugin string) (string, error) {
	match := pluginRegex.FindStringSubmatch(plugin)
	paramsMap := make(map[string]string)

	for i, name := range pluginRegex.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}

	if len(paramsMap) == 0 || paramsMap["editor"] == "" {
		return "", fmt.Errorf("plugin malformed: %s", plugin)
	}

	return paramsMap["editor"], nil
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
		apiKey = "..." + apiKey[:len(apiKey)-4]
	}

	return fmt.Sprintf(
		"api key: '%s', api url: '%s', backoff at: '%s', backoff retries: %d,"+
			" hostname: '%s', plugin: '%s', timeout: %s, disable ssl verify: %t,"+
			" proxy url: '%s', ssl cert filepath: '%s'",
		apiKey,
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

func (p FilterParams) String() string {
	return fmt.Sprintf(
		"exclude: '%s', exclude unknown project: %t, include: '%s', include only with project file: %t",
		p.Exclude,
		p.ExcludeUnknownProject,
		p.Include,
		p.IncludeOnlyWithProjectFile,
	)
}

func (p Heartbeat) String() string {
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

	var lineNumber string
	if p.LineNumber != nil {
		lineNumber = strconv.Itoa(*p.LineNumber)
	}

	var linesInFile string
	if p.LinesInFile != nil {
		linesInFile = strconv.Itoa(*p.LinesInFile)
	}

	return fmt.Sprintf(
		"category: '%s', cursor position: '%s', entity: '%s', entity type: '%s',"+
			" num extra heartbeats: %d, is write: %t, language: '%s',"+
			" line number: '%s', lines in file: '%s', time: %.5f,"+
			" filter params: (%s), project params: (%s), sanitize params: (%s)",
		p.Category,
		cursorPosition,
		p.Entity,
		p.EntityType,
		len(p.ExtraHeartbeats),
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
	return fmt.Sprintf(
		"disabled: %t, queue file: '%s', num sync max: %d",
		p.Disabled,
		p.QueueFile,
		p.SyncMax,
	)
}

// String implements fmt.Stringer interface.
func (p Params) String() string {
	return fmt.Sprintf(
		"api params: (%s), heartbeat params: (%s), offline params: (%s), status bar params: (%s)",
		p.API,
		p.Heartbeat,
		p.Offline,
		p.StatusBar,
	)
}

func (p ProjectParams) String() string {
	return fmt.Sprintf(
		"alternate: '%s', disable submodule: '%s', map patterns: '%s', override: '%s'",
		p.Alternate,
		p.DisableSubmodule,
		p.MapPatterns,
		p.Override,
	)
}

func (p SanitizeParams) String() string {
	return fmt.Sprintf(
		"hide branch names: '%s', hide project folder: %t, hide file names: '%s',"+
			" hide project names: '%s', project path override: '%s'",
		p.HideBranchNames,
		p.HideProjectFolder,
		p.HideFileNames,
		p.HideProjectNames,
		p.ProjectPathOverride,
	)
}

// String implements fmt.Stringer interface.
func (p StatusBar) String() string {
	return fmt.Sprintf(
		"hide categories: %t",
		p.HideCategories,
	)
}

func parseBoolOrRegexList(s string) ([]regex.Regex, error) {
	var patterns []regex.Regex

	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.Trim(s, "\n\t ")

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
			s = strings.Trim(s, "\n\t ")
			if s == "" {
				continue
			}

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
