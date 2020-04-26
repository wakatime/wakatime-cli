package commands

import (
	"github.com/alanhamlett/wakatime-cli/cmd/config"
	"github.com/alanhamlett/wakatime-cli/cmd/heartbeat"
	"github.com/alanhamlett/wakatime-cli/cmd/version"

	"github.com/spf13/cobra"
)

// AddCommands adds all the commands to the root command
func AddCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(
		// version
		version.NewVersionCommand(),

		// heartbeat
		heartbeat.NewHeartbeatCommand(),

		// config
		config.NewConfigCommand(),
	)
}
