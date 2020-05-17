package legacy

import (
	"errors"
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
	}

	if v.GetString("config-read") != "" {
		jww.DEBUG.Println("command: config-read")

		if err := runConfigRead(v); err != nil {
			jww.ERROR.Printf("err: %s", err)

			var cfrerr ErrConfigFileRead
			if errors.As(err, &cfrerr) {
				os.Exit(errCodeConfigFileRead)
			}

			os.Exit(errCodeDefault)
		}
	}

	os.Exit(successCode)
}

func setVerbose(v *viper.Viper) {
	if v.GetBool("verbose") {
		jww.SetStdoutThreshold(jww.LevelDebug)
	}
}
