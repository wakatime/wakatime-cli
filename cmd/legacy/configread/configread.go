package configread

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/spf13/viper"
)

// Params contains config read parameters.
type Params struct {
	Section string
	Key     string
}

// Run prints the value for the given config key.
func Run(v *viper.Viper) {
	output, err := Read(v)
	if err != nil {
		var cfrerr ErrFileRead
		if errors.As(err, &cfrerr) {
			log.Errorf(err.Error())
			os.Exit(exitcode.ErrConfigFileRead)
		}

		log.Fatalln(err)
	}

	fmt.Println(output)
	os.Exit(exitcode.Success)
}

// Read returns the value for the given config key.
func Read(v *viper.Viper) (string, error) {
	params, err := LoadParams(v)
	if err != nil {
		return "", err
	}

	value := strings.TrimSpace(vipertools.GetString(v, params.ViperKey()))
	if value == "" {
		return "", ErrFileRead(
			fmt.Sprintf("given section and key %q returned an empty string", params.ViperKey()),
		)
	}

	return value, nil
}

// LoadParams loads needed data from the configuration file.
func LoadParams(v *viper.Viper) (Params, error) {
	section := strings.TrimSpace(vipertools.GetString(v, "config-section"))
	key := strings.TrimSpace(vipertools.GetString(v, "config-read"))

	if section == "" || key == "" {
		return Params{},
			ErrFileRead("failed reading wakatime config file. neither section nor key can be empty")
	}

	return Params{
		Section: section,
		Key:     key,
	}, nil
}

// ViperKey formats to a string [section].[key].
func (c *Params) ViperKey() string {
	return fmt.Sprintf("%s.%s", c.Section, c.Key)
}
