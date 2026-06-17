package ai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/viper"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/diagnostic"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/params"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"
)

// Config contains filtering configurations.
type Config struct {
	SyncDisabled bool
	Plugin       string
	Project      params.ProjectParams
	Sanitize     params.SanitizeParams
	V            *viper.Viper
}

// ProjectInfo contains --project, --alternate-project, and --project-folder cli args.
type ProjectInfo struct {
	Alternate           string
	Override            string
	ProjectPathOverride string
}

// ParserConfig contains the arguments for each Parser implementation.
type ParserConfig struct {
	After             time.Time
	FallbackUserAgent string
	UserAgents        map[string]string
	ProjectInfo       ProjectInfo
}

const maxTranscriptLineSize = 10 * 1024 * 1024

const (
	// UnknownParser is the parser ID used when not detected.
	UnknownParser int = iota
	// ClaudeParser is the parser ID for Claude Code.
	ClaudeParser
	// CodexParser is the parser ID for Codex.
	CodexParser
	// ContinueParser is the parser ID for Continue.
	ContinueParser
	// CodyParser is the parser ID for Cody by Sourcegraph.
	CodyParser
	// RooParser is the parser ID for Roo Code.
	RooParser
	// OpenCodeParser is the parser ID for OpenCode.
	OpenCodeParser
	// CopilotParser is the parser ID for GitHub Copilot Chat.
	CopilotParser
	// CursorParser is the parser ID for Cursor.
	CursorParser
	// WindsurfParser is the parser ID for Windsurf.
	WindsurfParser
	// QoderParser is the parser ID for Qoder.
	QoderParser
	// KiroParser is the parser ID for Kiro.
	KiroParser
	// ClineParser is the parser ID for Cline.
	ClineParser
	// GeminiParser is the parser ID for Gemini.
	GeminiParser
	// QwenCodeParser is the parser ID for Qwen Code.
	QwenCodeParser
	// PiParser is the parser ID for Pi.
	PiParser
	// GooseParser is the parser ID for Goose.
	GooseParser
)

type (
	// Parser is a common interface for AI.
	Parser interface {
		Parse(context.Context) (Heartbeats, error)
		Name() string
	}

	// Heartbeats contains the parsed ai heartbeats from Parse().
	Heartbeats []heartbeat.Heartbeat
)

type aiPanicReport struct {
	Logs       string
	ParserName string
	Recovered  any
	Stack      string
}

// WithAISync initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to add heartbeats
// from AI tool transcript logs, and modify existing heartbeats to
// use the 'AI Coding' category when necessary.
func WithAISync(config Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(ctx context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			logger := log.Extract(ctx)
			// logger.Debugln("execute WithAISync")

			if config.SyncDisabled {
				return next(ctx, hh)
			}

			lock, err := AcquireSyncLock(ctx, config.V, SyncLockTimeout)
			if err != nil {
				if IsSyncLockBusy(err) {
					logger.Debugln("skipping ai sync because another wakatime-cli process is already syncing ai activity")
				} else {
					logger.Warnf("skipping ai sync because failed to acquire ai sync lock: %s", err)
				}

				return next(ctx, hh)
			}

			releaseLock := func() {
				if lock != nil {
					lock.Release()
					lock = nil
				}
			}
			defer releaseLock()

			lastParsedAt, err := getLastParsedAt(ctx, config.V)
			if err != nil {
				logger.Debugf("failed ai last parsed: %s", err)
				releaseLock()

				return next(ctx, hh)
			}

			userAgents := entityUserAgents(hh)

			heartbeats, err := parseAIHeartbeats(ctx, lastParsedAt, userAgents, config)
			if err != nil {
				logger.Errorf("failed to parse ai heartbeats: %s", err)
				releaseLock()

				return next(ctx, hh)
			}

			if len(heartbeats) == 0 {
				releaseLock()
				return next(ctx, hh)
			}

			minAIHeartbeatTime, maxAIHeartbeatTime := minMaxAIHeartbeatTimes(heartbeats)

			parsedAt := heartbeatTime(maxAIHeartbeatTime)
			if err := UpdateLastParsedAt(ctx, config.V, parsedAt); err != nil {
				log.Extract(ctx).Warnf("failed to update ai_logs_last_parsed_at: %s", err)
			}

			releaseLock()

			heartbeats = replaceAppHeartbeats(heartbeats)

			heartbeats, firstHumanEdit := preserveHumanAttributes(heartbeats, hh, config, maxAIHeartbeatTime)

			entities := entityToTimeMap(heartbeats)

			// Add back Human heartbeats unless they look like duplicate AI heartbeats
			for _, h := range hh {
				if sameEntityAIHeartbeatWithinWindow(h, entities, 5) && (firstHumanEdit == nil || h.Time < *firstHumanEdit) {
					continue
				}

				inRange := func(windowMinutes float64) bool {
					const secondsPerMinute = 60.0

					windowSeconds := windowMinutes * secondsPerMinute

					return h.Time > minAIHeartbeatTime-windowSeconds && h.Time < maxAIHeartbeatTime+windowSeconds
				}

				if (inRange(2) && (h.HumanLineChanges == nil || *h.HumanLineChanges == 0)) ||
					((firstHumanEdit == nil || h.Time < *firstHumanEdit) && inRange(30)) {
					h.Category = "ai coding"
				}

				// add back this human heartbeat
				heartbeats = append(heartbeats, h)
			}

			return next(ctx, heartbeats)
		}
	}
}

