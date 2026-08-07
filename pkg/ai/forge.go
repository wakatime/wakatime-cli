package ai

import (
	"context"
	"path/filepath"
)

// Forge contains parameters for detecting heartbeats from Forge logs.
type Forge ParserConfig

// Parse parses Forge logs for AI heartbeats.
func (g Forge) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	paths := genericAIPaths{sqliteRoots: []string{filepath.Join(home, ".forge", ".forge.db")}}

	return parseGenericProvider(ctx, g, ParserConfig(g), paths)
}

// Name returns the Forge parser name.
func (Forge) Name() string { return "Forge" }
