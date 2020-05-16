package legacy

import (
	"fmt"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

// ErrCodeConfigFileRead error number when reading wakatime config file
const ErrCodeConfigFileRead = 110

// ConfigReadParams contains config read parameters
type ConfigReadParams struct {
	Section string
	Key     string
}

func runConfigRead(v *viper.Viper) {
	params := LoadConfigReadParams(v)

	value := v.GetString(params.ViperKey())
	fmt.Println(value)
}

// LoadConfigReadParams loads needed data from the configuration file
func LoadConfigReadParams(v *viper.Viper) ConfigReadParams {
	section := v.GetString("config-section")
	key := v.GetString("config-read")

	jww.DEBUG.Println("section:", section)
	jww.DEBUG.Println("key:", key)

	return ConfigReadParams{
		Section: section,
		Key:     key,
	}
}

// ViperKey formats to a string [section].[key]
func (c *ConfigReadParams) ViperKey() string {
	return fmt.Sprintf("%s.%s", c.Section, c.Key)
}
