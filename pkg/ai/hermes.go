package ai

import (
	"context"
	"path/filepath"
)

// Hermes contains parameters for detecting heartbeats from Hermes logs.
type Hermes ParserConfig

// Parse parses Hermes logs for AI heartbeats.
func (g Hermes) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	root := envOrDefault("HERMES_HOME", filepath.Join(home, ".hermes"))

	stores, err := immediateChildPaths(filepath.Join(root, "profiles"), "state.db")
	if err != nil {
		return nil, err
	}

	stores = append([]string{filepath.Join(root, "state.db")}, stores...)

	return parseGenericAIProvider(ctx, genericAIProvider{
		parser: g, config: ParserConfig(g), sqliteRoots: stores,
		sqliteTables: []string{"sessions", "messages"}, sqliteEvents: hermesSQLiteEvents, sqliteParserVersion: 1,
	})
}

// Name returns the Hermes Agent parser name.
func (Hermes) Name() string { return "Hermes Agent" }

// Sessions own usage; messages supply prompts and completed file activity only.
func hermesSQLiteEvents(table string, row map[string]any) []genericAIEvent {
	event := genericAIEventFromValue(row)
	if table == "sessions" {
		event.sessionID = genericAIString(row["id"])
		if event.sessionID == "" {
			return nil
		}

		event.recordID = "session:" + event.sessionID
		event.input = genericAIInt64(row["input_tokens"]) + genericAIInt64(row["cache_write_tokens"])
		event.cachedInput = genericAIInt64(row["cache_read_tokens"])
		event.output = genericAIInt64(row["output_tokens"]) + genericAIInt64(row["reasoning_tokens"])
		event.tokensFound = true
		event.observeGrowth = true
		event.prompt = ""

		event.timestamp = genericAITime(row["ended_at"])
		if event.timestamp.IsZero() {
			event.timestamp = genericAITime(row["started_at"])
		}
	} else {
		event.input, event.cachedInput, event.output, event.tokensFound = 0, 0, 0, false
		if id := genericAIString(row["id"]); id != "" {
			event.recordID = "message:" + id
		}
	}

	return []genericAIEvent{event}
}
