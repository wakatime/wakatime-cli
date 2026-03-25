package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"runtime/debug"
	"strings"

	cmdapi "github.com/wakatime/wakatime-cli/cmd/api"
	"github.com/wakatime/wakatime-cli/cmd/configread"
	"github.com/wakatime/wakatime-cli/cmd/configwrite"
	"github.com/wakatime/wakatime-cli/cmd/fileexperts"
	cmdheartbeat "github.com/wakatime/wakatime-cli/cmd/heartbeat"
	cmdoffline "github.com/wakatime/wakatime-cli/cmd/offline"
	"github.com/wakatime/wakatime-cli/cmd/offlinecount"
	"github.com/wakatime/wakatime-cli/cmd/offlineprint"
	"github.com/wakatime/wakatime-cli/cmd/offlinesync"
	"github.com/wakatime/wakatime-cli/cmd/today"
	"github.com/wakatime/wakatime-cli/cmd/todaygoal"
	"github.com/wakatime/wakatime-cli/pkg/diagnostic"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/lexer"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/log/setup"
	"github.com/wakatime/wakatime-cli/pkg/metrics"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	pkgparams "github.com/wakatime/wakatime-cli/pkg/params"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"
	"github.com/wakatime/wakatime-cli/pkg/wakaerror"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

type diagnostics struct {
	Logs          string
	OriginalError any
	Panicked      bool
	Stack         string
}

// RunE executes commands parsed from a command line.
func RunE(cmd *cobra.Command, v *viper.Viper) error {
	ctx := context.Background()

	// extract logger from context despite it's not fully initialized yet
	logger := log.Extract(ctx)

	err := parseConfigFiles(ctx, v)
	if err != nil {
		logger.Errorf("failed to parse config files: %s", err)

		if v.IsSet("entity") {
			_ = fallbackSaveHeartbeats(ctx, v)

			return exitcode.Err{Code: exitcode.ErrConfigFileParse}
		}
	}

	logger, err = setup.Logging(ctx, v)
	if err != nil {
		// log to std out and exit, as logger instance failed to setup
		stdlog.Fatalf("failed to setup logging: %s", err)
	}

	// save logger to context
	ctx = log.ToContext(ctx, logger)

	// register all custom lexers
	if err := lexer.RegisterAll(); err != nil {
		logger.Fatalf("failed to register custom lexers: %s", err)
	}

	// start profiling if enabled
	if logger.IsMetricsEnabled() {
		stopProfiling, err := metrics.StartProfiling(ctx)
		if err != nil {
			logger.Errorf("failed to start profiling: %s", err)
		} else {
			defer stopProfiling()
		}
	}

	if v.GetBool("user-agent") {
		logger.Debugln("command: user-agent")

		fmt.Println(heartbeat.UserAgent(ctx, vipertools.GetString(v, "plugin")))

		return nil
	}

	if v.GetBool("version") {
		logger.Debugln("command: version")

		return RunCmd(ctx, v, logger.IsVerboseEnabled(), logger.SendDiagsOnErrors(), runVersion)
	}

	if v.IsSet("config-read") {
		logger.Debugln("command: config-read")

		return RunCmd(ctx, v, logger.IsVerboseEnabled(), logger.SendDiagsOnErrors(), configread.Run)
	}

	if v.IsSet("config-write") {
		logger.Debugln("command: config-write")

		return RunCmd(ctx, v, logger.IsVerboseEnabled(), logger.SendDiagsOnErrors(), configwrite.Run)
	}

	if v.GetBool("today") {
		logger.Debugln("command: today")

		return RunCmd(ctx, v, logger.IsVerboseEnabled(), logger.SendDiagsOnErrors(), today.Run)
	}

	if v.IsSet("today-goal") {
		logger.Debugln("command: today-goal")

		return RunCmd(ctx, v, logger.IsVerboseEnabled(), logger.SendDiagsOnErrors(), todaygoal.Run)
	}

	if v.GetBool("file-experts") {
		logger.Debugln("command: file-experts")

		return RunCmd(ctx, v, logger.IsVerboseEnabled(), logger.SendDiagsOnErrors(), fileexperts.Run)
	}

	if v.IsSet("entity") {
		logger.Debugln("command: heartbeat")

		return RunCmdWithOfflineSync(ctx, v, logger.IsVerboseEnabled(), logger.SendDiagsOnErrors(), cmdheartbeat.Run)
	}

	if v.IsSet("sync-offline-activity") {
		logger.Debugln("command: sync-offline-activity")

		return RunCmd(ctx, v, logger.IsVerboseEnabled(), logger.SendDiagsOnErrors(), offlinesync.RunWithoutRateLimiting)
	}

	if v.GetBool("offline-count") {
		logger.Debugln("command: offline-count")

		return RunCmd(ctx, v, logger.IsVerboseEnabled(), logger.SendDiagsOnErrors(), offlinecount.Run)
	}

	if v.IsSet("print-offline-heartbeats") {
		logger.Debugln("command: print-offline-heartbeats")

		return RunCmd(ctx, v, logger.IsVerboseEnabled(), logger.SendDiagsOnErrors(), offlineprint.Run)
	}

	logger.Warnf("one of the following parameters has to be provided: %s", strings.Join([]string{
		"--config-read",
		"--config-write",
		"--entity",
		"--file-experts",
		"--offline-count",
		"--print-offline-heartbeats",
		"--sync-offline-activity",
		"--today",
		"--today-goal",
		"--user-agent",
		"--version",
	}, ", "))

	_ = cmd.Help()

	return exitcode.Err{Code: exitcode.ErrGeneric}
}

