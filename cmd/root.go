package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/offline"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
)

// defaultConfigSection is the default section in the wakatime ini config file.
const defaultConfigSection = "settings"

// NewRootCMD creates a rootCmd, which represents the base command when called without any subcommands.
func NewRootCMD() *cobra.Command {
	iniOption := viper.IniLoadOptions(ini.LoadOptions{
		AllowPythonMultilineValues: true,
	})
	v := viper.NewWithOptions(iniOption)

	cmd := &cobra.Command{
		Use:   "wakatime-cli",
		Short: "Command line interface used by all WakaTime text editor plugins.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := RunE(cmd, v); err != nil {
				var errexitcode exitcode.Err

				if errors.As(err, &errexitcode) {
					os.Exit(errexitcode.Code)
				}

				os.Exit(exitcode.ErrGeneric)
			}

			os.Exit(exitcode.Success)

			return nil
		},
	}

	setFlags(cmd, v)

	return cmd
}

func setFlags(cmd *cobra.Command, v *viper.Viper) {
	flags := cmd.Flags()
	flags.String("alternate-branch", "", "Optional alternate branch name. Auto-detected branch takes priority.")
	flags.String("alternate-language", "", "Optional alternate language name. Auto-detected language takes priority.")
	flags.String("alternate-project", "", "Optional alternate project name. Auto-detected project takes priority.")
	flags.String(
		"api-url",
		"",
		"API base url used when sending heartbeats and fetching code stats. Defaults to https://api.wakatime.com/api/v1/.",
	)
	flags.String(
		"apiurl",
		"",
		"(deprecated) API base url used when sending heartbeats and fetching code stats. Defaults to"+
			" https://api.wakatime.com/api/v1/.",
	)
	flags.String(
		"category",
		"",
		"Category of this heartbeat activity. Can be \"coding\","+
			" \"building\", \"indexing\", \"debugging\", \"learning\","+
			" \"meeting\", \"planning\", \"researching\", \"communicating\", \"supporting\" "+
			" \"advising\", \"running tests\", \"writing tests\", \"manual testing\","+
			" \"writing docs\", \"code reviewing\", \"browsing\","+
			" \"translating\", or \"designing\". Defaults to \"coding\".",
	)
	flags.String("config", "", "Optional config file. Defaults to '~/.wakatime.cfg'.")
	flags.String("internal-config", "", "Optional internal config file. Defaults to '~/.wakatime/wakatime-internal.cfg'.")
	flags.String("config-read", "", "Prints value for the given config key, then exits.")
	flags.String(
		"config-section",
		defaultConfigSection,
		"Optional config section when reading or writing a config key. Defaults to [settings].",
	)
	flags.StringToString(
		"config-write",
		nil,
		"Writes value to a config key, then exits. Expects two arguments, key and value.",
	)
	flags.Int("cursorpos", 0, "Optional cursor position in the current file.")
	flags.Bool("disable-offline", false, "Disables offline time logging instead of queuing logged time.")
	flags.Bool("disableoffline", false, "(deprecated) Disables offline time logging instead of queuing logged time.")
	flags.String(
		"entity",
		"",
		"Absolute path to file for the heartbeat. Can also be a url, domain or app when --entity-type is not file.",
	)
	flags.String(
		"entity-type",
		"",
		"Entity type for this heartbeat. Can be \"file\", \"domain\", \"url\", or \"app\". Defaults to \"file\".",
	)
	flags.StringSlice(
		"exclude",
		nil,
		"Filename patterns to exclude from logging. POSIX regex syntax."+
			" Can be used more than once.",
	)
	flags.Bool(
		"exclude-unknown-project",
		false,
		"When set, any activity where the project cannot be detected will be ignored.",
	)
	flags.Bool("extra-heartbeats", false, "Reads extra heartbeats from STDIN as a JSON array until EOF.")
	flags.String(
		"file",
		"",
		"(deprecated) Absolute path to file for the heartbeat."+
			" Can also be a url, domain or app when --entity-type is not file.")
	flags.Bool("file-experts", false, "Prints the top developer within a team for the given entity, then exits.")
	flags.Bool(
		"guess-language",
		false,
		"Enable detecting language from file contents.")
	flags.Int(
		"heartbeat-rate-limit-seconds",
		offline.RateLimitDefaultSeconds,
		fmt.Sprintf("Only sync heartbeats to the API once per these seconds, instead"+
			" saving to the offline db. Defaults to %d. Use zero to disable.",
			offline.RateLimitDefaultSeconds),
	)
	flags.String("hide-branch-names", "", "Obfuscate branch names. Will not send revision control branch names to api.")
	flags.String("hide-file-names", "", "Obfuscate filenames. Will not send file names to api.")
	flags.String("hide-filenames", "", "(deprecated) Obfuscate filenames. Will not send file names to api.")
	flags.String("hidefilenames", "", "(deprecated) Obfuscate filenames. Will not send file names to api.")
	flags.Bool(
		"hide-project-folder",
		false,
		"When set, send the file's path relative to the project folder."+
			" For ex: /User/me/projects/bar/src/file.ts is sent as src/file.ts so the server never sees the full path."+
			" When the project folder cannot be detected, only the file name is sent. For ex: file.ts.")
	flags.String(
		"hide-project-names",
		"",
		"Obfuscate project names. When a project folder is detected instead of"+
			" using the folder name as the project, a .wakatime-project file is"+
			" created with a random project name.",
	)
	flags.String("hostname", "", "Optional name of local machine. Defaults to local machine name read from system.")
	flags.StringSlice(
		"include",
		nil,
		"Filename patterns to log. When used in combination with"+
			" --exclude, files matching include will still be logged."+
			" POSIX regex syntax. Can be used more than once.",
	)
	flags.Bool(
		"include-only-with-project-file",
		false,
		"Disables tracking folders unless they contain a .wakatime-project file. Defaults to false.",
	)
	flags.Bool(
		"is-unsaved-entity",
		false,
		"Normally files that don't exist on disk are skipped and not tracked. When this option is present,"+
			" the main heartbeat file will be tracked even if it doesn't exist. To set this flag on"+
			" extra heartbeats, use the 'is_unsaved_entity' json key.")
	flags.String("key", "", "Your wakatime api key; uses api_key from ~/.wakatime.cfg by default.")
	flags.String("language", "", "Optional language name. If valid, takes priority over auto-detected language.")
	flags.Int("lineno", 0, "Optional line number. This is the current line being edited.")
	flags.Int(
		"lines-in-file",
		0,
		"Optional lines in the file. Normally, this is detected automatically but"+
			" can be provided manually for performance, accuracy, or when using --local-file.")
	flags.Int("line-additions", 0, "Optional number of lines added since last heartbeat in the current file.")
	flags.Int("line-deletions", 0, "Optional number of lines deleted since last heartbeat in the current file.")
	flags.String(
		"local-file",
		"",
		"Absolute path to local file for the heartbeat. When --entity is a"+
			" remote file, this local file will be used for stats and just"+
			" the value of --entity is sent with the heartbeat.",
	)
	flags.String("log-file", "", "Optional log file. Defaults to '~/.wakatime/wakatime.log'.")
	flags.String("logfile", "", "(deprecated) Optional log file. Defaults to '~/.wakatime/wakatime.log'.")
	flags.Bool("log-to-stdout", false, "If enabled, logs will go to stdout. Will overwrite logfile configs.")
	flags.Bool(
		"metrics",
		false,
		"When set, collects metrics usage in '~/.wakatime/metrics' folder. Defaults to false.",
	)
	flags.Bool(
		"no-ssl-verify",
		false,
		"Disables SSL certificate verification for HTTPS requests. By default,"+
			" SSL certificates are verified.",
	)
	flags.String(
		"offline-queue-file",
		"",
		"(internal) Specify an offline queue file, which will be used instead of the default one.",
	)
	flags.String(
		"offline-queue-file-legacy",
		"",
		"(internal) Specify the legacy offline queue file, which will be used instead of the default one.",
	)
	flags.String(
		"output",
		"",
		"Format output. Can be \"text\", \"json\" or \"raw-json\". Defaults to \"text\".",
	)
	flags.String("plugin", "", "Optional text editor plugin name and version for User-Agent header.")
	flags.Int("print-offline-heartbeats", offline.PrintMaxDefault, "Prints offline heartbeats to stdout.")
	flags.String("project", "", "Override auto-detected project."+
		" Use --alternate-project to supply a fallback project if one can't be auto-detected.")
	flags.String(
		"project-folder",
		"",
		"Optional workspace path. Usually used when hiding the project folder, or when a project"+
			" root folder can't be auto detected.")
	flags.String(
		"proxy",
		"",
		"Optional proxy configuration. Supports HTTPS SOCKS and NTLM proxies."+
			" For example: 'https://user:pass@host:port' or 'socks5://user:pass@host:port'"+
			" or 'domain\\user:pass'",
	)
	flags.Bool(
		"send-diagnostics-on-errors",
		false,
		"When --verbose or debug enabled, also sends diagnostics on any error not just crashes.",
	)
	flags.String(
		"ssl-certs-file",
		"",
		"Override the bundled CA certs file. By default, uses"+
			" system ca certs.",
	)
	flags.Int(
		"sync-offline-activity",
		offline.SyncMaxDefault,
		fmt.Sprintf("Amount of offline activity to sync from your local ~/.wakatime/offline_heartbeats.bdb bolt"+
			" file to your WakaTime Dashboard before exiting. Can be zero or"+
			" a positive integer. Defaults to %d, meaning after sending a heartbeat"+
			" while online, all queued offline heartbeats are sent to WakaTime API, up"+
			" to a limit of 1000. Zero syncs all offline heartbeats. Can be used"+
			" without --entity to only sync offline activity without generating"+
			" new heartbeats.", offline.SyncMaxDefault),
	)
	flags.Bool("offline-count", false, "Prints the number of heartbeats in the offline db, then exits.")
	flags.Int(
		"timeout",
		api.DefaultTimeoutSecs,
		fmt.Sprintf(
			"Number of seconds to wait when sending heartbeats to api. Defaults to %d seconds.", api.DefaultTimeoutSecs),
	)
	flags.Float64("time", 0, "Optional floating-point unix epoch timestamp. Uses current time by default.")
	flags.Bool("today", false, "Prints dashboard time for today, then exits.")
	flags.String("today-hide-categories", "", "When optionally included with --today, causes output to"+
		" show total code time today without categories. Defaults to false.")
	flags.String(
		"today-goal",
		"",
		"Prints time for the given goal id today, then exits"+
			" Visit wakatime.com/api/v1/users/current/goals to find your goal id.")
	flags.Bool(
		"user-agent",
		false,
		"(internal) Prints the wakatime-cli useragent, as it will be sent to the api, then exits.",
	)
	flags.Bool("verbose", false, "Turns on debug messages in log file, and sends diagnostics if a crash occurs.")
	flags.Bool("version", false, "Prints the wakatime-cli version number, then exits.")
	flags.Bool("write", false, "When set, tells api this heartbeat was triggered from writing to a file.")

	// hide deprecated flags
	_ = flags.MarkHidden("apiurl")
	_ = flags.MarkHidden("disableoffline")
	_ = flags.MarkHidden("file")
	_ = flags.MarkHidden("hide-filenames")
	_ = flags.MarkHidden("hidefilenames")
	_ = flags.MarkHidden("logfile")

	// hide internal flags
	_ = flags.MarkHidden("offline-queue-file")
	_ = flags.MarkHidden("offline-queue-file-legacy")
	_ = flags.MarkHidden("user-agent")

	err := v.BindPFlags(flags)
	if err != nil {
		log.Fatalf("failed to bind cobra flags to viper: %s", err)
	}
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := NewRootCMD().Execute(); err != nil {
		log.Fatalf("failed to run wakatime-cli: %s", err)
	}
}
