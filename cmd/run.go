package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	cmdapi "github.com/wakatime/wakatime-cli/cmd/api"
	"github.com/wakatime/wakatime-cli/cmd/configread"
	"github.com/wakatime/wakatime-cli/cmd/configwrite"
	"github.com/wakatime/wakatime-cli/cmd/fileexperts"
	cmdheartbeat "github.com/wakatime/wakatime-cli/cmd/heartbeat"
	"github.com/wakatime/wakatime-cli/cmd/logfile"
	cmdoffline "github.com/wakatime/wakatime-cli/cmd/offline"
	"github.com/wakatime/wakatime-cli/cmd/offlinecount"
	"github.com/wakatime/wakatime-cli/cmd/offlineprint"
	"github.com/wakatime/wakatime-cli/cmd/offlinesync"
	"github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/cmd/today"
	"github.com/wakatime/wakatime-cli/cmd/todaygoal"
	"github.com/wakatime/wakatime-cli/pkg/diagnostic"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/lexer"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"
	"github.com/wakatime/wakatime-cli/pkg/wakaerror"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	iniv1 "gopkg.in/ini.v1"
)

type diagnostics struct {
	Logs          string
	OriginalError any
	Panicked      bool
	Stack         string
}

// Run executes commands parsed from a command line.
func Run(cmd *cobra.Command, v *viper.Viper) {
	// force setup logging otherwise log goes to std out
	_, err := SetupLogging(v)
	if err != nil {
		log.Fatalf("failed to setup logging: %s", err)
	}

	err = parseConfigFiles(v)
	if err != nil {
		log.Errorf("failed to parse config files: %s", err)

		if v.IsSet("entity") {
			saveHeartbeats(v)

			os.Exit(exitcode.ErrConfigFileParse)
		}
	}

	// setup logging again to use config file settings
	logFileParams, err := SetupLogging(v)
	if err != nil {
		log.Fatalf("failed to setup logging: %s", err)
	}

	// register all custom lexers
	if err := lexer.RegisterAll(); err != nil {
		log.Fatalf("failed to register custom lexers: %s", err)
	}

	if v.GetBool("user-agent") {
		log.Debugln("command: user-agent")

		fmt.Println(heartbeat.UserAgent(vipertools.GetString(v, "plugin")))

		os.Exit(exitcode.Success)
	}

	if v.GetBool("version") {
		log.Debugln("command: version")

		RunCmd(v, logFileParams.Verbose, logFileParams.SendDiagsOnErrors, runVersion)
	}

	if v.IsSet("config-read") {
		log.Debugln("command: config-read")

		RunCmd(v, logFileParams.Verbose, logFileParams.SendDiagsOnErrors, configread.Run)
	}

	if v.IsSet("config-write") {
		log.Debugln("command: config-write")

		RunCmd(v, logFileParams.Verbose, logFileParams.SendDiagsOnErrors, configwrite.Run)
	}

	if v.GetBool("today") {
		log.Debugln("command: today")

		RunCmd(v, logFileParams.Verbose, logFileParams.SendDiagsOnErrors, today.Run)
	}

	if v.IsSet("today-goal") {
		log.Debugln("command: today-goal")

		RunCmd(v, logFileParams.Verbose, logFileParams.SendDiagsOnErrors, todaygoal.Run)
	}

	if v.GetBool("file-experts") {
		log.Debugln("command: file-experts")

		RunCmd(v, logFileParams.Verbose, logFileParams.SendDiagsOnErrors, fileexperts.Run)
	}

	if v.IsSet("entity") {
		log.Debugln("command: heartbeat")

		RunCmdWithOfflineSync(v, logFileParams.Verbose, logFileParams.SendDiagsOnErrors, cmdheartbeat.Run)
	}

	if v.IsSet("sync-offline-activity") {
		log.Debugln("command: sync-offline-activity")

		RunCmd(v, logFileParams.Verbose, logFileParams.SendDiagsOnErrors, offlinesync.Run)
	}

	if v.GetBool("offline-count") {
		log.Debugln("command: offline-count")

		RunCmd(v, logFileParams.Verbose, logFileParams.SendDiagsOnErrors, offlinecount.Run)
	}

	if v.IsSet("print-offline-heartbeats") {
		log.Debugln("command: print-offline-heartbeats")

		RunCmd(v, logFileParams.Verbose, logFileParams.SendDiagsOnErrors, offlineprint.Run)
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
func SetupLogging(v *viper.Viper) (*logfile.Params, error) {
	logfileParams, err := logfile.LoadParams(v)
	if err != nil {
		return nil, fmt.Errorf("failed to load log params: %s", err)
	}

	logFile := os.Stdout

	if !logfileParams.ToStdout {
		dir := filepath.Dir(logfileParams.File)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0750)
			if err != nil {
				return nil, fmt.Errorf("error creating log file directory: %s", err)
			}
		}

		logFile, err = os.OpenFile(logfileParams.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return nil, fmt.Errorf("error opening log file: %s", err)
		}

		log.SetOutput(logFile)
	}

	log.SetVerbose(logfileParams.Verbose)
	log.SetJww(logfileParams.Verbose, logFile)

	return &logfileParams, nil
}

