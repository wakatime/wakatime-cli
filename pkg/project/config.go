package project

import (
	"fmt"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
	paramspkg "github.com/wakatime/wakatime-cli/pkg/params"

	"github.com/spf13/viper"
)

// defaultConfigFile is the name of the default wakatime config file per project.
const defaultConfigFile = ".wakatime"

// WithConfiguration initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to load project-level configuration params
// and override the loaded from ~/.wakatime.cfg file.
func WithConfiguration(v *viper.Viper, params *paramspkg.Params) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			for n, h := range hh {
				log.Debugf("execute load project configuration for: %s", h.Entity)

				fp, ok := FindFile(h.Entity, defaultConfigFile)
				if !ok {
					continue
				}

				_, err := LoadConfig(fp)
				if err != nil {
					log.Errorf("failed to load project configuration: %s", err)
					continue
				}

				hh[n].
			}

			return next(hh)
		}
	}
}

// LoadConfig loads project-level configuration params and overrides the loaded from ~/.wakatime.cfg file.
func LoadConfig(fp string, params *paramspkg.Params) error {
	if err := ini.ReadInConfig(c.vp, configFile); err != nil {
		return fmt.Errorf("failed to load configuration file: %s", err)
	}
	
	return nil
}
