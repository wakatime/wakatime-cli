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
	defaultConfigFile          = ".wakatime.cfg"
	errCodeConfigFileParse int = 103
	errCodeDefault         int = 1
	successCode                = 0
)

// ErrConfigFileParse handles a custom error while parsing wakatime config file
type ErrConfigFileParse string

func (e ErrConfigFileParse) Error() string {
	return string(e)
}

// NewRootCMD creates a rootCmd, which represents the base command when called without any subcommands.
func NewRootCMD() *cobra.Command {
	v := viper.GetViper()
	cmd := &cobra.Command{
		Use:   "wakatime-cli",
		Short: "Command line interface used by all WakaTime text editor plugins.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := loadConfigFile(v); err != nil {
				jww.CRITICAL.Printf("err: %s", err)
				var cfperr ErrConfigFileParse
				if errors.As(err, &cfperr) {
					os.Exit(errCodeConfigFileParse)
				}
				os.Exit(errCodeDefault)
			}
			if err := legacy.Run(v); err != nil {
				jww.CRITICAL.Printf("err: %s", err)
				os.Exit(errCodeDefault)
			}

			os.Exit(successCode)
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

func loadConfigFile(v *viper.Viper) error {
	var (
		configFilepath string
		err            error
	)

	configFilepath = v.GetString("config")

	if configFilepath == "" {
		configFilepath, err = getConfigFilePath()
		if err != nil {
			return ErrConfigFileParse(err.Error())
		}
	}

	jww.DEBUG.Println("wakatime path:", configFilepath)

	v.SetConfigType("ini")
	v.SetConfigFile(configFilepath)
	if err := v.ReadInConfig(); err != nil {
		return ErrConfigFileParse(err.Error())
	}

	return nil
}

func getConfigFilePath() (string, error) {
	home, exists := os.LookupEnv("WAKATIME_HOME")
	if exists {
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
