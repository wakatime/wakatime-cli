package configs

import (
	"fmt"
	"os"

	"github.com/go-ini/ini"
	"github.com/wakatime/wakatime-cli/constants"
)

// WakaTimeConfig WakaTime configuration file interface
type WakaTimeConfig interface {
	Get(string, string) (string, error)
	GetSectionMap(string) (map[string]string, error)
	GetConfigForPlugin(string) map[string]string
	Set(string, map[string]string) []string
}

// NewConfig returns a configFile
func NewConfig(path string) WakaTimeConfig {
	if path == "" {
		path = getConfigFile()
	}

	cfg, err := ini.Load(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(constants.ConfigFileParseError)
	}

	return &configFile{
		Config: cfg,
		Path:   path,
	}
}
