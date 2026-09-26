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

	for _, path := range paths {
		parsed, _, err := Claude(g).parseTranscript(ctx, path, g.After, claudeTranscriptCheckpoint{})
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

		for _, value := range values {
			raw, _ := json.Marshal(value)

			var line claudeLogLine
			if claudeJSONUnmarshal(raw, &line) != nil || line.Message == nil || line.Message.Usage == nil {
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
			if event.timestamp.IsZero() || !timestampAtOrAfterCutoff(event.timestamp, g.After) || !event.hasActivity() {
				continue
			}

			emitted := genericAIEventHeartbeats(g, ParserConfig(g), event)
			for i := range emitted {
				emitted[i].UserAgent = aiUserAgentWithModelAndEditor(emitted[i].Entity, g.UserAgents, g.FallbackUserAgent,
					event.model, "", "openclaude/"+unknownIfEmpty(line.Version))
			}

			result = append(result, emitted...)
		}
	}

	return result, nil
}