func parseAIHeartbeats(
	ctx context.Context,
	after time.Time,
	userAgents map[string]string,
	config Config,
) (Heartbeats, error) {
	logger := log.Extract(ctx)

	logs, resetLogs := captureAIParsingLogs(ctx)
	defer resetLogs()

	var parsers = []Parser{
		Claude{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: config.Plugin,
		},
		Codex{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: config.Plugin,
		},
		Continue{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: config.Plugin,
		},
		Cody{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: config.Plugin,
		},
		RooCode{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: config.Plugin,
		},
		OpenCode{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: config.Plugin,
		},
		Copilot{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: config.Plugin,
		},
		Cursor{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: config.Plugin,
		},
		Windsurf{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: config.Plugin,
		},
		Qoder{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: config.Plugin,
		},
		Kiro{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: config.Plugin,
		},
		Cline{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: config.Plugin,
		},
		Gemini{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: config.Plugin,
		},
		QwenCode{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: config.Plugin,
		},
		Pi{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: config.Plugin,
		},
		Goose{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: config.Plugin,
		},
	}

	var aiHeartbeats Heartbeats

	for _, p := range parsers {
		logger.Debugf("execute %s", p.Name())

		heartbeats, err := parseHeartbeats(ctx, p, func(report aiPanicReport) {
			if logger.IsVerboseEnabled() {
				report.Logs = logs.String()
			}

			if diagErr := sendAIPanicDiagnostics(ctx, config.V, report); diagErr != nil {
				logger.Warnf("failed to send ai parser panic diagnostics: %s", diagErr)
			}
		})
		if err != nil {
			logger.Errorf("unexpected error occurred at %q: %s", p.Name(), err)
			continue
		}

		if len(heartbeats) > 0 {
			aiHeartbeats = append(aiHeartbeats, heartbeats...)
		}
	}

	return aiHeartbeats, nil
}

func parseHeartbeats(
	ctx context.Context,
	parser Parser,
	reportPanic func(aiPanicReport),
) (heartbeats Heartbeats, err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Extract(ctx).Errorf("panicked: %v. Stack: %s", r, stack)

			if reportPanic != nil {
				reportPanic(aiPanicReport{
					ParserName: parser.Name(),
					Recovered:  r,
					Stack:      stack,
				})
			}

			heartbeats = nil
			err = fmt.Errorf("panicked: %v", r)
		}
	}()

	return parser.Parse(ctx)
}

func captureAIParsingLogs(ctx context.Context) (*bytes.Buffer, func()) {
	logger := log.Extract(ctx)
	logs := bytes.NewBuffer(nil)

	if !logger.IsVerboseEnabled() {
		return logs, func() {}
	}

	logOutput := logger.Output()
	logger.SetOutput(io.MultiWriter(logOutput, logs))

	return logs, func() {
		logger.SetOutput(logOutput)
	}
}

