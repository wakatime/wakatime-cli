package ai

import (
	"context"
	"path/filepath"
)

// Zerostack contains parameters for detecting heartbeats from Zerostack logs.
type Zerostack ParserConfig

// Parse parses Zerostack logs for AI heartbeats.
func (g Zerostack) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	root := envOrDefault("ZS_DATA_DIR", filepath.Join(dataHome(home), "zerostack"))

	paths := genericJSONPaths(filepath.Join(root, "sessions"), ".json")
	paths.tokenCounterMode = genericAICumulativeCounters

	return parseGenericProvider(ctx, g, ParserConfig(g), paths)
}

// Name returns the Zerostack parser name.
func (Zerostack) Name() string { return "Zerostack" }
