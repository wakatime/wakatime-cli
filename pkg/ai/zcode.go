package ai

import (
	"context"
	"path/filepath"
)

// ZCode contains parameters for detecting heartbeats from ZCode logs.
type ZCode ParserConfig

// Parse parses ZCode logs for AI heartbeats.
func (g ZCode) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(home, ".zcode", "cli", "db", "db.sqlite")

	return parseGenericProvider(ctx, g, ParserConfig(g), genericAIPaths{sqliteRoots: []string{dbPath}})
}

// Name returns the ZCode parser name.
func (ZCode) Name() string { return "ZCode" }
