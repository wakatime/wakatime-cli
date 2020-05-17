package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/wakatime/wakatime-cli/cmd/legacy"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

const (
	defaultConfigFile      = ".wakatime.cfg"
	errCodeConfigFileParse = 103
	errCodeDefault         = 1
	successCode            = 0
)

// NewRootCMD creates a rootCmd, which represents the base command when called without any subcommands.
func NewRootCMD() *cobra.Command {
	v := viper.GetViper()
	cmd := &cobra.Command{
		Use:   "wakatime-cli",
		Short: "Command line interface used by all WakaTime text editor plugins.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := ReadInConfig(v, ConfigFilePath); err != nil {
				jww.CRITICAL.Printf("err: %s", err)
				var cfperr ErrConfigFileParse
				if errors.As(err, &cfperr) {
					os.Exit(errCodeConfigFileParse)
				}
				os.Exit(errCodeDefault)
			}
			legacy.Run(v)
		},
	}

	setFlags(cmd, v)

	return cmd
}

func setFlags(cmd *cobra.Command, v *viper.Viper) {
	flags := cmd.Flags()
	flags.String("config", "", "Optional config file. Defaults to '~/.wakatime.cfg'.")
	flags.String("config-read", "", "Prints value for the given config key, then exits.")
	flags.String("config-section", "settings", "Optional config section when reading or writing a config key. Defaults to [settings].")
	flags.Bool("verbose", false, "Turns on debug messages in log file")
	flags.Bool("version", false, "") // help missing
	err := v.BindPFlags(flags)
	if err != nil {
		fmt.Printf("failed to bind cobra flags to viper: %s", err)
		os.Exit(errCodeDefault)
	}
}

// ReadInConfig reads wakatime config file in memory.
func ReadInConfig(v *viper.Viper, filepathFn func(v *viper.Viper) (string, error)) error {
	configFilepath, err := filepathFn(v)
	if err != nil {
		return ErrConfigFileParse(err.Error())
	}
	jww.DEBUG.Println("wakatime path:", configFilepath)

	v.SetConfigType("ini")
	v.SetConfigFile(configFilepath)
	if err := v.ReadInConfig(); err != nil {
		return ErrConfigFileParse(err.Error())
	}

	return nil
}

// ConfigFilePath returns the path for wakatime config file.
func ConfigFilePath(v *viper.Viper) (string, error) {
	configFilepath := v.GetString("config")
	if configFilepath != "" {
		p, err := homedir.Expand(configFilepath)
		if err != nil {
			return "", fmt.Errorf("failed parsing config flag variable: %s", err)
		}
		return p, nil
	}

	home, exists := os.LookupEnv("WAKATIME_HOME")
	if exists && home != "" {
		p, err := homedir.Expand(home)
		if err != nil {
			return "", fmt.Errorf("failed parsing WAKATIME_HOME environment variable: %s", err)
		}
		return path.Join(p, defaultConfigFile), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed getting user's home directory: %s", err)
	}

	return path.Join(home, defaultConfigFile), nil
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := NewRootCMD().Execute(); err != nil {
		jww.CRITICAL.Fatalf("failed to run wakatime-cli: %s", err)
		os.Exit(errCodeDefault)
	}
}
