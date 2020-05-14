package legacy

import (
	"os"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

// Run executes legacy commands following the interface of the old python implementation of the WakaTime script.
func Run(v *viper.Viper) {
	setVerbose(v)

	if v.GetBool("version") {
		jww.DEBUG.Println("command: version")
		runVersion()
		os.Exit(0)
	}

	if v.GetString("config-read") != "" {
		jww.DEBUG.Println("command: config-read")
		runConfigRead(v)
		os.Exit(0)
	}
}

func setVerbose(v *viper.Viper) {
	if v.GetBool("verbose") {
		jww.SetStdoutThreshold(jww.LevelDebug)
	}
}
