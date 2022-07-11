package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	cmdapi "github.com/wakatime/wakatime-cli/cmd/api"
	"github.com/wakatime/wakatime-cli/cmd/configread"
	"github.com/wakatime/wakatime-cli/cmd/configwrite"
	cmdheartbeat "github.com/wakatime/wakatime-cli/cmd/heartbeat"
	"github.com/wakatime/wakatime-cli/cmd/logfile"
	cmdoffline "github.com/wakatime/wakatime-cli/cmd/offline"
	"github.com/wakatime/wakatime-cli/cmd/offlinecount"
	"github.com/wakatime/wakatime-cli/cmd/offlineprint"
	"github.com/wakatime/wakatime-cli/cmd/offlinesync"
	"github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/cmd/today"
	"github.com/wakatime/wakatime-cli/cmd/todaygoal"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/diagnostic"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	iniv1 "gopkg.in/ini.v1"
)

// Run executes commands parsed from a command line.
func Run(cmd *cobra.Command, v *viper.Viper) {
	// force setup logging otherwise log goes to std out
	_ = SetupLogging(v)

	err := parseConfigFiles(v)
	if err != nil {
		log.Errorf("failed to parse config files: %s", err)

		if !v.IsSet("entity") {
			os.Exit(exitcode.ErrConfigFileParse)
		}

		queueFilepath, err := offline.QueueFilepath()
		if err != nil {
			log.Warnf("failed to load offline queue filepath: %s", err)
		}

		if err := cmdoffline.SaveHeartbeats(v, nil, queueFilepath); err != nil {
			log.Errorf("failed to save heartbeats to offline queue: %s", err)
		}

		os.Exit(exitcode.ErrConfigFileParse)
	}

	// setup logging again to use config file settings
	logFileParams := SetupLogging(v)

	if v.GetBool("user-agent") {
		log.Debugln("command: user-agent")

		fmt.Println(heartbeat.UserAgent(vipertools.GetString(v, "plugin")))

		os.Exit(exitcode.Success)
	}

	if v.GetBool("version") {
		log.Debugln("command: version")

		RunCmd(v, logFileParams.Verbose, runVersion)
	}

	if v.IsSet("config-read") {
		log.Debugln("command: config-read")

		RunCmd(v, logFileParams.Verbose, configread.Run)
	}

	if v.IsSet("config-write") {
		log.Debugln("command: config-write")

		RunCmd(v, logFileParams.Verbose, configwrite.Run)
	}

	if v.GetBool("today") {
		log.Debugln("command: today")

		RunCmd(v, logFileParams.Verbose, today.Run)
	}

	if v.IsSet("today-goal") {
		log.Debugln("command: today-goal")

		RunCmd(v, logFileParams.Verbose, todaygoal.Run)
	}

	if v.IsSet("entity") {
		log.Debugln("command: heartbeat")

		RunCmdWithOfflineSync(v, logFileParams.Verbose, cmdheartbeat.Run)
	}

	if v.IsSet("sync-offline-activity") {
		log.Debugln("command: sync-offline-activity")

		RunCmd(v, logFileParams.Verbose, offlinesync.Run)
	}

	if v.GetBool("offline-count") {
		log.Debugln("command: offline-count")

		RunCmd(v, logFileParams.Verbose, offlinecount.Run)
	}

	if v.IsSet("print-offline-heartbeats") {
		log.Debugln("command: print-offline-heartbeats")

		RunCmd(v, logFileParams.Verbose, offlineprint.Run)
	}

	log.Warnf("one of the following parameters has to be provided: %s", strings.Join([]string{
		"--config-read",
		"--config-write",
		"--entity",
		"--offline-count",
		"--print-offline-heartbeats",
		"--sync-offline-activity",
		"--today",
		"--today-goal",
		"--useragent",
		"--version",
	}, ", "))

	_ = cmd.Help()

	os.Exit(exitcode.ErrGeneric)
}

