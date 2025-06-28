package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadProjectConfig(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	// Create .wakatime config file
	configContent := `[settings]
api_key = waka_project-1234-4123-8123-123456789012
debug = true
exclude = .*project.*

[projectmap]
src/.* = project-core
`
	configFile := filepath.Join(projectDir, project.WakaTimeConfigFile)
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	// Create a test file
	testFile := filepath.Join(projectDir, "src", "main.go")
	require.NoError(t, os.MkdirAll(filepath.Dir(testFile), 0755))
	require.NoError(t, os.WriteFile(testFile, []byte("package main"), 0644))

	// Create global viper config
	globalViper := viper.New()
	globalViper.Set("settings.api_key", "waka_global-1234-4123-8123-123456789012")
	globalViper.Set("settings.debug", false)
	globalViper.Set("settings.timeout", 30)
	globalViper.Set("projectmap.lib/.*", "global-lib")

	tests := []struct {
		name       string
		filepath   string
		wantMerged bool
	}{
		{
			name:       "file in project with config",
			filepath:   testFile,
			wantMerged: true,
		},
		{
			name:       "file outside project",
			filepath:   filepath.Join(tmpDir, "other", "file.go"),
			wantMerged: false,
		},
		{
			name:       "empty filepath",
			filepath:   "",
			wantMerged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := project.LoadProjectConfig(t.Context(), globalViper, tt.filepath)

			if tt.wantMerged {
				// Should have project-specific overrides
				assert.Equal(t, "waka_project-1234-4123-8123-123456789012", result.GetString("settings.api_key"))
				assert.Equal(t, true, result.GetBool("settings.debug"))
				assert.Equal(t, ".*project.*", result.GetString("settings.exclude"))
				assert.Equal(t, "project-core", result.GetString("projectmap.src/.*"))

				// Should retain global settings not overridden
				assert.Equal(t, 30, result.GetInt("settings.timeout"))
				assert.Equal(t, "global-lib", result.GetString("projectmap.lib/.*"))
			} else {
				// Should be identical to global config
				assert.Equal(t, "waka_global-1234-4123-8123-123456789012", result.GetString("settings.api_key"))
				assert.Equal(t, false, result.GetBool("settings.debug"))
				assert.Equal(t, 30, result.GetInt("settings.timeout"))
				assert.Equal(t, "global-lib", result.GetString("projectmap.lib/.*"))
				assert.Equal(t, "", result.GetString("settings.exclude"))
			}
		})
	}
}

func TestGetProjectConfigPath(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	// Create .wakatime config file
	configFile := filepath.Join(projectDir, project.WakaTimeConfigFile)
	require.NoError(t, os.WriteFile(configFile, []byte("[project]\nname=test"), 0644))

	// Create a test file
	testFile := filepath.Join(projectDir, "src", "main.go")
	require.NoError(t, os.MkdirAll(filepath.Dir(testFile), 0755))
	require.NoError(t, os.WriteFile(testFile, []byte("package main"), 0644))

	tests := []struct {
		name     string
		filepath string
		wantPath string
	}{
		{
			name:     "file in project",
			filepath: testFile,
			wantPath: configFile,
		},
		{
			name:     "file outside project",
			filepath: filepath.Join(tmpDir, "other", "file.go"),
			wantPath: "",
		},
		{
			name:     "empty filepath",
			filepath: "",
			wantPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := project.GetProjectConfigPath(t.Context(), tt.filepath)
			assert.Equal(t, tt.wantPath, result)
		})
	}
}