// cmdFn represents a command function.
type cmdFn func(v *viper.Viper) (int, error)

// RunCmd runs a command function and exits with the exit code returned by
// the command function. Will send diagnostic on any errors or panics.
func RunCmd(v *viper.Viper, verbose bool, sendDiagsOnErrors bool, cmd cmdFn) {
	exitCode := runCmd(v, verbose, sendDiagsOnErrors, cmd)

	os.Exit(exitCode)
}

// RunCmdWithOfflineSync runs a command function and exits with the exit code
// returned by the command function. If command run was successful, it will execute
// offline sync command afterwards. Will send diagnostic on any errors or panics.
func RunCmdWithOfflineSync(v *viper.Viper, verbose bool, sendDiagsOnErrors bool, cmd cmdFn) {
	exitCode := runCmd(v, verbose, sendDiagsOnErrors, cmd)
	if exitCode != exitcode.Success {
		os.Exit(exitCode)
	}

	exitCode = runCmd(v, verbose, sendDiagsOnErrors, offlinesync.Run)

	os.Exit(exitCode)
}

// runCmd contains the main logic of RunCmd.
// It will send diagnostic on any errors or panics.
// On panic, it will send diagnostic and exit with ErrGeneric exit code.
// On error, it will only send diagnostic if sendDiagsOnErrors and verbose is true.
func runCmd(v *viper.Viper, verbose bool, sendDiagsOnErrors bool, cmd cmdFn) int {
	logs := bytes.NewBuffer(nil)
	resetLogs := captureLogs(logs)

	// catch panics
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("panicked: %v. Stack: %s", err, string(debug.Stack()))

			resetLogs()

			diags := diagnostics{
				OriginalError: err,
				Panicked:      true,
				Stack:         string(debug.Stack()),
			}

			if verbose {
				diags.Logs = logs.String()
			}

			if err := sendDiagnostics(v, diags); err != nil {
				log.Warnf("failed to send diagnostics: %s", err)
			}

			os.Exit(exitcode.ErrGeneric)
		}
	}()

	// run command
	exitCode, err := cmd(v)
	// nolint:nestif
	if err != nil {
		if errwaka, ok := err.(wakaerror.Error); ok {
			sendDiagsOnErrors = sendDiagsOnErrors || errwaka.SendDiagsOnErrors()
			// if verbose is not set, use the value from the error
			verbose = verbose || errwaka.ShouldLogError()
		}

		if verbose {
			log.Errorf("failed to run command: %s", err)
		}

		resetLogs()

		if verbose && sendDiagsOnErrors {
			if err := sendDiagnostics(v,
				diagnostics{
					Logs:          logs.String(),
					OriginalError: err.Error(),
					Stack:         string(debug.Stack()),
				}); err != nil {
				log.Warnf("failed to send diagnostics: %s", err)
			}
		}
	}

	return exitCode
}

func saveHeartbeats(v *viper.Viper) {
	queueFilepath, err := offline.QueueFilepath()
	if err != nil {
		log.Warnf("failed to load offline queue filepath: %s", err)
	}

	if err := cmdoffline.SaveHeartbeats(v, nil, queueFilepath); err != nil {
		log.Errorf("failed to save heartbeats to offline queue: %s", err)
	}
}

func sendDiagnostics(v *viper.Viper, d diagnostics) error {
	paramAPI, err := params.LoadAPIParams(v)
	if err != nil {
		return fmt.Errorf("failed to load API parameters: %s", err)
	}

	c, err := cmdapi.NewClient(paramAPI)
	if err != nil {
		return fmt.Errorf("failed to initialize api client: %s", err)
	}

	diagnostics := []diagnostic.Diagnostic{
		diagnostic.Error(d.OriginalError),
		diagnostic.Logs(d.Logs),
		diagnostic.Stack(d.Stack),
	}

	err = c.SendDiagnostics(paramAPI.Plugin, d.Panicked, diagnostics...)
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
