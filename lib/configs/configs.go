package configs

import (
	"fmt"
	"os"
	"path"

	"github.com/go-ini/ini"
	"github.com/mitchellh/go-homedir"
	"github.com/wakatime/wakatime-cli/constants"
	"github.com/wakatime/wakatime-cli/lib/system"
)

// ConfigFile ConfigFile
type ConfigFile struct {
	Config *ini.File
	Path   string
}

//NewConfig returns a configFile
func NewConfig(path string) *ConfigFile {
	if path == "" {
		path = getConfigFile()
	}

	cfg, err := ini.Load(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(constants.ConfigFileParseError)
	}

	return &ConfigFile{
		Config: cfg,
		Path:   path,
	}
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

//Get Get an option value for a given section
func (cf ConfigFile) Get(section string, key string) (*string, error) {
	if section == "" {
		section = "settings"
	}

	s, err := cf.Config.GetSection(section)
	if err != nil {
		return nil, err
	}

	hasKey := s.HasKey(key)
	if !hasKey {
		return nil, fmt.Errorf("The given key '%v' was not found on section '%v'", key, section)
	}

	v := s.Key(key).Value()

	return &v, nil
}

// Set Set an option
func (cf ConfigFile) Set(section string, keyValue map[string]string) []string {
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
