package legacy

import (
	"fmt"
	"os"
	"path"

	"github.com/wakatime/wakatime-cli/pkg/config"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// ReadInConfig reads wakatime config file in memory.
func ReadInConfig(v *viper.Viper, filepathFn func(v *viper.Viper) (string, error)) error {
	configFilepath, err := filepathFn(v)
	if err != nil {
		return ErrConfigFileParse(err.Error())
	}

	v.SetConfigType("ini")
	v.SetConfigFile(configFilepath)

	if err := v.ReadInConfig(); err != nil {
		return ErrConfigFileParse(err.Error())
	}

	return nil
}

// ConfigFilePath returns the path for wakatime config file.
func ConfigFilePath(v *viper.Viper) (string, error) {
	configFilepath := v.GetString("config")
	if configFilepath != "" {
		p, err := homedir.Expand(configFilepath)
		if err != nil {
			return "", fmt.Errorf("failed parsing config flag variable: %s", err)
		}

		return p, nil
	}

	home, exists := os.LookupEnv("WAKATIME_HOME")
	if exists && home != "" {
		p, err := homedir.Expand(home)
		if err != nil {
			return "", fmt.Errorf("failed parsing WAKATIME_HOME environment variable: %s", err)
		}

		return path.Join(p, config.DefaultFile), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed getting user's home directory: %s", err)
	}

	return path.Join(home, config.DefaultFile), nil
}
