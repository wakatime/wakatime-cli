package heartbeat

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
	"github.com/wakatime/wakatime-cli/constants"
	"github.com/wakatime/wakatime-cli/lib/arguments"
	"github.com/wakatime/wakatime-cli/lib/configs"
)

type heartbeatOptions struct {
	Entity                     string
	File                       string
	Key                        string
	Write                      bool
	Plugin                     string
	Time                       int64
	Lineo                      int32
	CursorPos                  int32
	EntityType                 string
	Category                   string
	Proxy                      string
	NoSslVerify                bool
	SslCertsFile               string
	Project                    string
	AlternateProject           string
	AlternateLanguage          string
	Language                   string
	LocalFile                  string
	Hostname                   string
	DisableOffline             bool
	HideFileNames              bool
	HiddenFileNames            []string
	HideProjectNames           bool
	HiddenProjectNames         []string
	HideBranchNames            bool
	HiddenBranchNames          []string
	Exclude                    []string
	ExcludeUnknownProject      bool
	Include                    []string
	IncludeOnlyWithProjectFile bool
	Ignore                     []string
	ExtraHeartbeats            bool
	LogFile                    string
	ApiUrl                     string
	Timeout                    int
	SyncOfflineActivity        int
	ConfigPath                 string
	Verbose                    bool
}

