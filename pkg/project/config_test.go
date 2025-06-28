package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDetector_DetectConfig(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	// Create .wakatime config file
	configContent := `[project]
name = test-cli-project
branch = feature/test

[settings]
api_key = waka_12345678-1234-4123-8123-123456789012
debug = true
exclude = .*test.*

[projectmap]
src/.* = core-module
`
	configFile := filepath.Join(projectDir, project.WakaTimeConfigFile)
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	// Create a test file
	testFile := filepath.Join(projectDir, "src", "main.go")
	require.NoError(t, os.MkdirAll(filepath.Dir(testFile), 0755))
	require.NoError(t, os.WriteFile(testFile, []byte("package main"), 0644))

	tests := []struct {
		name           string
		filepath       string
		wantFound      bool
		wantProject    string
		wantBranch     string
		wantConfigKeys []string
	}{
		{
			name:           "file in project directory",
			filepath:       testFile,
			wantFound:      true,
			wantProject:    "test-cli-project",
			wantBranch:     "feature/test",
			wantConfigKeys: []string{"settings.api_key", "settings.debug", "settings.exclude", "projectmap.src/.*"},
		},
		{
			name:      "file outside project",
			filepath:  filepath.Join(tmpDir, "other", "file.go"),
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := project.ConfigDetector{
				Filepath: tt.filepath,
			}

			result, err := detector.DetectConfig(t.Context())
			require.NoError(t, err)

			assert.Equal(t, tt.wantFound, result.Found)

			if tt.wantFound {
				assert.Equal(t, tt.wantProject, result.Project)
				assert.Equal(t, tt.wantBranch, result.Branch)
				assert.Equal(t, projectDir, result.Folder)

				// Check that config values are loaded
				for _, key := range tt.wantConfigKeys {
					assert.True(t, result.Config.IsSet(key), "expected config key %s to be set", key)
				}

				// Verify specific values
				assert.Equal(t, "waka_12345678-1234-4123-8123-123456789012", result.Config.GetString("settings.api_key"))
				assert.Equal(t, "true", result.Config.GetString("settings.debug"))
				assert.Equal(t, ".*test.*", result.Config.GetString("settings.exclude"))
				assert.Equal(t, "core-module", result.Config.GetString("projectmap.src/.*"))
			}
		})
	}
}

func TestConfigDetector_Detect(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	// Create .wakatime config file with minimal project info
	configContent := `[project]
name = minimal-project
branch = main
`
	configFile := filepath.Join(projectDir, project.WakaTimeConfigFile)
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	testFile := filepath.Join(projectDir, "file.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	detector := project.ConfigDetector{
		Filepath: testFile,
	}

	result, found, err := detector.Detect(t.Context())
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "minimal-project", result.Project)
	assert.Equal(t, "main", result.Branch)
	assert.Equal(t, projectDir, result.Folder)
}

func TestConfigDetector_NoConfigFile(t *testing.T) {
	// Create temporary directory without config file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "file.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	detector := project.ConfigDetector{
		Filepath: testFile,
	}

	result, err := detector.DetectConfig(t.Context())
	require.NoError(t, err)
	assert.False(t, result.Found)
}
