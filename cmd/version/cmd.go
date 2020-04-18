package version

import (
	"fmt"

	"github.com/alanhamlett/wakatime-cli/constants"

	"github.com/spf13/cobra"
)

// NewVersionCommand returns a cobra command for `version` subcommands
func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information and quit",
		Run: func(cmd *cobra.Command, args []string) {
			runVersion()
		},
	}
	return cmd
}

func runVersion() {
	fmt.Println(constants.Version)
}
