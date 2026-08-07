package ai

import (
	"context"
	"path/filepath"
)

// QuickDesk contains parameters for detecting heartbeats from Quick Desktop logs.
type QuickDesk ParserConfig

// Parse parses Quick Desktop logs for AI heartbeats.
func (g QuickDesk) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	root := envOrDefault("QUICKWORK_HOME", filepath.Join(home, ".quickwork"))
	paths := genericAIPaths{
		roots:       []string{root},
		extensions:  []string{".jsonl"},
		sqliteRoots: []string{root},
	}

	return parseGenericProvider(ctx, g, ParserConfig(g), paths)
}

// Name returns the Quick Desktop parser name.
func (QuickDesk) Name() string { return "Quick Desktop" }
