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

	transcriptRoots, err := immediateChildPaths(filepath.Join(root, "projects"), "agent-transcripts")
	if err != nil {
		return nil, err
	}

	paths := genericAIPaths{
		roots:       transcriptRoots,
		extensions:  []string{".jsonl", ".txt"},
		sqliteRoots: []string{filepath.Join(root, "ai-tracking", "ai-code-tracking.db")},
		// Cursor's tracking database contains code attribution in ai_code_hashes.
		// The other tables include full tracked-file snapshots and commit history,
		// neither of which produces agent activity heartbeats.
		sqliteTables:        []string{"ai_code_hashes"},
		sqliteParserVersion: 1,
	}

	return parseGenericProvider(ctx, g, ParserConfig(g), paths)
}

// Name returns the Cursor Agent parser name.
func (CursorAgent) Name() string { return "Cursor Agent" }
