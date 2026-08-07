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

	return parseGenericProvider(ctx, g, ParserConfig(g), genericAIPaths{sqliteRoots: []string{root}})
}

// Name returns the Hermes Agent parser name.
func (Hermes) Name() string { return "Hermes Agent" }
