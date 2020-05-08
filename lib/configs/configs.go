package configs

import (
	"os"
	"path"

	"github.com/alanhamlett/wakatime-cli/lib/system"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type configFile struct {
	Config *viper.Viper
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

// Get Get an option value for a given section.key
func (cf configFile) Get(keys ...string) *string {
	for _, k := range keys {
		if cf.Config.IsSet(k) {
			value := cf.Config.GetString(k)
			return &value
		}
	}
	return nil
}
