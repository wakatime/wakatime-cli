package legacy

import (
	"fmt"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

const (
	configFileParseError int = 103
)

func runConfigRead(v *viper.Viper) {
	loadConfigFile(v)

	section := v.GetString("config-section")
	key := v.GetString("config-read")

	jww.DEBUG.Println("section:", section)
	jww.DEBUG.Println("key:", key)

	value := v.GetString(section + "." + key)
	fmt.Println(value)
}

func loadConfigFile(v *viper.Viper) {
	p := v.GetString("config")

	if p == "" {
		p = getConfigFile()
	}

	jww.DEBUG.Println("wakatime path:", p)

	v.SetConfigType("ini")
	v.SetConfigFile(p)
	if err := v.ReadInConfig(); err != nil {
		jww.CRITICAL.Panicf("Error reading config file, %s", err)
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
	home = getHomeDirectory()

	return path.Join(home, fileName)
}

func getHomeDirectory() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return home
}
