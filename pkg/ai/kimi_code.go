package ai

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// KimiCode contains parameters for detecting heartbeats from Kimi Code logs.
type KimiCode ParserConfig

// Parse parses Kimi Code logs for AI heartbeats.
func (g KimiCode) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	paths := genericAIPaths{roots: kimiCodeRoots(home), fileNames: []string{"wire.jsonl"}}

	return parseGenericProvider(ctx, g, ParserConfig(g), paths)
}

func kimiCodeRoots(home string) []string {
	if override := strings.TrimSpace(os.Getenv("KIMI_CODE_HOME")); override != "" {
		return []string{override}
	}

	return []string{
		filepath.Join(home, ".kimi-code"),
		filepath.Join(
			home,
			"Library", "Application Support", "kimi-desktop", "daimon-share", "daimon",
			"runtime", "kimi-code", "home",
		),
	}
}

// Name returns the Kimi Code parser name.
func (KimiCode) Name() string { return "Kimi Code" }
