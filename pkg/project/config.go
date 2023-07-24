package project

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// defaultConfigFile is the name of the default wakatime config file per project.
const defaultConfigFile = ".wakatime"

// WithConfiguration initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to load project-level configuration params
// and override the loaded from ~/.wakatime.cfg file.
func WithConfiguration() heartbeat.HandleOption {
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
func LoadConfig(fp string) (Config, error) {
	return Config{}, nil
}
