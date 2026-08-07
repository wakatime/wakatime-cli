package ai

import (
	"context"
	"path/filepath"
)

// Devin contains parameters for detecting heartbeats from Devin logs.
type Devin ParserConfig

// Parse parses Devin logs for AI heartbeats.
func (g Devin) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	root := filepath.Join(home, ".local", "share", "devin", "transcripts")
	paths := genericJSONPaths(root, ".json")
	paths.containerKey = "steps"

	return parseGenericProvider(ctx, g, ParserConfig(g), paths)
}

// Name returns the Devin parser name.
func (Devin) Name() string { return "Devin" }
