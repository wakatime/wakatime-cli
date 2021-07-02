package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
)

// defaultFile is the name of the default wakatime config file.
const defaultFile = ".wakatime.cfg"

// Writer defines the methods to write to config file.
type Writer interface {
	Write(section string, keyValue map[string]string) error
}

// IniWriter stores the configuration necessary to write to config file.
type IniWriter struct {
	File           *ini.File
	ConfigFilepath string
}

// NewIniWriter creates a new IniWriter instance.
func NewIniWriter(v *viper.Viper, filepathFn func(v *viper.Viper) (string, error)) (*IniWriter, error) {
	configFilepath, err := filepathFn(v)
	if err != nil {
		return nil, fmt.Errorf("error getting filepath: %s", err)
	}

	ini, err := ini.Load(configFilepath)
	if err != nil {
		return nil, fmt.Errorf("error loading config file: %s", err)
	}

	return &IniWriter{
		File:           ini,
		ConfigFilepath: configFilepath,
	}, nil
}

// Write persists key(s) and value(s) on disk.
func (w *IniWriter) Write(section string, keyValue map[string]string) error {
	if w.File == nil || w.ConfigFilepath == "" {
		return errors.New("got undefined wakatime config file instance")
	}

	for key, value := range keyValue {
		w.File.Section(section).Key(key).SetValue(value)
	}

	if err := w.File.SaveTo(w.ConfigFilepath); err != nil {
		return fmt.Errorf("error saving wakatime config: %s", err)
	}

	return nil
}

// ReadInConfig reads wakatime config file in memory.
func ReadInConfig(v *viper.Viper, filepathFn func(v *viper.Viper) (string, error)) error {
	configFilepath, err := filepathFn(v)
	if err != nil {
		return fmt.Errorf("error getting filepath: %s", err)
	}

	v.SetConfigType("ini")
	v.SetConfigFile(configFilepath)

	// check if file exists
	if _, err := os.Stat(configFilepath); os.IsNotExist(err) {
		log.Debugf("config file not present or not accessible")

		return nil
	}

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("error parsing config file: %s", err)
	}

	return nil
}

// FilePath returns the path for wakatime config file.
func FilePath(v *viper.Viper) (string, error) {
	configFilepath := vipertools.GetString(v, "config")
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

		return filepath.Join(p, defaultFile), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed getting user's home directory: %s", err)
	}

	return filepath.Join(home, defaultFile), nil
}
