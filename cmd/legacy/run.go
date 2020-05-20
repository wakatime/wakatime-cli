package legacy

import (
	"errors"
	"os"

	"github.com/wakatime/wakatime-cli/cmd/legacy/config"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

// Run executes legacy commands following the interface of the old python implementation of the WakaTime script.
func Run(v *viper.Viper) {
	setVerbose(v)

	if v.GetBool("version") {
		jww.DEBUG.Println("command: version")
		runVersion()

		os.Exit(exitcode.Success)
	}

	if err := ReadInConfig(v, ConfigFilePath); err != nil {
		jww.CRITICAL.Printf("err: %s", err)

		var cfperr ErrConfigFileParse
		if errors.As(err, &cfperr) {
			os.Exit(exitcode.ErrConfigFileParse)
		}

		os.Exit(exitcode.ErrDefault)
	}

	if v.GetString("config-read") != "" {
		jww.DEBUG.Println("command: config-read")

		if err := config.RunRead(v); err != nil {
			jww.ERROR.Printf("err: %s", err)

			var cfrerr config.ErrFileRead
			if errors.As(err, &cfrerr) {
				os.Exit(exitcode.ErrConfigFileRead)
			}

			os.Exit(exitcode.ErrDefault)
		}
	}

	os.Exit(exitcode.Success)
}

func setVerbose(v *viper.Viper) {
	if v.GetBool("verbose") {
		jww.SetStdoutThreshold(jww.LevelDebug)
	}
}
