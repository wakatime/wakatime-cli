package legacy

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/go-homedir"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

const (
	errCodeConfigFileRead = 110
	defaultConfigFile     = ".wakatime.cfg"
)

// ConfigReadParams contains config read parameters.
type ConfigReadParams struct {
	Section string
	Key     string
}

// RunConfigRead prints value for the given config key.
func RunConfigRead(v *viper.Viper) error {
	params, err := LoadConfigReadParams(v)
	if err != nil {
		return err
	}

	value := strings.TrimSpace(v.GetString(params.ViperKey()))
	if value == "" {
		return ErrConfigFileRead(
			fmt.Errorf("given section and key \"%s\" returned an empty string", params.ViperKey()).Error(),
		)
	}

	fmt.Println(value)

	return nil
}

// LoadConfigReadParams loads needed data from the configuration file.
func LoadConfigReadParams(v *viper.Viper) (ConfigReadParams, error) {
	section := strings.TrimSpace(v.GetString("config-section"))
	key := strings.TrimSpace(v.GetString("config-read"))

	jww.DEBUG.Println("section:", section)
	jww.DEBUG.Println("key:", key)

	if section == "" || key == "" {
		return ConfigReadParams{},
			ErrConfigFileRead(fmt.Errorf("failed reading wakatime config file. neither section nor key can be empty").Error())
	}

	return ConfigReadParams{
		Section: section,
		Key:     key,
	}, nil
}

// ViperKey formats to a string [section].[key].
func (c *ConfigReadParams) ViperKey() string {
	return fmt.Sprintf("%s.%s", c.Section, c.Key)
}

// ReadInConfig reads wakatime config file in memory.
func ReadInConfig(v *viper.Viper, filepathFn func(v *viper.Viper) (string, error)) error {
	configFilepath, err := filepathFn(v)
	if err != nil {
		return ErrConfigFileParse(err.Error())
	}

	jww.DEBUG.Println("wakatime path:", configFilepath)

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

		return path.Join(p, defaultConfigFile), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed getting user's home directory: %s", err)
	}

	return path.Join(home, defaultConfigFile), nil
}
