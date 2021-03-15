package configread

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/exitcode"

	jww "github.com/spf13/jwalterweatherman"
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
		jww.ERROR.Printf("err: %s", err)

		var cfrerr ErrFileRead
		if errors.As(err, &cfrerr) {
			os.Exit(exitcode.ErrConfigFileRead)
		}

		os.Exit(exitcode.ErrDefault)
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

	value := strings.TrimSpace(v.GetString(params.ViperKey()))
	if value == "" {
		return "", ErrFileRead(
			fmt.Sprintf("given section and key %q returned an empty string", params.ViperKey()),
		)
	}

	return value, nil
}

// LoadParams loads needed data from the configuration file.
func LoadParams(v *viper.Viper) (Params, error) {
	section := strings.TrimSpace(v.GetString("config-section"))
	key := strings.TrimSpace(v.GetString("config-read"))

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
