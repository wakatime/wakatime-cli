package legacy

import (
	"errors"
	"os"

	"github.com/wakatime/wakatime-cli/cmd/legacy/configread"
	"github.com/wakatime/wakatime-cli/cmd/legacy/configwrite"
<<<<<<< HEAD
	"github.com/wakatime/wakatime-cli/cmd/legacy/today"
=======
>>>>>>> 7b81186... add config-write flag
	"github.com/wakatime/wakatime-cli/pkg/config"
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

	if err := config.ReadInConfig(v, config.FilePath); err != nil {
		jww.CRITICAL.Printf("err: %s", err)

		var cfperr ErrConfigFileParse
		if errors.As(err, &cfperr) {
			os.Exit(exitcode.ErrConfigFileParse)
		}

		os.Exit(exitcode.ErrDefault)
	}

	if v.GetString("config-read") != "" {
		jww.DEBUG.Println("command: config-read")

		configread.Run(v)
	}

	if v.GetStringMapString("config-write") != nil {
		jww.DEBUG.Println("command: config-write")

		configwrite.Run(v)
	}

	if v.GetBool("today") {
		today.Run(v)
	}

	os.Exit(exitcode.Success)
}

func setVerbose(v *viper.Viper) {
	if v.GetBool("verbose") {
		jww.SetStdoutThreshold(jww.LevelDebug)
	}
}
