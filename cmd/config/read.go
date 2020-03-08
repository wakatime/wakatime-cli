package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wakatime/wakatime-cli/lib/configs"
)

type readOptions struct {
	Path    string
	Section string
	Key     string
}

func newConfigReadCommand() *cobra.Command {
	options := readOptions{}

	cmd := &cobra.Command{
		Use:   "read",
		Short: "Prints value for the given config key.",
		Run: func(cmd *cobra.Command, args []string) {
			runConfigRead(options)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.Path, "config", "c", "", "Optional config file. Defaults to '~/.wakatime.cfg'.")
	flags.StringVarP(&options.Section, "section", "s", "settings", "Optional config section.")
	flags.StringVarP(&options.Key, "key", "k", "", "config key.")
	cmd.MarkFlagRequired("key")

	return cmd
}

func runConfigRead(options readOptions) {
	c := configs.NewConfig(options.Path)
	v, err := c.Get(options.Section, options.Key)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(*v)
}
