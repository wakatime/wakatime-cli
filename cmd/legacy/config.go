package legacy

import (
	"fmt"
	"strings"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

const errCodeConfigFileRead = 110

// ConfigReadParams contains config read parameters.
type ConfigReadParams struct {
	Section string
	Key     string
}

func runConfigRead(v *viper.Viper) error {
	params, err := LoadConfigReadParams(v)
	if err != nil {
		return err
	}

	value := v.GetString(params.ViperKey())
	fmt.Println(value)

	return nil
}

// LoadConfigReadParams loads needed data from the configuration file.
func LoadConfigReadParams(v *viper.Viper) (ConfigReadParams, error) {
	section := strings.TrimSpace(v.GetString("config-section"))
	key := strings.TrimSpace(v.GetString("config-read"))

	jww.DEBUG.Println("section:", section)
	jww.DEBUG.Println("key:", key)

	if section == "" || key == "" {
		return ConfigReadParams{},
			ErrConfigFileRead(fmt.Errorf("failed reading wakatime config file. neither section nor key can be empty").Error())
	}

	return ConfigReadParams{
		Section: section,
		Key:     key,
	}, nil
}

// ViperKey formats to a string [section].[key].
func (c *ConfigReadParams) ViperKey() string {
	return fmt.Sprintf("%s.%s", c.Section, c.Key)
}
