package ai

import (
	"context"
	"path/filepath"
	"time"
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
		sqliteQuery:         cursorAgentSQLiteQuery,
		sqliteParserVersion: 2,
	}

	return parseGenericProvider(ctx, g, ParserConfig(g), paths)
}

// Attribution records do not need historical token or prompt baselines. Filter
// their millisecond timestamps before transferring rows to the shared scanner.
func cursorAgentSQLiteQuery(_ string, projection string, after time.Time) (string, []any) {
	query := "SELECT " + projection + " FROM ai_code_hashes WHERE 1=1"
	if after.IsZero() {
		return query, nil
	}

	return query + " AND timestamp >= ?", []any{after.UnixMilli()}
}

// Name returns the Cursor Agent parser name.
func (CursorAgent) Name() string { return "Cursor Agent" }
