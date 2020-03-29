package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/guid"
	"github.com/spf13/cobra"
	"github.com/wakatime/wakatime-cli/cmd/legacy"
	"github.com/wakatime/wakatime-cli/constants"
	"github.com/wakatime/wakatime-cli/lib/arguments"
	"github.com/wakatime/wakatime-cli/lib/configs"
)

var args = legacy.Arguments{}
var obsoleteArgs = legacy.ObsoleteArguments{}

var rootCmd = &cobra.Command{
	Use:   "wakatime-cli",
	Short: "Command line interface used by all WakaTime text editor plugins.",
	Run: func(cmd *cobra.Command, args2 []string) {
		runLegacy(args, obsoleteArgs, cmd)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	flags := rootCmd.Flags()

	flags.StringVar(&args.Entity, "entity", "", "Absolute path to file for the heartbeat. Can also be a url, domain or app when --entity-type is not file.")
	flags.StringVar(&obsoleteArgs.File, "file", "", "") //help missing
	flags.StringVar(&args.Key, "key", "", "Your wakatime api key; uses api_key from ~/.wakatime.cfg by default.")
	flags.BoolVar(&args.Write, "write", false, "When set, tells api this heartbeat was triggered from writing to a file.")
	flags.StringVar(&args.Plugin, "plugin", "", "Optional text editor plugin name and version for User-Agent header.")
	flags.Int64Var(&args.Time, "time", math.MinInt64, "Optional floating-point unix epoch timestamp. Uses current time by default.")
	flags.Int32Var(&args.Lineo, "lineo", math.MinInt32, "Optional line number. This is the current line being edited.")
	flags.Int32Var(&args.CursorPos, "curorpos", math.MinInt32, "Optional cursor position in the current file.")
	flags.StringVar(&args.EntityType, "entity-type", "file", "Entity type for this heartbeat. Can be 'file', 'domain' or 'app'.")
	flags.StringVar(&args.Category, "category", "coding", "Category of this heartbeat activity. Can be 'coding', 'building', 'indexing', 'debugging', 'running tests', 'writing tests', 'manual testing', 'code reviewing', 'browsing', or 'designing'.")
	flags.StringVar(&args.Proxy, "proxy", "", "Optional proxy configuration. Supports HTTPS and SOCKS proxies. For example: https://user:pass@host:port or socks5://user:pass@host:port or domain\\user:pass")
	flags.BoolVar(&args.NoSslVerify, "no-ssl-verify", false, "Disables SSL certificate verification for HTTPS requests. By default, SSL certificates are verified.")
	flags.StringVar(&args.SslCertsFile, "ssl-certs-file", "", "Override the bundled Python Requests CA certs file. By default, uses certifi for ca certs.") //Rename python in the string as soon as it's implemented
	flags.StringVar(&args.Project, "project", "", "Optional project name.")
	flags.StringVar(&args.Branch, "branch", "", "Optional branch name.")
	flags.StringVar(&args.AlternateProject, "alternate-project", "", "Optional alternate project name. Auto-discovered project takes priority.")
	flags.StringVar(&args.AlternateLanguage, "alternate-language", "", "") //help missing
	flags.StringVar(&args.Language, "language", "", "Optional language name. If valid, takes priority over auto-detected language.")
	flags.StringVar(&args.LocalFile, "local-file", "", "Absolute path to local file for the heartbeat. When --entity is a remote file, this local file will be used for stats and just the value of --entity sent with heartbeat.")
	flags.StringVar(&args.Hostname, "hostname", "", "Hostname of current machine.")
	flags.BoolVar(&args.DisableOffline, "disable-offline", true, "Disables offline time logging instead of queuing logged time.")
	flags.BoolVar(&args.HideFileNames, "hide-file-names", false, "Obfuscate filenames. Will not send file names to api.")
	flags.BoolVar(&obsoleteArgs.HideFilenames1, "hide-filenames", false, "") //help missing
	flags.BoolVar(&obsoleteArgs.HideFilenames2, "hidefilenames", false, "")  //help missing
	flags.BoolVar(&args.HideProjectNames, "hide-project-names", false, "Obfuscate project names. When a project older is detected instead of using the folder name as the project, a .wakatime-project file is created with a random project name.")
	flags.BoolVar(&args.HideBranchNames, "hide-branch-names", false, "Obfuscate branch names. Will not send revision control branch names to api.")
	flags.StringSliceVar(&args.Exclude, "exclude", []string{}, "Filename patterns to exclude from logging. POSIX regex syntax. Can be used more than once.")
	flags.BoolVar(&args.ExcludeUnknownProject, "exclude-unknown-project", false, "When set, any activity where the project cannot be detected will be ignored.")
	flags.StringSliceVar(&args.Include, "include", []string{}, "Filename patterns to log. When used in combination with --exclude, files matching include will still be logged. POSIX regex syntax. Can be used more than once.")
	flags.BoolVar(&args.IncludeOnlyWithProjectFile, "include-only-with-project-file", false, "Disables tracking folders unless they contain a .wakatime-project file.")
	flags.StringSliceVar(&args.Ignore, "ignore", []string{}, "") //help missing
	flags.BoolVar(&args.ExtraHeartbeats, "extra-heartbeats", false, "Reads extra heartbeats from STDIN as a JSON array until EOF.")
	flags.StringVar(&args.LogFile, "log-file", "~/.wakatime.log", "") //help missing
	flags.StringVar(&obsoleteArgs.LogFile, "logfile", "", "")         //help missing
	flags.StringVar(&args.APIURL, "api-url", "", "Heartbeats api url. For debugging with a local server.")
	flags.StringVar(&obsoleteArgs.APIURL, "apiurl", "", "") //help missing
	flags.IntVar(&args.Timeout, "timeout", constants.DefaultTimeout, "Number of seconds to wait when sending heartbeats to api.")
	flags.IntVar(&args.SyncOfflineActivity, "sync-offline-activity", constants.DefaultSyncOfflineActivity, "Amount of offline activity to sync from your local ~/.wakatime.db sqlite3 file to your WakaTime Dashboard before exiting. Can be a positive integer number. It means for every heartbeat sent while online, 100 offline heartbeats are synced. Can be used without --entity to only sync offline activity without generating new heartbeats.") //should be revisited when the 'provider' for local database has already been decided
	flags.BoolVar(&args.Today, "today", false, "Prints dashboard time for Today, then exits.")
	flags.StringVar(&args.ConfigPath, "config", "", "Optional config file. Defaults to '~/.wakatime.cfg'.")
	flags.StringVar(&args.ConfigSection, "config-section", "settings", "Optional config section when reading or writing a config key. Defaults to [settings].")
	flags.StringVar(&args.ConfigRead, "config-read", "", "Prints value for the given config key, then exits.")
	flags.StringToStringVar(&args.ConfigWrite, "config-write", map[string]string{}, "Writes value to a config key, then exits. Expects two arguments, key and value.")
	flags.BoolVar(&args.Verbose, "verbose", false, "Turns on debug messages in log file")
	flags.BoolVar(&args.Version, "version", false, "") //help missing

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runLegacy(args legacy.Arguments, obsoleteArgs legacy.ObsoleteArguments, cmd *cobra.Command) {

	flags := cmd.Flags()

	if flags.Changed("version") {
		runVersion()
		os.Exit(constants.Success)
	}

	//For debugging purposes, might be removed later
	if args.Verbose {
		json, _ := json.Marshal(args)
		fmt.Println(string(json))
	}

	cfg := configs.NewConfig(args.ConfigPath)

	if flags.Changed("config-read") {
		runConfigRead(args.ConfigSection, args.ConfigRead, cfg)
		os.Exit(constants.Success)
	}

	if flags.Changed("config-write") {
		runConfigWrite(args.ConfigSection, args.ConfigWrite, cfg)
		os.Exit(constants.Success)
	}

	// use current unix epoch timestamp by default
	if !flags.Changed("time") {
		args.Time = time.Now().Unix()
	}

	// update args from configs
	if !flags.Changed("hostname") {
		hostname, err := cfg.Get("settings", "hostname")
		if err == nil {
			args.Hostname = *hostname
		}
	}

	if !flags.Changed("key") {
		key, err := cfg.Get("settings", "api_key")
		if err != nil {
			key, err = cfg.Get("settings", "apikey")
			if err != nil {
				panic("Missing api key. Find your api key from wakatime.com/settings/api-key.")
			}
		}
		args.Key = *key
	}

	// validate the api key
	if !guid.IsGuid(args.Key) {
		panic("Invalid api key. Find your api key from wakatime.com/settings/api-key.")
	}

	// validate entity
	if !flags.Changed("entity") {
		if flags.Changed("file") {
			args.Entity = obsoleteArgs.File
		} else if !flags.Changed("sync-offline-activity") && !flags.Changed("today") {
			panic("argument --entity is required.")
		}
	}

	if args.SyncOfflineActivity < 0 {
		panic("argument --sync-offline-activity must be a positive integer number.")
	}

	if !flags.Changed("language") && flags.Changed("alternate-language") {
		args.Language = args.AlternateLanguage
	}

	ignore, err := cfg.Get("settings", "ignore")
	if err == nil {
		parts := strings.Split(*ignore, "\n")

		for _, part := range parts {
			part := strings.TrimSpace(part)

			if len(part) > 0 {
				args.Exclude = append(args.Exclude, part)
			}
		}
	}

	if !flags.Changed("include-only-with-project-file") {
		includeOnlyWithProjectFile, err := cfg.Get("settings", "include_only_with_project_file")
		if err == nil {
			b, err := strconv.ParseBool(*includeOnlyWithProjectFile)
			if err != nil {
				b = false
			}
			args.IncludeOnlyWithProjectFile = b == true
		}
	}

	include, err := cfg.Get("settings", "include")
	if err == nil {
		parts := strings.Split(*include, "\n")

		for _, part := range parts {
			part = strings.TrimSpace(part)

			if len(part) > 0 {
				args.Include = append(args.Include, part)
			}
		}
	}

	if !flags.Changed("exclude-unknown-project") {
		excludeUnknownProject, err := cfg.Get("settings", "exclude_unknown_project")
		if err == nil {
			b, err := strconv.ParseBool(*excludeUnknownProject)
			if err != nil {
				b = false
			}
			args.ExcludeUnknownProject = b
		}
	}

	if flags.Changed("hide-file-names") {
		args.HiddenFileNames = append(args.HiddenFileNames, ".*")
	} else {
		args.HiddenFileNames = arguments.GetBooleanOrList("settings", "hide-file-names", nil, cfg)
	}

	if flags.Changed("hide-project-names") {
		args.HiddenProjectNames = append(args.HiddenProjectNames, ".*")
	} else {
		args.HiddenProjectNames = arguments.GetBooleanOrList("settings", "hide-project-names", nil, cfg)
	}

	if flags.Changed("hide-branch-names") {
		args.HiddenBranchNames = append(args.HiddenBranchNames, ".*")
	} else {
		args.HiddenBranchNames = arguments.GetBooleanOrList("settings", "hide-branch-names", nil, cfg)
	}

	if flags.Changed("offline") {
		offline, err := cfg.Get("settings", "offline")
		if err == nil {
			b, err := strconv.ParseBool(*offline)
			if err != nil {
				b = false
			}
			args.DisableOffline = b
		}
	}

	if !flags.Changed("proxy") {
		proxy, err := cfg.Get("settings", "proxy")
		if err == nil {
			args.Proxy = *proxy
		}
	}

	if len(strings.TrimSpace(args.Proxy)) > 0 {
		pattern := "^(?i)((https?|socks5)://)?([^:@]+(:([^:@])+)?@)?[^:]+(:\\d+)?$"

		if strings.Contains(args.Proxy, "\\\\") {
			pattern = "^.*\\.+$"
		}

		isValid, err := regexp.MatchString(pattern, args.Proxy)

		if !isValid || err != nil {
			panic("Invalid proxy. Must be in format https://user:pass@host:port or socks5://user:pass@host:port or domain\\user:pass.")
		}
	}

	noSslVerify, err := cfg.Get("settings", "no_ssl_verify")
	if err == nil {
		b, err := strconv.ParseBool(*noSslVerify)
		if err != nil {
			b = false
		}
		args.NoSslVerify = b
	}

	sslCertsFile, err := cfg.Get("settings", "ssl_certs_file")
	if err == nil {
		args.SslCertsFile = *sslCertsFile
	}

	if !flags.Changed("verbose") {
		verbose, err := cfg.Get("settings", "verbose")
		if err == nil {
			b, err := strconv.ParseBool(*verbose)
			if err != nil {
				b = false
			}
			args.Verbose = b
		} else {
			debug, err := cfg.Get("settings", "debug")
			if err == nil {
				b, err := strconv.ParseBool(*debug)
				if err != nil {
					b = false
				}
				args.Verbose = b
			}
		}
	}

	if !flags.Changed("log-file") {
		if flags.Changed("logfile") {
			args.LogFile = obsoleteArgs.LogFile
		} else if len(args.LogFile) == 0 {
			logFile, err := cfg.Get("settings", "log_file")
			if err == nil {
				args.LogFile = *logFile
			} else {
				home, err := os.LookupEnv("WAKATIME_HOME")

				if !err {
					// Should find a way to 'expanduser'
					args.LogFile = path.Join(home, ".wakatime.log")
				}
			}
		}
	}

	if !flags.Changed("api-url") {
		if flags.Changed("apiurl") {
			args.APIURL = obsoleteArgs.APIURL
		} else if len(args.APIURL) == 0 {
			apiURL, err := cfg.Get("settings", "api_url")
			if err == nil {
				args.APIURL = *apiURL
			}
		}
	}

	if !flags.Changed("timeout") {
		timeout, err := cfg.Get("settings", "timeout")
		if err == nil {
			t, err := strconv.Atoi(*timeout)
			if err == nil {
				args.Timeout = t
			} else {
				fmt.Printf("Error converting to integer the timeout value '%v'", timeout)
			}
		}
	}
}

func runVersion() {
	fmt.Println(constants.Version)
}

func runConfigRead(section string, key string, cfg *configs.ConfigFile) {
	v, err := cfg.Get(section, key)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(*v)
}

func runConfigWrite(section string, keyValue map[string]string, cfg *configs.ConfigFile) {
	v := cfg.Set(section, keyValue)

	fmt.Println(strings.Join(v, "\n"))
}
