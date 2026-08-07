package ai

import (
	"context"
	"path/filepath"
)

// Mux contains parameters for detecting heartbeats from Mux logs.
type Mux ParserConfig

// Parse parses Mux logs for AI heartbeats.
func (g Mux) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	root := envOrDefault("MUX_ROOT", filepath.Join(home, ".mux"))
	paths := genericAIPaths{
		roots:     []string{filepath.Join(root, "sessions")},
		fileNames: []string{"chat.jsonl"},
	}

	return parseGenericProvider(ctx, g, ParserConfig(g), paths)
}

// Name returns the Mux parser name.
func (Mux) Name() string { return "Mux" }
