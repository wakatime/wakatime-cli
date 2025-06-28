package project

import (
	"context"
	"path/filepath"

	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/spf13/viper"
)

// LoadProjectConfig attempts to load project-specific configuration for the given file path.
// It searches for .wakatime files in the file's directory hierarchy and merges the configuration
// with the provided global viper instance.
// Returns a new viper instance with merged configuration, or the original viper if no project config found.
func LoadProjectConfig(ctx context.Context, globalViper *viper.Viper, filePath string) *viper.Viper {
	if filePath == "" {
		return globalViper
	}

	logger := log.Extract(ctx)

	// Try to detect project configuration
	configDetector := ConfigDetector{
		Filepath: filePath,
	}

	configResult, err := configDetector.DetectConfig(ctx)
	if err != nil {
		logger.Debugf("failed to detect project config for %s: %s", filePath, err)
		return globalViper
	}

	if !configResult.Found {
		logger.Debugf("no project config found for %s", filePath)
		return globalViper
	}

	logger.Debugf("merging project config from %s", filepath.Join(filepath.Dir(configResult.Folder), WakaTimeConfigFile))

	// Create a new viper instance for merged configuration
	mergedViper := viper.New()

	// First, copy all settings from global config
	allGlobalSettings := globalViper.AllSettings()
	mergedSettings := deepCopyMap(allGlobalSettings)

	// Then, merge project-specific settings
	allProjectSettings := configResult.Config.AllSettings()
	deepMergeMap(mergedSettings, allProjectSettings)

	// Set all merged settings in new viper
	for key, value := range mergedSettings {
		mergedViper.Set(key, value)
	}

	return mergedViper
}

// GetProjectConfigPath returns the path to the project config file for the given file path,
// or empty string if no project config is found.
func GetProjectConfigPath(ctx context.Context, filePath string) string {
	if filePath == "" {
		return ""
	}

	configPath, found := FindFileOrDirectory(ctx, filePath, WakaTimeConfigFile)
	if !found {
		return ""
	}

	return configPath
}

// deepCopyMap creates a deep copy of a map[string]interface{}.
func deepCopyMap(original map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for key, value := range original {
		copy[key] = deepCopyValue(value)
	}

	return copy
}

// deepCopyValue creates a deep copy of an interface{} value.
func deepCopyValue(original interface{}) interface{} {
	switch original := original.(type) {
	case map[string]interface{}:
		return deepCopyMap(original)
	case []interface{}:
		copy := make([]interface{}, len(original))
		for i, v := range original {
			copy[i] = deepCopyValue(v)
		}

		return copy
	default:
		return original
	}
}

// deepMergeMap merges source map into destination map recursively.
func deepMergeMap(dest, source map[string]interface{}) {
	for key, sourceValue := range source {
		if destValue, exists := dest[key]; exists {
			// Both source and dest have this key
			destMap, destIsMap := destValue.(map[string]interface{})
			sourceMap, sourceIsMap := sourceValue.(map[string]interface{})

			if destIsMap && sourceIsMap {
				// Both are maps, merge recursively
				deepMergeMap(destMap, sourceMap)
			} else {
				// One or both are not maps, override with source value
				dest[key] = deepCopyValue(sourceValue)
			}
		} else {
			// Key only exists in source, copy it
			dest[key] = deepCopyValue(sourceValue)
		}
	}
}
