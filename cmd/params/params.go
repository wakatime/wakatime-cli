package params

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/apikey"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/output"
	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/regex"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"golang.org/x/net/http/httpproxy"
)

const errMsgTemplate = "invalid url %q. Must be in format" +
	"'https://user:pass@host:port' or " +
	"'socks5://user:pass@host:port' or " +
	"'domain\\\\user:pass.'"

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
	pluginRegex = regexp.MustCompile(`(?i)([a-z\/0-9.]+\s)?(?P<editor>[a-z-]+)\-wakatime\/[0-9.]+`)
	// nolint
	proxyRegex = regexp.MustCompile(`^((https?|socks5)://)?([^:@]+(:([^:@])+)?@)?([^:]+|(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])))(:\d+)?$`)
)

type (
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
		KeyPatterns      []apikey.MapPattern
		Plugin           string
		ProxyURL         string
		SSLCertFilepath  string
		Timeout          time.Duration
		URL              string
	}

	// ExtraHeartbeat contains extra heartbeat.
	ExtraHeartbeat struct {
		BranchAlternate   string             `json:"alternate_branch"`
		Category          heartbeat.Category `json:"category"`
		CursorPosition    any                `json:"cursorpos"`
		Entity            string             `json:"entity"`
		EntityType        string             `json:"entity_type"`
		Type              string             `json:"type"`
		IsUnsavedEntity   any                `json:"is_unsaved_entity"`
		IsWrite           any                `json:"is_write"`
		Language          *string            `json:"language"`
		LanguageAlternate string             `json:"alternate_language"`
		LineNumber        any                `json:"lineno"`
		Lines             any                `json:"lines"`
		Project           string             `json:"project"`
		ProjectAlternate  string             `json:"alternate_project"`
		Time              any                `json:"time"`
		Timestamp         any                `json:"timestamp"`
	}

	// Heartbeat contains heartbeat command parameters.
	Heartbeat struct {
		Category          heartbeat.Category
		CursorPosition    *int
		Entity            string
		EntityType        heartbeat.EntityType
		ExtraHeartbeats   []heartbeat.Heartbeat
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

	// Offline contains offline related parameters.
	Offline struct {
		Disabled  bool
		QueueFile string
		PrintMax  int
		SyncMax   int
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
		HideFileNames       []regex.Regex
		HideProjectFolder   bool
		HideProjectNames    []regex.Regex
		ProjectPathOverride string
	}

	// StatusBar contains status bar related parameters.
	StatusBar struct {
		HideCategories bool
		Output         output.Output
	}
)

