package ai

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
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

// ParserID represents an AI Parser ID.
type ParserID int

const maxTranscriptLineSize = 10 * 1024 * 1024

const (
	// UnknownParser is the parser ID used when not detected.
	UnknownParser ParserID = iota
	// ClaudeParser is the parser ID for Claude Code.
	ClaudeParser
	// CodexParser is the parser ID for Codex.
	CodexParser
	// CopilotParser is the parser ID for GitHub Copilot Chat.
	CopilotParser
	// CursorParser is the parser ID for Cursor.
	CursorParser
)

const (
	claudeParserString  = "claude-parser"
	codexParserString   = "codex-parser"
	copilotParserString = "copilot-parser"
	cursorParserString  = "cursor-parser"
)

// String implements fmt.Stringer interface.
func (d ParserID) String() string {
	switch d {
	case ClaudeParser:
		return claudeParserString
	case CodexParser:
		return codexParserString
	case CopilotParser:
		return copilotParserString
	case CursorParser:
		return cursorParserString
	case UnknownParser:
		fallthrough
	default:
		return ""
	}
}

type (
	// Parser is a common interface for AI.
	Parser interface {
		Parse(context.Context) (Heartbeats, error)
		ID() ParserID
	}

	// Heartbeats contains the parsed ai heartbeats from Parse().
	Heartbeats []heartbeat.Heartbeat
)

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

			lastParsedAt, err := getLastParsedAt(ctx, config.V)
			if err != nil {
				logger.Debugf("failed ai last parsed: %s", err)
				return next(ctx, hh)
			}

			userAgents := entityUserAgents(hh)

			heartbeats, err := parseAIHeartbeats(ctx, lastParsedAt, userAgents, config.Plugin)
			if err != nil {
				logger.Errorf("failed to parse ai heartbeats: %s", err)
				return next(ctx, hh)
			}

			if len(heartbeats) == 0 {
				return next(ctx, hh)
			}

			heartbeats = preserveAttributes(heartbeats, hh)

			heartbeats = applyProject(heartbeats, config)

			minHeartbeatTime := heartbeats[0].Time

			maxHeartbeatTime := heartbeats[0].Time
			for i := 1; i < len(heartbeats); i++ {
				t := heartbeats[i].Time
				if t < minHeartbeatTime {
					minHeartbeatTime = t
				} else if t > maxHeartbeatTime {
					maxHeartbeatTime = t
				}
			}

			entities := make(map[string][]float64, len(heartbeats))
			for _, h := range heartbeats {
				entities[h.Entity] = append(entities[h.Entity], h.Time)
			}

			for _, h := range hh {
				found := false

				for _, t := range entities[h.Entity] {
					if t >= h.Time-5 && t <= h.Time+5 {
						found = true
						break
					}
				}

				if found {
					continue // remove this human heartbeat, it's actually AI
				}

				inRange := func(windowMinutes float64) bool {
					const secondsPerMinute = 60.0

					windowSeconds := windowMinutes * secondsPerMinute

					return h.Time > minHeartbeatTime-windowSeconds && h.Time < maxHeartbeatTime+windowSeconds
				}

				if inRange(2) {
					h.Category = "ai coding"
				}

				if (h.HumanLineChanges == nil || *h.HumanLineChanges == 0) && inRange(30) {
					h.Category = "ai coding"
				}

				// add this human heartbeat
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
	fallbackUserAgent string,
) (Heartbeats, error) {
	logger := log.Extract(ctx)

	var parsers = []Parser{
		Claude{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: fallbackUserAgent,
		},
		Codex{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: fallbackUserAgent,
		},
		Copilot{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: fallbackUserAgent,
		},
		Cursor{
			After:             after,
			UserAgents:        userAgents,
			FallbackUserAgent: fallbackUserAgent,
		},
	}

	var aiHeartbeats Heartbeats

	for _, p := range parsers {
		logger.Debugf("execute %s", p.ID().String())

		heartbeats, err := p.Parse(ctx)
		if err != nil {
			logger.Errorf("unexpected error occurred at %q: %s", p.ID().String(), err)
			continue
		}

		if len(heartbeats) > 0 {
			aiHeartbeats = append(aiHeartbeats, heartbeats...)
		}
	}

	return aiHeartbeats, nil
}

func applyProject(heartbeats Heartbeats, config Config) Heartbeats {
	for i := range heartbeats {
		if heartbeats[i].ProjectOverride == "" {
			heartbeats[i].ProjectOverride = config.Project.Override
		}

		if heartbeats[i].ProjectAlternate == "" {
			heartbeats[i].ProjectAlternate = config.Project.Alternate
		}

		if heartbeats[i].BranchAlternate == "" {
			heartbeats[i].BranchAlternate = config.Project.BranchAlternate
		}

		if heartbeats[i].ProjectPathOverride == "" {
			heartbeats[i].ProjectPathOverride = config.Sanitize.ProjectPathOverride
		}
	}

	return heartbeats
}

func getLastParsedAt(ctx context.Context, v *viper.Viper) (time.Time, error) {
	lastParsedAt := time.Now().Add(-1 * time.Minute)

	if v == nil {
		return lastParsedAt, fmt.Errorf("missing viper instance")
	}

	logger := log.Extract(ctx)

	lastParsedAtStr := vipertools.GetString(v, "internal.ai_heartbeats_last_parsed_at")
	if lastParsedAtStr != "" {
		parsed, err := vipertools.SafeTimeParse(ini.DateFormat, lastParsedAtStr)
		// nolint:gocritic
		if err != nil {
			logger.Warnf("failed to parse ai_heartbeats_last_parsed_at: %s", err)
		} else if parsed.After(time.Now()) {
			lastParsedAt = time.Now()
		} else {
			lastParsedAt = parsed
		}
	}

	w, err := ini.NewWriter(ctx, v, ini.InternalFilePath)
	if err != nil {
		return lastParsedAt, fmt.Errorf("failed to parse internal config file: %s", err)
	}

	keyValue := map[string]string{
		"ai_heartbeats_last_parsed_at": time.Now().Format(ini.DateFormat),
	}

	if err := w.Write(ctx, "internal", keyValue); err != nil {
		return lastParsedAt, fmt.Errorf("failed to write to internal config file: %s", err)
	}

	return lastParsedAt, nil
}

// preserveAttributes mutates aiHeartbeats pulling in the attributes from
// humanHeartbeats, which should normally have more details already populated
// from the IDE than available on aiHeartbeats.
func preserveAttributes(aiHeartbeats []heartbeat.Heartbeat, humanHeartbeats []heartbeat.Heartbeat) Heartbeats {
	if len(humanHeartbeats) == 0 {
		return aiHeartbeats
	}

	originals := make(map[string][]heartbeat.Heartbeat, len(humanHeartbeats))
	fallbackProjectFolder := ""

	for _, h := range humanHeartbeats {
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

		if aiHeartbeat.EntityType == heartbeat.AppType && aiHeartbeat.ProjectPathOverride == "" {
			aiHeartbeat.ProjectPathOverride = fallbackProjectFolder
		}
	}

	return aiHeartbeats
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

func aiUserAgent(ctx context.Context, entity string, userAgents map[string]string, fallback string, parser string) string {
	existing := fallback
	if fromHeartbeat, found := userAgents[entity]; found && fromHeartbeat != "" {
		existing = fromHeartbeat
	}

	if existing != "" {
		return heartbeat.UserAgent(ctx, parser+" "+existing)
	}

	return heartbeat.UserAgent(ctx, parser)
}
