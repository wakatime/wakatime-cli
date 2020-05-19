package cmd

import (
	"fmt"
	"os"

	"github.com/wakatime/wakatime-cli/cmd/legacy"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
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
	flags.String("config", "", "Optional config file. Defaults to '~/.wakatime.cfg'.")
	flags.String("config-read", "", "Prints value for the given config key, then exits.")
	flags.String(
		"config-section",
		"settings",
		"Optional config section when reading or writing a config key. Defaults to [settings].",
	)
	flags.Bool("verbose", false, "Turns on debug messages in log file")
	flags.Bool("version", false, "") // help missing

	err := v.BindPFlags(flags)
	if err != nil {
		fmt.Printf("failed to bind cobra flags to viper: %s", err)
		os.Exit(errCodeDefault)
	}
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := NewRootCMD().Execute(); err != nil {
		jww.CRITICAL.Fatalf("failed to run wakatime-cli: %s", err)
		os.Exit(errCodeDefault)
	}
}
