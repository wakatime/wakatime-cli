package project

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
)

const (
	// WakaTimeConfigFile is the new INI config file which contains project settings
	// and can override global configuration per project.
	WakaTimeConfigFile = ".wakatime"
)

// ConfigDetector detects project configuration from .wakatime files.
type ConfigDetector struct {
	Filepath string
}

// ConfigResult contains both project information and configuration data.
type ConfigResult struct {
	Result
	Config *viper.Viper
	Found  bool
}

// DetectConfig finds project configuration from .wakatime INI files.
// It reads the full configuration file and returns both project info and config.
func (c ConfigDetector) DetectConfig(ctx context.Context) (ConfigResult, error) {
	fp, found := FindFileOrDirectory(ctx, c.Filepath, WakaTimeConfigFile)
	if !found {
		return ConfigResult{Found: false}, nil
	}

	logger := log.Extract(ctx)
	logger.Debugf("wakatime config file found at: %s", fp)

	// Load the INI file
	cfg, err := ini.LoadSources(ini.LoadOptions{
		AllowPythonMultilineValues: true,
		SkipUnrecognizableLines:    true,
	}, fp)
	if err != nil {
		return ConfigResult{}, fmt.Errorf("failed to load config file %s: %s", fp, err)
	}

	// Create a new viper instance for this project config
	projectViper := viper.New()
	projectViper.SetConfigType("ini")

	// Convert INI sections to viper format
	for _, section := range cfg.Sections() {
		sectionName := section.Name()
		if sectionName == "DEFAULT" {
			// Skip the default section
			continue
		}

		for _, key := range section.Keys() {
			viperKey := fmt.Sprintf("%s.%s", sectionName, key.Name())
			projectViper.Set(viperKey, key.Value())
		}
	}

	// Extract project information from the config
	result := Result{
		Folder:  filepath.Dir(fp),
		Project: filepath.Base(filepath.Dir(fp)),
	}

	// Check for project section with name and branch
	if projectSection, err := cfg.GetSection("project"); err == nil {
		if nameKey, err := projectSection.GetKey("name"); err == nil {
			result.Project = strings.TrimSpace(nameKey.Value())
		}

		if branchKey, err := projectSection.GetKey("branch"); err == nil {
			result.Branch = strings.TrimSpace(branchKey.Value())
		}
	}

	return ConfigResult{
		Result: result,
		Config: projectViper,
		Found:  true,
	}, nil
}

// Detect implements the Detecter interface for backwards compatibility.
// It only returns project information without configuration data.
func (c ConfigDetector) Detect(ctx context.Context) (Result, bool, error) {
	configResult, err := c.DetectConfig(ctx)
	if err != nil {
		return Result{}, false, err
	}

	return configResult.Result, configResult.Found, nil
}

// ID returns the detector ID.
func (ConfigDetector) ID() DetectorID {
	return FileDetector // Use same ID as File detector for now
}
