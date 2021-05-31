package legacy

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/wakatime/wakatime-cli/cmd/legacy/configread"
	"github.com/wakatime/wakatime-cli/cmd/legacy/configwrite"
	heartbeatcmd "github.com/wakatime/wakatime-cli/cmd/legacy/heartbeat"
	"github.com/wakatime/wakatime-cli/cmd/legacy/logfile"
	"github.com/wakatime/wakatime-cli/cmd/legacy/offlinesync"
	"github.com/wakatime/wakatime-cli/cmd/legacy/today"
	"github.com/wakatime/wakatime-cli/cmd/legacy/todaygoal"
	"github.com/wakatime/wakatime-cli/pkg/config"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Run executes legacy commands following the interface of the old python implementation of the WakaTime script.
func Run(cmd *cobra.Command, v *viper.Viper) {
	logfileParams, err := logfile.LoadParams(v)
	if err != nil {
		log.Fatalf("failed to load log params: %s", err)
	}

	logFile := os.Stdout

	if !logfileParams.ToStdout {
		log.Debugf("log to file %s", logfileParams.File)

		logFile, err = os.OpenFile(logfileParams.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("error opening log file: %s", err)
		}

		log.SetOutput(logFile)
	}

	log.SetVerbose(logfileParams.Verbose)
	log.SetJww(logfileParams.Verbose, logFile)

	if v.GetBool("useragent") {
		log.Debugln("command: useragent")

		if plugin := vipertools.GetString(v, "plugin"); plugin != "" {
			fmt.Println(heartbeat.UserAgent(plugin))

			os.Exit(exitcode.Success)
		}

		fmt.Println(heartbeat.UserAgentUnknownPlugin())

		os.Exit(exitcode.Success)
	}

	if v.GetBool("version") {
		log.Debugln("command: version")

		runVersion(v.GetBool("verbose"))

		os.Exit(exitcode.Success)
	}

	if err := config.ReadInConfig(v, config.FilePath); err != nil {
		log.Errorf("failed to load configuration file: %s", err)

		var cfperr ErrConfigFileParse
		if errors.As(err, &cfperr) {
			os.Exit(exitcode.ErrConfigFileParse)
		}

		os.Exit(exitcode.ErrDefault)
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

	if v.IsSet("entity") {
		log.Debugln("command: heartbeat")

		heartbeatcmd.Run(v)
	}

	if v.IsSet("sync-offline-activity") {
		log.Debugln("command: sync-offline-activity")

		offlinesync.Run(v)
	}

	log.Warnf("One of the following parameters has to be provided: %s", strings.Join([]string{
		"--config-read",
		"--config-write",
		"--entity",
		"--sync-offline-activity",
		"--today",
		"--today-goal",
		"--useragent",
		"--version",
	}, ", "))

	_ = cmd.Help()

	os.Exit(exitcode.ErrDefault)
}
