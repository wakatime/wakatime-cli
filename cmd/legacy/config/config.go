package config

import (
	"fmt"
	"strings"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

// ReadParams contains config read parameters.
type ReadParams struct {
	Section string
	Key     string
}

// RunRead prints value for the given config key.
func RunRead(v *viper.Viper) error {
	params, err := LoadReadParams(v)
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

// LoadReadParams loads needed data from the configuration file.
func LoadReadParams(v *viper.Viper) (ReadParams, error) {
	section := strings.TrimSpace(v.GetString("config-section"))
	key := strings.TrimSpace(v.GetString("config-read"))

	jww.DEBUG.Println("section:", section)
	jww.DEBUG.Println("key:", key)

	if section == "" || key == "" {
		return ReadParams{},
			ErrFileRead(fmt.Errorf("failed reading wakatime config file. neither section nor key can be empty").Error())
	}

	return ReadParams{
		Section: section,
		Key:     key,
	}, nil
}

// ViperKey formats to a string [section].[key].
func (c *ReadParams) ViperKey() string {
	return fmt.Sprintf("%s.%s", c.Section, c.Key)
}
