package legacy

import (
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

// Run executes legacy commands following the interface of the old python implementation of the WakaTime script.
func Run(v *viper.Viper) error {
	setVerbose(v)

	if v.GetBool("version") {
		jww.DEBUG.Println("command: version")
		runVersion()
	}

	if v.GetString("config-read") != "" {
		jww.DEBUG.Println("command: config-read")
		runConfigRead(v)
	}

	return nil
}

func setVerbose(v *viper.Viper) {
	if v.GetBool("verbose") {
		jww.SetStdoutThreshold(jww.LevelDebug)
	}
}
