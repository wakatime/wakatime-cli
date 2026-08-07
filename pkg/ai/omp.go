package ai

import (
	"context"
	"path/filepath"
)

// OMP contains parameters for detecting heartbeats from OMP logs.
type OMP ParserConfig

// Parse parses OMP logs for AI heartbeats.
func (g OMP) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	root := filepath.Join(home, ".omp", "agent", "sessions")

	return parseGenericProvider(ctx, g, ParserConfig(g), genericJSONPaths(root, ".jsonl"))
}

// Name returns the OMP parser name.
func (OMP) Name() string { return "OMP" }
