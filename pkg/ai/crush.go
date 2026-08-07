package ai

import (
	"context"
	"path/filepath"
)

// Crush contains parameters for detecting heartbeats from Crush logs.
type Crush ParserConfig

// Parse parses Crush logs for AI heartbeats.
func (g Crush) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	root := envOrDefault("CRUSH_DATA_DIR", filepath.Join(dataHome(home), "crush"))

	return parseGenericProvider(ctx, g, ParserConfig(g), genericAIPaths{sqliteRoots: []string{root}})
}

// Name returns the Crush parser name.
func (Crush) Name() string { return "Crush" }