func sendAIPanicDiagnostics(ctx context.Context, v *viper.Viper, report aiPanicReport) error {
	if v == nil {
		return fmt.Errorf("missing viper instance")
	}

	paramAPI, err := params.LoadAPIParams(ctx, v, params.FlagReadOrderFlagPrecedence)
	if err != nil {
		return fmt.Errorf("failed to load API parameters: %s", err)
	}

	opts := []api.Option{
		api.WithTimeout(paramAPI.Timeout),
		api.WithHostname(strings.TrimSpace(paramAPI.Hostname)),
		api.WithUserAgent(ctx, paramAPI.Plugin),
	}

	withAuth, err := api.WithAuth(api.BasicAuth{
		Secret: paramAPI.Key,
	})
	if err != nil {
		return fmt.Errorf("failed to set up auth option on api client: %w", err)
	}

	opts = append(opts, withAuth)

	if paramAPI.DisableSSLVerify {
		opts = append(opts, api.WithDisableSSLVerify())
	}

	if !paramAPI.DisableSSLVerify && paramAPI.SSLCertFilepath != "" {
		withSSLCert, err := api.WithSSLCertFile(ctx, paramAPI.SSLCertFilepath)
		if err != nil {
			return fmt.Errorf("failed to set up ssl cert file option on api client: %s", err)
		}

		opts = append(opts, withSSLCert)
	} else if !paramAPI.DisableSSLVerify {
		opts = append(opts, api.WithSSLCertPool(api.CACerts(ctx)))
	}

	if paramAPI.ProxyURL != "" {
		withProxy, err := api.WithProxy(paramAPI.ProxyURL)
		if err != nil {
			return fmt.Errorf("failed to set up proxy option on api client: %w", err)
		}

		opts = append(opts, withProxy)

		if strings.Contains(paramAPI.ProxyURL, `\\`) {
			withNTLMRetry, err := api.WithNTLMRequestRetry(ctx, paramAPI.ProxyURL)
			if err != nil {
				return fmt.Errorf("failed to set up ntlm request retry option on api client: %w", err)
			}

			opts = append(opts, withNTLMRetry)
		}
	}

	c := api.NewClient(paramAPI.URL, opts...)

	return c.SendDiagnostics(
		ctx,
		paramAPI.Plugin,
		true,
		diagnostic.Error(fmt.Sprintf("ai parser %q panicked: %v", report.ParserName, report.Recovered)),
		diagnostic.Logs(report.Logs),
		diagnostic.Stack(report.Stack),
	)
}

func getLastParsedAt(ctx context.Context, v *viper.Viper) (time.Time, error) {
	lastParsedAt := time.Now().Add(-2 * time.Minute)

	if v == nil {
		return lastParsedAt, fmt.Errorf("missing viper instance")
	}

	logger := log.Extract(ctx)

	lastParsedAtStr := vipertools.GetString(v, "internal.ai_logs_last_parsed_at")
	if lastParsedAtStr != "" {
		parsed, err := vipertools.SafeTimeParse(ini.DateFormat, lastParsedAtStr)
		// nolint:gocritic
		if err != nil {
			logger.Warnf("failed to parse ai_logs_last_parsed_at: %s", err)
		} else if parsed.After(time.Now()) {
			lastParsedAt = time.Now()
		} else {
			lastParsedAt = parsed
		}
	} else {
		lastParsedAt = time.Date(2025, time.February, 24, 0, 0, 0, 0, time.UTC)
	}

	return lastParsedAt, nil
}

// UpdateLastParsedAt stores the latest AI transcript timestamp covered by
// generated AI heartbeats.
func UpdateLastParsedAt(ctx context.Context, v *viper.Viper, parsedAt time.Time) error {
	if v == nil {
		return fmt.Errorf("missing viper instance")
	}

	w, err := ini.NewWriter(ctx, v, ini.InternalFilePath)
	if err != nil {
		return fmt.Errorf("failed to parse internal config file: %s", err)
	}

	keyValue := map[string]string{
		"ai_logs_last_parsed_at": parsedAt.Format(time.RFC3339Nano),
	}

	if err := w.Write(ctx, "internal", keyValue); err != nil {
		return fmt.Errorf("failed to write to internal config file: %s", err)
	}

	return nil
}

