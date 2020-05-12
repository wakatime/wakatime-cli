package legacy

import (
	"os"

	"github.com/spf13/viper"
)

// Run executes legacy commands following the interface of the old python implementation of the WakaTime script.
func Run(v *viper.Viper) {
	if v.GetBool("version") {
		runVersion()
		os.Exit(0)
	}
}
