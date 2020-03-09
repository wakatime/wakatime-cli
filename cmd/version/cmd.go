package version

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wakatime/wakatime-cli/constants"
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
