package ai

import (
	"context"
	"path/filepath"
)

// OpenClaw contains parameters for detecting heartbeats from OpenClaw logs.
type OpenClaw ParserConfig

// Parse parses OpenClaw logs for AI heartbeats.
func (g OpenClaw) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	paths := genericAIPaths{
		roots: []string{
			filepath.Join(home, ".openclaw", "agents"),
			filepath.Join(home, ".clawdbot", "agents"),
		},
		extensions: []string{".jsonl"},
	}

	return parseGenericProvider(ctx, g, ParserConfig(g), paths)
}

// Name returns the OpenClaw parser name.
func (OpenClaw) Name() string { return "OpenClaw" }