// preserveHumanAttributes mutates aiHeartbeats pulling in the attributes from
// humanHeartbeats, which should normally have more details already populated
// from the IDE than available on aiHeartbeats.
func preserveHumanAttributes(
	aiHeartbeats []heartbeat.Heartbeat,
	humanHeartbeats []heartbeat.Heartbeat,
	config Config,
	after float64,
) (Heartbeats, *float64) {
	originals := make(map[string][]heartbeat.Heartbeat, len(humanHeartbeats))
	fallbackProjectFolder := ""

	var firstHumanEdit *float64

	for _, h := range humanHeartbeats {
		if h.Time > after+1 &&
			(firstHumanEdit == nil || h.Time < *firstHumanEdit) &&
			h.HumanLineChanges != nil &&
			*h.HumanLineChanges != 0 {
			firstHumanEdit = &h.Time
		}

		originals[h.Entity] = append(originals[h.Entity], h)

		if fallbackProjectFolder == "" {
			fallbackProjectFolder = firstNonEmptyString(h.ProjectPathOverride, h.ProjectPath)
			if h.EntityType == heartbeat.FileType && fallbackProjectFolder == "" {
				fallbackProjectFolder = h.Entity
			}
		}
	}

	for i := range aiHeartbeats {
		aiHeartbeat := &aiHeartbeats[i]

		for _, h := range originals[aiHeartbeat.Entity] {
			preserveAttributesFromHumanHeartbeat(aiHeartbeat, h)
		}

		if aiHeartbeat.ProjectOverride == "" {
			aiHeartbeat.ProjectOverride = config.Project.Override
		}

		if aiHeartbeat.ProjectAlternate == "" {
			aiHeartbeat.ProjectAlternate = config.Project.Alternate
		}

		if aiHeartbeat.BranchAlternate == "" {
			aiHeartbeat.BranchAlternate = config.Project.BranchAlternate
		}

		if aiHeartbeat.ProjectPathOverride == "" {
			aiHeartbeat.ProjectPathOverride = config.Sanitize.ProjectPathOverride
		}

		if aiHeartbeat.EntityType == heartbeat.AppType && aiHeartbeat.ProjectPathOverride == "" {
			aiHeartbeat.ProjectPathOverride = fallbackProjectFolder
		}
	}

	return aiHeartbeats, firstHumanEdit
}

func minMaxAIHeartbeatTimes(aiHeartbeats []heartbeat.Heartbeat) (float64, float64) {
	minHeartbeatTime := aiHeartbeats[0].Time

	maxHeartbeatTime := aiHeartbeats[0].Time
	for i := 1; i < len(aiHeartbeats); i++ {
		t := aiHeartbeats[i].Time
		if t < minHeartbeatTime {
			minHeartbeatTime = t
		} else if t > maxHeartbeatTime {
			maxHeartbeatTime = t
		}
	}

	return minHeartbeatTime, maxHeartbeatTime
}

func heartbeatTime(timestamp float64) time.Time {
	if math.IsNaN(timestamp) || math.IsInf(timestamp, 0) {
		return time.Time{}
	}

	seconds, fraction := math.Modf(timestamp)
	sec := int64(seconds)

	nsec := int64(math.Round(fraction * float64(time.Second)))
	if nsec >= int64(time.Second) {
		sec++
		nsec -= int64(time.Second)
	}

	return time.Unix(sec, nsec).UTC()
}

func entityToTimeMap(aiHeartbeats []heartbeat.Heartbeat) map[string][]float64 {
	entities := make(map[string][]float64, len(aiHeartbeats))
	for _, h := range aiHeartbeats {
		entities[h.Entity] = append(entities[h.Entity], h.Time)
	}

	return entities
}

func sameEntityAIHeartbeatWithinWindow(
	h heartbeat.Heartbeat,
	aiEntityTimes map[string][]float64,
	windowSeconds float64,
) bool {
	for _, t := range aiEntityTimes[h.Entity] {
		if t >= h.Time-windowSeconds && t <= h.Time+windowSeconds {
			return true
		}
	}

	return false
}

func preserveAttributesFromHumanHeartbeat(aiHeartbeat *heartbeat.Heartbeat, human heartbeat.Heartbeat) {
	preserveStringPointer(&aiHeartbeat.Project, human.Project)
	preserveStringValue(&aiHeartbeat.ProjectAlternate, human.ProjectAlternate)
	preserveStringPointer(&aiHeartbeat.Branch, human.Branch)
	preserveStringValue(&aiHeartbeat.BranchAlternate, human.BranchAlternate)
	preserveStringPointer(&aiHeartbeat.Language, human.Language)
	preserveStringValue(&aiHeartbeat.LanguageAlternate, human.LanguageAlternate)
	preservePositiveIntPointer(&aiHeartbeat.Lines, human.Lines)
	preserveStringValue(&aiHeartbeat.ProjectOverride, human.ProjectOverride)
	preserveStringValue(&aiHeartbeat.ProjectPath, human.ProjectPath)
	preserveStringValue(&aiHeartbeat.ProjectPathOverride, human.ProjectPathOverride)
	preserveIntPointer(&aiHeartbeat.ProjectRootCount, human.ProjectRootCount)
}

