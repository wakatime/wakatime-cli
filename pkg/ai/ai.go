package ai

import (
	"context"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// Config contains filtering configurations.
type Config struct {
	SyncAfterTime time.Time
	SyncDisabled  bool
}

// ParserID represents an AI Parser ID.
type ParserID int

const (
	// UnknownParser is the parser ID used when not detected.
	UnknownParser ParserID = iota
	// ClaudeParser is the parser ID for Claude Code.
	ClaudeParser
	// CodexParser is the parser ID for Codex.
	CodexParser
	// CursorParser is the parser ID for Cursor.
	CursorParser
)

const (
	claudeParserString = "claude-parser"
	codexParserString  = "codex-parser"
	cursorParserString = "cursor-parser"
)

// String implements fmt.Stringer interface.
func (d ParserID) String() string {
	switch d {
	case ClaudeParser:
		return claudeParserString
	case CodexParser:
		return codexParserString
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

			heartbeats, err := parseAIHeartbeats(ctx, config)
			if err != nil {
				logger.Errorf("failed to parse ai heartbeats: %s", err)
				return next(ctx, hh)
			}

			if len(heartbeats) == 0 {
				return next(ctx, hh)
			}

			heartbeats = PreserveAttributes(heartbeats, hh)

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

				if h.Time > minHeartbeatTime-1 && h.Time < maxHeartbeatTime+1 {
					h.Category = "ai coding"
				}

				// add this human heartbeat
				heartbeats = append(heartbeats, h)
			}

			return next(ctx, heartbeats)
		}
	}
}

func parseAIHeartbeats(ctx context.Context, config Config) (Heartbeats, error) {
	logger := log.Extract(ctx)

	var parsers = []Parser{
		Claude{
			After: config.SyncAfterTime,
		},
		// Codex{},
		// Cursor{},
	}

	for _, p := range parsers {
		logger.Debugf("execute %s", p.ID().String())

		heartbeats, err := p.Parse(ctx)
		if err != nil {
			logger.Errorf("unexpected error occurred at %q: %s", p.ID().String(), err)
			continue
		}

		if len(heartbeats) > 0 {
			return heartbeats, nil
		}
	}

	return nil, nil
}

// PreserveAttributes mutates aiHeartbeats pulling in the attributes from
// humanHeartbeats, which should normally have more details already populated
// from the IDE than available on aiHeartbeats.
func PreserveAttributes(aiHeartbeats []heartbeat.Heartbeat, humanHeartbeats []heartbeat.Heartbeat) Heartbeats {
	originals := make(map[string][]heartbeat.Heartbeat, len(humanHeartbeats))
	for _, h := range humanHeartbeats {
		originals[h.Entity] = append(originals[h.Entity], h)
	}

	for i := range aiHeartbeats {
		for _, h := range originals[aiHeartbeats[i].Entity] {
			if h.Project != nil && aiHeartbeats[i].Project == nil {
				aiHeartbeats[i].Project = h.Project
			}

			if h.ProjectAlternate != "" && aiHeartbeats[i].ProjectAlternate == "" {
				aiHeartbeats[i].ProjectAlternate = h.ProjectAlternate
			}

			if h.Branch != nil && aiHeartbeats[i].Branch == nil {
				aiHeartbeats[i].Branch = h.Branch
			}

			if h.BranchAlternate != "" && aiHeartbeats[i].BranchAlternate == "" {
				aiHeartbeats[i].BranchAlternate = h.BranchAlternate
			}

			if h.Language != nil && aiHeartbeats[i].Language == nil {
				aiHeartbeats[i].Language = h.Language
			}

			if h.LanguageAlternate != "" && aiHeartbeats[i].LanguageAlternate == "" {
				aiHeartbeats[i].LanguageAlternate = h.LanguageAlternate
			}

			if h.Lines != nil && *h.Lines > 0 {
				aiHeartbeats[i].Lines = h.Lines
			}

			if h.ProjectOverride != "" && aiHeartbeats[i].ProjectOverride == "" {
				aiHeartbeats[i].ProjectOverride = h.ProjectOverride
			}

			if h.ProjectPath != "" && aiHeartbeats[i].ProjectPath == "" {
				aiHeartbeats[i].ProjectPath = h.ProjectPath
			}

			if h.ProjectPathOverride != "" && aiHeartbeats[i].ProjectPathOverride == "" {
				aiHeartbeats[i].ProjectPathOverride = h.ProjectPathOverride
			}

			if h.ProjectRootCount != nil && aiHeartbeats[i].ProjectRootCount == nil {
				aiHeartbeats[i].ProjectRootCount = h.ProjectRootCount
			}
		}
	}

	return aiHeartbeats
}
