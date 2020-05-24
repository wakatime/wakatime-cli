package config

import (
	"fmt"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
)

// defaultFile is the name of the default wakatime config file.
const defaultFile = ".wakatime.cfg"

// ReadInConfig reads wakatime config file in memory.
func ReadInConfig(v *viper.Viper, filepathFn func(v *viper.Viper) (string, error)) error {
	configFilepath, err := filepathFn(v)
	if err != nil {
		return ErrFileParse(err.Error())
	}

	jww.DEBUG.Println("wakatime path:", configFilepath)

	v.SetConfigType("ini")
	v.SetConfigFile(configFilepath)

	if err := v.ReadInConfig(); err != nil {
		return ErrFileParse(fmt.Errorf("error parsing config file: %s", err).Error())
	}

	return nil
}

// FilePath returns the path for wakatime config file.
func FilePath(v *viper.Viper) (string, error) {
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

		return path.Join(p, defaultFile), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed getting user's home directory: %s", err)
	}

	return path.Join(home, defaultFile), nil
}

// LoadIni is an alternative solution for viper `WriteConfig()`.
// see issue https://github.com/spf13/viper/issues/858
func LoadIni(v *viper.Viper, filepathFn func(v *viper.Viper) (string, error)) (*ini.File, error) {
	configFilepath, err := filepathFn(v)
	if err != nil {
		return nil, ErrFileParse(err.Error())
	}

	cfg, err := ini.Load(configFilepath)
	if err != nil {
		return nil, ErrFileParse(fmt.Errorf("error parsing config file: %s", err).Error())
	}

	return cfg, nil
}
