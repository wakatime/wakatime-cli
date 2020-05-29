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
		return nil, ErrFileParse(err.Error())
	}

	ini, err := ini.Load(configFilepath)
	if err != nil {
		return nil, ErrFileParse(fmt.Sprintf("error parsing config file: %s", err))
	}

	return &IniWriter{
		File:           ini,
		ConfigFilepath: configFilepath,
	}, nil
}

// Write persists key(s) and value(s) on disk.
func (w *IniWriter) Write(section string, keyValue map[string]string) error {
	if w.File == nil || w.ConfigFilepath == "" {
		return ErrFileWrite(fmt.Errorf("got undefined wakatime config file instance").Error())
	}

	for key, value := range keyValue {
		w.File.Section(section).Key(key).SetValue(value)
	}

	if err := w.File.SaveTo(w.ConfigFilepath); err != nil {
		return ErrFileWrite(fmt.Errorf("error saving wakatime config: %s", err).Error())
	}

	return nil
}

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
