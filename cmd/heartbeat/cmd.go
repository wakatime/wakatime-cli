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

	"github.com/alanhamlett/wakatime-cli/constants"
	"github.com/alanhamlett/wakatime-cli/lib/arguments"
	"github.com/alanhamlett/wakatime-cli/lib/configs"

	"github.com/beevik/guid"
	"github.com/spf13/cobra"
)

// NewHeartbeatCommand returns a cobra command for `heartbeat` subcommands
func NewHeartbeatCommand() *cobra.Command {
	heartbeat := Heartbeat{}

	cmd := &cobra.Command{
		Use:   "heartbeat",
		Short: "Send a heartbeat",
		Run: func(cmd *cobra.Command, args []string) {
			runHeartbeat(heartbeat, cmd)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&heartbeat.Entity, "entity", "", "Absolute path to file for the heartbeat. Can also be a url, domain or app when --entity-type is not file.")
	flags.StringVar(&heartbeat.Key, "key", "", "Your wakatime api key; uses api_key from ~/.wakatime.cfg by default.")
	flags.BoolVar(&heartbeat.Write, "write", false, "When set, tells api this heartbeat was triggered from writing to a file.")
	flags.StringVar(&heartbeat.Plugin, "plugin", "", "Optional text editor plugin name and version for User-Agent header.")
	flags.Int64Var(&heartbeat.Time, "time", math.MinInt64, "Optional floating-point unix epoch timestamp. Uses current time by default.")
	flags.Int32Var(&heartbeat.Lineo, "lineo", math.MinInt32, "Optional line number. This is the current line being edited.")
	flags.Int32Var(&heartbeat.CursorPos, "curorpos", math.MinInt32, "Optional cursor position in the current file.")
	flags.StringVar(&heartbeat.EntityType, "entity-type", "file", "Entity type for this heartbeat. Can be 'file', 'domain' or 'app'.")
	flags.StringVar(&heartbeat.Category, "category", "coding", "Category of this heartbeat activity. Can be 'coding', 'building', 'indexing', 'debugging', 'running tests', 'writing tests', 'manual testing', 'code reviewing', 'browsing', or 'designing'.")
	flags.StringVar(&heartbeat.Proxy, "proxy", "", "Optional proxy configuration. Supports HTTPS and SOCKS proxies. For example: https://user:pass@host:port or socks5://user:pass@host:port or domain\\user:pass")
	flags.BoolVar(&heartbeat.NoSslVerify, "no-ssl-verify", false, "Disables SSL certificate verification for HTTPS requests. By default, SSL certificates are verified.")
	flags.StringVar(&heartbeat.SslCertsFile, "ssl-certs-file", "", "Override the bundled Python Requests CA certs file. By default, uses certifi for ca certs.") //Rename python in the string as soon as it's implemented
	flags.StringVar(&heartbeat.Project, "project", "", "Optional project name.")
	flags.StringVar(&heartbeat.Branch, "branch", "", "Optional branch name.")
	flags.StringVar(&heartbeat.AlternateProject, "alternate-project", "", "Optional alternate project name. Auto-discovered project takes priority.")
	flags.StringVar(&heartbeat.AlternateLanguage, "alternate-language", "", "") //help missing
	flags.StringVar(&heartbeat.Language, "language", "", "Optional language name. If valid, takes priority over auto-detected language.")
	flags.StringVar(&heartbeat.LocalFile, "local-file", "", "Absolute path to local file for the heartbeat. When --entity is a remote file, this local file will be used for stats and just the value of --entity sent with heartbeat.")
	flags.StringVar(&heartbeat.Hostname, "hostname", "", "Hostname of current machine.")
	flags.BoolVar(&heartbeat.DisableOffline, "disable-offline", true, "Disables offline time logging instead of queuing logged time.")
	flags.BoolVar(&heartbeat.HideFileNames, "hide-file-names", false, "Obfuscate filenames. Will not send file names to api.")
	flags.BoolVar(&heartbeat.HideProjectNames, "hide-project-names", false, "Obfuscate project names. When a project older is detected instead of using the folder name as the project, a .wakatime-project file is created with a random project name.")
	flags.BoolVar(&heartbeat.HideBranchNames, "hide-branch-names", false, "Obfuscate branch names. Will not send revision control branch names to api.")
	flags.StringSliceVar(&heartbeat.Exclude, "exclude", []string{}, "Filename patterns to exclude from logging. POSIX regex syntax. Can be used more than once.")
	flags.BoolVar(&heartbeat.ExcludeUnknownProject, "exclude-unknown-project", false, "When set, any activity where the project cannot be detected will be ignored.")
	flags.StringSliceVar(&heartbeat.Include, "include", []string{}, "Filename patterns to log. When used in combination with --exclude, files matching include will still be logged. POSIX regex syntax. Can be used more than once.")
	flags.BoolVar(&heartbeat.IncludeOnlyWithProjectFile, "include-only-with-project-file", false, "Disables tracking folders unless they contain a .wakatime-project file.")
	flags.StringSliceVar(&heartbeat.Ignore, "ignore", []string{}, "") //help missing
	flags.BoolVar(&heartbeat.ExtraHeartbeats, "extra-heartbeats", false, "Reads extra heartbeats from STDIN as a JSON array until EOF.")
	flags.StringVar(&heartbeat.LogFile, "log-file", "~/.wakatime.log", "") //help missing
	flags.StringVar(&heartbeat.ApiUrl, "api-url", "", "Heartbeats api url. For debugging with a local server.")
	flags.IntVar(&heartbeat.Timeout, "timeout", constants.DefaultTimeout, "Number of seconds to wait when sending heartbeats to api.")
	flags.IntVar(&heartbeat.SyncOfflineActivity, "sync-offline-activity", constants.DefaultSyncOfflineActivity, "Amount of offline activity to sync from your local ~/.wakatime.db sqlite3 file to your WakaTime Dashboard before exiting. Can be a positive integer number. It means for every heartbeat sent while online, 100 offline heartbeats are synced. Can be used without --entity to only sync offline activity without generating new heartbeats.") //should be revisited when the 'provider' for local database has already been decided
	flags.StringVar(&heartbeat.ConfigPath, "config", "", "Optional config file. Defaults to '~/.wakatime.cfg'.")
	flags.BoolVar(&heartbeat.Verbose, "verbose", false, "Turns on debug messages in log file")

	return cmd
}

func runHeartbeat(heartbeat Heartbeat, cmd *cobra.Command) {
	if heartbeat.Verbose {
		json, _ := json.Marshal(heartbeat)
		fmt.Println(string(json))
	}

	cfg := configs.NewConfig(heartbeat.ConfigPath)
	flags := cmd.Flags()

	// use current unix epoch timestamp by default
	if !flags.Changed("time") {
		heartbeat.Time = time.Now().Unix()
	}

	// update args from configs
	if !flags.Changed("hostname") {
		hostname, err := cfg.Get("settings", "hostname")
		if err == nil {
			heartbeat.Hostname = *hostname
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
		heartbeat.Key = *key
	}

	// validate the api key
	if !guid.IsGuid(heartbeat.Key) {
		panic("Invalid api key. Find your api key from wakatime.com/settings/api-key.")
	}

	// validate entity
	if !flags.Changed("entity") && !flags.Changed("sync-offline-activity") {
		panic("argument --entity is required.")
	}

	if heartbeat.SyncOfflineActivity < 0 {
		panic("argument --sync-offline-activity must be a positive integer number.")
	}

	if !flags.Changed("language") && flags.Changed("alternate-language") {
		heartbeat.Language = heartbeat.AlternateLanguage
	}

	ignore, err := cfg.Get("settings", "ignore")
	if err == nil {
		parts := strings.Split(*ignore, "\n")

		for _, part := range parts {
			part := strings.TrimSpace(part)

			if len(part) > 0 {
				heartbeat.Exclude = append(heartbeat.Exclude, part)
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
			heartbeat.IncludeOnlyWithProjectFile = b == true
		}
	}

	include, err := cfg.Get("settings", "include")
	if err == nil {
		parts := strings.Split(*include, "\n")

		for _, part := range parts {
			part = strings.TrimSpace(part)

			if len(part) > 0 {
				heartbeat.Include = append(heartbeat.Include, part)
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
			heartbeat.ExcludeUnknownProject = b
		}
	}

	if flags.Changed("hide-file-names") {
		heartbeat.HiddenFileNames = append(heartbeat.HiddenFileNames, ".*")
	} else {
		heartbeat.HiddenFileNames = arguments.GetBooleanOrList("settings", "hide-file-names", cfg)
	}

	if flags.Changed("hide-project-names") {
		heartbeat.HiddenProjectNames = append(heartbeat.HiddenProjectNames, ".*")
	} else {
		heartbeat.HiddenProjectNames = arguments.GetBooleanOrList("settings", "hide-project-names", cfg)
	}

	if flags.Changed("hide-branch-names") {
		heartbeat.HiddenBranchNames = append(heartbeat.HiddenBranchNames, ".*")
	} else {
		heartbeat.HiddenBranchNames = arguments.GetBooleanOrList("settings", "hide-branch-names", cfg)
	}

	if flags.Changed("offline") {
		offline, err := cfg.Get("settings", "offline")
		if err == nil {
			b, err := strconv.ParseBool(*offline)
			if err != nil {
				b = false
			}
			heartbeat.DisableOffline = b
		}
	}

	if !flags.Changed("proxy") {
		proxy, err := cfg.Get("settings", "proxy")
		if err == nil {
			heartbeat.Proxy = *proxy
		}
	}

	if len(strings.TrimSpace(heartbeat.Proxy)) > 0 {
		pattern := "^(?i)((https?|socks5)://)?([^:@]+(:([^:@])+)?@)?[^:]+(:\\d+)?$"

		if strings.Contains(heartbeat.Proxy, "\\\\") {
			pattern = "^.*\\.+$"
		}

		isValid, err := regexp.MatchString(pattern, heartbeat.Proxy)

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
		heartbeat.NoSslVerify = b
	}

	sslCertsFile, err := cfg.Get("settings", "ssl_certs_file")
	if err == nil {
		heartbeat.SslCertsFile = *sslCertsFile
	}

	if !flags.Changed("verbose") {
		verbose, err := cfg.Get("settings", "verbose")
		if err == nil {
			b, err := strconv.ParseBool(*verbose)
			if err != nil {
				b = false
			}
			heartbeat.Verbose = b
		} else {
			debug, err := cfg.Get("settings", "debug")
			if err == nil {
				b, err := strconv.ParseBool(*debug)
				if err != nil {
					b = false
				}
				heartbeat.Verbose = b
			}
		}
	}

	if !flags.Changed("log-file") {
		logFile, err := cfg.Get("settings", "log_file")
		if err == nil {
			heartbeat.LogFile = *logFile
		} else {
			home, err := os.LookupEnv("WAKATIME_HOME")

			if !err {
				// Should find a way to 'expanduser'
				heartbeat.LogFile = path.Join(home, ".wakatime.log")
			}
		}
	}

	if !flags.Changed("api-url") {
		apiUrl, err := cfg.Get("settings", "api_url")
		if err == nil {
			heartbeat.ApiUrl = *apiUrl
		}
	}

	if !flags.Changed("timeout") {
		timeout, err := cfg.Get("settings", "timeout")
		if err == nil {
			t, err := strconv.Atoi(*timeout)
			if err == nil {
				heartbeat.Timeout = t
			} else {
				fmt.Printf("Error converting to integer the timeout value '%v'", timeout)
			}
		}
	}
}
