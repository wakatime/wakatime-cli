package ai

import (
	"context"
	"path/filepath"
)

// CursorAgent contains parameters for detecting heartbeats from Cursor Agent logs.
type CursorAgent ParserConfig

// Parse parses Cursor Agent logs for AI heartbeats.
func (g CursorAgent) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	root := filepath.Join(home, ".cursor")
	paths := genericAIPaths{
		roots:       []string{filepath.Join(root, "projects")},
		extensions:  []string{".jsonl", ".txt"},
		sqliteRoots: []string{filepath.Join(root, "ai-tracking", "ai-code-tracking.db")},
	}

	return parseGenericProvider(ctx, g, ParserConfig(g), paths)
}

// Name returns the Cursor Agent parser name.
func (CursorAgent) Name() string { return "Cursor Agent" }
