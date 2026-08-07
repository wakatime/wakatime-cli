package ai

import (
	"context"
	"path/filepath"
)

// Droid contains parameters for detecting heartbeats from Droid logs.
type Droid ParserConfig

// Parse parses Droid logs for AI heartbeats.
func (g Droid) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	factory := envOrDefault("FACTORY_DIR", filepath.Join(home, ".factory"))

	return parseGenericProvider(ctx, g, ParserConfig(g), genericJSONPaths(filepath.Join(factory, "sessions"), ".jsonl"))
}

// Name returns the Droid parser name.
func (Droid) Name() string { return "Droid" }