func parseConfigFiles(ctx context.Context, v *viper.Viper) error {
	logger := log.Extract(ctx)

	var configFiles = []struct {
		filePathFn func(context.Context, *viper.Viper) (string, error)
	}{
		{
			filePathFn: ini.FilePath,
		},
		{
			filePathFn: ini.ImportFilePath,
		},
		{
			filePathFn: ini.InternalFilePath,
		},
	}

	for _, c := range configFiles {
		configFile, err := c.filePathFn(ctx, v)
		if err != nil {
			return fmt.Errorf("error getting config file path: %s", err)
		}

		if configFile == "" {
			continue
		}

		// check if file exists
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			logger.Debugf("config file %q not present or not accessible", configFile)
			continue
		}

		if err := ini.ReadInConfig(v, configFile); err != nil {
			return fmt.Errorf("failed to load configuration file: %s", err)
		}
	}

	return nil
}

// cmdFn represents a command function.
type cmdFn func(ctx context.Context, v *viper.Viper) (int, error)

// RunCmd runs a command function and exits with the exit code returned by
// the command function. Will send diagnostic on any errors or panics.
func RunCmd(ctx context.Context, v *viper.Viper, verbose bool, sendDiagsOnErrors bool, cmd cmdFn) error {
	return runCmd(ctx, v, verbose, sendDiagsOnErrors, cmd)
}

// RunCmdWithOfflineSync runs a command function and exits with the exit code
// returned by the command function. If command run was successful, it will execute
// offline sync command afterwards. Will send diagnostic on any errors or panics.
func RunCmdWithOfflineSync(ctx context.Context, v *viper.Viper, verbose bool, sendDiagsOnErrors bool, cmd cmdFn) error {
	if err := runCmd(ctx, v, verbose, sendDiagsOnErrors, cmd); err != nil {
		return err
	}

	return runCmd(ctx, v, verbose, sendDiagsOnErrors, offlinesync.RunWithRateLimiting)
}

