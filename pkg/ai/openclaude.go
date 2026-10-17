package ai

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
)

// OpenClaude reads the Claude-compatible transcripts written by OpenClaude.
type OpenClaude ParserConfig

// Name identifies OpenClaude separately from Claude Code.
func (OpenClaude) Name() string { return "OpenClaude" }

// Parse reuses the Claude schema without changing Claude's configured roots.
func (g OpenClaude) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	provider := genericAIProvider{parser: g, config: ParserConfig(g), roots: []string{filepath.Join(home,
		".openclaude", "projects")}, extensions: stringSet(".jsonl")}

	paths, err := provider.transcriptPaths()
	if err != nil {
		return nil, err
	}

	var result Heartbeats

	claude := Claude(g)

	caller := userAgentProductFields(g.FallbackUserAgent)
	if len(caller) > 0 && strings.EqualFold(userAgentProduct(caller[0]), g.Name()) {
		claude.FallbackUserAgent = strings.Replace(g.FallbackUserAgent,
			userAgentProduct(caller[0])+"/", "claude-code/", 1)
	}

	for _, path := range paths {
		parsed, err := claude.parseTranscriptWithPrefix(ctx, path, "openclaude:")
		if err != nil {
			return nil, err
		}

		for i := range parsed {
			parsed[i].AIInputTokens, parsed[i].AICachedInputTokens, parsed[i].AIOutputTokens = 0, 0, 0

			parsed[i].Entity = strings.Replace(parsed[i].Entity, "Claude ", "OpenClaude ", 1)
			if parsed[i].Entity == "Claude" {
				parsed[i].Entity = g.Name()
			}

			parsed[i].UserAgent = strings.ReplaceAll(parsed[i].UserAgent, "claude-code/", "openclaude/")
			parsed[i].AISession = "openclaude:" + parsed[i].AISession
		}

		result = append(result, parsed...)

		values, err := genericAIJSONLines(ctx, g.Name(), path)
		if err != nil {
			return nil, err
		}

		state := genericAIParseState{}
		ideSession := false

		for _, value := range values {
			raw, _ := json.Marshal(value)

			var line claudeLogLine
			if claudeJSONUnmarshal(raw, &line) != nil {
				continue
			}

			if claudeHasIDEContext(line) {
				ideSession = true
			}

			if line.Message == nil || line.Message.Usage == nil {
				continue
			}

			event := genericAIEvent{sessionID: "openclaude:" + firstNonEmptyString(line.SessionID, genericAISessionID(path)),
				timestamp: line.Timestamp, model: line.Message.Model, recordID: line.Message.ID, tokensFound: true}
			usage := line.Message.Usage
			event.input = claudeTokenValue(usage.InputTokens) + claudeTokenValue(usage.CacheCreationInputTokens)
			event.cachedInput = claudeTokenValue(usage.CacheReadInputTokens)

			event.output = claudeTokenValue(usage.OutputTokens)
			if line.Cwd != nil {
				event.cwd = *line.Cwd
			}

			event = genericAIRecordDelta(event, &state)
			if event.timestamp.IsZero() ||
				!timestampAtOrAfterCutoff(event.timestamp, ParserConfig(g).sessionAfter(event.sessionID)) ||
				!event.hasActivity() {
				continue
			}

			emitted := genericAIEventHeartbeats(g, ParserConfig(g), event)
			for i := range emitted {
				userAgent := claude.userAgent(emitted[i].Entity, line.Version, event.model, "", ideSession)
				emitted[i].UserAgent = strings.ReplaceAll(userAgent, "claude-code/", "openclaude/")
			}

			result = append(result, emitted...)
		}
	}

	return result, nil
}
