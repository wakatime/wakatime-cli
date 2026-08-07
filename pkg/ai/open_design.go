package ai

import (
	"context"
	"path/filepath"
	"runtime"
)

// OpenDesign contains parameters for detecting heartbeats from Open Design logs.
type OpenDesign ParserConfig

// Parse parses Open Design logs for AI heartbeats.
func (g OpenDesign) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	paths := genericAIPaths{
		roots:     []string{openDesignRoot(home)},
		fileNames: []string{"events.jsonl"},
	}

	return parseGenericProvider(ctx, g, ParserConfig(g), paths)
}

func openDesignRoot(home string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Open Design")
	case "windows":
		appData := envOrDefault("APPDATA", filepath.Join(home, "AppData", "Roaming"))
		return filepath.Join(appData, "Open Design")
	default:
		return filepath.Join(home, ".config", "Open Design")
	}
}

// Name returns the Open Design parser name.
func (OpenDesign) Name() string { return "Open Design" }
