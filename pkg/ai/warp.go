package ai

import (
	"context"
	"path/filepath"
	"runtime"
)

// Warp contains parameters for detecting heartbeats from Warp logs.
type Warp ParserConfig

// Parse parses Warp logs for AI heartbeats.
func (g Warp) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	paths := genericAIPaths{sqliteRoots: []string{warpDBPath(home)}}

	return parseGenericProvider(ctx, g, ParserConfig(g), paths)
}

func warpDBPath(home string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "dev.warp.Warp-Stable", "warp.sqlite")
	case "windows":
		localAppData := envOrDefault("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
		return filepath.Join(localAppData, "warp", "Warp", "data", "warp.sqlite")
	default:
		return filepath.Join(dataHome(home), "warp-terminal", "warp.sqlite")
	}
}

// Name returns the Warp parser name.
func (Warp) Name() string { return "Warp" }
