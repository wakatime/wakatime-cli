package ai

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// CodeWhale contains parameters for detecting heartbeats from CodeWhale logs.
type CodeWhale ParserConfig

// Parse parses CodeWhale logs for AI heartbeats.
func (g CodeWhale) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	return parseGenericProvider(ctx, g, ParserConfig(g), codeWhalePaths(home))
}

func codeWhalePaths(home string) genericAIPaths {
	if override := strings.TrimSpace(os.Getenv("CODEWHALE_HOME")); override != "" {
		paths := genericJSONPaths(filepath.Join(override, "sessions"), ".json")
		paths.tokenCounterMode = genericAICumulativeCounters

		return paths
	}

	return genericAIPaths{
		roots: []string{
			filepath.Join(home, ".codewhale", "sessions"),
			filepath.Join(home, ".deepseek", "sessions"),
		},
		extensions:       []string{".json"},
		tokenCounterMode: genericAICumulativeCounters,
	}
}

// Name returns the CodeWhale parser name.
func (CodeWhale) Name() string { return "CodeWhale" }
