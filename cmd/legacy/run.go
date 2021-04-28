package legacy

import (
	"errors"
	"os"

	"github.com/wakatime/wakatime-cli/cmd/legacy/configread"
	"github.com/wakatime/wakatime-cli/cmd/legacy/configwrite"
	"github.com/wakatime/wakatime-cli/cmd/legacy/heartbeat"
	"github.com/wakatime/wakatime-cli/cmd/legacy/logfile"
	"github.com/wakatime/wakatime-cli/cmd/legacy/today"
	"github.com/wakatime/wakatime-cli/cmd/legacy/todaygoal"
	"github.com/wakatime/wakatime-cli/pkg/config"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/spf13/viper"
)

// Run executes legacy commands following the interface of the old python implementation of the WakaTime script.
func Run(v *viper.Viper) {
	if err := config.ReadInConfig(v, config.FilePath); err != nil {
		log.Errorf("failed to load configuration file: %s", err)

		var cfperr ErrConfigFileParse
		if errors.As(err, &cfperr) {
			os.Exit(exitcode.ErrConfigFileParse)
		}

		os.Exit(exitcode.ErrDefault)
	}

	logfileParams, err := logfile.LoadParams(v)
	if err != nil {
		log.Fatalf("failed to load log params: %s", err)
	}

	f, err := os.OpenFile(logfileParams.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %s", err)
	}

	log.SetOutput(f)
	log.SetVerbose(logfileParams.Verbose)

	if v.GetBool("version") {
		log.Debugln("command: version")

		runVersion(v.GetBool("verbose"))

		os.Exit(exitcode.Success)
	}

	if v.IsSet("config-read") {
		log.Debugln("command: config-read")

		configread.Run(v)
	}

	if v.IsSet("config-write") {
		log.Debugln("command: config-write")

		configwrite.Run(v)
	}

	if v.GetBool("today") {
		log.Debugln("command: today")

		today.Run(v)
	}

	if v.IsSet("today-goal") {
		log.Debugln("command: today-goal")

		todaygoal.Run(v)
	}

	log.Debugln("command: heartbeat")

	heartbeat.Run(v)
}