// LoadAPIParams loads API params from viper.Viper instance. Returns ErrAuth
// if failed to retrieve api key.
func LoadAPIParams(v *viper.Viper) (API, error) {
	apiKey, ok := vipertools.FirstNonEmptyString(v, "key", "settings.api_key", "settings.apikey")
	if ok && !apiKeyRegex.MatchString(apiKey) {
		return API{}, api.ErrAuth{Err: errors.New("invalid api key format")}
	}

	var err error

	if !ok {
		apiKey, err = readAPIKeyFromCommand(vipertools.GetString(v, "settings.api_key_vault_cmd"))
		if err != nil {
			return API{}, api.ErrAuth{Err: fmt.Errorf("failed to read api key from vault: %s", err)}
		}

		if apiKey == "" {
			return API{}, api.ErrAuth{Err: errors.New("api key not found or empty")}
		}

		if !apiKeyRegex.MatchString(apiKey) {
			return API{}, api.ErrAuth{Err: errors.New("invalid api key format")}
		}

		log.Debugln("loaded api key from vault")
	}

	var apiKeyPatterns []apikey.MapPattern

	apiKeyMap := vipertools.GetStringMapString(v, "project_api_key")

	for k, s := range apiKeyMap {
		// make all regex case insensitive
		if !strings.HasPrefix(k, "(?i)") {
			k = "(?i)" + k
		}

		compiled, err := regex.Compile(k)
		if err != nil {
			log.Warnf("failed to compile project_api_key regex pattern %q", k)
			continue
		}

		if !apiKeyRegex.MatchString(s) {
			return API{}, api.ErrAuth{Err: fmt.Errorf("invalid api key format for %q", k)}
		}

		if s == apiKey {
			continue
		}

		apiKeyPatterns = append(apiKeyPatterns, apikey.MapPattern{
			APIKey: s,
			Regex:  compiled,
		})
	}

	apiURLStr := api.BaseURL

	if u, ok := vipertools.FirstNonEmptyString(v, "api-url", "apiurl", "settings.api_url"); ok {
		apiURLStr = u
	}

	// remove endpoint from api base url to support legacy api_url param
	apiURLStr = strings.TrimSuffix(apiURLStr, "/")
	apiURLStr = strings.TrimSuffix(apiURLStr, ".bulk")
	apiURLStr = strings.TrimSuffix(apiURLStr, "/users/current/heartbeats")
	apiURLStr = strings.TrimSuffix(apiURLStr, "/heartbeats")
	apiURLStr = strings.TrimSuffix(apiURLStr, "/heartbeat")

	apiURL, err := url.Parse(apiURLStr)
	if err != nil {
		return API{}, api.ErrAuth{Err: fmt.Errorf("invalid api url: %s", err)}
	}

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

	var backoffRetries = 0

	backoffRetriesStr := vipertools.GetString(v, "internal.backoff_retries")
	if backoffRetriesStr != "" {
		parsed, err := strconv.Atoi(backoffRetriesStr)
		if err != nil {
			log.Warnf("failed to parse backoff_retries: %s", err)
		} else {
			backoffRetries = parsed
		}
	}

	var hostname string

	hostname, ok = vipertools.FirstNonEmptyString(v, "hostname", "settings.hostname")
	if !ok {
		hostname, err = os.Hostname()
		if err != nil {
			log.Warnf("failed to retrieve hostname from system: %s", err)
		}
	}

	proxyURL, _ := vipertools.FirstNonEmptyString(v, "proxy", "settings.proxy")

	rgx := proxyRegex
	if strings.Contains(proxyURL, `\\`) {
		rgx = ntlmProxyRegex
	}

	if proxyURL != "" && !rgx.MatchString(proxyURL) {
		return API{}, api.ErrAuth{Err: fmt.Errorf(errMsgTemplate, proxyURL)}
	}

	proxyEnv := httpproxy.FromEnvironment()

	proxyEnvURL, err := proxyEnv.ProxyFunc()(apiURL)
	if err != nil {
		log.Warnf("failed to get proxy url from environment for api url: %s", err)
	}

	// try use proxy from environment if no custom proxy is set
	if proxyURL == "" && proxyEnvURL != nil {
		proxyURL = proxyEnvURL.String()
	}

	var sslCertFilepath string

	sslCertFilepath, ok = vipertools.FirstNonEmptyString(v, "ssl-certs-file", "settings.ssl_certs_file")
	if ok {
		sslCertFilepath, err = homedir.Expand(sslCertFilepath)
		if err != nil {
			return API{}, api.ErrAuth{Err: fmt.Errorf("failed expanding ssl certs file: %s", err)}
		}
	}

	var timeout time.Duration

	if timeoutSecs, ok := vipertools.FirstNonEmptyInt(v, "timeout", "settings.timeout"); ok {
		timeout = time.Duration(timeoutSecs) * time.Second
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
		Timeout:          timeout,
		URL:              apiURL.String(),
	}, nil
}

// LoadHeartbeatParams loads heartbeats params from viper.Viper instance.
func LoadHeartbeatParams(v *viper.Viper) (Heartbeat, error) {
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
		isWrite = heartbeat.PointerTo(b)
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
		IsUnsavedEntity:   v.GetBool("is-unsaved-entity"),
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
		// make all regex case insensitive
		if !strings.HasPrefix(s, "(?i)") {
			s = "(?i)" + s
		}

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
		// make all regex case insensitive
		if !strings.HasPrefix(s, "(?i)") {
			s = "(?i)" + s
		}

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
	submodulesDisabled, err := parseBoolOrRegexList(vipertools.GetString(v, "git.submodules_disabled"))
	if err != nil {
		return ProjectParams{}, fmt.Errorf(
			"failed to parse regex submodules disabled param: %s",
			err,
		)
	}

	return ProjectParams{
		Alternate:            vipertools.GetString(v, "alternate-project"),
		BranchAlternate:      vipertools.GetString(v, "alternate-branch"),
		MapPatterns:          loadProjectMapPatterns(v, "projectmap"),
		Override:             vipertools.GetString(v, "project"),
		ProjectFromGitRemote: v.GetBool("git.project_from_git_remote"),
		SubmodulesDisabled:   submodulesDisabled,
		SubmoduleMapPatterns: loadProjectMapPatterns(v, "git_submodule_projectmap"),
	}, nil
}

func loadProjectMapPatterns(v *viper.Viper, prefix string) []project.MapPattern {
	var mapPatterns []project.MapPattern

	values := vipertools.GetStringMapString(v, prefix)

	for k, s := range values {
		// make all regex case insensitive
		if !strings.HasPrefix(k, "(?i)") {
			k = "(?i)" + k
		}

		compiled, err := regex.Compile(k)
		if err != nil {
			log.Warnf("failed to compile projectmap regex pattern %q", k)
			continue
		}

		mapPatterns = append(mapPatterns, project.MapPattern{
			Name:  s,
			Regex: compiled,
		})
	}

	return mapPatterns
}