// NewHeartbeatCommand returns a cobra command for `heartbeat` subcommands
func NewHeartbeatCommand() *cobra.Command {
	options := heartbeatOptions{}

	cmd := &cobra.Command{
		Use:   "heartbeat",
		Short: "Send a heartbeat",
		Run: func(cmd *cobra.Command, args []string) {
			runHeartbeat(options, cmd)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&options.Entity, "entity", "", "Absolute path to file for the heartbeat. Can also be a url, domain or app when --entity-type is not file.")
	flags.StringVar(&options.File, "file", "", "Obsolete. Use --entity instead.")
	flags.StringVar(&options.Key, "key", "", "Your wakatime api key; uses api_key from ~/.wakatime.cfg by default.")
	flags.BoolVar(&options.Write, "write", false, "When set, tells api this heartbeat was triggered from writing to a file.")
	flags.StringVar(&options.Plugin, "plugin", "", "Optional text editor plugin name and version for User-Agent header.")
	flags.Int64Var(&options.Time, "time", math.MinInt64, "Optional floating-point unix epoch timestamp. Uses current time by default.")
	flags.Int32Var(&options.Lineo, "lineo", math.MinInt32, "Optional line number. This is the current line being edited.")
	flags.Int32Var(&options.CursorPos, "curorpos", math.MinInt32, "Optional cursor position in the current file.")
	flags.StringVar(&options.EntityType, "entity-type", "file", "Entity type for this heartbeat. Can be 'file', 'domain' or 'app'.")
	flags.StringVar(&options.Category, "category", "coding", "Category of this heartbeat activity. Can be 'coding', 'building', 'indexing', 'debugging', 'running tests', 'writing tests', 'manual testing', 'code reviewing', 'browsing', or 'designing'.")
	flags.StringVar(&options.Proxy, "proxy", "", "Optional proxy configuration. Supports HTTPS and SOCKS proxies. For example: https://user:pass@host:port or socks5://user:pass@host:port or domain\\user:pass")
	flags.BoolVar(&options.NoSslVerify, "no-ssl-verify", false, "Disables SSL certificate verification for HTTPS requests. By default, SSL certificates are verified.")
	flags.StringVar(&options.SslCertsFile, "ssl-certs-file", "", "Override the bundled Python Requests CA certs file. By default, uses certifi for ca certs.") //Rename python in the string as soon as it's implemented
	flags.StringVar(&options.Project, "project", "", "Optional project name.")
	flags.StringVar(&options.AlternateProject, "alternate-project", "", "Optional alternate project name. Auto-discovered project takes priority.")
	flags.StringVar(&options.AlternateLanguage, "alternate-language", "", "") //help missing
	flags.StringVar(&options.Language, "language", "", "Optional language name. If valid, takes priority over auto-detected language.")
	flags.StringVar(&options.LocalFile, "local-file", "", "Absolute path to local file for the heartbeat. When --entity is a remote file, this local file will be used for stats and just the value of --entity sent with heartbeat.")
	flags.StringVar(&options.Hostname, "hostname", "", "Hostname of current machine.")
	flags.BoolVar(&options.DisableOffline, "disable-offline", true, "Disables offline time logging instead of queuing logged time.")
	flags.BoolVar(&options.HideFileNames, "hide-file-names", false, "Obfuscate filenames. Will not send file names to api.")
	flags.BoolVar(&options.HideProjectNames, "hide-project-names", false, "Obfuscate project names. When a project older is detected instead of using the folder name as the project, a .wakatime-project file is created with a random project name.")
	flags.BoolVar(&options.HideBranchNames, "hide-branch-names", false, "Obfuscate branch names. Will not send revision control branch names to api.")
	flags.StringSliceVar(&options.Exclude, "exclude", []string{}, "Filename patterns to exclude from logging. POSIX regex syntax. Can be used more than once.")
	flags.BoolVar(&options.ExcludeUnknownProject, "exclude-unknown-project", false, "When set, any activity where the project cannot be detected will be ignored.")
	flags.StringSliceVar(&options.Include, "include", []string{}, "Filename patterns to log. When used in combination with --exclude, files matching include will still be logged. POSIX regex syntax. Can be used more than once.")
	flags.BoolVar(&options.IncludeOnlyWithProjectFile, "include-only-with-project-file", false, "Disables tracking folders unless they contain a .wakatime-project file.")
	flags.StringSliceVar(&options.Ignore, "ignore", []string{}, "") //help missing
	flags.BoolVar(&options.ExtraHeartbeats, "extra-heartbeats", false, "Reads extra heartbeats from STDIN as a JSON array until EOF.")
	flags.StringVar(&options.LogFile, "log-file", "~/.wakatime.log", "") //help missing
	flags.StringVar(&options.ApiUrl, "api-url", "", "Heartbeats api url. For debugging with a local server.")
	flags.IntVar(&options.Timeout, "timeout", constants.DefaultTimeout, "Number of seconds to wait when sending heartbeats to api.")
	flags.IntVar(&options.SyncOfflineActivity, "sync-offline-activity", constants.DefaultSyncOfflineActivity, "Amount of offline activity to sync from your local ~/.wakatime.db sqlite3 file to your WakaTime Dashboard before exiting. Can be a positive integer number. It means for every heartbeat sent while online, 100 offline heartbeats are synced. Can be used without --entity to only sync offline activity without generating new heartbeats.") //should be revisited when the 'provider' for local database has already been decided
	flags.StringVar(&options.ConfigPath, "config", "", "Optional config file. Defaults to '~/.wakatime.cfg'.")
	flags.BoolVar(&options.Verbose, "verbose", false, "Turns on debug messages in log file")

	return cmd
}

func runHeartbeat(options heartbeatOptions, cmd *cobra.Command) {
	if options.Verbose {
		json, _ := json.Marshal(options)
		fmt.Println(string(json))
	}

	cfg := configs.NewConfig(options.ConfigPath)
	flags := cmd.Flags()

	// use current unix epoch timestamp by default
	if !flags.Changed("time") {
		options.Time = time.Now().Unix()
	}

	// update args from configs
	if !flags.Changed("hostname") {
		hostname, err := cfg.Get("settings", "hostname")
		if err == nil {
			options.Hostname = *hostname
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
		options.Key = *key
	}

	// validate the api key
	if !guid.IsGuid(options.Key) {
		panic("Invalid api key. Find your api key from wakatime.com/settings/api-key.")
	}

	// validate entity || file
	if !flags.Changed("entity") {
		if flags.Changed("file") {
			options.Entity = options.File
		} else if !flags.Changed("sync-offline-activity") {
			panic("argument --entity is required.")
		}
	}

	if options.SyncOfflineActivity < 0 {
		panic("argument --sync-offline-activity must be a positive integer number.")
	}

	if !flags.Changed("language") && flags.Changed("alternate-language") {
		options.Language = options.AlternateLanguage
	}

	ignore, err := cfg.Get("settings", "ignore")
	if err == nil {
		parts := strings.Split(*ignore, "\n")

		for _, part := range parts {
			part := strings.TrimSpace(part)

			if len(part) > 0 {
				options.Exclude = append(options.Exclude, part)
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
			options.IncludeOnlyWithProjectFile = b == true
		}
	}

	include, err := cfg.Get("settings", "include")
	if err == nil {
		parts := strings.Split(*include, "\n")

		for _, part := range parts {
			part = strings.TrimSpace(part)

			if len(part) > 0 {
				options.Include = append(options.Include, part)
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
			options.ExcludeUnknownProject = b
		}
	}

	if flags.Changed("hide-file-names") {
		options.HiddenFileNames = append(options.HiddenFileNames, ".*")
	} else {
		options.HiddenFileNames = arguments.GetBooleanOrList("settings", "hide-file-names", cfg)
	}

	if flags.Changed("hide-project-names") {
		options.HiddenProjectNames = append(options.HiddenProjectNames, ".*")
	} else {
		options.HiddenProjectNames = arguments.GetBooleanOrList("settings", "hide-project-names", cfg)
	}

	if flags.Changed("hide-branch-names") {
		options.HiddenBranchNames = append(options.HiddenBranchNames, ".*")
	} else {
		options.HiddenBranchNames = arguments.GetBooleanOrList("settings", "hide-branch-names", cfg)
	}

	if flags.Changed("offline") {
		offline, err := cfg.Get("settings", "offline")
		if err == nil {
			b, err := strconv.ParseBool(*offline)
			if err != nil {
				b = false
			}
			options.DisableOffline = b
		}
	}

	if !flags.Changed("proxy") {
		proxy, err := cfg.Get("settings", "proxy")
		if err == nil {
			options.Proxy = *proxy
		}
	}

	if len(strings.TrimSpace(options.Proxy)) > 0 {
		pattern := "^(?i)((https?|socks5)://)?([^:@]+(:([^:@])+)?@)?[^:]+(:\\d+)?$"

		if strings.Contains(options.Proxy, "\\\\") {
			pattern = "^.*\\.+$"
		}

		isValid, err := regexp.MatchString(pattern, options.Proxy)

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
		options.NoSslVerify = b
	}

	sslCertsFile, err := cfg.Get("settings", "ssl_certs_file")
	if err == nil {
		options.SslCertsFile = *sslCertsFile
	}

	if !flags.Changed("verbose") {
		verbose, err := cfg.Get("settings", "verbose")
		if err == nil {
			b, err := strconv.ParseBool(*verbose)
			if err != nil {
				b = false
			}
			options.Verbose = b
		} else {
			debug, err := cfg.Get("settings", "debug")
			if err == nil {
				b, err := strconv.ParseBool(*debug)
				if err != nil {
					b = false
				}
				options.Verbose = b
			}
		}
	}

	if !flags.Changed("log-file") {
		logFile, err := cfg.Get("settings", "log_file")
		if err == nil {
			options.LogFile = *logFile
		} else {
			home, err := os.LookupEnv("WAKATIME_HOME")

			if !err {
				// Should find a way to 'expanduser'
				options.LogFile = path.Join(home, ".wakatime.log")
			}
		}
	}

	if !flags.Changed("api-url") {
		apiUrl, err := cfg.Get("settings", "api_url")
		if err == nil {
			options.ApiUrl = *apiUrl
		}
	}

	if !flags.Changed("timeout") {
		timeout, err := cfg.Get("settings", "timeout")
		if err == nil {
			t, err := strconv.Atoi(*timeout)
			if err == nil {
				options.Timeout = t
			} else {
				fmt.Printf("Error converting to integer the timeout value '%v'", timeout)
			}
		}
	}
}
