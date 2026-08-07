package ai

import (
	"context"
	"path/filepath"
	"runtime"
)

// IBMBob contains parameters for detecting heartbeats from IBM Bob logs.
type IBMBob ParserConfig

// Parse parses IBM Bob logs for AI heartbeats.
func (g IBMBob) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	paths := genericAIPaths{roots: ibmBobRoots(home), fileNames: []string{"ui_messages.json"}}

	return parseGenericProvider(ctx, g, ParserConfig(g), paths)
}

func ibmBobRoots(home string) []string {
	var bases []string

	switch runtime.GOOS {
	case "darwin":
		bases = []string{
			filepath.Join(home, "Library", "Application Support", "IBM Bob"),
			filepath.Join(home, "Library", "Application Support", "Bob-IDE"),
		}
	case "windows":
		appData := envOrDefault("APPDATA", filepath.Join(home, "AppData", "Roaming"))
		bases = []string{filepath.Join(appData, "IBM Bob"), filepath.Join(appData, "Bob-IDE")}
	default:
		config := envOrDefault("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
		bases = []string{filepath.Join(config, "IBM Bob"), filepath.Join(config, "Bob-IDE")}
	}

	roots := make([]string, 0, len(bases))
	for _, base := range bases {
		roots = append(roots, filepath.Join(base, "User", "globalStorage", "ibm.bob-code", "tasks"))
	}

	return roots
}

// Name returns the IBM Bob parser name.
func (IBMBob) Name() string { return "IBM Bob" }
