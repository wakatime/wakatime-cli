package ai

import (
	"context"
	"path/filepath"
)

// ClineCLI contains parameters for detecting heartbeats from Cline CLI logs.
type ClineCLI ParserConfig

// Parse parses Cline CLI logs for AI heartbeats.
func (g ClineCLI) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	root := envOrDefault("CLINE_DIR", filepath.Join(home, ".cline"))
	sessions := envOrDefault("CLINE_SESSION_DATA_DIR", filepath.Join(root, "data", "sessions"))

	paths := genericJSONPaths(sessions, ".json")
	paths.containerKey = "messages"
	paths.preferMessagesFile = true

	return parseGenericProvider(ctx, g, ParserConfig(g), paths)
}

// Name returns the Cline CLI parser name.
func (ClineCLI) Name() string { return "Cline CLI" }
