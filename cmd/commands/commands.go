package commands

import (
	"github.com/spf13/cobra"
	"github.com/wakatime/wakatime-cli/cmd/config"
	"github.com/wakatime/wakatime-cli/cmd/heartbeat"
	"github.com/wakatime/wakatime-cli/cmd/version"
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
