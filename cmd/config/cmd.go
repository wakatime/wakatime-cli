package config

import "github.com/spf13/cobra"

// NewConfigCommand returns a cobra command for `config` subcommands
func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manipulate the configuration file",
	}

	cmd.AddCommand(
		newConfigReadCommand(),
		newConfigWriteCommand(),
	)

	return cmd
}
