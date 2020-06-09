package cmd

import (
	"fmt"
	"os"

	"github.com/wakatime/wakatime-cli/cmd/legacy"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

const (
	// defaultConfigSection is the default section in the wakatime ini config file
	defaultConfigSection = "settings"
	// defaultTimeoutSecs is the default timeout used for requests to the wakatime api
	defaultTimeoutSecs = 60
)

// NewRootCMD creates a rootCmd, which represents the base command when called without any subcommands.
func NewRootCMD() *cobra.Command {
	v := viper.GetViper()
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
	flags.String("file", "", "(deprecated) Absolute path to file for the heartbeat.")
	flags.String("hostname", "", "Optional name of local machine. Defaults to local machine name read from system")
	flags.String("key", "", "Your wakatime api key; uses api_key from ~/.wakatime.cfg by default.")
	flags.String("log-file", "", "Optional log file. Defaults to '~/.wakatime.log'.")
	flags.String("logfile", "", "(deprecated) Optional log file. Defaults to '~/.wakatime.log'.")
	flags.Bool(
		"no-ssl-verify",
		false,
		"Disables SSL certificate verification for HTTPS requests. By default,"+
			" SSL certificates are verified.",
	)
	flags.String("plugin", "", "Optional text editor plugin name and version for User-Agent header.")
	flags.String(
		"proxy",
		"",
		"Optional proxy configuration. Supports HTTPS and SOCKS proxies."+
			" For example: 'https://user:pass@host:port' or 'socks5://user:pass@host:port'"+
			" or 'domain\\user:pass'",
	)
	flags.String(
		"ssl-certs-file",
		"",
		"Override the bundled Python Requests CA certs file. By default, uses"+
			"system ca certs.",
	)
	flags.Int(
		"timeout",
		defaultTimeoutSecs,
		"Number of seconds to wait when sending heartbeats to api. Defaults to 60 seconds.",
	)
	flags.Float64("time", 0, "Optional floating-point unix epoch timestamp. Uses current time by default.")
	flags.Bool("today", false, "Prints dashboard time for Today, then exits.")
	flags.Bool("verbose", false, "Turns on debug messages in log file.")
	flags.Bool("version", false, "Prints the wakatime-cli version number, then exits.")
	flags.Bool("write", false, "When set, tells api this heartbeat was triggered from writing to a file.")

	err := v.BindPFlags(flags)
	if err != nil {
		fmt.Printf("failed to bind cobra flags to viper: %s", err)
		os.Exit(exitcode.ErrDefault)
	}
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := NewRootCMD().Execute(); err != nil {
		jww.CRITICAL.Fatalf("failed to run wakatime-cli: %s", err)
		os.Exit(exitcode.ErrDefault)
	}
}
