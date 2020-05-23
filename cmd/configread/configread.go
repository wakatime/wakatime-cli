package configread

import (
	"fmt"
	"strings"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

// Params contains config read parameters.
type Params struct {
	Section string
	Key     string
}

// Run prints value for the given config key.
func Run(v *viper.Viper) error {
	params, err := LoadParams(v)
	if err != nil {
		return err
	}

	value := strings.TrimSpace(v.GetString(params.ViperKey()))
	if value == "" {
		return ErrFileRead(
			fmt.Errorf("given section and key \"%s\" returned an empty string", params.ViperKey()).Error(),
		)
	}

	fmt.Println(value)

	return nil
}

// LoadParams loads needed data from the configuration file.
func LoadParams(v *viper.Viper) (Params, error) {
	section := strings.TrimSpace(v.GetString("config-section"))
	key := strings.TrimSpace(v.GetString("config-read"))

	jww.DEBUG.Println("section:", section)
	jww.DEBUG.Println("key:", key)

	if section == "" || key == "" {
		return Params{},
			ErrFileRead(fmt.Errorf("failed reading wakatime config file. neither section nor key can be empty").Error())
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
