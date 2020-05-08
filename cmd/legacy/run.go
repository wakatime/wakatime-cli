package legacy

import (
	"fmt"
	"os"

	"github.com/alanhamlett/wakatime-cli/constants"
	"github.com/alanhamlett/wakatime-cli/lib/configs"
	"github.com/spf13/viper"
)

// Run Run legacy commands
func Run(cfg configs.WakaTimeConfig, v *viper.Viper) {
	if cfg == nil {
		configPath := v.GetString("config")
		cfg = configs.NewConfig(configPath, v)
	}

	if v.GetBool("version") {
		runVersion()
		os.Exit(constants.Success)
	}
}

func runVersion() {
	fmt.Println(constants.Version)
}