func parseConfigFiles(v *viper.Viper) error {
	var configFiles = []struct {
		fn    func(v *viper.Viper) (string, error)
		vp    *viper.Viper
		merge bool
	}{
		{
			fn: ini.FilePath,
			vp: v,
		},
		{
			fn: ini.ImportFilePath,
			vp: v,
		},
		{
			fn:    ini.InternalFilePath,
			vp:    viper.NewWithOptions(viper.IniLoadOptions(iniv1.LoadOptions{SkipUnrecognizableLines: true})),
			merge: true,
		},
	}

	for _, c := range configFiles {
		configFile, err := c.fn(v)
		if err != nil {
			return fmt.Errorf("error getting config file path: %s", err)
		}

		if configFile == "" {
			continue
		}

		// check if file exists
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			log.Debugf("config file %q not present or not accessible", configFile)
			continue
		}

		if err := ini.ReadInConfig(c.vp, configFile); err != nil {
			return fmt.Errorf("failed to load configuration file: %s", err)
		}

		if c.merge {
			err = v.MergeConfigMap(c.vp.AllSettings())
			if err != nil {
				log.Warnf("failed to merge configuration file: %s", err)
			}
		}
	}

	return nil
}

// SetupLogging uses the --log-file param to configure logging to file or stdout.
func SetupLogging(v *viper.Viper) *logfile.Params {
	logfileParams, err := logfile.LoadParams(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load log params: %s", err)
		log.Fatalf("failed to load log params: %s", err)
	}

	logFile := os.Stdout

	if !logfileParams.ToStdout {
		logFile, err = os.OpenFile(logfileParams.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening log file: %s", err)
			log.Fatalf("error opening log file: %s", err)
		}

		log.SetOutput(logFile)
	}

	log.SetVerbose(logfileParams.Verbose)
	log.SetJww(logfileParams.Verbose, logFile)

	return &logfileParams
}

// cmdFn represents a command function.
type cmdFn func(v *viper.Viper) (int, error)

// RunCmd runs a command function and exits with the exit code returned by
// the command function. Will send diagnostic on any errors or panics.
func RunCmd(v *viper.Viper, verbose bool, cmd cmdFn) {
	exitCode := runCmd(v, verbose, cmd)

	os.Exit(exitCode)
}

// RunCmdWithOfflineSync runs a command function and exits with the exit code
// returned by the command function. If command run was successful, it will execute
// offline sync command afterwards. Will send diagnostic on any errors or panics.
func RunCmdWithOfflineSync(v *viper.Viper, verbose bool, cmd cmdFn) {
	exitCode := runCmd(v, verbose, cmd)
	if exitCode != exitcode.Success {
		os.Exit(exitCode)
	}

	os.Exit(runCmd(v, verbose, offlinesync.Run))
}

// runCmd contains the main logic of RunCmd.
func runCmd(v *viper.Viper, verbose bool, cmd cmdFn) int {
	logs := bytes.NewBuffer(nil)
	resetLogs := captureLogs(logs)

	// catch panics
	defer func() {
		if err := recover(); err != nil {
			resetLogs()

			if verbose {
				if err := sendDiagnostics(v, logs.String(), string(debug.Stack())); err != nil {
					log.Warnf("failed to send diagnostics: %s", err)
				}
			}

			os.Exit(exitcode.ErrGeneric)
		}
	}()

	// run command
	exitCode, err := cmd(v)
	if err != nil {
		log.Errorf("failed to run command: %s", err)

		resetLogs()

		if exitCode != exitcode.ErrAuth && exitCode != exitcode.ErrBackoff && verbose {
			if err := sendDiagnostics(v, logs.String(), string(debug.Stack())); err != nil {
				log.Warnf("failed to send diagnostics: %s", err)
			}
		}
	}

	return exitCode
}

func sendDiagnostics(v *viper.Viper, logs, stack string) error {
	paramAPI, err := params.LoadAPIParams(v)
	if err != nil {
		var errauth api.ErrAuth

		// api.ErrAuth represents an error when parsing api key.
		// In this context api key is not required to send diagnostics.
		if !errors.As(err, &errauth) {
			return fmt.Errorf("failed to load API parameters: %s", err)
		}
	}

	c, err := cmdapi.NewClientWithoutAuth(paramAPI)
	if err != nil {
		return fmt.Errorf("failed to initialize api client: %s", err)
	}

	diagnostics := []diagnostic.Diagnostic{
		diagnostic.Logs(logs),
		diagnostic.Stack(stack),
	}

	api.WithDisableSSLVerify()(c)

	err = c.SendDiagnostics(paramAPI.Plugin, diagnostics...)
	if err != nil {
		return fmt.Errorf("failed to send diagnostics to the API: %s", err)
	}

	log.Debugln("successfully sent diagnostics")

	return nil
}

func captureLogs(dest io.Writer) func() {
	logOutput := log.Output()

	// will write to log output and dest
	mw := io.MultiWriter(logOutput, dest)

	log.SetOutput(mw)

	return func() {
		log.SetOutput(logOutput)
	}
}
