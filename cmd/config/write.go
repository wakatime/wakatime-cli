package config

import (
	"fmt"
	"strings"

	"github.com/alanhamlett/wakatime-cli/lib/configs"

	"github.com/spf13/cobra"
)

type writeOptions struct {
	Path     string
	Section  string
	KeyValue map[string]string
}

func newConfigWriteCommand() *cobra.Command {
	options := writeOptions{}

	cmd := &cobra.Command{
		Use:   "write",
		Short: "Writes value to a config key. Expects two arguments, key and value.",
		Run: func(cmd *cobra.Command, args []string) {
			runConfigWrite(options)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.Path, "config", "c", "", "Optional config file. Defaults to '~/.wakatime.cfg'.")
	flags.StringVarP(&options.Section, "section", "s", "settings", "Optional config section.")
	flags.StringToStringVarP(&options.KeyValue, "key-value", "v", map[string]string{}, "key value pair.")
	cmd.MarkFlagRequired("key-value")

	return cmd
}

func runConfigWrite(options writeOptions) {
	c := configs.NewConfig(options.Path)
	v := c.Set(options.Section, options.KeyValue)

	fmt.Println(strings.Join(v, "\n"))
}
