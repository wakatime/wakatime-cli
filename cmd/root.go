package cmd

import (
	"github.com/wakatime/wakatime-cli/cmd/legacy"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
)

const (
	// defaultConfigSection is the default section in the wakatime ini config file.
	defaultConfigSection = "settings"
	// defaultTimeoutSecs is the default timeout used for requests to the wakatime api.
	defaultTimeoutSecs = 60
	// defaultOfflineSync is the default maximum number of heartbeats from the
	// offline queue, which will be synced upon sending heartbeats to the API.
	defaultOfflineSync = "100"
)

// NewRootCMD creates a rootCmd, which represents the base command when called without any subcommands.
func NewRootCMD() *cobra.Command {
	multilineOption := viper.IniLoadOptions(ini.LoadOptions{AllowPythonMultilineValues: true})
	v := viper.NewWithOptions(multilineOption)

	cmd := &cobra.Command{
		Use:   "wakatime-cli",
		Short: "Command line interface used by all WakaTime text editor plugins.",
		Run: func(cmd *cobra.Command, args []string) {
			legacy.Run(v)
		},
	}

	setFlags(cmd, v)

	return cmd
}

func setFlags(cmd *cobra.Command, v *viper.Viper) {
	flags := cmd.Flags()
	flags.String("alternate-language", "", "Optional alternate language name. Auto-detected language takes priority.")
	flags.String("alternate-project", "", "Optional alternate project name. Auto-detected project takes priority.")
	flags.String("api-url", "", "Heartbeats api url. For debugging with a local server.")
	flags.String("apiurl", "", "(deprecated) Heartbeats api url. For debugging with a local server.")
	flags.String(
		"category",
		"",
		"Category of this heartbeat activity. Can be \"coding\","+
			" \"building\", \"indexing\", \"debugging\", \"running tests\","+
			" \"writing tests\", \"manual testing\", \"code reviewing\","+
			" \"browsing\", or \"designing\". Defaults to \"coding\".",
	)
	flags.String("config", "", "Optional config file. Defaults to '~/.wakatime.cfg'.")
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
		"Entity type for this heartbeat. Can be \"file\", \"domain\" or \"app\". Defaults to \"file\".",
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
	flags.String("hide-branch-names", "", "Obfuscate branch names. Will not send revision control branch names to api.")
	flags.String("hide-file-names", "", "Obfuscate filenames. Will not send file names to api.")
	flags.String("hide-filenames", "", "(deprecated) Obfuscate filenames. Will not send file names to api.")
	flags.String("hidefilenames", "", "(deprecated) Obfuscate filenames. Will not send file names to api.")
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
	flags.String("key", "", "Your wakatime api key; uses api_key from ~/.wakatime.cfg by default.")
	flags.String("language", "", "Optional language name. If valid, takes priority over auto-detected language.")
	flags.Int("lineno", 0, "Optional line number. This is the current line being edited.")
	flags.Int(
		"lines-in-file",
		0,
		"Optional lines in the file. Normally, this is detected automatically but"+
			" can be provided manually for performance, accuracy, or when using --local-file.")
	flags.String(
		"local-file",
		"",
		"Absolute path to local file for the heartbeat. When --entity is a"+
			" remote file, this local file will be used for stats and just"+
			" the value of --entity is sent with the heartbeat.",
	)
	flags.String("log-file", "", "Optional log file. Defaults to '~/.wakatime.log'.")
	flags.String("logfile", "", "(deprecated) Optional log file. Defaults to '~/.wakatime.log'.")
	flags.Bool("log-to-stdout", false, "If enabled, logs will go to stdout. Will overwrite logfile configs.")
	flags.Bool(
		"no-ssl-verify",
		false,
		"Disables SSL certificate verification for HTTPS requests. By default,"+
			" SSL certificates are verified.",
	)
	flags.String("plugin", "", "Optional text editor plugin name and version for User-Agent header.")
	flags.String("project", "", "Override auto-detected project."+
		" Use --alternate-project to supply a fallback project if one can't be auto-detected.")
	flags.String(
		"proxy",
		"",
		"Optional proxy configuration. Supports HTTPS SOCKS and NTLM proxies."+
			" For example: 'https://user:pass@host:port' or 'socks5://user:pass@host:port'"+
			" or 'domain\\user:pass'",
	)
	flags.String(
		"ssl-certs-file",
		"",
		"Override the bundled CA certs file. By default, uses"+
			" system ca certs.",
	)
	flags.String(
		"sync-offline-activity",
		defaultOfflineSync,
		"Amount of offline activity to sync from your local ~/.wakatime.bdb bolt"+
			" file to your WakaTime Dashboard before exiting. Can be \"none\" or"+
			" a positive integer. Defaults to 100, meaning for every heartbeat sent"+
			" while online, 100 offline heartbeats are synced. Can be used without"+
			" --entity to only sync offline activity without generating new heartbeats.",
	)
	flags.Int(
		"timeout",
		defaultTimeoutSecs,
		"Number of seconds to wait when sending heartbeats to api. Defaults to 60 seconds.",
	)
	flags.Float64("time", 0, "Optional floating-point unix epoch timestamp. Uses current time by default.")
	flags.Bool("today", false, "Prints dashboard time for Today, then exits.")
	flags.String(
		"today-goal",
		"",
		"Prints time for the given goal id Today, then exits"+
			" Visit wakatime.com/api/v1/users/current/goals to find your goal id.")
	flags.Bool("useragent", false, "Prints the wakatime-cli useragent, as it will be sent to the api, then exits.")
	flags.Bool("verbose", false, "Turns on debug messages in log file.")
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
	_ = flags.MarkHidden("useragent")

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
