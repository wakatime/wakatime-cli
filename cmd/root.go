package cmd

import (
	"fmt"
	"math"
	"os"

	"github.com/spf13/cobra"
	"github.com/wakatime/wakatime-cli/cmd/legacy"
	"github.com/wakatime/wakatime-cli/constants"
	"github.com/wakatime/wakatime-cli/lib/arguments"
	"github.com/wakatime/wakatime-cli/lib/configs"
)

// NewRootCmd New root cobra command
func NewRootCmd(args *arguments.Arguments, cfg configs.WakaTimeConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wakatime-cli",
		Short: "Command line interface used by all WakaTime text editor plugins.",
		Run: func(cmd *cobra.Command, args2 []string) {
			legacy.Run(args, cfg, cmd)
		},
	}

	flags := cmd.Flags()

	flags.StringVar(&args.Entity.Entity, "entity", "", "Absolute path to file for the heartbeat. Can also be a url, domain or app when --entity-type is not file.")
	flags.StringVar(&args.ObsoleteArgs.File, "file", "", "") //help missing
	flags.StringVar(&args.Key, "key", "", "Your wakatime api key; uses api_key from ~/.wakatime.cfg by default.")
	flags.BoolVar(&args.Entity.IsWrite, "write", false, "When set, tells api this heartbeat was triggered from writing to a file.")
	flags.StringVar(&args.Editor.Plugin, "plugin", "", "Optional text editor plugin name and version for User-Agent header.")
	flags.Int64Var(&args.Time, "time", math.MinInt64, "Optional floating-point unix epoch timestamp. Uses current time by default.")
	flags.Int32Var(&args.Editor.LineNo, "lineno", math.MinInt32, "Optional line number. This is the current line being edited.")
	flags.Int32Var(&args.Editor.CursorPos, "cursorpos", math.MinInt32, "Optional cursor position in the current file.")
	flags.StringVar(&args.Entity.Type, "entity-type", "file", "Entity type for this heartbeat. Can be 'file', 'domain' or 'app'.")
	flags.StringVar(&args.Category, "category", "coding", "Category of this heartbeat activity. Can be 'coding', 'building', 'indexing', 'debugging', 'running tests', 'writing tests', 'manual testing', 'code reviewing', 'browsing', or 'designing'.")
	flags.StringVar(&args.Proxy.Address, "proxy", "", "Optional proxy configuration. Supports HTTPS and SOCKS proxies. For example: https://user:pass@host:port or socks5://user:pass@host:port or domain\\user:pass")
	flags.BoolVar(&args.Proxy.NoSslVerify, "no-ssl-verify", false, "Disables SSL certificate verification for HTTPS requests. By default, SSL certificates are verified.")
	flags.StringVar(&args.Proxy.SslCertsFile, "ssl-certs-file", "", "Override the bundled Python Requests CA certs file. By default, uses certifi for ca certs.") //TODO: Rename python in the string as soon as it's implemented
	flags.StringVar(&args.Project.Name, "project", "", "Optional project name.")
	flags.StringVar(&args.Project.Branch, "branch", "", "Optional branch name.")
	flags.StringVar(&args.Project.AlternateProject, "alternate-project", "", "Optional alternate project name. Auto-discovered project takes priority.")
	flags.StringVar(&args.AlternateLanguage, "alternate-language", "", "") //help missing
	flags.StringVar(&args.Language, "language", "", "Optional language name. If valid, takes priority over auto-detected language.")
	flags.StringVar(&args.Entity.LocalFile, "local-file", "", "Absolute path to local file for the heartbeat. When --entity is a remote file, this local file will be used for stats and just the value of --entity sent with heartbeat.")
	flags.StringVar(&args.Hostname, "hostname", "", "Hostname of current machine.")
	flags.BoolVar(&args.DisableOffline, "disable-offline", true, "Disables offline time logging instead of queuing logged time.")
	flags.BoolVar(&args.Obfuscate.HideFileNames, "hide-file-names", false, "Obfuscate filenames. Will not send file names to api.")
	flags.BoolVar(&args.ObsoleteArgs.HideFilenames1, "hide-filenames", false, "") //help missing
	flags.BoolVar(&args.ObsoleteArgs.HideFilenames2, "hidefilenames", false, "")  //help missing
	flags.BoolVar(&args.Obfuscate.HideProjectNames, "hide-project-names", false, "Obfuscate project names. When a project folder is detected instead of using the folder name as the project, a .wakatime-project file is created with a random project name.")
	flags.BoolVar(&args.Obfuscate.HideBranchNames, "hide-branch-names", false, "Obfuscate branch names. Will not send revision control branch names to api.")
	flags.StringSliceVar(&args.Exclude.Exclude, "exclude", []string{}, "Filename patterns to exclude from logging. POSIX regex syntax. Can be used more than once.")
	flags.BoolVar(&args.Exclude.ExcludeUnknownProject, "exclude-unknown-project", false, "When set, any activity where the project cannot be detected will be ignored.")
	flags.StringSliceVar(&args.Include.Include, "include", []string{}, "Filename patterns to log. When used in combination with --exclude, files matching include will still be logged. POSIX regex syntax. Can be used more than once.")
	flags.BoolVar(&args.Include.IncludeOnlyWithProjectFile, "include-only-with-project-file", false, "Disables tracking folders unless they contain a .wakatime-project file.")
	flags.StringSliceVar(&args.Exclude.Ignore, "ignore", []string{}, "") //help missing
	flags.BoolVar(&args.ExtraHeartbeats, "extra-heartbeats", false, "Reads extra heartbeats from STDIN as a JSON array until EOF.")
	flags.StringVar(&args.LogFile, "log-file", "~/.wakatime.log", "") //help missing
	flags.StringVar(&args.ObsoleteArgs.LogFile, "logfile", "", "")    //help missing
	flags.StringVar(&args.APIURL, "api-url", "", "Heartbeats api url. For debugging with a local server.")
	flags.StringVar(&args.ObsoleteArgs.APIURL, "apiurl", "", "") //help missing
	flags.IntVar(&args.Timeout, "timeout", constants.DefaultTimeout, "Number of seconds to wait when sending heartbeats to api.")
	flags.IntVar(&args.SyncOfflineActivity, "sync-offline-activity", constants.DefaultSyncOfflineActivity, "Amount of offline activity to sync from your local ~/.wakatime.db sqlite3 file to your WakaTime Dashboard before exiting. Can be a positive integer number. It means for every heartbeat sent while online, 100 offline heartbeats are synced. Can be used without --entity to only sync offline activity without generating new heartbeats.") //should be revisited when the 'provider' for local database has already been decided
	flags.BoolVar(&args.Today, "today", false, "Prints dashboard time for Today, then exits.")
	flags.StringVar(&args.Config.Path, "config", "", "Optional config file. Defaults to '~/.wakatime.cfg'.")
	flags.StringVar(&args.Config.Section, "config-section", "settings", "Optional config section when reading or writing a config key. Defaults to [settings].")
	flags.StringVar(&args.Config.Read, "config-read", "", "Prints value for the given config key, then exits.")
	flags.StringToStringVar(&args.Config.Write, "config-write", map[string]string{}, "Writes value to a config key, then exits. Expects two arguments, key and value.")
	flags.BoolVar(&args.Verbose, "verbose", false, "Turns on debug messages in log file")
	flags.BoolVar(&args.Version, "version", false, "") //help missing

	return cmd
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	args := arguments.NewArguments()

	if err := NewRootCmd(args, nil).Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
