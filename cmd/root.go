package cmd

import (
	"fmt"
	"os"

	"github.com/wakatime/wakatime-cli/cmd/commands"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wakatime-cli",
	Short: "Command line interface used by all WakaTime text editor plugins.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("For help run: wakatime-cli --help")
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	commands.AddCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
