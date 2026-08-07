package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/ini"
)

type genericAIPaths struct {
	roots              []string
	fileNames          []string
	extensions         []string
	sqliteRoots        []string
	containerKey       string
	preferMessagesFile bool
	tokenCounterMode   genericAICounterMode
	lineCounterMode    genericAICounterMode
}

type genericAICounterMode uint8

const (
	genericAIPerEventCounters genericAICounterMode = iota
	genericAICumulativeCounters
)

func parseGenericProvider(
	ctx context.Context,
	parser Parser,
	config ParserConfig,
	paths genericAIPaths,
) (Heartbeats, error) {
	return parseGenericAIProvider(ctx, genericAIProvider{
		parser:             parser,
		config:             config,
		roots:              paths.roots,
		fileNames:          stringSet(paths.fileNames...),
		extensions:         stringSet(paths.extensions...),
		sqliteRoots:        paths.sqliteRoots,
		containerKey:       paths.containerKey,
		preferMessagesFile: paths.preferMessagesFile,
		tokenCounterMode:   paths.tokenCounterMode,
		lineCounterMode:    paths.lineCounterMode,
	})
}

func genericJSONPaths(root string, extension string) genericAIPaths {
	return genericAIPaths{
		roots:            []string{root},
		extensions:       []string{extension},
		tokenCounterMode: genericAIPerEventCounters,
		lineCounterMode:  genericAIPerEventCounters,
	}
}

func aiUserHome(ctx context.Context) (string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to find user home dir: %s", err)
	}

	return home, nil
}

func vscodeExtensionTaskRoots(home string, extensionID string) []string {
	applications := []string{"Code", "Cursor", "Windsurf", "VSCodium"}
	roots := make([]string, 0, len(applications)*4)

	for _, application := range applications {
		macApplicationSupport := filepath.Join(home, "Library", "Application Support", application)
		roots = append(roots,
			filepath.Join(macApplicationSupport, "User", "globalStorage", extensionID, "tasks"),
			filepath.Join(home, "AppData", "Roaming", application, "User", "globalStorage", extensionID, "tasks"),
			filepath.Join(home, ".config", application, "User", "globalStorage", extensionID, "tasks"),
		)
	}

	return append(roots,
		filepath.Join(home, ".vscode-server", "data", "User", "globalStorage", extensionID, "tasks"),
		filepath.Join(home, ".cursor-server", "data", "User", "globalStorage", extensionID, "tasks"),
	)
}

func dataHome(home string) string {
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support")
	}

	if runtime.GOOS == "windows" {
		return envOrDefault("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	}

	return envOrDefault("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
}

func envOrDefault(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}

func stringSet(values ...string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}

	return result
}
