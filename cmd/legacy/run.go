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

		os.Exit(successCode)
	}

	if err := ReadInConfig(v, ConfigFilePath); err != nil {
		jww.CRITICAL.Printf("err: %s", err)

		var cfperr ErrConfigFileParse
		if errors.As(err, &cfperr) {
			os.Exit(errCodeConfigFileParse)
		}

		os.Exit(errCodeDefault)
	}

	if v.GetString("config-read") != "" {
		jww.DEBUG.Println("command: config-read")

		if err := RunConfigRead(v); err != nil {
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
