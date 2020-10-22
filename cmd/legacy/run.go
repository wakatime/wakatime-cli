package legacy

import (
	"errors"
	"os"

	"github.com/wakatime/wakatime-cli/cmd/legacy/configread"
	"github.com/wakatime/wakatime-cli/cmd/legacy/configwrite"
	"github.com/wakatime/wakatime-cli/cmd/legacy/heartbeat"
	"github.com/wakatime/wakatime-cli/cmd/legacy/logfile"
	"github.com/wakatime/wakatime-cli/cmd/legacy/today"
	"github.com/wakatime/wakatime-cli/pkg/config"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

// Run executes legacy commands following the interface of the old python implementation of the WakaTime script.
func Run(v *viper.Viper) {
	logfile.Set(v)
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

	if v.IsSet("config-read") {
		jww.DEBUG.Println("command: config-read")

		configread.Run(v)
	}

	if v.IsSet("config-write") {
		jww.DEBUG.Println("command: config-write")

		configwrite.Run(v)
	}

	if v.GetBool("today") {
		jww.DEBUG.Println("command: today")

		today.Run(v)
	}

	jww.DEBUG.Println("command: heartbeat")

	heartbeat.Run(v)
}

func setVerbose(v *viper.Viper) {
	if v.GetBool("verbose") {
		jww.SetStdoutThreshold(jww.LevelDebug)
	}
}
