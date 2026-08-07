package ai

import (
	"context"
	"path/filepath"
)

// MistralVibe contains parameters for detecting heartbeats from Mistral Vibe logs.
type MistralVibe ParserConfig

// Parse parses Mistral Vibe logs for AI heartbeats.
func (g MistralVibe) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	root := envOrDefault("VIBE_HOME", filepath.Join(home, ".vibe"))
	paths := genericAIPaths{
		roots:            []string{filepath.Join(root, "logs", "session")},
		fileNames:        []string{"messages.jsonl", "meta.json"},
		tokenCounterMode: genericAICumulativeCounters,
	}

	return parseGenericProvider(ctx, g, ParserConfig(g), paths)
}

// Name returns the Mistral Vibe parser name.
func (MistralVibe) Name() string { return "Mistral Vibe" }
