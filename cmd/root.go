package cmd

import (
	"fmt"
	"os"

	"github.com/alanhamlett/wakatime-cli/cmd/legacy"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRootCMD creates a rootCmd, which represents the base command when called without any subcommands
func NewRootCMD() *cobra.Command {
	v := viper.GetViper()
	cmd := &cobra.Command{
		Use:   "wakatime-cli",
		Short: "Command line interface used by all WakaTime text editor plugins.",
		Run: func(cmd *cobra.Command, args []string) {
			legacy.Run(v)
		},
	}

	// set flags
	flags := cmd.Flags()
	flags.Bool("version", false, "") // help missing
	v.BindPFlags(flags)

	return cmd
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := NewRootCMD().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