// LoadOfflineParams loads offline params from viper.Viper instance.
func LoadOfflineParams(v *viper.Viper) Offline {
	disabled := vipertools.FirstNonEmptyBool(v, "disable-offline", "disableoffline")
	if b := v.GetBool("settings.offline"); v.IsSet("settings.offline") {
		disabled = !b
	}

	syncMax := v.GetInt("sync-offline-activity")
	if syncMax < 0 {
		log.Warnf("argument --sync-offline-activity must be zero or a positive integer number, got %d", syncMax)

		syncMax = 0
	}

	return Offline{
		Disabled:  disabled,
		QueueFile: vipertools.GetString(v, "offline-queue-file"),
		PrintMax:  v.GetInt("print-offline-heartbeats"),
		SyncMax:   syncMax,
	}
}

// LoadStatusBarParams loads status bar params from viper.Viper instance.
func LoadStatusBarParams(v *viper.Viper) (StatusBar, error) {
	var hideCategories bool

	if hideCategoriesStr, ok := vipertools.FirstNonEmptyString(
		v,
		"today-hide-categories",
		"settings.status_bar_hide_categories",
	); ok {
		val, err := strconv.ParseBool(hideCategoriesStr)
		if err != nil {
			return StatusBar{}, fmt.Errorf("failed to parse today-hide-categories: %s", err)
		}

		hideCategories = val
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

	cmdParts := strings.Split(cmdStr, " ")
	if len(cmdParts) == 0 {
		return "", nil
	}

	cmdName := cmdParts[0]
	cmdArgs := cmdParts[1:]

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...) // nolint:gosec
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func readExtraHeartbeats() ([]heartbeat.Heartbeat, error) {
	in := bufio.NewReader(os.Stdin)

	input, err := in.ReadString('\n')
	if err != nil {
		log.Debugf("failed to read data from stdin: %s", err)
	}

	heartbeats, err := parseExtraHeartbeats(input)
	if err != nil {
		return nil, fmt.Errorf("failed parsing: %s", err)
	}

	return heartbeats, nil
}

func parseExtraHeartbeats(data string) ([]heartbeat.Heartbeat, error) {
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
	var err error

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
		cursorPosition = heartbeat.PointerTo(int(cursorPositionVal))
	case string:
		val, err := strconv.Atoi(cursorPositionVal)
		if err != nil {
			return nil, fmt.Errorf("failed to convert cursor position to int: %s", err)
		}

		cursorPosition = heartbeat.PointerTo(val)
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

	var lineNumber *int

	switch lineNumberVal := h.LineNumber.(type) {
	case float64:
		lineNumber = heartbeat.PointerTo(int(lineNumberVal))
	case string:
		val, err := strconv.Atoi(lineNumberVal)
		if err != nil {
			return nil, fmt.Errorf("failed to convert line number to int: %s", err)
		}

		lineNumber = heartbeat.PointerTo(val)
	}

	var lines *int

	switch linesVal := h.Lines.(type) {
	case float64:
		lines = heartbeat.PointerTo(int(linesVal))
	case string:
		val, err := strconv.Atoi(linesVal)
		if err != nil {
			return nil, fmt.Errorf("failed to convert lines to int: %s", err)
		}

		lines = heartbeat.PointerTo(val)
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

	return &heartbeat.Heartbeat{
		BranchAlternate:   h.BranchAlternate,
		Category:          h.Category,
		CursorPosition:    cursorPosition,
		Entity:            h.Entity,
		EntityType:        entityType,
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
			" num extra heartbeats: %d, is unsaved entity: %t, is write: %t,"+
			" language: '%s', line number: '%s', lines in file: '%s', time: %.5f,"+
			" filter params: (%s), project params: (%s), sanitize params: (%s)",
		p.Category,
		cursorPosition,
		p.Entity,
		p.EntityType,
		len(p.ExtraHeartbeats),
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
	return fmt.Sprintf(
		"disabled: %t, print max: %d, queue file: '%s', num sync max: %d",
		p.Disabled,
		p.PrintMax,
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
		"hide categories: %t, output: '%s'",
		p.HideCategories,
		p.Output,
	)
}

func parseBoolOrRegexList(s string) ([]regex.Regex, error) {
	var patterns []regex.Regex

	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.Trim(s, "\n\t ")

	switch {
	case s == "":
	case strings.ToLower(s) == "false":
		patterns = []regex.Regex{matchNoneRegex}
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