func preserveStringPointer(dst **string, src *string) {
	if *dst == nil && src != nil {
		*dst = src
	}
}

func preserveStringValue(dst *string, src string) {
	if *dst == "" && src != "" {
		*dst = src
	}
}

func preserveIntPointer(dst **int, src *int) {
	if *dst == nil && src != nil {
		*dst = src
	}
}

func preservePositiveIntPointer(dst **int, src *int) {
	if src != nil && *src > 0 {
		*dst = src
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}

func entityUserAgents(hh []heartbeat.Heartbeat) map[string]string {
	userAgents := make(map[string]string, len(hh))

	for _, h := range hh {
		if h.Entity == "" || h.UserAgent == "" {
			continue
		}

		userAgents[h.Entity] = h.UserAgent
	}

	return userAgents
}

func aiPlugin(plugin Parser, version string) string {
	if version == "" {
		return plugin.Name()
	}

	return plugin.Name() + "/" + version
}

func aiUserAgent(entity string, userAgents map[string]string, fallback string, parser string) string {
	existing := fallback
	if fromHeartbeat, found := userAgents[entity]; found && fromHeartbeat != "" {
		existing = fromHeartbeat
	}

	if existing != "" {
		return userAgentWithPrependedParser(existing, parser)
	}

	return parser
}

func aiUserAgentWithAgentPrefix(
	entity string,
	userAgents map[string]string,
	fallback string,
	parser string,
	agent string,
) string {
	userAgent := aiUserAgent(entity, userAgents, fallback, parser)

	return userAgentWithPrependedAgent(userAgent, aiAgentUserAgentToken(agent))
}

func aiUserAgentWithEditor(
	entity string,
	userAgents map[string]string,
	fallback string,
	parser string,
	editor string,
) string {
	existing := fallback
	if fromHeartbeat, found := userAgents[entity]; found && fromHeartbeat != "" {
		existing = fromHeartbeat
	}

	existing = userAgentWithPrependedEditor(existing, editor)

	if existing != "" {
		return userAgentWithPrependedParser(existing, parser)
	}

	return parser
}

func userAgentWithPrependedAgent(userAgent string, agent string) string {
	if userAgent == "" {
		return agent
	}

	if agent == "" || userAgentHasToken(userAgent, agent) {
		return userAgent
	}

	return strings.TrimSpace(agent + " " + userAgent)
}

func userAgentWithPrependedParser(userAgent string, parser string) string {
	if userAgentHasProduct(userAgent, userAgentProduct(parser)) {
		return userAgent
	}

	return strings.TrimSpace(parser + " " + userAgent)
}

func userAgentWithPrependedEditor(userAgent string, editor string) string {
	if editor == "" {
		return userAgent
	}

	if userAgentHasProduct(userAgent, userAgentProduct(editor)) {
		return userAgent
	}

	return strings.TrimSpace(editor + " " + userAgent)
}

func aiAgentUserAgentToken(agent string) string {
	agent = strings.Join(strings.Fields(strings.TrimSpace(agent)), "-")

	agent = strings.Trim(agent, "/")
	if agent == "" {
		return ""
	}

	product, version, found := strings.Cut(agent, "/")
	if found {
		product = strings.Trim(product, "-_.")

		version = strings.Trim(version, "-_.")
		if product == "" || version == "" {
			return ""
		}

		return product + "/" + version
	}

	for i, r := range agent {
		if i == 0 || (r < '0' || r > '9') {
			continue
		}

		separator := agent[i-1]
		if separator != '-' && separator != '_' {
			continue
		}

		product = strings.Trim(agent[:i-1], "-_.")

		version = strings.Trim(agent[i:], "-_.")
		if product == "" || version == "" {
			return ""
		}

		return product + "/" + version
	}

	return agent
}

func userAgentHasProduct(userAgent string, product string) bool {
	if userAgent == "" || product == "" {
		return false
	}

	for _, field := range userAgentProductFields(userAgent) {
		if strings.EqualFold(userAgentProduct(field), product) {
			return true
		}
	}

	return false
}

func userAgentHasToken(userAgent string, token string) bool {
	if userAgent == "" || token == "" {
		return false
	}

	for _, field := range userAgentProductFields(userAgent) {
		if strings.EqualFold(field, token) {
			return true
		}
	}

	return false
}

func userAgentProduct(token string) string {
	product, _, _ := strings.Cut(token, "/")

	return product
}

func userAgentProductFields(userAgent string) []string {
	fields := strings.Fields(userAgent)
	if len(fields) == 0 {
		return nil
	}

	if strings.HasPrefix(fields[0], "wakatime/") {
		if len(fields) < 4 {
			return nil
		}

		fields = fields[3:]
	}

	return fields
}

func appHeartbeatEntity(parserName string, rawEntity string) string {
	entity := strings.TrimSpace(rawEntity)
	if entity == "" {
		return parserName
	}

	entity = filepath.Base(entity)

	entity = strings.TrimSuffix(entity, filepath.Ext(entity))
	if entity == "" || strings.EqualFold(entity, parserName) {
		return parserName
	}

	return parserName + " " + entity
}

func countStringLines(content string) int {
	lineChanges := 1

	for _, char := range content {
		if char == '\n' {
			lineChanges++
		}
	}

	return lineChanges
}

func promptLength(text string) int {
	if strings.TrimSpace(text) == "" {
		return 0
	}

	return utf8.RuneCountInString(text)
}

func replaceAppHeartbeats(heartbeats []heartbeat.Heartbeat) []heartbeat.Heartbeat {
	var entity string

	for _, h := range heartbeats {
		if h.EntityType == heartbeat.FileType && h.Entity != "" {
			entity = h.Entity
			break
		}
	}

	if entity == "" {
		return heartbeats
	}

	deduped := make([]heartbeat.Heartbeat, 0, len(heartbeats))

	for i, h := range heartbeats {
		if h.EntityType == heartbeat.FileType && h.Entity != "" {
			entity = h.Entity
		}

		if h.EntityType == heartbeat.AppType {
			h.EntityType = heartbeat.FileType
			h.Entity = entity

			if i > 0 && i+1 < len(heartbeats) && sameHeartbeat(h, heartbeats[i+1]) {
				mergeHeartbeatCounts(&heartbeats[i+1], h)
				continue
			}

			if i > 0 && len(deduped) > 0 && sameHeartbeat(h, deduped[len(deduped)-1]) {
				mergeHeartbeatCounts(&deduped[len(deduped)-1], h)
				continue
			}
		}

		deduped = append(deduped, h)
	}

	return deduped
}

func sameHeartbeat(a, b heartbeat.Heartbeat) bool {
	return a.AISession == b.AISession &&
		a.Category == b.Category &&
		a.Entity == b.Entity &&
		a.EntityType == b.EntityType &&
		absFloat64(a.Time-b.Time) < 60
}

func mergeHeartbeatCounts(dst *heartbeat.Heartbeat, src heartbeat.Heartbeat) {
	dst.AILineChanges = addIntPointers(dst.AILineChanges, src.AILineChanges)
	dst.HumanLineChanges = addIntPointers(dst.HumanLineChanges, src.HumanLineChanges)
	dst.AIPromptLength += src.AIPromptLength
	dst.AIInputTokens += src.AIInputTokens

	dst.AIOutputTokens += src.AIOutputTokens
	if dst.AISubscriptionPlan == "" {
		dst.AISubscriptionPlan = src.AISubscriptionPlan
	}

	if dst.Project == nil || *dst.Project == "" {
		dst.Project = src.Project
	}

	if dst.ProjectAlternate == "" {
		dst.ProjectAlternate = src.ProjectAlternate
	}

	if dst.Branch == nil || *dst.Branch == "" {
		dst.Branch = src.Branch
	}

	if dst.BranchAlternate == "" {
		dst.BranchAlternate = src.BranchAlternate
	}

	if dst.Language == nil || *dst.Language == "" {
		dst.Language = src.Language
	}

	if dst.LanguageAlternate == "" {
		dst.LanguageAlternate = src.LanguageAlternate
	}

	if dst.ProjectOverride == "" {
		dst.ProjectOverride = src.ProjectOverride
	}

	if dst.ProjectPath == "" {
		dst.ProjectPath = src.ProjectPath
	}

	if dst.ProjectPathOverride == "" {
		dst.ProjectPathOverride = src.ProjectPathOverride
	}
}

func addIntPointers(dst *int, src *int) *int {
	switch {
	case dst == nil:
		return src
	case src == nil:
		return dst
	default:
		sum := *dst + *src

		return &sum
	}
}

func absFloat64(v float64) float64 {
	if v < 0 {
		return -v
	}

	return v
}
