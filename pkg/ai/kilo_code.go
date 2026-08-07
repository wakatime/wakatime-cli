package ai

import (
	"context"
	"path/filepath"
)

// KiloCode contains parameters for detecting heartbeats from KiloCode logs.
type KiloCode ParserConfig

// Parse parses KiloCode logs for AI heartbeats.
func (g KiloCode) Parse(ctx context.Context) (Heartbeats, error) {
	paths, err := kiloCodePaths(ctx)
	if err != nil {
		return nil, err
	}

	return parseGenericProvider(ctx, g, ParserConfig(g), paths)
}

func kiloCodePaths(ctx context.Context) (genericAIPaths, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return genericAIPaths{}, err
	}

	return genericAIPaths{
		roots:       vscodeExtensionTaskRoots(home, "kilocode.kilo-code"),
		fileNames:   []string{"ui_messages.json"},
		sqliteRoots: []string{filepath.Join(dataHome(home), "kilo")},
	}, nil
}

// Name returns the KiloCode parser name.
func (KiloCode) Name() string { return "KiloCode" }
