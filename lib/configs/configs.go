package configs

import (
	"fmt"
	"os"
	"path"

	"github.com/go-ini/ini"
	"github.com/mitchellh/go-homedir"
	"github.com/wakatime/wakatime-cli/lib/system"
)

type configFile struct {
	Config *ini.File
	Path   string
}

func getConfigFile() string {
	fileName := ".wakatime.cfg"
	home, exists := os.LookupEnv("WAKATIME_HOME")

	if exists {
		p, err := homedir.Expand(home)
		if err != nil {
			panic(err)
		}
		return path.Join(p, fileName)
	}

	home = system.GetHomeDirectory()

	return path.Join(home, fileName)
}

// Get Get an option value for a given section
func (cf configFile) Get(section string, key string) (string, error) {
	if section == "" {
		section = "settings"
	}

	s, err := cf.Config.GetSection(section)
	if err != nil {
		return "", err
	}

	hasKey := s.HasKey(key)
	if !hasKey {
		return "", fmt.Errorf("The given key '%v' was not found on section '%v'", key, section)
	}

	v := s.Key(key).Value()

	return v, nil
}

// GetSectionMap Get a map for a given section
func (cf configFile) GetSectionMap(section string) (map[string]string, error) {
	s, err := cf.Config.GetSection(section)
	if err != nil {
		return nil, err
	}

	return s.KeysHash(), nil
}

// Set Set an option
func (cf configFile) Set(section string, keyValue map[string]string) []string {
	if section == "" {
		section = "settings"
	}

	messages := []string{}
	for key, value := range keyValue {
		s, err := cf.Config.GetSection(section)
		if err != nil {
			messages = append(messages, err.Error())
		}

		hasKey := s.HasKey(key)
		if !hasKey {
			s.NewKey(key, value)
			messages = append(messages, fmt.Sprintf("Key '%v' successfully created", key))
			continue
		}

		v, _ := s.GetKey(key)
		v.SetValue(value)
		messages = append(messages, fmt.Sprintf("Key '%v' changed successfully", key))
	}

	cf.Config.SaveTo(cf.Path)

	return messages
}

// GetConfigForPlugin GetConfigForPlugin
func (cf configFile) GetConfigForPlugin(pluginName string) map[string]string {
	values, _ := cf.GetSectionMap(pluginName)
	return values
}
