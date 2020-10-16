package language

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	jww "github.com/spf13/jwalterweatherman"
)

// Config contains configurations for language detection.
type Config struct {
	Alternate string
	Override  string
}

// WithDetection initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to detect and add programming
// language info to heartbeats of entity type 'file'.
func WithDetection(config Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			for n := range hh {
				if config.Alternate != "" {
					parsed, ok := heartbeat.ParseLanguage(config.Alternate)
					if !ok {
						jww.WARN.Printf("Failed to parse alternate language %q", config.Alternate)
					}

					hh[n].Language = parsed
				}

				if config.Override != "" {
					parsed, ok := heartbeat.ParseLanguage(config.Override)
					if !ok {
						jww.WARN.Printf("Failed to parse override language %q", config.Alternate)
					}

					hh[n].Language = parsed
				}
			}

			return next(hh)
		}
	}
}
