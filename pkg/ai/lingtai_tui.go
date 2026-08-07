package ai

import (
	"context"
	"os"
	"path/filepath"
)

// LingTaiTUI contains parameters for detecting heartbeats from LingTai TUI logs.
type LingTaiTUI ParserConfig

// Parse parses LingTai TUI logs for AI heartbeats.
func (g LingTaiTUI) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	return parseGenericProvider(ctx, g, ParserConfig(g), lingTaiPaths(home))
}

func lingTaiPaths(home string) genericAIPaths {
	override := firstNonEmptyString(os.Getenv("LINGTAI_HOME"), os.Getenv("LINGTAI_TUI_HOME"))
	if override != "" {
		return genericAIPaths{roots: filepath.SplitList(override), fileNames: []string{"token_ledger.jsonl"}}
	}

	return genericAIPaths{
		roots:     []string{filepath.Join(home, ".lingtai")},
		fileNames: []string{"token_ledger.jsonl"},
	}
}

// Name returns the LingTai TUI parser name.
func (LingTaiTUI) Name() string { return "LingTai TUI" }
