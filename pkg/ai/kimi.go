package ai

import (
	"context"
	"path/filepath"
)

// Kimi contains parameters for detecting heartbeats from Kimi logs.
type Kimi ParserConfig

// Parse parses Kimi logs for AI heartbeats.
func (g Kimi) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	root := envOrDefault("KIMI_SHARE_DIR", filepath.Join(home, ".kimi"))
	paths := genericAIPaths{
		roots:     []string{filepath.Join(root, "sessions")},
		fileNames: []string{"wire.jsonl"},
	}

	return parseGenericProvider(ctx, g, ParserConfig(g), paths)
}

// Name returns the Kimi parser name.
func (Kimi) Name() string { return "Kimi" }
