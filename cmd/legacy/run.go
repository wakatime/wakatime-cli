package legacy

import (
	"encoding/json"
	"fmt"
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

// Run Run legacy commands
func Run(args *arguments.Arguments, cfg configs.WakaTimeConfig, cmd *cobra.Command) {
	if cfg == nil {
		cfg = configs.NewConfig(args.Config.Path)
	}

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

	if flags.Changed("config-read") {
		runConfigRead(args.Config.Section, args.Config.Read, cfg)
		os.Exit(constants.Success)
	}

	if flags.Changed("config-write") {
		runConfigWrite(args.Config.Section, args.Config.Write, cfg)
		os.Exit(constants.Success)
	}

	runSendHeartbeat(args, cfg, cmd)
}

func runVersion() {
	fmt.Println(constants.Version)
}

func runConfigRead(section string, key string, cfg configs.WakaTimeConfig) {
	v, err := cfg.Get(section, key)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(v)
}

func runConfigWrite(section string, keyValue map[string]string, cfg configs.WakaTimeConfig) {
	v := cfg.Set(section, keyValue)

	fmt.Println(strings.Join(v, "\n"))
}

func runSendHeartbeat(args *arguments.Arguments, cfg configs.WakaTimeConfig, cmd *cobra.Command) {
	flags := cmd.Flags()

	// use current unix epoch timestamp by default
	if !flags.Changed("time") {
		args.Time = time.Now().Unix()
	}

	// update args from configs
	if !flags.Changed("hostname") {
		hostname, err := cfg.Get("settings", "hostname")
		if err == nil {
			args.Hostname = hostname
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
		args.Key = key
	}

	// validate the api key
	if !guid.IsGuid(args.Key) {
		panic("Invalid api key. Find your api key from wakatime.com/settings/api-key.")
	}

	// validate entity
	if !flags.Changed("entity") {
		if flags.Changed("file") {
			args.Entity.Entity = args.ObsoleteArgs.File
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
		parts := strings.Split(ignore, "\n")

		for _, part := range parts {
			part = strings.TrimSpace(part)

			if len(part) > 0 {
				args.Exclude.Exclude = append(args.Exclude.Exclude, part)
			}
		}
	}

	exclude, err := cfg.Get("settings", "exclude")
	if err == nil {
		parts := strings.Split(exclude, "\n")

		for _, part := range parts {
			part = strings.TrimSpace(part)

			if len(part) > 0 {
				args.Exclude.Exclude = append(args.Exclude.Exclude, part)
			}
		}
	}

	if !flags.Changed("include-only-with-project-file") {
		includeOnlyWithProjectFile, err := cfg.Get("settings", "include_only_with_project_file")
		if err == nil {
			b, err := strconv.ParseBool(includeOnlyWithProjectFile)
			if err != nil {
				b = false
			}
			args.Include.IncludeOnlyWithProjectFile = b
		}
	}

	include, err := cfg.Get("settings", "include")
	if err == nil {
		parts := strings.Split(include, "\n")

		for _, part := range parts {
			part = strings.TrimSpace(part)

			if len(part) > 0 {
				args.Include.Include = append(args.Include.Include, part)
			}
		}
	}

	if !flags.Changed("exclude-unknown-project") {
		excludeUnknownProject, err := cfg.Get("settings", "exclude_unknown_project")
		if err == nil {
			b, err := strconv.ParseBool(excludeUnknownProject)
			if err != nil {
				b = false
			}
			args.Exclude.ExcludeUnknownProject = b
		}
	}

	if flags.Changed("hide-file-names") {
		args.Obfuscate.HiddenFileNames = append(args.Obfuscate.HiddenFileNames, ".*")
	} else {
		args.Obfuscate.HiddenFileNames = arguments.GetBooleanOrList("settings", "hide-file-names", nil, cfg)
	}

	if flags.Changed("hide-project-names") {
		args.Obfuscate.HiddenProjectNames = append(args.Obfuscate.HiddenProjectNames, ".*")
	} else {
		args.Obfuscate.HiddenProjectNames = arguments.GetBooleanOrList("settings", "hide-project-names", nil, cfg)
	}

	if flags.Changed("hide-branch-names") {
		args.Obfuscate.HiddenBranchNames = append(args.Obfuscate.HiddenBranchNames, ".*")
	} else {
		args.Obfuscate.HiddenBranchNames = arguments.GetBooleanOrList("settings", "hide-branch-names", nil, cfg)
	}

	if flags.Changed("disable-offline") {
		offline, err := cfg.Get("settings", "offline")
		if err == nil {
			b, err := strconv.ParseBool(offline)
			if err != nil {
				b = false
			}
			args.DisableOffline = b
		}
	}

	if !flags.Changed("proxy") {
		proxy, err := cfg.Get("settings", "proxy")
		if err == nil {
			args.Proxy.Address = proxy
		}
	}

	if len(strings.TrimSpace(args.Proxy.Address)) > 0 {
		pattern := "^(?i)((https?|socks5)://)?([^:@]+(:([^:@])+)?@)?[^:]+(:\\d+)?$"

		if strings.Contains(args.Proxy.Address, "\\\\") {
			pattern = "^.*\\.+$"
		}

		isValid, err := regexp.MatchString(pattern, args.Proxy.Address)
		if !isValid || err != nil {
			panic("Invalid proxy. Must be in format https://user:pass@host:port or socks5://user:pass@host:port or domain\\user:pass.")
		}
	}

	noSslVerify, err := cfg.Get("settings", "no_ssl_verify")
	if err == nil {
		b, err := strconv.ParseBool(noSslVerify)
		if err != nil {
			b = false
		}
		args.Proxy.NoSslVerify = b
	}

	sslCertsFile, err := cfg.Get("settings", "ssl_certs_file")
	if err == nil {
		args.Proxy.SslCertsFile = sslCertsFile
	}

	if !flags.Changed("verbose") {
		verbose, err := cfg.Get("settings", "verbose")
		if err == nil {
			b, err := strconv.ParseBool(verbose)
			if err != nil {
				b = false
			}
			args.Verbose = b
		} else {
			debug, err := cfg.Get("settings", "debug")
			if err == nil {
				b, err := strconv.ParseBool(debug)
				if err != nil {
					b = false
				}
				args.Verbose = b
			}
		}
	}

	if !flags.Changed("log-file") {
		if flags.Changed("logfile") {
			args.LogFile = args.ObsoleteArgs.LogFile
		} else {
			logFile, err := cfg.Get("settings", "log_file")
			if err != nil {
				home, err := os.LookupEnv("WAKATIME_HOME")
				if !err {
					// Should find a way to 'expanduser'
					args.LogFile = path.Join(home, ".wakatime.log")
				}
			} else {
				args.LogFile = logFile
			}
		}
	}

	if !flags.Changed("api-url") {
		if flags.Changed("apiurl") {
			args.APIURL = args.ObsoleteArgs.APIURL
		} else {
			apiURL, err := cfg.Get("settings", "api_url")
			if err == nil {
				args.APIURL = apiURL
			}
		}
	}

	if !flags.Changed("timeout") {
		timeout, err := cfg.Get("settings", "timeout")
		if err == nil {
			t, err := strconv.Atoi(timeout)
			if err != nil {
				fmt.Printf("Error converting to integer the timeout value '%v'", timeout)
			} else {
				args.Timeout = t
			}
		}
	}
}
