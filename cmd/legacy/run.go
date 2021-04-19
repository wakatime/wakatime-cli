package legacy

import (
	"errors"
	"os"

	"github.com/wakatime/wakatime-cli/cmd/legacy/configread"
	"github.com/wakatime/wakatime-cli/cmd/legacy/configwrite"
	"github.com/wakatime/wakatime-cli/cmd/legacy/heartbeat"
	"github.com/wakatime/wakatime-cli/cmd/legacy/today"
	"github.com/wakatime/wakatime-cli/cmd/legacy/todaygoal"
	"github.com/wakatime/wakatime-cli/pkg/config"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	log "github.com/wakatime/wakatime-cli/pkg/logfile2"

	"github.com/spf13/viper"
)

// Run executes legacy commands following the interface of the old python implementation of the WakaTime script.
func Run(v *viper.Viper) {
	if err := config.ReadInConfig(v, config.FilePath); err != nil {
		log.LogEntry.Errorf("failed to load configuration file: %s", err)

		var cfperr ErrConfigFileParse
		if errors.As(err, &cfperr) {
			os.Exit(exitcode.ErrConfigFileParse)
		}

		os.Exit(exitcode.ErrDefault)
	}

	log.SetOutput(v)
	log.SetVerbose(v)

	if v.GetBool("version") {
		log.LogEntry.Debugln("command: version")

		runVersion(v.GetBool("verbose"))

		os.Exit(exitcode.Success)
	}

	if v.IsSet("config-read") {
		log.LogEntry.Debugln("command: config-read")

		configread.Run(v)
	}

	if v.IsSet("config-write") {
		log.LogEntry.Debugln("command: config-write")

		configwrite.Run(v)
	}

	if v.GetBool("today") {
		log.LogEntry.Debugln("command: today")

		today.Run(v)
	}

	if v.IsSet("today-goal") {
		log.LogEntry.Debugln("command: today-goal")

		todaygoal.Run(v)
	}

	log.LogEntry.Debugln("command: heartbeat")

	heartbeat.Run(v)
}
