package configs

import (
	"fmt"
	"os"

	"github.com/alanhamlett/wakatime-cli/constants"
	"github.com/spf13/viper"
)

// WakaTimeConfig WakaTime configuration file interface
type WakaTimeConfig interface {
	Get(...string) *string
}

// NewConfig returns a configFile
func NewConfig(path string, v *viper.Viper) WakaTimeConfig {
	if path == "" {
		path = getConfigFile()
	}

	// load config file
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(constants.ConfigFileParseError)
	}
	defer f.Close()
	err = v.ReadConfig(f)
	if err != nil {
		fmt.Println(err)
		os.Exit(constants.ConfigFileParseError)
	}

	return &configFile{
		Config: v,
		Path:   path,
	}
}