// runCmd contains the main logic of RunCmd.
// It will send diagnostic on any errors or panics.
// On panic, it will send diagnostic and exit with ErrGeneric exit code.
// On error, it will only send diagnostic if sendDiagsOnErrors and verbose is true.
func runCmd(ctx context.Context, v *viper.Viper, verbose bool, sendDiagsOnErrors bool, cmd cmdFn) (errresponse error) {
	logs := bytes.NewBuffer(nil)
	resetLogs := captureLogs(ctx, logs)

	logger := log.Extract(ctx)

	// catch panics
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("panicked: %v. Stack: %s", r, string(debug.Stack()))

			resetLogs()

			diags := diagnostics{
				OriginalError: r,
				Panicked:      true,
				Stack:         string(debug.Stack()),
			}

			if verbose {
				diags.Logs = logs.String()
			}

			if err := sendDiagnostics(ctx, v, diags); err != nil {
				logger.Warnf("failed to send diagnostics: %s", err)
			}

			errresponse = exitcode.Err{Code: exitcode.ErrGeneric}
		}
	}()

	var err error

	// run command
	exitCode, err := cmd(ctx, v)
	// nolint:nestif
	if err != nil {
		if errwaka, ok := err.(wakaerror.Error); ok {
			sendDiagsOnErrors = sendDiagsOnErrors || errwaka.SendDiagsOnErrors()
			// if verbose is not set, use the value from the error
			verbose = verbose || errwaka.ShouldLogError()
		}

		if errloglevel, ok := err.(wakaerror.LogLevel); ok {
			logger.Logf(zapcore.Level(errloglevel.LogLevel()), "failed to run command: %s", err)
		} else if verbose {
			logger.Errorf("failed to run command: %s", err)
		}

		resetLogs()

		if verbose && sendDiagsOnErrors {
			if err := sendDiagnostics(ctx, v,
				diagnostics{
					Logs:          logs.String(),
					OriginalError: err.Error(),
					Stack:         string(debug.Stack()),
				}); err != nil {
				logger.Warnf("failed to send diagnostics: %s", err)
			}
		}
	}

	if exitCode != exitcode.Success {
		logger.Debugf("command failed with exit code %d", exitCode)

		errresponse = exitcode.Err{Code: exitCode}
	}

	return errresponse
}

func fallbackSaveHeartbeats(ctx context.Context, v *viper.Viper) int {
	logger := log.Extract(ctx)

	queueFilepath, err := offline.QueueFilepath(ctx, v)
	if err != nil {
		logger.Warnf("failed to load offline queue filepath: %s", err)
	}

	apiParams, _ := pkgparams.LoadAPIParams(ctx, v, pkgparams.FlagReadOrderFlagPrecedence)

	heartbeatParams, err := pkgparams.LoadHeartbeatParams(ctx, v, pkgparams.FlagReadOrderFlagPrecedence)
	if err != nil {
		logger.Errorf("failed to load command parameters: %s", err)
	}

	heartbeats := cmdheartbeat.BuildHeartbeats(ctx, apiParams.Plugin, heartbeatParams)

	if err := cmdoffline.SaveHeartbeats(ctx, v, queueFilepath, heartbeats); err != nil {
		logger.Errorf("failed to save heartbeats to offline queue: %s", err)

		return exitcode.ErrGeneric
	}

	return exitcode.Success
}

func sendDiagnostics(ctx context.Context, v *viper.Viper, d diagnostics) error {
	paramAPI, err := pkgparams.LoadAPIParams(ctx, v, pkgparams.FlagReadOrderFlagPrecedence)
	if err != nil {
		return fmt.Errorf("failed to load API parameters: %s", err)
	}

	c, err := cmdapi.NewClient(ctx, paramAPI)
	if err != nil {
		return fmt.Errorf("failed to initialize api client: %s", err)
	}

	diagnostics := []diagnostic.Diagnostic{
		diagnostic.Error(d.OriginalError),
		diagnostic.Logs(d.Logs),
		diagnostic.Stack(d.Stack),
	}

	err = c.SendDiagnostics(ctx, paramAPI.Plugin, d.Panicked, diagnostics...)
	if err != nil {
		return fmt.Errorf("failed to send diagnostics to the API: %s", err)
	}

	logger := log.Extract(ctx)
	logger.Debugln("successfully sent diagnostics")

	return nil
}

func captureLogs(ctx context.Context, dest io.Writer) func() {
	logger := log.Extract(ctx)

	logOutput := logger.Output()

	// will write to log output and dest
	mw := io.MultiWriter(logOutput, dest)

	logger.SetOutput(mw)

	return func() {
		logger.SetOutput(logOutput)
	}
}
