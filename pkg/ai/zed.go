package ai

import (
	"context"
	"path/filepath"
	"runtime"
)

// Zed contains parameters for detecting heartbeats from Zed logs.
type Zed ParserConfig

// Parse parses Zed logs for AI heartbeats.
func (g Zed) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	return parseGenericProvider(ctx, g, ParserConfig(g), genericAIPaths{sqliteRoots: []string{zedDBPath(home)}})
}

func zedDBPath(home string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Zed", "threads", "threads.db")
	case "windows":
		return filepath.Join(home, "AppData", "Local", "Zed", "threads", "threads.db")
	default:
		return filepath.Join(dataHome(home), "zed", "threads", "threads.db")
	}
}

// Name returns the Zed parser name.
func (Zed) Name() string { return "Zed" }
