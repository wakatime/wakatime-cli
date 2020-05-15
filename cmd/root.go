package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/wakatime/wakatime-cli/cmd/legacy"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

const configFileParseError int = 103

// NewRootCMD creates a rootCmd, which represents the base command when called without any subcommands.
func NewRootCMD() *cobra.Command {
	v := viper.GetViper()
	cmd := &cobra.Command{
		Use:   "wakatime-cli",
		Short: "Command line interface used by all WakaTime text editor plugins.",
		Run: func(cmd *cobra.Command, args []string) {
			loadConfigFile(v)
			legacy.Run(v)
		},
	}

	setFlags(cmd, v)

	return cmd
}

func setFlags(cmd *cobra.Command, v *viper.Viper) {
	flags := cmd.Flags()
	flags.Bool("version", false, "") // help missing
	flags.String("config", "", "Optional config file. Defaults to '~/.wakatime.cfg'.")
	flags.String("config-section", "settings", "Optional config section when reading or writing a config key. Defaults to [settings].")
	flags.String("config-read", "", "Prints value for the given config key, then exits.")
	flags.Bool("verbose", false, "Turns on debug messages in log file")
	err := v.BindPFlags(flags)
	if err != nil {
		fmt.Printf("failed to bind cobra flags to viper: %s", err)
		os.Exit(1)
	}
}

func loadConfigFile(v *viper.Viper) {
	var configFilepath string
	var err error

	configFilepath = v.GetString("config")

	if configFilepath == "" {
		configFilepath, err = getConfigFile()
		if err != nil {
			jww.CRITICAL.Panicf("Error loading config file, %s", err)
			os.Exit(configFileParseError)
		}
	}

	jww.DEBUG.Println("wakatime path:", configFilepath)

	v.SetConfigType("ini")
	v.SetConfigFile(configFilepath)
	if err := v.ReadInConfig(); err != nil {
		jww.CRITICAL.Panicf("Error reading config file, %s", err)
		os.Exit(configFileParseError)
	}
}

func getConfigFile() (string, error) {
	fileName := ".wakatime.cfg"
	home, exists := os.LookupEnv("WAKATIME_HOME")

	if exists {
		p, err := homedir.Expand(home)
		if err != nil {
			return "", err
		}
		return path.Join(p, fileName), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(home, fileName), nil
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := NewRootCMD().Execute(); err != nil {
		jww.CRITICAL.Fatalln(err)
		os.Exit(1)
	}
}
