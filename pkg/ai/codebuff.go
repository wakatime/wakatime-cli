package ai

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// Codebuff contains parameters for detecting heartbeats from Codebuff logs.
type Codebuff ParserConfig

// Parse parses Codebuff logs for AI heartbeats.
func (g Codebuff) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	return parseGenericProvider(ctx, g, ParserConfig(g), codebuffPaths(home))
}

func codebuffPaths(home string) genericAIPaths {
	if override := strings.TrimSpace(os.Getenv("CODEBUFF_DATA_DIR")); override != "" {
		return genericAIPaths{roots: []string{override}, fileNames: []string{"chat-messages.json"}}
	}

	config := filepath.Join(home, ".config")

	return genericAIPaths{
		roots: []string{
			filepath.Join(config, "manicode"),
			filepath.Join(config, "manicode-dev"),
			filepath.Join(config, "manicode-staging"),
		},
		fileNames: []string{"chat-messages.json"},
	}
}

// Name returns the Codebuff parser name.
func (Codebuff) Name() string { return "Codebuff" }
