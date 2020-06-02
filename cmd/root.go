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
	flags.String("key", "", "Your wakatime api key; uses api_key from ~/.wakatime.cfg by default.")
	flags.String("log-file", "", "Optional log file. Defaults to '~/.wakatime.log'.")
	flags.String("logfile", "", "(deprecated) Optional log file. Defaults to '~/.wakatime.log'.")
	flags.String("plugin", "", "Optional text editor plugin name and version for User-Agent header.")
	flags.Int(
		"timeout",
		defaultTimeoutSecs,
		"Number of seconds to wait when sending heartbeats to api. Defaults to 60 seconds.",
	)
	flags.Bool("today", false, "Prints dashboard time for Today, then exits.")
	flags.Bool("verbose", false, "Turns on debug messages in log file.")
	flags.Bool("version", false, "Prints the wakatime-cli version number, then exits.")

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
